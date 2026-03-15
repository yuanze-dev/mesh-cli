// Package initcmd 实现 mesh init 命令。
package initcmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"mesh-cli/internal/collect"
	"mesh-cli/internal/config"
	"mesh-cli/pkg/types"
)

// Run 执行 mesh init 命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	syncSpace := fs.String("sync-space", "", "同步空间路径（必填）")
	fs.Usage = func() { printUsage(fs.Output()) }

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "init: 不支持的位置参数: %s\n", strings.Join(fs.Args(), " "))
		printUsage(stderr)
		return 2
	}
	if strings.TrimSpace(*syncSpace) == "" {
		_, _ = fmt.Fprintln(stderr, "init: 缺少必填参数 --sync-space")
		printUsage(stderr)
		return 2
	}

	absSync, err := filepath.Abs(strings.TrimSpace(*syncSpace))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "init: 解析同步路径失败: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(absSync, 0o755); err != nil {
		_, _ = fmt.Fprintf(stderr, "init: 创建同步目录失败: %v\n", err)
		return 1
	}

	dbPath, err := collect.ResolveDBPathForReuse()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "init: 解析数据库路径失败: %v\n", err)
		return 1
	}
	if strings.TrimSpace(dbPath) == "" {
		_, _ = fmt.Fprintln(stderr, "init: 数据库路径为空")
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		_, _ = fmt.Fprintf(stderr, "init: 创建本地目录失败: %v\n", err)
		return 1
	}

	cfg := types.Config{SyncSpace: absSync, DBPath: dbPath}
	cfgPath, err := config.Save(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "init: 保存配置失败: %v\n", err)
		return 1
	}
	if err := validate(cfg); err != nil {
		_, _ = fmt.Fprintf(stderr, "init: 配置校验失败: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(stdout, "✅ 初始化完成\n配置: %s\n同步目录: %s\n", cfgPath, absSync)
	return 0
}

func validate(cfg types.Config) error {
	if strings.TrimSpace(cfg.SyncSpace) == "" {
		return errors.New("sync_space is empty")
	}
	if strings.TrimSpace(cfg.DBPath) == "" {
		return errors.New("db_path is empty")
	}
	if st, err := os.Stat(cfg.SyncSpace); err != nil || !st.IsDir() {
		return errors.New("sync_space not exists")
	}
	if _, err := os.Stat(filepath.Dir(cfg.DBPath)); err != nil {
		return fmt.Errorf("db directory not exists: %w", err)
	}
	return nil
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh init --sync-space <path>")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --sync-space   同步空间路径（必填）")
}
