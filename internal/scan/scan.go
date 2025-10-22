package scan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
)

// FindCandidates returns a list of TypeScript/TSX files that don't have corresponding test files.
func FindCandidates(root string, changedOnly bool) ([]string, error) {
	var candidates []string

	if changedOnly {
		changed, err := ChangedFiles(root)
		if err != nil {
			return nil, fmt.Errorf("failed to get changed files: %w", err)
		}
		candidates = changed
	} else {
		all, err := AllTypeScriptFiles(root)
		if err != nil {
			return nil, fmt.Errorf("failed to scan files: %w", err)
		}
		candidates = all
	}

	// Filter: keep only files without tests
	var result []string
	for _, candidate := range candidates {
		if HasTest(candidate) {
			continue
		}
		result = append(result, candidate)
	}

	return result, nil
}

// AllTypeScriptFiles returns all TypeScript/TSX files in the root, excluding node_modules, .d.ts, and test files.
func AllTypeScriptFiles(root string) ([]string, error) {
	pattern := filepath.Join(root, "**/*.{ts,tsx}")
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, match := range matches {
		// Skip node_modules
		if strings.Contains(match, "node_modules") {
			continue
		}
		// Skip .d.ts files
		if strings.HasSuffix(match, ".d.ts") {
			continue
		}
		// Skip test files
		if strings.Contains(match, ".test.") || strings.Contains(match, ".spec.") {
			continue
		}
		// Skip build/dist
		if strings.Contains(match, "/build/") || strings.Contains(match, "/dist/") {
			continue
		}
		result = append(result, match)
	}

	return result, nil
}

// HasTest checks if a TypeScript file has a corresponding test file.
func HasTest(tsPath string) bool {
	base := strings.TrimSuffix(tsPath, filepath.Ext(tsPath))
	testPaths := []string{
		base + ".test.ts",
		base + ".test.tsx",
		base + ".spec.ts",
		base + ".spec.tsx",
	}

	for _, testPath := range testPaths {
		if _, err := os.Stat(testPath); err == nil {
			return true
		}
	}

	return false
}

// DefaultTestPath returns the default test file path for a given source file.
// If outDir is provided, mirrors the structure under outDir; otherwise places next to source.
func DefaultTestPath(tsPath string, framework string, outDir string) string {
	ext := ".test.ts"
	if framework == "vitest" {
		ext = ".spec.ts"
	}

	base := strings.TrimSuffix(tsPath, filepath.Ext(tsPath))

	if outDir == "" {
		return base + ext
	}

	// Mirror structure under outDir
	rel, err := filepath.Rel(".", tsPath)
	if err != nil {
		rel = tsPath
	}
	base = strings.TrimSuffix(rel, filepath.Ext(rel))
	return filepath.Join(outDir, base+ext)
}

// ChangedFiles returns TypeScript/TSX files changed against origin/main.
func ChangedFiles(root string) ([]string, error) {
	repo, err := git.PlainOpen(root)
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	// Try to get origin/main
	remoteRef, err := repo.Reference("refs/remotes/origin/main", true)
	if err != nil {
		// Fallback to origin/master
		remoteRef, err = repo.Reference("refs/remotes/origin/master", true)
		if err != nil {
			return nil, fmt.Errorf("failed to find origin/main or origin/master: %w", err)
		}
	}

	// Use git diff to get changed files
	cmd := exec.Command("git", "-C", root, "diff", "--name-only", remoteRef.Hash().String()+"...HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	var result []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Only include TypeScript/TSX files
		if strings.HasSuffix(line, ".ts") || strings.HasSuffix(line, ".tsx") {
			// Skip .d.ts, test files, node_modules, build/dist
			if strings.HasSuffix(line, ".d.ts") || strings.Contains(line, ".test.") || strings.Contains(line, ".spec.") {
				continue
			}
			if strings.Contains(line, "node_modules") || strings.Contains(line, "/build/") || strings.Contains(line, "/dist/") {
				continue
			}
			fullPath := filepath.Join(root, line)
			result = append(result, fullPath)
		}
	}

	return result, nil
}

// IsWorkingTreeDirty checks if the git working tree has uncommitted changes.
func IsWorkingTreeDirty(root string) (bool, error) {
	repo, err := git.PlainOpen(root)
	if err != nil {
		return false, fmt.Errorf("not a git repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return !status.IsClean(), nil
}
