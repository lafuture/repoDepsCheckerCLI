package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"repoDepsCheckerCLI/internal/checker"
	"repoDepsCheckerCLI/internal/cloner"
	"repoDepsCheckerCLI/internal/output"
	"repoDepsCheckerCLI/internal/parser"
	"repoDepsCheckerCLI/internal/types"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	token          string
	format         string
	outputFile     string
	maxConcurrency int
	maxRetries     int
	noProgress     bool
	noColor        bool
	noCache        bool
)

var Version = "dev"

func init() {
	rootCmd.Flags().StringVarP(&token, "token", "t", "", "Токен для приватных репозиториев")

	rootCmd.Flags().StringVarP(&format, "format", "f", "table", "Формат вывода: table, json, simple")

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Указать файл для сохранения результата")

	rootCmd.Flags().BoolVar(&noProgress, "no-progress", false, "Отключить progress bar")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Отключить цветной вывод")
	rootCmd.Flags().BoolVar(&noCache, "no-cache", false, "Игнорировать кеш, всегда запрашивать с proxy")

	rootCmd.Flags().IntVar(&maxConcurrency, "concurrency", 10, "Максимальное количество одновременных запросов к прокси")
	rootCmd.Flags().IntVar(&maxRetries, "retries", 3, "Максимальное количество повторных попыток проверки зависимости")
}

var rootCmd = &cobra.Command{
	Use:   "go-repo-deps-checker <repo-url>",
	Short: "Проверка зависимостей на обновления в Go репозитории",
	Args:  cobra.ExactArgs(1),
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(),
			syscall.SIGINT,
			syscall.SIGTERM,
		)
		defer cancel()

		repoURL := args[0]

		authToken := token
		if authToken == "" {
			authToken = os.Getenv("GITHUB_TOKEN")
		}

		tmpDir, cleanup, err := cloner.Clone(ctx, repoURL, authToken)
		if err != nil {
			return fmt.Errorf("clone failed: %w", err)
		}
		defer cleanup()

		info, err := parser.Parse(ctx, tmpDir)
		if err != nil {
			return fmt.Errorf("parse failed: %w", err)
		}

		total := len(info.DepVersions)

		cacheDir := ""
		if !noCache {
			if dir, err := os.UserCacheDir(); err == nil {
				cacheDir = filepath.Join(dir, "repo-deps-checker")
			}
		}

		opts := types.CheckOptions{
			MaxConcurrency: maxConcurrency,
			MaxRetries:     maxRetries,
			HTTPClient: &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 10,
					IdleConnTimeout:     90 * time.Second,
				},
			},
			CacheDir: cacheDir,
			CacheTTL: time.Hour,
			NoCache:  noCache,
		}

		if !noProgress {
			bar := progressbar.NewOptions(total,
				progressbar.OptionSetDescription("Проверяются зависимости..."),
				progressbar.OptionShowCount(),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "█",
					SaucerPadding: "░",
					BarStart:      "[",
					BarEnd:        "]",
				}),
			)
			opts.OnProgress = func(done, total int) {
				bar.Add(1)
			}
		}

		updates, checkErr := checker.CheckUpdates(ctx, info, opts)

		writer := os.Stdout
		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer f.Close()
			writer = f
		}

		if updates != nil {
			useColor := !noColor && outputFile == ""
			if err := output.Print(info, updates, format, writer, useColor); err != nil {
				return fmt.Errorf("output: %w", err)
			}
		}

		if checkErr != nil {
			fmt.Fprintf(os.Stderr, "\nПредупреждения:\n\n%v\n", checkErr)
			return fmt.Errorf("check updates failed: %w", checkErr)
		}

		return nil
	},
}

func Execute() {
	rootCmd.InitDefaultHelpFlag()
	if f := rootCmd.Flags().Lookup("help"); f != nil {
		f.Usage = "Показать справку по команде"
	}
	rootCmd.InitDefaultVersionFlag()
	if f := rootCmd.Flags().Lookup("version"); f != nil {
		f.Usage = "Показать версию"
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
