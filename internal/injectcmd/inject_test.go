package injectcmd

import (
	"bytes"
	"strings"
	"testing"

	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

func TestRunInjectMarkdown(t *testing.T) {
	dbPath := t.TempDir() + "/mesh.db"
	t.Setenv("MESH_DB_PATH", dbPath)

	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer func() { _ = st.Close() }()

	_ = st.Insert(&types.Insight{ID: "i1", Source: "codex", Content: "实现了 inject 命令", Summary: "inject", Tags: "mesh,cli", CreatedAt: 1710000000})

	var out bytes.Buffer
	var errBuf bytes.Buffer
	code := Run([]string{"inject", "--format", "markdown", "--max-tokens", "200"}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "# Context Injection") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}
