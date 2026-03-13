package store

import (
	"os"
	"path/filepath"
	"testing"

	"mesh-cli/pkg/types"
)

// TestNewStoreCreatesDatabase 验证 NewStore 会创建数据库文件。
func TestNewStoreCreatesDatabase(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "data", "mesh.db")

	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected db file to exist, stat error = %v", err)
	}
}

// TestInsertAndList 验证插入与列表查询行为。
func TestInsertAndList(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	first := &types.Insight{
		ID:        "1",
		Source:    "claude",
		Content:   "first content",
		Summary:   "first summary",
		Tags:      "a,b",
		AgentID:   "agent-1",
		CreatedAt: 100,
		IsShared:  0,
	}
	second := &types.Insight{
		ID:        "2",
		Source:    "cursor",
		Content:   "second content",
		Summary:   "second summary",
		Tags:      "c,d",
		AgentID:   "agent-2",
		CreatedAt: 200,
		IsShared:  1,
	}

	if err := s.Insert(first); err != nil {
		t.Fatalf("Insert(first) error = %v", err)
	}
	if err := s.Insert(second); err != nil {
		t.Fatalf("Insert(second) error = %v", err)
	}

	items, err := s.List(10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("List() len = %d, want 2", len(items))
	}
	if items[0].ID != "2" || items[1].ID != "1" {
		t.Fatalf("List() order got [%s, %s], want [2, 1]", items[0].ID, items[1].ID)
	}
}

// TestQuery 验证关键词查询行为。
func TestQuery(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	if err := s.Insert(&types.Insight{
		ID:        "a",
		Source:    "claude",
		Content:   "定价策略采用分层方案",
		Summary:   "产品策略",
		Tags:      "定价,产品",
		CreatedAt: 100,
	}); err != nil {
		t.Fatalf("Insert(a) error = %v", err)
	}
	if err := s.Insert(&types.Insight{
		ID:        "b",
		Source:    "cursor",
		Content:   "实现同步命令",
		Summary:   "工程计划",
		Tags:      "同步,开发",
		CreatedAt: 200,
	}); err != nil {
		t.Fatalf("Insert(b) error = %v", err)
	}

	items, err := s.Query("定价", 5)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Query() len = %d, want 1", len(items))
	}
	if items[0].ID != "a" {
		t.Fatalf("Query() ID = %s, want a", items[0].ID)
	}
}

// TestInsertEscapesSingleQuote 验证包含单引号的内容可正确写入与查询。
func TestInsertEscapesSingleQuote(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	in := &types.Insight{
		ID:        "quote-1",
		Source:    "claude",
		Content:   "it's a test",
		Summary:   "summary's",
		Tags:      "tag's",
		CreatedAt: 100,
	}
	if err := s.Insert(in); err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	items, err := s.Query("it's", 5)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Query() len = %d, want 1", len(items))
	}
	if items[0].Content != "it's a test" {
		t.Fatalf("Query() content = %q, want %q", items[0].Content, "it's a test")
	}
}

// TestQueryEmptyEquivalentToList 验证空查询等价于 List(limit)。
func TestQueryEmptyEquivalentToList(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	mustInsert := func(insight *types.Insight) {
		t.Helper()
		if err := s.Insert(insight); err != nil {
			t.Fatalf("Insert(%s) error = %v", insight.ID, err)
		}
	}

	mustInsert(&types.Insight{
		ID:        "1",
		Source:    "claude",
		Content:   "first",
		CreatedAt: 100,
	})
	mustInsert(&types.Insight{
		ID:        "2",
		Source:    "cursor",
		Content:   "second",
		CreatedAt: 200,
	})

	queryItems, err := s.Query("", 1)
	if err != nil {
		t.Fatalf("Query(\"\") error = %v", err)
	}
	listItems, err := s.List(1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(queryItems) != len(listItems) {
		t.Fatalf("len(Query(\"\")) = %d, len(List()) = %d", len(queryItems), len(listItems))
	}
	if len(queryItems) > 0 && queryItems[0].ID != listItems[0].ID {
		t.Fatalf("first item mismatch, query=%s list=%s", queryItems[0].ID, listItems[0].ID)
	}
}

// TestQueryWithOptions 验证 source/tag/limit 组合过滤行为。
func TestQueryWithOptions(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	mustInsert := func(insight *types.Insight) {
		t.Helper()
		if err := s.Insert(insight); err != nil {
			t.Fatalf("Insert(%s) error = %v", insight.ID, err)
		}
	}

	mustInsert(&types.Insight{
		ID:        "1",
		Source:    "claude",
		Content:   "产品定价采用分层策略",
		Summary:   "定价策略",
		Tags:      "产品,定价",
		CreatedAt: 100,
	})
	mustInsert(&types.Insight{
		ID:        "2",
		Source:    "claude",
		Content:   "技术选型总结",
		Summary:   "Go 与 SQLite",
		Tags:      "技术,架构",
		CreatedAt: 200,
	})
	mustInsert(&types.Insight{
		ID:        "3",
		Source:    "cursor",
		Content:   "产品上线节奏",
		Summary:   "发布计划",
		Tags:      "产品,计划",
		CreatedAt: 300,
	})

	items, err := s.QueryWithOptions(types.QueryOptions{
		Source: "claude",
		Tags:   "产品",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("QueryWithOptions(source+tag) error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "1" {
		t.Fatalf("QueryWithOptions(source+tag) got %+v, want only ID=1", ids(items))
	}

	items, err = s.QueryWithOptions(types.QueryOptions{
		Source: "claude",
		Limit:  1,
	})
	if err != nil {
		t.Fatalf("QueryWithOptions(source+limit) error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "2" {
		t.Fatalf("QueryWithOptions(source+limit) got %+v, want only ID=2", ids(items))
	}

	items, err = s.QueryWithOptions(types.QueryOptions{
		Query: "产品",
		Tags:  "计划",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("QueryWithOptions(query+tag) error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "3" {
		t.Fatalf("QueryWithOptions(query+tag) got %+v, want only ID=3", ids(items))
	}
}

// newTestStore 创建测试专用 Store。
func newTestStore(t *testing.T) *Store {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "mesh.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	t.Cleanup(func() {
		_ = s.Close()
	})
	return s
}

func ids(items []*types.Insight) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}
