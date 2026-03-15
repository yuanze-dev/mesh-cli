package mdimport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMarkdownFile_MultiLineContent(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "sample.md")
	content := `## 2026-03-10

### 决策：使用 Go 开发 Mesh
标签: 技术,决策
内容: 第一行
第二行
第三行
`
	if err := os.WriteFile(md, []byte(content), 0o644); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	entries, err := parseMarkdownFile(md)
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}
	if got := len(entries); got != 1 {
		t.Fatalf("expected 1 entry, got %d", got)
	}
	if !strings.Contains(entries[0].Content, "第一行\n第二行\n第三行") {
		t.Fatalf("unexpected content: %q", entries[0].Content)
	}
}
