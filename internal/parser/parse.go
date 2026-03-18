package parser

import (
	"context"
	"os"
	"path/filepath"
	"repoDepsCheckerCLI/internal/types"

	"golang.org/x/mod/modfile"
)

func Parse(ctx context.Context, dirpath string) (*types.Info, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	gomodpath := filepath.Join(dirpath, "go.mod")

	data, err := os.ReadFile(gomodpath)
	if err != nil {
		return nil, err
	}

	f, err := modfile.ParseLax(gomodpath, data, nil)
	if err != nil {
		return nil, err
	}

	deps := make(map[string]string, len(f.Require))

	for _, req := range f.Require {
		deps[req.Mod.Path] = req.Mod.Version
	}

	info := &types.Info{
		ModuleName:  f.Module.Mod.Path,
		GoVersion:   func() string {
			if f.Go != nil { return f.Go.Version }
			return ""
		}(),
		DepVersions: deps,
	}

	return info, nil
}
