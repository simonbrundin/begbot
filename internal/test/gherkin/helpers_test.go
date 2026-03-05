//go:build gherkin
// +build gherkin

package gherkin

import (
	"os"
	"path/filepath"
)

// getFeaturesPath returns the absolute path to a feature file inside the features/ directory.
// When running "go test", the working directory is the package directory, so we can resolve
// the features folder relative to the current working directory.
func getFeaturesPath(filename string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Join("features", filename)
	}
	return filepath.Join(cwd, "features", filename)
}
