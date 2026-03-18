package cloner

import (
	"context"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func Clone(ctx context.Context, url string, token string) (string, func(), error) {
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	default:
	}

	tempDir, err := os.MkdirTemp("", "repo-deps-checker-")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	ins := &git.CloneOptions{
		URL:          url,
		Progress:     nil,
		Depth:        1,
		SingleBranch: true,
	}

	if token != "" {
		ins.Auth = &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	}

	if _, err := git.PlainClone(tempDir, false, ins); err != nil {
		cleanup()
		return "", nil, err
	}

	return tempDir, cleanup, nil
}
