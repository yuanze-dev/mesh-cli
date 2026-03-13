// Package store 提供 SQLite 存储实现。
package store

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"mesh-cli/pkg/types"
)

const defaultListLimit = 20

// Store 表示 SQLite 存储。
type Store struct {
	dbPath string
}

// NewStore 初始化 SQLite 存储，并自动创建数据库和表结构。
func NewStore(dbPath string) (*Store, error) {
	path := strings.TrimSpace(dbPath)
	if path == "" {
		return nil, errors.New("db path cannot be empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	s := &Store{dbPath: path}
	if err := s.createSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

// Insert 插入一条 Insight 记录。
func (s *Store) Insert(insight *types.Insight) error {
	if s == nil {
		return errors.New("store is not initialized")
	}
	if err := validateInsight(insight); err != nil {
		return err
	}

	updatedAt := "NULL"
	if insight.UpdatedAt > 0 {
		updatedAt = strconv.FormatInt(insight.UpdatedAt, 10)
	}
	stmt := fmt.Sprintf(`
	INSERT INTO insights (id, source, content, summary, tags, agent_id, created_at, updated_at, is_shared)
	VALUES ('%s', '%s', '%s', '%s', '%s', '%s', %d, %s, %d);
	`,
		escapeSQL(insight.ID), escapeSQL(insight.Source), escapeSQL(insight.Content),
		escapeSQL(insight.Summary), escapeSQL(insight.Tags), escapeSQL(insight.AgentID),
		insight.CreatedAt, updatedAt, insight.IsShared,
	)
	return s.exec(stmt)
}

// Query 按关键词查询记录，按创建时间倒序返回。
// 若 query 为空字符串，等价于 List(limit)。
func (s *Store) Query(query string, limit int) ([]*types.Insight, error) {
	if strings.TrimSpace(query) == "" {
		return s.List(limit)
	}
	return s.QueryWithOptions(types.QueryOptions{
		Query: query,
		Limit: limit,
	})
}

// QueryWithOptions 按条件查询记录，按创建时间倒序返回。
func (s *Store) QueryWithOptions(opts types.QueryOptions) ([]*types.Insight, error) {
	if s == nil {
		return nil, errors.New("store is not initialized")
	}

	conditions := make([]string, 0, 3)
	if keyword := strings.TrimSpace(opts.Query); keyword != "" {
		like := escapeSQL("%" + keyword + "%")
		conditions = append(conditions, fmt.Sprintf(
			"(content LIKE '%[1]s' OR summary LIKE '%[1]s' OR tags LIKE '%[1]s' OR source LIKE '%[1]s')",
			like,
		))
	}
	if source := strings.TrimSpace(opts.Source); source != "" {
		conditions = append(conditions, fmt.Sprintf("source = '%s'", escapeSQL(source)))
	}
	if tag := strings.TrimSpace(opts.Tags); tag != "" {
		conditions = append(conditions, fmt.Sprintf("tags LIKE '%%%s%%'", escapeSQL(tag)))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	querySQL := fmt.Sprintf(`
	SELECT id, source, content, summary, tags, agent_id, created_at, IFNULL(updated_at, ''), is_shared
	FROM insights
	%[1]s
	ORDER BY created_at DESC
	LIMIT %d;
	`, whereClause, normalizeLimit(opts.Limit))
	return s.queryRows(querySQL)
}

// List 列出最近的记录，按创建时间倒序返回。
func (s *Store) List(limit int) ([]*types.Insight, error) {
	if s == nil {
		return nil, errors.New("store is not initialized")
	}

	querySQL := fmt.Sprintf(`
	SELECT id, source, content, summary, tags, agent_id, created_at, IFNULL(updated_at, ''), is_shared
	FROM insights
	ORDER BY created_at DESC
	LIMIT %d;
	`, normalizeLimit(limit))
	return s.queryRows(querySQL)
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	return nil
}

// createSchema 确保 insights 表存在。
func (s *Store) createSchema() error {
	const schema = `
CREATE TABLE IF NOT EXISTS insights (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    content TEXT NOT NULL,
    summary TEXT,
    tags TEXT,
    agent_id TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER,
    is_shared INTEGER DEFAULT 0
);`
	return s.exec(schema)
}

// queryRows 执行查询并将结果转换为 Insight 切片。
func (s *Store) queryRows(querySQL string) ([]*types.Insight, error) {
	output, err := s.execWithOutput(querySQL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(output) == "" {
		return []*types.Insight{}, nil
	}

	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse query result: %w", err)
	}

	insights := make([]*types.Insight, 0, len(records))
	for _, record := range records {
		insight, err := parseInsightRecord(record)
		if err != nil {
			return nil, err
		}
		insights = append(insights, insight)
	}
	return insights, nil
}

// exec 执行无输出 SQL。
func (s *Store) exec(stmt string) error {
	_, err := s.execWithOutput(stmt)
	return err
}

// execWithOutput 调用 sqlite3 命令执行 SQL。
func (s *Store) execWithOutput(stmt string) (string, error) {
	cmd := exec.Command("sqlite3", "-batch", "-csv", "-noheader", s.dbPath, stmt)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("sqlite command failed: %w, stderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// parseInsightRecord 将 CSV 记录解析为 Insight。
func parseInsightRecord(record []string) (*types.Insight, error) {
	if len(record) != 9 {
		return nil, fmt.Errorf("invalid record column count: %d", len(record))
	}

	createdAt, err := strconv.ParseInt(record[6], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	isShared, err := strconv.Atoi(record[8])
	if err != nil {
		return nil, fmt.Errorf("parse is_shared: %w", err)
	}

	insight := &types.Insight{
		ID:        record[0],
		Source:    record[1],
		Content:   record[2],
		Summary:   record[3],
		Tags:      record[4],
		AgentID:   record[5],
		CreatedAt: createdAt,
		IsShared:  isShared,
	}

	if record[7] != "" {
		updatedAt, err := strconv.ParseInt(record[7], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at: %w", err)
		}
		insight.UpdatedAt = updatedAt
	}
	return insight, nil
}

// validateInsight 验证输入完整性。
func validateInsight(insight *types.Insight) error {
	if insight == nil {
		return errors.New("insight cannot be nil")
	}
	if strings.TrimSpace(insight.ID) == "" {
		return errors.New("insight id cannot be empty")
	}
	if strings.TrimSpace(insight.Source) == "" {
		return errors.New("insight source cannot be empty")
	}
	if strings.TrimSpace(insight.Content) == "" {
		return errors.New("insight content cannot be empty")
	}
	if insight.CreatedAt <= 0 {
		return errors.New("insight created_at must be greater than 0")
	}
	return nil
}

// normalizeLimit 规范化 limit 参数。
func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	return limit
}

// escapeSQL 转义单引号，防止 SQL 字符串断裂。
func escapeSQL(v string) string {
	return strings.ReplaceAll(v, "'", "''")
}
