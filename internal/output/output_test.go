package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"repoDepsCheckerCLI/internal/types"
)

func TestPrint(t *testing.T) {
	info := &types.Info{
		ModuleName:  "example.com/project",
		GoVersion:   "1.21",
		DepVersions: map[string]string{"github.com/foo/bar": "v1.0.0"},
	}

	updates := &types.Updates{
		DepsToUpdate: &[]types.DepToUpdate{
			{Dep: "github.com/foo/bar", CurrentVersion: "v1.0.0", LatestVersion: "v1.2.0"},
		},
		CheckedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := Print(info, updates, "table", &buf, false)
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "example.com/project") || !strings.Contains(out, "1.21") || !strings.Contains(out, "github.com/foo/bar") {
		t.Errorf("output should contain module, version and dep: %s", out)
	}
}
