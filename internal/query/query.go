// Package query 实现 mesh query 命令。
package query

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

const defaultDBFile = ".mesh/mesh.db"

// Run 执行 mesh query 命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	fs.SetOutput(stderr)

	source := fs.String("source", "", "按来源过滤（可选）")
	tag := fs.String("tag", "", "按标签过滤（可选）")
	limit := fs.Int("limit", 20, "限制返回条数（可选）")
	fs.Usage = func() {
		printUsage(fs.Output())
	}

	normalizedArgs, err := normalizeArgs(args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "query: %v\n", err)
		printUsage(stderr)
		return 2
	}

	if err := fs.Parse(normalizedArgs); err != nil {
		return 2
	}
	if fs.NArg() > 1 {
		_, _ = fmt.Fprintf(stderr, "query: 最多只支持 1 个关键词参数，收到 %d 个\n", fs.NArg())
		printUsage(stderr)
		return 2
	}

	keyword := ""
	if fs.NArg() == 1 {
		keyword = strings.TrimSpace(fs.Arg(0))
	}
	if err := validateInput(keyword, *source, *tag, *limit); err != nil {
		_, _ = fmt.Fprintf(stderr, "query: %v\n", err)
		printUsage(stderr)
		return 2
	}

	dbPath, err := resolveDBPath()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "query: 解析数据库路径失败: %v\n", err)
		return 1
	}

	dataStore, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "query: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = dataStore.Close() }()

	items, err := dataStore.QueryWithOptions(types.QueryOptions{
		Query:  keyword,
		Source: strings.TrimSpace(*source),
		Tags:   strings.TrimSpace(*tag),
		Limit:  *limit,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "query: 查询失败: %v\n", err)
		return 1
	}

	printResult(stdout, items)
	return 0
}

// normalizeArgs 兼容 `mesh query <keyword> --flag` 与 `mesh query --flag <keyword>` 两种输入顺序。
func normalizeArgs(args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

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
		if current == "--source" || current == "--tag" || current == "--limit" {
			flags = append(flags, current)
			expectValue = true
			continue
		}
		if strings.HasPrefix(current, "--source=") || strings.HasPrefix(current, "--tag=") || strings.HasPrefix(current, "--limit=") {
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

// validateInput 验证输入参数。
func validateInput(keyword string, source string, tag string, limit int) error {
	if limit <= 0 {
		return errors.New("--limit 必须大于 0")
	}
	if strings.TrimSpace(keyword) == "" && strings.TrimSpace(source) == "" && strings.TrimSpace(tag) == "" {
		return errors.New("请至少提供关键词、--source 或 --tag 之一")
	}
	return nil
}

// resolveDBPath 解析数据库文件路径，优先使用 MESH_DB_PATH。
func resolveDBPath() (string, error) {
	if fromEnv := strings.TrimSpace(os.Getenv("MESH_DB_PATH")); fromEnv != "" {
		return fromEnv, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(homeDir, defaultDBFile), nil
}

// printResult 输出格式化查询结果。
func printResult(w io.Writer, items []*types.Insight) {
	if len(items) == 0 {
		_, _ = fmt.Fprintln(w, "🔍 未找到匹配的认知")
		return
	}

	_, _ = fmt.Fprintf(w, "🔍 查询到 %d 条认知\n", len(items))
	for idx, item := range items {
		_, _ = fmt.Fprintln(w, "----------------------------------------")
		_, _ = fmt.Fprintf(w, "[%d] %s\n", idx+1, formatTime(item.CreatedAt))
		_, _ = fmt.Fprintf(w, "来源: %s\n", fallback(item.Source))
		_, _ = fmt.Fprintf(w, "标签: %s\n", fallback(item.Tags))
		_, _ = fmt.Fprintf(w, "内容: %s\n", fallback(item.Content))
		if strings.TrimSpace(item.Summary) != "" {
			_, _ = fmt.Fprintf(w, "摘要: %s\n", item.Summary)
		}
	}
}

// formatTime 将时间戳格式化为可读字符串。
func formatTime(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

// fallback 返回展示字段的回退值。
func fallback(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

// printUsage 输出 query 命令帮助信息。
func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh query [keyword] [--source <source>] [--tag <tag>] [--limit <n>]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --source    按来源过滤（可选）")
	_, _ = fmt.Fprintln(w, "  --tag       按标签过滤（可选）")
	_, _ = fmt.Fprintln(w, "  --limit     限制返回条数，默认 20")
}
