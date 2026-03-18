package checker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"repoDepsCheckerCLI/internal/types"
)

func TestCheckUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.Result{Version: "v0.15.0", Time: time.Now()}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	info := &types.Info{
		ModuleName: "example.com/project",
		GoVersion:  "1.21",
		DepVersions: map[string]string{"golang.org/x/mod": "v0.10.0"},
	}

	opts := types.CheckOptions{
		MaxConcurrency: 5,
		MaxRetries:     1,
		HTTPClient:     server.Client(),
		ProxyBaseURL:   server.URL,
		NoCache:        true,
	}

	updates, err := CheckUpdates(context.Background(), info, opts)
	if err != nil {
		t.Fatalf("CheckUpdates failed: %v", err)
	}
	if updates == nil || updates.DepsToUpdate == nil {
		t.Fatal("updates should not be nil")
	}
	deps := *updates.DepsToUpdate
	if len(deps) != 1 || deps[0].Dep != "golang.org/x/mod" || deps[0].LatestVersion != "v0.15.0" {
		t.Errorf("unexpected result: %v", deps)
	}
}
