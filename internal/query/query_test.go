package query

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

// TestRunQuerySuccess 验证 query 成功返回格式化结果。
func TestRunQuerySuccess(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "mesh.db")
	t.Setenv("MESH_DB_PATH", dbPath)
	seedInsights(t, dbPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"定价", "--source", "claude", "--tag", "产品", "--limit", "5"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "查询到 1 条认知") {
		t.Fatalf("expected query result count, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "来源: claude") {
		t.Fatalf("expected source field, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "标签: 产品,定价") {
		t.Fatalf("expected tags field, got: %s", stdout.String())
	}
}

// TestRunQueryMissingCondition 验证未提供查询条件时报错。
func TestRunQueryMissingCondition(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "请至少提供关键词、--source 或 --tag 之一") {
		t.Fatalf("expected missing condition error, got: %s", stderr.String())
	}
}

// TestRunQueryInvalidLimit 验证非法 limit 报错。
func TestRunQueryInvalidLimit(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"定价", "--limit", "0"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--limit 必须大于 0") {
		t.Fatalf("expected invalid limit error, got: %s", stderr.String())
	}
}

func seedInsights(t *testing.T, dbPath string) {
	t.Helper()

	dataStore, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() {
		_ = dataStore.Close()
	})

	items := []*types.Insight{
		{
			ID:        "i-1",
			Source:    "claude",
			Content:   "定价策略采用阶梯方案",
			Summary:   "定价决策",
			Tags:      "产品,定价",
			CreatedAt: 100,
		},
		{
			ID:        "i-2",
			Source:    "cursor",
			Content:   "登录模块开发完成",
			Summary:   "开发进展",
			Tags:      "工程,进度",
			CreatedAt: 200,
		},
	}
	for _, item := range items {
		if err := dataStore.Insert(item); err != nil {
			t.Fatalf("insert seed insight %s: %v", item.ID, err)
		}
	}
}
