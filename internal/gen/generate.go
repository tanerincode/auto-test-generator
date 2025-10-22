package gen

import (
	"fmt"
	"regexp"
	"strings"
)

// TestResult holds the result of test generation for a single file.
type TestResult struct {
	SourcePath string
	TestPath   string
	TestCode   string
	Error      error
}

// GenerateTest generates a test file for the given TypeScript source code.
// Uses basic regex-based analysis.
func GenerateTest(tsPath string, code string, framework string) (string, error) {
	// Extract exported symbols
	exports := extractExports(code)
	if len(exports) == 0 {
		return "", fmt.Errorf("no exported symbols found in %s", tsPath)
	}

	// Determine test framework syntax
	var testSyntax string
	if framework == "vitest" {
		testSyntax = "vitest"
	} else {
		testSyntax = "jest"
	}

	// Generate test code
	testCode := generateTestCode(tsPath, exports, code, testSyntax)
	return testCode, nil
}

// GenerateTestWithContext generates a test file using Augment CLI for code understanding.
// This provides more intelligent test generation based on actual code analysis.
func GenerateTestWithContext(tsPath string, code string, framework string, projectRoot string) (string, error) {
	return GenerateTestWithAugment(tsPath, code, framework, projectRoot)
}

// exportedSymbol represents an exported function or class.
type exportedSymbol struct {
	name      string
	kind      string // "function", "class", "const", "interface", "type"
	isAsync   bool
	params    []string
	isDefault bool
}

// extractExports parses TypeScript code and extracts exported symbols.
func extractExports(code string) []exportedSymbol {
	var exports []exportedSymbol

	// Match: export function name(...) or export const name = ...
	funcPattern := regexp.MustCompile(`export\s+(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)`)
	for _, match := range funcPattern.FindAllStringSubmatch(code, -1) {
		name := match[1]
		params := parseParams(match[2])
		isAsync := strings.Contains(match[0], "async")
		exports = append(exports, exportedSymbol{
			name:    name,
			kind:    "function",
			isAsync: isAsync,
			params:  params,
		})
	}

	// Match: export const name = (...) => or export const name = async (...) =>
	constPattern := regexp.MustCompile(`export\s+const\s+(\w+)\s*=\s*(?:async\s*)?\(([^)]*)\)\s*=>`)
	for _, match := range constPattern.FindAllStringSubmatch(code, -1) {
		name := match[1]
		params := parseParams(match[2])
		isAsync := strings.Contains(match[0], "async")
		exports = append(exports, exportedSymbol{
			name:    name,
			kind:    "const",
			isAsync: isAsync,
			params:  params,
		})
	}

	// Match: export class Name
	classPattern := regexp.MustCompile(`export\s+class\s+(\w+)`)
	for _, match := range classPattern.FindAllStringSubmatch(code, -1) {
		name := match[1]
		exports = append(exports, exportedSymbol{
			name: name,
			kind: "class",
		})
	}

	// Match: export default ...
	defaultPattern := regexp.MustCompile(`export\s+default\s+(?:function|class)?\s*(\w+)?`)
	if matches := defaultPattern.FindAllStringSubmatch(code, -1); len(matches) > 0 {
		match := matches[0]
		name := match[1]
		if name == "" {
			name = "default"
		}
		exports = append(exports, exportedSymbol{
			name:      name,
			kind:      "default",
			isDefault: true,
		})
	}

	return exports
}

// parseParams extracts parameter names from a parameter list.
func parseParams(paramStr string) []string {
	if strings.TrimSpace(paramStr) == "" {
		return nil
	}

	var params []string
	for _, param := range strings.Split(paramStr, ",") {
		param = strings.TrimSpace(param)
		// Extract name before : or =
		if idx := strings.IndexAny(param, ":="); idx != -1 {
			param = param[:idx]
		}
		param = strings.TrimSpace(param)
		if param != "" {
			params = append(params, param)
		}
	}
	return params
}

