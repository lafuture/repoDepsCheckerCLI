package output

import (
	"encoding/json"
	"fmt"
	"io"
	"repoDepsCheckerCLI/internal/types"
	"strings"
	"text/tabwriter"
)

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiGreen   = "\033[32m"
)

func Print(info *types.Info, updates *types.Updates, format string, w io.Writer, useColor bool) error {
	switch format {
	case "table":
		return printTable(info, updates, w, useColor)
	case "json":
		return printJSON(info, updates, w)
	case "simple":
		return printSimple(info, updates, w, useColor)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func color(s, code string, useColor bool) string {
	if !useColor {
		return s
	}
	return code + s + ansiReset
}

func printTable(info *types.Info, updates *types.Updates, w io.Writer, useColor bool) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintf(tw, "\n\n%s\t%s\n", color("Название модуля:", ansiBold, useColor), info.ModuleName)
	fmt.Fprintf(tw, "%s\t%s\n\n", color("Версия Go:", ansiBold, useColor), info.GoVersion)

	deps := getDepsToUpdate(updates)
	if len(deps) == 0 {
		fmt.Fprintln(tw, color("Все зависимости актуальны.", ansiGreen, useColor))
		return nil
	}

	fmt.Fprintf(tw, "Доступны обновления:\n\n")

	fmt.Fprintf(tw, "%-45s\t%-18s\t%-18s\n", "Зависимости", "Текущие", "Последние")
	fmt.Fprintln(tw, strings.Repeat("-", 85))

	for _, d := range deps {
		fmt.Fprintf(tw, "%-45s\t%-18s\t%-18s\n", d.Dep, d.CurrentVersion, color(d.LatestVersion, ansiGreen, useColor))
	}

	return nil
}

func printJSON(info *types.Info, updates *types.Updates, w io.Writer) error {
	deps := getDepsToUpdate(updates)
	updatesMap := make(map[string]string, len(deps))
	for _, d := range deps {
		updatesMap[d.Dep] = d.LatestVersion
	}

	result := struct {
		Module    string            `json:"module"`
		GoVersion string            `json:"go_version"`
		Updates   map[string]string  `json:"updates"`
		CheckedAt string            `json:"checked_at"`
	}{
		Module:    info.ModuleName,
		GoVersion: info.GoVersion,
		Updates:   updatesMap,
		CheckedAt: updates.CheckedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func printSimple(info *types.Info, updates *types.Updates, w io.Writer, useColor bool) error {
	fmt.Fprintf(w, "%s\n", color(info.ModuleName, ansiBold, useColor))
	fmt.Fprintf(w, "%s\n", info.GoVersion)
	for _, d := range getDepsToUpdate(updates) {
		fmt.Fprintf(w, "%s: %s -> %s\n", d.Dep, d.CurrentVersion, color(d.LatestVersion, ansiGreen, useColor))
	}
	return nil
}

func getDepsToUpdate(updates *types.Updates) []types.DepToUpdate {
	if updates == nil || updates.DepsToUpdate == nil {
		return nil
	}
	return *updates.DepsToUpdate
}
