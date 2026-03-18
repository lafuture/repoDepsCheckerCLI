package parser

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	dir := t.TempDir()
	gomod := `module example.com/myproject

go 1.21

require (
	github.com/foo/bar v1.2.3
	golang.org/x/mod v0.10.0
)`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := Parse(context.Background(), dir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if info.ModuleName != "example.com/myproject" {
		t.Errorf("ModuleName = %q, want example.com/myproject", info.ModuleName)
	}
	if info.GoVersion != "1.21" {
		t.Errorf("GoVersion = %q, want 1.21", info.GoVersion)
	}
	if len(info.DepVersions) != 2 {
		t.Errorf("DepVersions length = %d, want 2", len(info.DepVersions))
	}
}
