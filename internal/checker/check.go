package checker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"repoDepsCheckerCLI/internal/types"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
)

type cacheEntry struct {
	Version  string    `json:"version"`
	CachedAt time.Time `json:"cached_at"`
}

func cacheKey(modulePath string) string {
	h := sha256.Sum256([]byte(modulePath))
	return hex.EncodeToString(h[:16])
}

func isLastVersion(cur, last string) bool {
	cmp := semver.Compare(cur, last)

	if cmp < 0 {
		return false
	}

	return true
}

func getLastVersion(ctx context.Context, url string, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result types.Result

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Version, nil
}

func getLastVersionWithRetry(ctx context.Context, url string, client *http.Client, maxRetries int) (string, error) {
	var err error
	for i := 0; i < maxRetries; i++ {
		var version string
		version, err = getLastVersion(ctx, url, client)
		if err == nil {
			return version, nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(1+i) * time.Second):
		}
	}
	return "", fmt.Errorf("after %d retries: %w", maxRetries, err)
}

func getLastVersionCached(ctx context.Context, dep, url string, client *http.Client, maxRetries int, cacheDir string, cacheTTL time.Duration, noCache bool) (string, error) {
	if !noCache && cacheDir != "" && cacheTTL > 0 {
		key := cacheKey(dep)
		path := filepath.Join(cacheDir, key+".json")
		if data, err := os.ReadFile(path); err == nil {
			var entry cacheEntry
			if json.Unmarshal(data, &entry) == nil && entry.Version != "" {
				if time.Since(entry.CachedAt) < cacheTTL {
					return entry.Version, nil
				}
			}
		}
	}

	version, err := getLastVersionWithRetry(ctx, url, client, maxRetries)
	if err != nil {
		return "", err
	}

	if !noCache && cacheDir != "" {
		key := cacheKey(dep)
		path := filepath.Join(cacheDir, key+".json")
		entry := cacheEntry{Version: version, CachedAt: time.Now()}
		if data, err := json.Marshal(entry); err == nil {
			_ = os.MkdirAll(cacheDir, 0755)
			_ = os.WriteFile(path, data, 0644)
		}
	}

	return version, nil
}

func CheckUpdates(ctx context.Context, info *types.Info, opts types.CheckOptions) (*types.Updates, error) {
	if opts.MaxConcurrency == 0 {
		opts.MaxConcurrency = 10
	}

	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}

	depsToUpdate := make([]types.DepToUpdate, 0)
	sem := make(chan struct{}, opts.MaxConcurrency)
	total := len(info.DepVersions)

	var done atomic.Int64
	var errsMu sync.Mutex
	var errs []types.DepError
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)

	proxyBase := opts.ProxyBaseURL
	if proxyBase == "" {
		proxyBase = "https://proxy.golang.org"
	}
	for dep, version := range info.DepVersions {
		dep, version := dep, version
		url := proxyBase + "/" + dep + "/@latest"
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			last, err := getLastVersionCached(ctx, dep, url, opts.HTTPClient, opts.MaxRetries, opts.CacheDir, opts.CacheTTL, opts.NoCache)

			if opts.OnProgress != nil {
				opts.OnProgress(int(done.Add(1)), total)
			}

			if err != nil {
				errsMu.Lock()
				errs = append(errs, types.DepError{dep, err})
				errsMu.Unlock()
				return nil
			}

			if !isLastVersion(version, last) {
				mu.Lock()
				depsToUpdate = append(depsToUpdate, types.DepToUpdate{Dep: dep, CurrentVersion: version, LatestVersion: last})
				mu.Unlock()
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if len(errs) > 0 {
		var errMsgs []string
		for _, e := range errs {
			errMsgs = append(errMsgs, fmt.Sprintf("%s: %v", e.Dep, e.Err))
		}
		combinedErr := fmt.Errorf("failed to check: %s", strings.Join(errMsgs, "; "))

		return &types.Updates{DepsToUpdate: &depsToUpdate, CheckedAt: time.Now()}, combinedErr
	}

	return &types.Updates{DepsToUpdate: &depsToUpdate, CheckedAt: time.Now()}, nil
}
