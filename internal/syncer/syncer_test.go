package syncer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupLocalDBIfNeeded_CreateBackupWhenDifferent(t *testing.T) {
	dir := t.TempDir()
	local := filepath.Join(dir, "local.db")
	remote := filepath.Join(dir, "remote.db")

	if err := os.WriteFile(local, []byte("local"), 0o644); err != nil {
		t.Fatalf("write local: %v", err)
	}
	if err := os.WriteFile(remote, []byte("remote-content"), 0o644); err != nil {
		t.Fatalf("write remote: %v", err)
	}
	// 确保 mtime 有差异，避免同一秒写入导致误判相同。
	time.Sleep(1100 * time.Millisecond)
	if err := os.Chtimes(remote, time.Now(), time.Now()); err != nil {
		t.Fatalf("chtimes remote: %v", err)
	}

	backup, err := backupLocalDBIfNeeded(local, remote)
	if err != nil {
		t.Fatalf("backupLocalDBIfNeeded: %v", err)
	}
	if backup == "" {
		t.Fatalf("expected backup path, got empty")
	}
	if !strings.Contains(backup, ".bak.") {
		t.Fatalf("unexpected backup path: %s", backup)
	}
	b, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(b) != "local" {
		t.Fatalf("backup content mismatch: %q", string(b))
	}
}
