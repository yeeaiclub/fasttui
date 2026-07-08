//go:build integration

package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExamplesBuild(t *testing.T) {
	root := moduleRoot(t)
	examples := []string{"chat", "input", "key", "select"}

	for _, name := range examples {
		t.Run(name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), name)
			cmd := exec.Command("go", "build", "-o", out, "./_examples/"+name)
			cmd.Dir = root
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build _examples/%s failed: %v\n%s", name, err, output)
			}
		})
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
