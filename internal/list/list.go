// Package list 实现 mesh list 命令。
package list

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

const defaultDBFile = ".mesh/mesh.db"
const defaultLimit = 20

// Run 执行 mesh list 命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)

	source := fs.String("source", "", "过滤来源（可选）")
	limit := fs.Int("limit", defaultLimit, "限制返回数量（默认 20）")
	fs.Usage = func() {
		printUsage(fs.Output())
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "list: 不支持的位置参数: %s\n", strings.Join(fs.Args(), " "))
		printUsage(stderr)
		return 2
	}

	dbPath, err := resolveDBPath()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "list: 解析数据库路径失败: %v\n", err)
		return 1
	}

	dataStore, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "list: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = dataStore.Close() }()

	opts := types.QueryOptions{
		Source: strings.TrimSpace(*source),
		Limit:  *limit,
	}

	insights, err := dataStore.QueryWithOptions(opts)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "list: 查询失败: %v\n", err)
		return 1
	}

	if len(insights) == 0 {
		_, _ = fmt.Fprintln(stdout, "暂无数据")
		return 0
	}

	printTable(stdout, insights)
	return 0
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
	return homeDir + "/" + defaultDBFile, nil
}

// printTable 以表格格式输出记录。
func printTable(w io.Writer, insights []*types.Insight) {
	// 计算每列的最大宽度
	maxID := 8
	maxSource := 6
	maxContent := 7
	maxTags := 4

	for _, insight := range insights {
		if len(insight.ID) > maxID {
			maxID = len(insight.ID)
		}
		if len(insight.Source) > maxSource {
			maxSource = len(insight.Source)
		}
		if len(truncateString(insight.Content, 50)) > maxContent {
			maxContent = len(truncateString(insight.Content, 50))
		}
		if len(insight.Tags) > maxTags {
			maxTags = len(insight.Tags)
		}
	}

	// 输出表头
	header := fmt.Sprintf(
		"%-*s  %-*s  %-*s  %-*s  %s",
		maxID, "ID",
		maxSource, "来源",
		maxContent, "内容",
		maxTags, "标签",
		"时间",
	)
	separator := strings.Repeat("-", len(header))

	_, _ = fmt.Fprintln(w, header)
	_, _ = fmt.Fprintln(w, separator)

	// 输出数据行
	for _, insight := range insights {
		timestamp := formatTimestamp(insight.CreatedAt)
		content := truncateString(insight.Content, 50)
		tags := insight.Tags
		if tags == "" {
			tags = "-"
		}

		row := fmt.Sprintf(
			"%-*s  %-*s  %-*s  %-*s  %s",
			maxID, truncateString(insight.ID, maxID),
			maxSource, insight.Source,
			maxContent, content,
			maxTags, tags,
			timestamp,
		)
		_, _ = fmt.Fprintln(w, row)
	}

	// 输出统计信息
	_, _ = fmt.Fprintf(w, "\n共 %d 条记录\n", len(insights))
}

// truncateString 截断字符串到指定长度，并添加省略号。
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatTimestamp 格式化时间戳为可读字符串。
func formatTimestamp(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04")
}

// printUsage 输出 list 命令帮助信息。
func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh list [--source <source>] [--limit <n>]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --source    过滤来源（可选）")
	_, _ = fmt.Fprintln(w, "  --limit     限制返回数量（默认 20）")
}
