// Package mdexport 实现 mesh export 命令。
package mdexport

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mesh-cli/internal/collect"
	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

const exportLimit = 100000

// Run 执行 mesh export 命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("output", "", "输出文件路径（可选，默认 stdout）")
	fs.Usage = func() { printUsage(fs.Output()) }

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "export: 不支持的位置参数: %s\n", strings.Join(fs.Args(), " "))
		printUsage(stderr)
		return 2
	}

	dbPath, err := collect.ResolveDBPathForReuse()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "export: 解析数据库路径失败: %v\n", err)
		return 1
	}

	st, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "export: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = st.Close() }()

	insights, err := st.List(exportLimit)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "export: 查询失败: %v\n", err)
		return 1
	}

	md := renderMarkdown(insights)

	if strings.TrimSpace(*output) == "" {
		_, _ = io.WriteString(stdout, md)
		return 0
	}

	outPath := filepath.Clean(*output)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		_, _ = fmt.Fprintf(stderr, "export: 创建目录失败: %v\n", err)
		return 1
	}
	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		_, _ = fmt.Fprintf(stderr, "export: 写文件失败: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "✅ 已导出 %d 条到 %s\n", len(insights), outPath)
	return 0
}

func renderMarkdown(insights []*types.Insight) string {
	if len(insights) == 0 {
		return ""
	}

	var b strings.Builder
	currentDate := ""
	for _, it := range insights {
		date := time.Unix(it.CreatedAt, 0).Format("2006-01-02")
		if date != currentDate {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString("## ")
			b.WriteString(date)
			b.WriteString("\n\n")
			currentDate = date
		}

		title := strings.TrimSpace(it.Summary)
		if title == "" {
			title = summarizeTitle(it.Content)
		}
		b.WriteString("### ")
		b.WriteString(title)
		b.WriteString("\n")

		tags := strings.TrimSpace(it.Tags)
		if tags != "" {
			b.WriteString("标签: ")
			b.WriteString(tags)
			b.WriteString("\n")
		}

		b.WriteString("内容: ")
		b.WriteString(strings.TrimSpace(it.Content))
		b.WriteString("\n\n")
	}
	return b.String()
}

func summarizeTitle(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "未命名记录"
	}
	runes := []rune(content)
	if len(runes) <= 24 {
		return content
	}
	return string(runes[:24]) + "..."
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh export [--output <file>]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --output    输出文件路径（可选，默认 stdout）")
}
