// Package injectcmd 实现 mesh inject 命令。
package injectcmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"mesh-cli/internal/collect"
	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

const (
	defaultLimit     = 8
	defaultMaxTokens = 1200
)

// Run 执行 mesh inject 命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("inject", flag.ContinueOnError)
	fs.SetOutput(stderr)
	source := fs.String("source", "", "按来源过滤（可选）")
	tag := fs.String("tag", "", "按标签过滤（可选）")
	limit := fs.Int("limit", defaultLimit, "限制返回条数（默认 8）")
	maxTokens := fs.Int("max-tokens", defaultMaxTokens, "输出最大 token 预算（默认 1200）")
	format := fs.String("format", "text", "输出格式：text|markdown")
	fs.Usage = func() { printUsage(fs.Output()) }

	normalizedArgs, err := normalizeArgs(args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "inject: %v\n", err)
		printUsage(stderr)
		return 2
	}
	if err := fs.Parse(normalizedArgs); err != nil {
		return 2
	}
	if fs.NArg() > 1 {
		_, _ = fmt.Fprintf(stderr, "inject: 最多只支持 1 个关键词参数，收到 %d 个\n", fs.NArg())
		printUsage(stderr)
		return 2
	}

	keyword := ""
	if fs.NArg() == 1 {
		keyword = strings.TrimSpace(fs.Arg(0))
	}
	if err := validateInput(keyword, *source, *tag, *limit, *maxTokens, *format); err != nil {
		_, _ = fmt.Fprintf(stderr, "inject: %v\n", err)
		printUsage(stderr)
		return 2
	}

	dbPath, err := collect.ResolveDBPathForReuse()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "inject: 解析数据库路径失败: %v\n", err)
		return 1
	}
	st, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "inject: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = st.Close() }()

	items, err := st.QueryWithOptions(types.QueryOptions{Query: keyword, Source: strings.TrimSpace(*source), Tags: strings.TrimSpace(*tag), Limit: *limit})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "inject: 查询失败: %v\n", err)
		return 1
	}
	if len(items) == 0 {
		_, _ = fmt.Fprintln(stdout, "# Context Injection\n\n未找到匹配认知。")
		return 0
	}

	result := render(items, *format)
	trimmed := trimByTokenBudget(result, *maxTokens)
	_, _ = io.WriteString(stdout, trimmed)
	if trimmed != result {
		_, _ = fmt.Fprintf(stdout, "\n\n[truncated by --max-tokens=%d]\n", *maxTokens)
	}
	return 0
}

func normalizeArgs(args []string) ([]string, error) {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, 1)
	expectValue := false
	for i := 0; i < len(args); i++ {
		current := args[i]
		if expectValue {
			flags = append(flags, current)
			expectValue = false
			continue
		}
		if current == "--source" || current == "--tag" || current == "--limit" || current == "--max-tokens" || current == "--format" {
			flags = append(flags, current)
			expectValue = true
			continue
		}
		if strings.HasPrefix(current, "--source=") || strings.HasPrefix(current, "--tag=") || strings.HasPrefix(current, "--limit=") || strings.HasPrefix(current, "--max-tokens=") || strings.HasPrefix(current, "--format=") {
			flags = append(flags, current)
			continue
		}
		positionals = append(positionals, current)
	}
	if expectValue {
		return nil, errors.New("flag 缺少对应值")
	}
	return append(flags, positionals...), nil
}

func validateInput(keyword, source, tag string, limit, maxTokens int, format string) error {
	if strings.TrimSpace(keyword) == "" && strings.TrimSpace(source) == "" && strings.TrimSpace(tag) == "" {
		return errors.New("请至少提供关键词、--source 或 --tag 之一")
	}
	if limit <= 0 {
		return errors.New("--limit 必须大于 0")
	}
	if maxTokens <= 0 {
		return errors.New("--max-tokens 必须大于 0")
	}
	f := strings.ToLower(strings.TrimSpace(format))
	if f != "text" && f != "markdown" {
		return errors.New("--format 仅支持 text|markdown")
	}
	return nil
}

func render(items []*types.Insight, format string) string {
	if strings.ToLower(strings.TrimSpace(format)) == "markdown" {
		return renderMarkdown(items)
	}
	return renderText(items)
}

func renderText(items []*types.Insight) string {
	var b strings.Builder
	b.WriteString("Context Injection\n")
	b.WriteString("=================\n")
	for i, it := range items {
		b.WriteString(fmt.Sprintf("[%d] %s\n", i+1, time.Unix(it.CreatedAt, 0).Format("2006-01-02 15:04:05")))
		b.WriteString("来源: " + fallback(it.Source) + "\n")
		b.WriteString("标签: " + fallback(it.Tags) + "\n")
		if strings.TrimSpace(it.Summary) != "" {
			b.WriteString("摘要: " + it.Summary + "\n")
		}
		b.WriteString("内容: " + fallback(it.Content) + "\n\n")
	}
	return b.String()
}

func renderMarkdown(items []*types.Insight) string {
	var b strings.Builder
	b.WriteString("# Context Injection\n\n")
	for i, it := range items {
		b.WriteString(fmt.Sprintf("## [%d] %s\n", i+1, time.Unix(it.CreatedAt, 0).Format("2006-01-02 15:04:05")))
		b.WriteString("- 来源: " + fallback(it.Source) + "\n")
		b.WriteString("- 标签: " + fallback(it.Tags) + "\n")
		if strings.TrimSpace(it.Summary) != "" {
			b.WriteString("- 摘要: " + it.Summary + "\n")
		}
		b.WriteString("\n")
		b.WriteString(it.Content + "\n\n")
	}
	return b.String()
}

func trimByTokenBudget(s string, maxTokens int) string {
	// 粗略估算：1 token ≈ 4 chars。
	maxChars := maxTokens * 4
	runes := []rune(s)
	if len(runes) <= maxChars {
		return s
	}
	if maxChars <= 3 {
		return string(runes[:maxChars])
	}
	return string(runes[:maxChars-3]) + "..."
}

func fallback(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh inject [keyword] [--source <source>] [--tag <tag>] [--limit <n>] [--max-tokens <n>] [--format text|markdown]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --source      按来源过滤（可选）")
	_, _ = fmt.Fprintln(w, "  --tag         按标签过滤（可选）")
	_, _ = fmt.Fprintln(w, "  --limit       限制返回条数（默认 8）")
	_, _ = fmt.Fprintln(w, "  --max-tokens  输出最大 token 预算（默认 1200）")
	_, _ = fmt.Fprintln(w, "  --format      输出格式：text|markdown（默认 text）")
}
