// Package syncer 实现 mesh push/pull/sync 命令。
package syncer

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mesh-cli/internal/config"
)

const (
	syncDBName = "mesh.db"
	lockName   = ".mesh.sync.lock"
)

// RunPush 执行 mesh push。
func RunPush(args []string, stdout io.Writer, stderr io.Writer) int {
	if !parseNoArg("push", args, stderr) {
		return 2
	}
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "push: 读取配置失败，请先执行 mesh init: %v\n", err)
		return 1
	}
	lockPath := filepath.Join(cfg.SyncSpace, lockName)
	release, err := acquireLock(lockPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "push: 获取文件锁失败: %v\n", err)
		return 1
	}
	defer release()

	bytes, err := copyFile(cfg.DBPath, filepath.Join(cfg.SyncSpace, syncDBName))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "push: 复制失败: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "✅ push 完成，已推送 %d bytes\n", bytes)
	return 0
}

// RunPull 执行 mesh pull。
func RunPull(args []string, stdout io.Writer, stderr io.Writer) int {
	if !parseNoArg("pull", args, stderr) {
		return 2
	}
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "pull: 读取配置失败，请先执行 mesh init: %v\n", err)
		return 1
	}
	lockPath := filepath.Join(cfg.SyncSpace, lockName)
	release, err := acquireLock(lockPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "pull: 获取文件锁失败: %v\n", err)
		return 1
	}
	defer release()

	bytes, err := copyFile(filepath.Join(cfg.SyncSpace, syncDBName), cfg.DBPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "pull: 复制失败: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "✅ pull 完成，已拉取 %d bytes\n", bytes)
	return 0
}

// RunSync 执行 mesh sync（先 pull 再 push）。
func RunSync(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintln(fs.Output(), "Usage:")
		_, _ = fmt.Fprintln(fs.Output(), "  mesh sync")
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "sync: 不支持的位置参数: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "sync: 读取配置失败，请先执行 mesh init: %v\n", err)
		return 1
	}

	lockPath := filepath.Join(cfg.SyncSpace, lockName)
	release, err := acquireLock(lockPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "sync: 获取文件锁失败: %v\n", err)
		return 1
	}
	defer release()

	remoteDB := filepath.Join(cfg.SyncSpace, syncDBName)
	pullBytes := int64(0)
	if _, err := os.Stat(remoteDB); err == nil {
		backupPath, err := backupLocalDBIfNeeded(cfg.DBPath, remoteDB)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "sync: 备份本地库失败: %v\n", err)
			return 1
		}
		if backupPath != "" {
			_, _ = fmt.Fprintf(stdout, "ℹ️ 检测到潜在冲突，已备份本地库: %s\n", backupPath)
		}

		pullBytes, err = copyFile(remoteDB, cfg.DBPath)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "sync: pull 失败: %v\n", err)
			return 1
		}
	}

	pushBytes, err := copyFile(cfg.DBPath, remoteDB)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "sync: push 失败: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "✅ sync 完成（pull: %d bytes, push: %d bytes）\n", pullBytes, pushBytes)
	return 0
}

func parseNoArg(name string, args []string, stderr io.Writer) bool {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return false
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "%s: 不支持的位置参数: %s\n", name, strings.Join(fs.Args(), " "))
		return false
	}
	return true
}

func copyFile(src string, dst string) (int64, error) {
	if strings.TrimSpace(src) == "" || strings.TrimSpace(dst) == "" {
		return 0, errors.New("empty path")
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, fmt.Errorf("create destination directory: %w", err)
	}
	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return 0, err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return 0, err
	}
	defer func() { _ = out.Close() }()

	n, err := io.Copy(out, in)
	if err != nil {
		return 0, err
	}
	if err := out.Sync(); err != nil {
		return 0, err
	}
	return n, nil
}

func backupLocalDBIfNeeded(localDB string, remoteDB string) (string, error) {
	localInfo, err := os.Stat(localDB)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	remoteInfo, err := os.Stat(remoteDB)
	if err != nil {
		return "", err
	}

	// 简单冲突判定：双方都存在且大小或修改时间不一致时，先备份本地库。
	if localInfo.Size() == remoteInfo.Size() && localInfo.ModTime().Equal(remoteInfo.ModTime()) {
		return "", nil
	}

	backupPath := fmt.Sprintf("%s.bak.%d", localDB, time.Now().Unix())
	if _, err := copyFile(localDB, backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

func acquireLock(lockPath string) (func(), error) {
	deadline := time.Now().Add(3 * time.Second)
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = f.WriteString(fmt.Sprintf("pid=%d\ntime=%d\n", os.Getpid(), time.Now().Unix()))
			_ = f.Close()
			return func() { _ = os.Remove(lockPath) }, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, errors.New("lock busy")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
