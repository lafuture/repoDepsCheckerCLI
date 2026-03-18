package types

import (
	"net/http"
	"time"
)

type Info struct {
	ModuleName string
	GoVersion  string

	DepVersions map[string]string
}

type Updates struct {
	DepsToUpdate *[]DepToUpdate
	CheckedAt    time.Time
}

type Result struct {
	Version string
	Time    time.Time
}

type DepError struct {
	Dep string
	Err error
}

type CheckOptions struct {
	MaxConcurrency int
	MaxRetries     int
	HTTPClient     *http.Client
	OnProgress     func(done, total int)
	CacheDir       string
	CacheTTL       time.Duration
	NoCache        bool
	ProxyBaseURL   string
}

type DepToUpdate struct {
	Dep            string
	CurrentVersion string
	LatestVersion  string
}
