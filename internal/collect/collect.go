// Package collect 实现 mesh collect 命令。
package collect

import (
	"crypto/rand"
	"encoding/hex"
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

// Run 执行 mesh collect 命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("collect", flag.ContinueOnError)
	fs.SetOutput(stderr)

	source := fs.String("source", "", "数据来源（必填）")
	content := fs.String("content", "", "认知内容（必填）")
	tag := fs.String("tag", "", "标签（可选，逗号分隔）")
	summary := fs.String("summary", "", "摘要（可选）")
	fs.Usage = func() {
		printUsage(fs.Output())
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "collect: 不支持的位置参数: %s\n", strings.Join(fs.Args(), " "))
		printUsage(stderr)
		return 2
	}

	if err := validateRequired(*source, *content); err != nil {
		_, _ = fmt.Fprintf(stderr, "collect: %v\n", err)
		printUsage(stderr)
		return 2
	}

	dbPath, err := resolveDBPath()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "collect: 解析数据库路径失败: %v\n", err)
		return 1
	}

	dataStore, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "collect: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = dataStore.Close() }()

	id, err := newUUID()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "collect: 生成 ID 失败: %v\n", err)
		return 1
	}

	insight := &types.Insight{
		ID:        id,
		Source:    strings.TrimSpace(*source),
		Content:   strings.TrimSpace(*content),
		Summary:   strings.TrimSpace(*summary),
		Tags:      strings.TrimSpace(*tag),
		CreatedAt: time.Now().Unix(),
	}
	if err := dataStore.Insert(insight); err != nil {
		_, _ = fmt.Fprintf(stderr, "collect: 写入失败: %v\n", err)
		return 1
	}

	printSuccess(stdout, insight)
	return 0
}

// validateRequired 验证必填参数。
func validateRequired(source string, content string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("缺少必填参数 --source")
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("缺少必填参数 --content")
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

// newUUID 生成 UUID v4 字符串。
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(b[:])
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32],
	), nil
}

// ResolveDBPathForReuse 暴露数据库路径解析能力给其他命令复用。
func ResolveDBPathForReuse() (string, error) {
	return resolveDBPath()
}

// NewUUIDForReuse 暴露 UUID 生成能力给其他命令复用。
func NewUUIDForReuse() (string, error) {
	return newUUID()
}

// printSuccess 输出采集成功提示。
func printSuccess(w io.Writer, insight *types.Insight) {
	_, _ = fmt.Fprintln(w, "✅ 已采集 1 条认知")
	_, _ = fmt.Fprintf(w, "   来源: %s\n", insight.Source)
	if insight.Tags == "" {
		_, _ = fmt.Fprintln(w, "   标签: -")
		return
	}
	_, _ = fmt.Fprintf(w, "   标签: %s\n", insight.Tags)
}

// printUsage 输出 collect 命令帮助信息。
func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh collect --source <source> --content <content> [--tag <tags>] [--summary <summary>]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --source    数据来源（必填）")
	_, _ = fmt.Fprintln(w, "  --content   认知内容（必填）")
	_, _ = fmt.Fprintln(w, "  --tag       标签（可选，逗号分隔）")
	_, _ = fmt.Fprintln(w, "  --summary   摘要（可选）")
}
