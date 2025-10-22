package exec

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// DetectFramework detects the test framework (jest or vitest) from package.json and lockfiles.
func DetectFramework(root string) (string, error) {
	pkgPath := filepath.Join(root, "package.json")
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return "", fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(content, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Check devDependencies
	devDeps, ok := pkg["devDependencies"].(map[string]interface{})
	if !ok {
		devDeps = make(map[string]interface{})
	}

	deps, ok := pkg["dependencies"].(map[string]interface{})
	if !ok {
		deps = make(map[string]interface{})
	}

	hasVitest := false
	hasJest := false

	if _, ok := devDeps["vitest"]; ok {
		hasVitest = true
	}
	if _, ok := devDeps["jest"]; ok {
		hasJest = true
	}
	if _, ok := deps["vitest"]; ok {
		hasVitest = true
	}
	if _, ok := deps["jest"]; ok {
		hasJest = true
	}

	// Prefer vitest if both are present
	if hasVitest {
		return "vitest", nil
	}
	if hasJest {
		return "jest", nil
	}

	// Check lockfiles as fallback
	if _, err := os.Stat(filepath.Join(root, "pnpm-lock.yaml")); err == nil {
		// pnpm lockfile exists; check for vitest/jest in it
		if hasFrameworkInLockfile(root, "vitest") {
			return "vitest", nil
		}
		if hasFrameworkInLockfile(root, "jest") {
			return "jest", nil
		}
	}

	return "", fmt.Errorf("no test framework detected (jest or vitest required)")
}

// hasFrameworkInLockfile checks if a framework is mentioned in lockfiles.
func hasFrameworkInLockfile(root string, framework string) bool {
	lockfiles := []string{
		filepath.Join(root, "pnpm-lock.yaml"),
		filepath.Join(root, "yarn.lock"),
		filepath.Join(root, "package-lock.json"),
	}

	for _, lockfile := range lockfiles {
		content, err := os.ReadFile(lockfile)
		if err != nil {
			continue
		}
		if strings.Contains(string(content), framework) {
			return true
		}
	}

	return false
}

// RunTests runs the test suite on the specified test files.
func RunTests(testPaths []string, framework string, root string) error {
	if len(testPaths) == 0 {
		return nil
	}

	// Determine the test command
	var cmd *exec.Cmd
	if framework == "vitest" {
		cmd = exec.Command("npm", "run", "test", "--", testPaths[0])
	} else {
		cmd = exec.Command("npm", "run", "test", "--", testPaths[0])
	}

	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test run failed: %w", err)
	}

	return nil
}

// GetCoverage retrieves the test coverage percentage.
func GetCoverage(root string, framework string) (float64, error) {
	// Try to run coverage command
	var cmd *exec.Cmd

	// Check if test:coverage script exists
	if hasScript(root, "test:coverage") {
		cmd = exec.Command("npm", "run", "test:coverage")
	} else if framework == "vitest" {
		cmd = exec.Command("npm", "run", "test", "--", "--coverage")
	} else {
		cmd = exec.Command("npm", "run", "test", "--", "--coverage")
	}

	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("coverage command failed: %w", err)
	}

	// Parse coverage from output (simple heuristic)
	coverage := parseCoverageFromOutput(string(output))
	return coverage, nil
}

// hasScript checks if a script exists in package.json.
func hasScript(root string, scriptName string) bool {
	pkgPath := filepath.Join(root, "package.json")
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(content, &pkg); err != nil {
		return false
	}

	scripts, ok := pkg["scripts"].(map[string]interface{})
	if !ok {
		return false
	}

	_, exists := scripts[scriptName]
	return exists
}

// parseCoverageFromOutput extracts coverage percentage from test output.
func parseCoverageFromOutput(output string) float64 {
	// Look for patterns like "Coverage: 85.5%" or "Statements: 92.3%" or "85.5% coverage"
	patterns := []string{
		"Coverage: ",
		"coverage: ",
		"Statements: ",
		"Statements   : ",
		"% coverage",
		"% Statements",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(output, pattern); idx != -1 {
			// Extract number before or after pattern
			start := idx
			if strings.HasSuffix(pattern, " ") || strings.HasSuffix(pattern, ": ") {
				start = idx + len(pattern)
			}

			// Find the number
			var numStr string
			for i := start; i < len(output) && i < start+10; i++ {
				ch := output[i]
				if (ch >= '0' && ch <= '9') || ch == '.' {
					numStr += string(ch)
				} else if numStr != "" {
					break
				}
			}

			if numStr != "" {
				if val, err := strconv.ParseFloat(numStr, 64); err == nil {
					return val
				}
			}
		}
	}

	return 0
}
