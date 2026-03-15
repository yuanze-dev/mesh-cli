// Package mdimport 实现 mesh import 命令。
package mdimport

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"mesh-cli/internal/collect"
	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

const defaultSource = "markdown"

var (
	dateHeaderPattern  = regexp.MustCompile(`^##\s+(\d{4}-\d{2}-\d{2})\s*$`)
	titleHeaderPattern = regexp.MustCompile(`^###\s+(.+?)\s*$`)
	tagsLinePattern    = regexp.MustCompile(`^标签[:：]\s*(.*)$`)
	contentLinePattern = regexp.MustCompile(`^内容[:：]\s*(.*)$`)
)

// Run 执行 mesh import 命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.SetOutput(stderr)
	source := fs.String("source", defaultSource, "数据来源（可选，默认 markdown）")
	fs.Usage = func() { printUsage(fs.Output()) }

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		_, _ = fmt.Fprintln(stderr, "import: 需要且仅需要一个 Markdown 文件路径")
		printUsage(stderr)
		return 2
	}

	filePath := strings.TrimSpace(fs.Arg(0))
	if filePath == "" {
		_, _ = fmt.Fprintln(stderr, "import: 文件路径不能为空")
		return 2
	}

	dbPath, err := collect.ResolveDBPathForReuse()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "import: 解析数据库路径失败: %v\n", err)
		return 1
	}

	st, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "import: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = st.Close() }()

	entries, err := parseMarkdownFile(filePath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "import: 解析 Markdown 失败: %v\n", err)
		return 1
	}

	imported := 0
	skipped := 0
	for _, e := range entries {
		if strings.TrimSpace(e.Content) == "" {
			skipped++
			continue
		}
		exists, err := st.ExistsByContent(e.Content)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "import: 检查重复失败: %v\n", err)
			return 1
		}
		if exists {
			skipped++
			continue
		}

		id, err := collect.NewUUIDForReuse()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "import: 生成 ID 失败: %v\n", err)
			return 1
		}
		insight := &types.Insight{
			ID:        id,
			Source:    strings.TrimSpace(*source),
			Content:   e.Content,
			Summary:   e.Title,
			Tags:      e.Tags,
			CreatedAt: e.CreatedAt,
		}
		if err := st.Insert(insight); err != nil {
			_, _ = fmt.Fprintf(stderr, "import: 写入失败: %v\n", err)
			return 1
		}
		imported++
	}

	_, _ = fmt.Fprintf(stdout, "✅ 导入完成: 新增 %d 条，跳过 %d 条\n", imported, skipped)
	return 0
}

type markdownEntry struct {
	Date      string
	Title     string
	Tags      string
	Content   string
	CreatedAt int64
}

func parseMarkdownFile(path string) ([]markdownEntry, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	entries := make([]markdownEntry, 0)
	var currentDate string
	var current *markdownEntry
	collectingContent := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if m := dateHeaderPattern.FindStringSubmatch(line); len(m) == 2 {
			currentDate = m[1]
			collectingContent = false
			continue
		}
		if m := titleHeaderPattern.FindStringSubmatch(line); len(m) == 2 {
			if current != nil {
				current.Content = strings.TrimSpace(current.Content)
				entries = append(entries, *current)
			}
			createdAt := time.Now().Unix()
			if currentDate != "" {
				if ts, err := parseDateToUnix(currentDate); err == nil {
					createdAt = ts
				}
			}
			current = &markdownEntry{Date: currentDate, Title: strings.TrimSpace(m[1]), CreatedAt: createdAt}
			collectingContent = false
			continue
		}
		if current == nil {
			continue
		}
		if m := tagsLinePattern.FindStringSubmatch(line); len(m) == 2 {
			current.Tags = normalizeTags(m[1])
			continue
		}
		if m := contentLinePattern.FindStringSubmatch(line); len(m) == 2 {
			current.Content = strings.TrimSpace(m[1])
			collectingContent = true
			continue
		}

		if collectingContent {
			if line == "" {
				current.Content += "\n"
				continue
			}
			if current.Content == "" {
				current.Content = raw
			} else {
				current.Content += "\n" + raw
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if current != nil {
		current.Content = strings.TrimSpace(current.Content)
		entries = append(entries, *current)
	}
	if len(entries) == 0 {
		return nil, errors.New("未识别到可导入条目")
	}
	return entries, nil
}

func parseDateToUnix(date string) (int64, error) {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func normalizeTags(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "，", ",")
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return strings.Join(out, ",")
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh import [--source <source>] <markdown-file>")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --source    数据来源（可选，默认 markdown）")
}
