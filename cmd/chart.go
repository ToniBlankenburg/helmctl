package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// resolveChartPath returns the path to the chart archive for appName inside
// chartsDir. It prefers <name>.tgz (the helmctl convention) but falls back to
// the lexicographically last <name>-[0-9]*.tgz it finds, which is the default
// output of `helm package`. An error is returned when neither form exists.
func resolveChartPath(chartsDir, appName string) (string, error) {
	plain := filepath.Join(chartsDir, appName+".tgz")
	if _, err := os.Stat(plain); err == nil {
		return plain, nil
	}

	// Standard helm package naming: <name>-<semver>.tgz
	pattern := filepath.Join(chartsDir, appName+"-[0-9]*.tgz")
	matches, _ := filepath.Glob(pattern)
	if len(matches) > 0 {
		// Lexicographic sort gives us the latest semver when the major/minor
		// prefix is the same (e.g. 0.1.0 < 0.2.0 < 1.0.0).
		return matches[len(matches)-1], nil
	}

	return "", fmt.Errorf("chart not found: %s — did you build the chart?", plain)
}
