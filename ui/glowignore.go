package ui

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// readGlowignore reads a .glowignore file from the given directory
// and returns the patterns found. Returns an empty slice if the file
// doesn't exist or can't be read.
func readGlowignore(dir string) []string {
	path := filepath.Join(dir, ".glowignore")
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}

	return patterns
}
