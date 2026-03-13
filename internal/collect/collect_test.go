package collect

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"mesh-cli/internal/store"
)

// TestRunSuccessInsertsInsight 验证 collect 成功写入数据库。
func TestRunSuccessInsertsInsight(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "mesh.db")
	t.Setenv("MESH_DB_PATH", dbPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--source", "claude",
		"--content", "决策：使用Go开发",
		"--tag", "技术,决策",
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "已采集 1 条认知") {
		t.Fatalf("expected success output, got: %s", stdout.String())
	}

	dataStore, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer func() { _ = dataStore.Close() }()

	items, err := dataStore.List(10)
	if err != nil {
		t.Fatalf("list insights: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(items))
	}
	if items[0].Source != "claude" {
		t.Fatalf("expected source claude, got %s", items[0].Source)
	}
	if items[0].Content != "决策：使用Go开发" {
		t.Fatalf("unexpected content: %s", items[0].Content)
	}
	if items[0].Tags != "技术,决策" {
		t.Fatalf("unexpected tags: %s", items[0].Tags)
	}
}

// TestRunMissingRequiredFlags 验证缺少必填参数时返回参数错误。
func TestRunMissingRequiredFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"--source", "claude"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "缺少必填参数 --content") {
		t.Fatalf("expected missing content error, got: %s", stderr.String())
	}
}

// TestRunMissingSourceFlag 验证仅传 content 时返回缺少 source。
func TestRunMissingSourceFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"--content", "hello"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "缺少必填参数 --source") {
		t.Fatalf("expected missing source error, got: %s", stderr.String())
	}
}

// TestRunRejectsPositionalArgs 验证存在多余位置参数时返回参数错误。
func TestRunRejectsPositionalArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--source", "claude",
		"--content", "ok",
		"extra-arg",
	}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "不支持的位置参数") {
		t.Fatalf("expected positional args error, got: %s", stderr.String())
	}
}
