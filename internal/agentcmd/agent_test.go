package agentcmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRegisterAndList(t *testing.T) {
	t.Setenv("MESH_DB_PATH", t.TempDir()+"/mesh.db")

	var out bytes.Buffer
	var err bytes.Buffer
	code := Run([]string{"register", "--id", "a1", "--name", "Codex", "--type", "dev", "--device", "mini2"}, &out, &err)
	if code != 0 {
		t.Fatalf("register code=%d stderr=%s", code, err.String())
	}

	out.Reset()
	err.Reset()
	code = Run([]string{"list"}, &out, &err)
	if code != 0 {
		t.Fatalf("list code=%d stderr=%s", code, err.String())
	}
	if !strings.Contains(out.String(), "Codex") {
		t.Fatalf("expected Codex in list, got: %s", out.String())
	}
}