// generateTestCode creates the test file content.
func generateTestCode(tsPath string, exports []exportedSymbol, sourceCode string, testSyntax string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("/**\n")
	sb.WriteString(" * Auto-generated test file\n")
	sb.WriteString(" * Source: " + tsPath + "\n")
	sb.WriteString(" */\n\n")

	// Import statement
	importPath := strings.TrimSuffix(tsPath, ".ts")
	importPath = strings.TrimSuffix(importPath, ".tsx")
	sb.WriteString("import { ")

	for i, exp := range exports {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(exp.name)
	}
	sb.WriteString(" } from '../" + importPath + "';\n\n")

	// Test framework setup
	if testSyntax == "vitest" {
		sb.WriteString("import { describe, it, expect, beforeEach, afterEach } from 'vitest';\n\n")
	} else {
		sb.WriteString("import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';\n\n")
	}

	// Generate tests for each export
	for _, exp := range exports {
		sb.WriteString(generateTestForSymbol(exp))
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateTestForSymbol generates test cases for a single exported symbol.
func generateTestForSymbol(sym exportedSymbol) string {
	var sb strings.Builder

	sb.WriteString("describe('" + sym.name + "', () => {\n")

	switch sym.kind {
	case "function", "const":
		sb.WriteString(generateFunctionTests(sym))
	case "class":
		sb.WriteString(generateClassTests(sym))
	case "default":
		sb.WriteString(generateDefaultTests(sym))
	}

	sb.WriteString("});\n")
	return sb.String()
}

// generateFunctionTests generates test cases for a function.
func generateFunctionTests(sym exportedSymbol) string {
	var sb strings.Builder

	// Basic happy path test
	sb.WriteString("  it('should be defined', () => {\n")
	sb.WriteString("    expect(" + sym.name + ").toBeDefined();\n")
	sb.WriteString("  });\n\n")

	// If function has parameters, add a basic call test
	if len(sym.params) > 0 {
		sb.WriteString("  it('should handle basic input', () => {\n")
		sb.WriteString("    // TODO: Replace with actual test data\n")
		sb.WriteString("    const result = " + sym.name + "(")
		for i := range sym.params {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("null")
		}
		sb.WriteString(");\n")
		sb.WriteString("    expect(result).toBeDefined();\n")
		sb.WriteString("  });\n\n")
	}

	// If async, add async test
	if sym.isAsync {
		sb.WriteString("  it('should handle async operations', async () => {\n")
		sb.WriteString("    // TODO: Replace with actual test data\n")
		sb.WriteString("    const result = await " + sym.name + "(")
		for i := range sym.params {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("null")
		}
		sb.WriteString(");\n")
		sb.WriteString("    expect(result).toBeDefined();\n")
		sb.WriteString("  });\n\n")
	}

	// Edge case: null/undefined handling
	sb.WriteString("  it('should handle edge cases', () => {\n")
	sb.WriteString("    // TODO: Add edge case tests\n")
	sb.WriteString("    expect(true).toBe(true);\n")
	sb.WriteString("  });\n")

	return sb.String()
}

// generateClassTests generates test cases for a class.
func generateClassTests(sym exportedSymbol) string {
	var sb strings.Builder

	sb.WriteString("  it('should be instantiable', () => {\n")
	sb.WriteString("    const instance = new " + sym.name + "();\n")
	sb.WriteString("    expect(instance).toBeDefined();\n")
	sb.WriteString("  });\n\n")

	sb.WriteString("  it('should have expected methods', () => {\n")
	sb.WriteString("    const instance = new " + sym.name + "();\n")
	sb.WriteString("    // TODO: Add method existence checks\n")
	sb.WriteString("    expect(instance).toBeDefined();\n")
	sb.WriteString("  });\n")

	return sb.String()
}

// generateDefaultTests generates test cases for default exports.
func generateDefaultTests(sym exportedSymbol) string {
	var sb strings.Builder

	sb.WriteString("  it('should be defined', () => {\n")
	sb.WriteString("    expect(" + sym.name + ").toBeDefined();\n")
	sb.WriteString("  });\n")

	return sb.String()
}
