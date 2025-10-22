package gen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"bytes"
)

// AugmentCodeAnalysis represents the analysis result from Augment CLI
type AugmentCodeAnalysis struct {
	FilePath      string
	Exports       []ExportedFunction
	Dependencies  []string
	Complexity    string
	Description   string
	TestScenarios []TestScenario
}

// ExportedFunction represents a function/class exported from the module
type ExportedFunction struct {
	Name        string
	Type        string // "function", "class", "const", "interface"
	IsAsync     bool
	Parameters  []Parameter
	ReturnType  string
	Description string
}

// Parameter represents a function parameter
type Parameter struct {
	Name     string
	Type     string
	Optional bool
	Default  string
}

// TestScenario represents a test case scenario
type TestScenario struct {
	Name        string
	Description string
	Inputs      map[string]interface{}
	Expected    interface{}
	EdgeCase    bool
}

// AnalyzeWithAugment uses Augment CLI to analyze TypeScript code
func AnalyzeWithAugment(filePath string, code string, projectRoot string) (*AugmentCodeAnalysis, error) {
	// Check if augment CLI is available
	if _, err := exec.LookPath("augment"); err != nil {
		return nil, fmt.Errorf("augment CLI not found in PATH: %w", err)
	}

	// Create a temporary file with the code for analysis
	tmpFile, err := os.CreateTemp("", "augment-*.ts")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Run augment CLI to analyze the code
	cmd := exec.Command("augment", "analyze", tmpFile.Name(), "--json", "--project-root", projectRoot)
	output, err := cmd.Output()
	if err != nil {
		// If augment fails, fall back to basic analysis
		return analyzeBasic(filePath, code), nil
	}

	// Parse the JSON output
	var analysis AugmentCodeAnalysis
	if err := json.Unmarshal(output, &analysis); err != nil {
		// Fall back to basic analysis if JSON parsing fails
		return analyzeBasic(filePath, code), nil
	}

	analysis.FilePath = filePath
	return &analysis, nil
}

// analyzeBasic performs basic code analysis without Augment
func analyzeBasic(filePath string, code string) *AugmentCodeAnalysis {
	// Convert exportedSymbol to ExportedFunction
	symbols := extractExports(code)
	exports := make([]ExportedFunction, len(symbols))
	for i, sym := range symbols {
		exports[i] = ExportedFunction{
			Name:    sym.name,
			Type:    sym.kind,
			IsAsync: sym.isAsync,
		}
	}

	analysis := &AugmentCodeAnalysis{
		FilePath:     filePath,
		Exports:      exports,
		Dependencies: extractDependencies(code),
		Complexity:   "medium",
		Description:  "Auto-analyzed module",
	}

	// Generate basic test scenarios
	for _, exp := range analysis.Exports {
		scenarios := generateBasicScenarios(exp)
		analysis.TestScenarios = append(analysis.TestScenarios, scenarios...)
	}

	return analysis
}

// extractDependencies extracts import statements from code
func extractDependencies(code string) []string {
	var deps []string
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
			// Extract module name from import statement
			if idx := strings.Index(line, "from"); idx != -1 {
				rest := line[idx+4:]
				rest = strings.TrimSpace(rest)
				rest = strings.Trim(rest, "'\"")
				if rest != "" {
					deps = append(deps, rest)
				}
			}
		}
	}

	return deps
}

// generateBasicScenarios creates basic test scenarios for an exported function
func generateBasicScenarios(exp ExportedFunction) []TestScenario {
	var scenarios []TestScenario

	// Happy path scenario
	scenarios = append(scenarios, TestScenario{
		Name:        fmt.Sprintf("%s - happy path", exp.Name),
		Description: fmt.Sprintf("Test %s with valid inputs", exp.Name),
		Inputs:      generateSampleInputs(exp.Parameters),
		Expected:    "success",
		EdgeCase:    false,
	})

	// Edge case: null/undefined
	if len(exp.Parameters) > 0 {
		scenarios = append(scenarios, TestScenario{
			Name:        fmt.Sprintf("%s - null input", exp.Name),
			Description: fmt.Sprintf("Test %s with null input", exp.Name),
			Inputs:      map[string]interface{}{"input": nil},
			Expected:    "error or default",
			EdgeCase:    true,
		})
	}

	// Edge case: empty input
	scenarios = append(scenarios, TestScenario{
		Name:        fmt.Sprintf("%s - empty input", exp.Name),
		Description: fmt.Sprintf("Test %s with empty input", exp.Name),
		Inputs:      map[string]interface{}{},
		Expected:    "error or default",
		EdgeCase:    true,
	})

	// Async scenario
	if exp.IsAsync {
		scenarios = append(scenarios, TestScenario{
			Name:        fmt.Sprintf("%s - async resolution", exp.Name),
			Description: fmt.Sprintf("Test %s async behavior", exp.Name),
			Inputs:      generateSampleInputs(exp.Parameters),
			Expected:    "resolved promise",
			EdgeCase:    false,
		})
	}

	return scenarios
}

// generateSampleInputs creates sample input values based on parameter types
func generateSampleInputs(params []Parameter) map[string]interface{} {
	inputs := make(map[string]interface{})

	for _, param := range params {
		inputs[param.Name] = generateSampleValue(param.Type)
	}

	return inputs
}

// generateSampleValue generates a sample value for a given type
func generateSampleValue(typeStr string) interface{} {
	typeStr = strings.ToLower(typeStr)

	switch {
	case strings.Contains(typeStr, "string"):
		return "sample"
	case strings.Contains(typeStr, "number"):
		return 42
	case strings.Contains(typeStr, "boolean"):
		return true
	case strings.Contains(typeStr, "array"):
		return []interface{}{}
	case strings.Contains(typeStr, "object"):
		return map[string]interface{}{}
	default:
		return nil
	}
}

// LoginToAuggie handles user login to Augment Code
func LoginToAuggie() error {
	// Ensure Auggie is installed first
	if err := EnsureAuggieCLIInstalled(); err != nil {
		return err
	}

	fmt.Println("\n🔐 Logging in to Augment Code...")
	fmt.Println("Opening browser for authentication...")

	// Call auggie with a simple instruction to trigger authentication
	// Auggie will automatically prompt for login if not authenticated
	cmd := exec.Command("auggie", "Hello, I'm ready to generate tests")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	fmt.Println("\n✓ Successfully logged in to Augment Code!")
	fmt.Println("\nYou can now generate tests with:")
	fmt.Println("  ./autotest -root <project-path> -allow-dirty")
	fmt.Println("\nExample:")
	fmt.Println("  ./autotest -root ./my-project -allow-dirty")
	fmt.Println("  ./autotest -root ./my-project -allow-dirty -dry-run  # Preview first")
	return nil
}

// EnsureAuggieCLIInstalled checks if Auggie CLI is installed, and installs it if not
func EnsureAuggieCLIInstalled() error {
	// Check if auggie is already installed
	cmd := exec.Command("auggie", "--version")
	if err := cmd.Run(); err == nil {
		// Already installed
		return nil
	}

	fmt.Println("📦 Auggie CLI not found. Installing...")

	// Try to install with npm
	installCmd := exec.Command("npm", "install", "-g", "@augmentcode/auggie")
	var stderr bytes.Buffer
	installCmd.Stderr = &stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install Auggie CLI: %v\nstderr: %s", err, stderr.String())
	}

	fmt.Println("✓ Auggie CLI installed successfully!")
	return nil
}

// EnsureAuggieCLILoggedIn checks if user is logged in to Auggie, and prompts login if not
func EnsureAuggieCLILoggedIn() error {
	// Try a simple auggie command to check if logged in
	cmd := exec.Command("auggie", "--help")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If help fails, user might not be logged in
		fmt.Println("\n⚠️  Auggie CLI requires authentication")
		fmt.Println("Please login to Augment Code:")
		fmt.Println("\n  auggie --login")
		return fmt.Errorf("auggie CLI not authenticated. Please run 'auggie --login' first")
	}

	return nil
}

// GenerateTestWithAugmentCLI generates tests using Auggie CLI with project context
func GenerateTestWithAugmentCLI(filePath string, code string, framework string, projectContext string) (string, error) {
	// Ensure Auggie CLI is installed
	if err := EnsureAuggieCLIInstalled(); err != nil {
		return "", fmt.Errorf("auggie CLI setup failed: %w", err)
	}

	// Ensure user is logged in
	if err := EnsureAuggieCLILoggedIn(); err != nil {
		return "", err
	}

	// Build the prompt for Auggie
	prompt := buildAugmentPrompt(filePath, code, framework, projectContext)

	fmt.Printf("  ⏳ Generating tests for %s with Auggie...\n", filePath)

	// Call Auggie CLI with -p (print mode) flag
	// Show Auggie's reasoning in real-time
	cmd := exec.Command("auggie", "-p", prompt)
	cmd.Stderr = os.Stderr

	// Capture stdout to get the generated test code
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  ❌ Failed to generate tests for %s: %v\n", filePath, err)
		return "", fmt.Errorf("auggie CLI failed: %v", err)
	}

	testCode := string(output)
	if testCode == "" {
		return "", fmt.Errorf("auggie CLI returned empty output")
	}

	fmt.Printf("  ✓ Generated tests for %s\n", filePath)
	return testCode, nil
}

// buildAugmentPrompt creates a detailed prompt for Auggie CLI
func buildAugmentPrompt(filePath string, code string, framework string, projectContext string) string {
	var prompt strings.Builder

	prompt.WriteString("Generate comprehensive Jest/Vitest tests for the following TypeScript file:\n\n")
	prompt.WriteString("## File: " + filePath + "\n\n")

	prompt.WriteString("## Source Code:\n")
	prompt.WriteString("```typescript\n")
	prompt.WriteString(code)
	prompt.WriteString("\n```\n\n")

	prompt.WriteString("## Test Framework: " + framework + "\n\n")

	if projectContext != "" {
		prompt.WriteString("## Project Context:\n")
		prompt.WriteString(projectContext)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("## Requirements:\n")
	prompt.WriteString("1. Generate comprehensive test cases covering:\n")
	prompt.WriteString("   - Happy path scenarios\n")
	prompt.WriteString("   - Edge cases (null, undefined, empty inputs)\n")
	prompt.WriteString("   - Error handling\n")
	prompt.WriteString("   - Async operations (if applicable)\n")
	prompt.WriteString("2. Use proper mocking for dependencies\n")
	prompt.WriteString("3. Include descriptive test names\n")
	prompt.WriteString("4. Add comments explaining complex test logic\n")
	prompt.WriteString("5. Return ONLY the test code, no explanations\n")

	return prompt.String()
}

// GenerateTestWithAugment generates a test file using Augment analysis
func GenerateTestWithAugment(tsPath string, code string, framework string, projectRoot string) (string, error) {
	// Analyze the code with Augment
	analysis, err := AnalyzeWithAugment(tsPath, code, projectRoot)
	if err != nil {
		return "", fmt.Errorf("augment analysis failed: %w", err)
	}

	if len(analysis.Exports) == 0 {
		return "", fmt.Errorf("no exported symbols found in %s", tsPath)
	}

	// Generate test code based on analysis
	testCode := generateTestCodeFromAnalysis(tsPath, analysis, framework)
	return testCode, nil
}

// generateTestCodeFromAnalysis creates test code from Augment analysis
func generateTestCodeFromAnalysis(tsPath string, analysis *AugmentCodeAnalysis, framework string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("/**\n")
	sb.WriteString(" * Auto-generated test file (powered by Augment)\n")
	sb.WriteString(" * Source: " + tsPath + "\n")
	sb.WriteString(" * Description: " + analysis.Description + "\n")
	sb.WriteString(" */\n\n")

	// Import statement
	importPath := strings.TrimSuffix(tsPath, filepath.Ext(tsPath))
	importPath = strings.TrimPrefix(importPath, "./")
	sb.WriteString("import { ")

	for i, exp := range analysis.Exports {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(exp.Name)
	}
	sb.WriteString(" } from '../" + importPath + "';\n\n")

	// Test framework setup
	if framework == "vitest" {
		sb.WriteString("import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';\n\n")
	} else {
		sb.WriteString("import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';\n\n")
	}

	// Generate tests for each export
	for _, exp := range analysis.Exports {
		sb.WriteString(generateTestsForExport(exp, analysis.TestScenarios, framework))
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateTestsForExport generates test cases for a single export
func generateTestsForExport(exp ExportedFunction, scenarios []TestScenario, framework string) string {
	var sb strings.Builder

	sb.WriteString("describe('" + exp.Name + "', () => {\n")

	// Add setup/teardown if needed
	sb.WriteString("  let result: any;\n\n")

	// Generate tests for each scenario
	for _, scenario := range scenarios {
		if strings.Contains(scenario.Name, exp.Name) {
			sb.WriteString(generateTestCase(exp, scenario, framework))
			sb.WriteString("\n")
		}
	}

	// Add a basic existence test
	sb.WriteString("  it('should be defined', () => {\n")
	sb.WriteString("    expect(" + exp.Name + ").toBeDefined();\n")
	sb.WriteString("  });\n\n")

	// Add type check
	sb.WriteString("  it('should be a " + exp.Type + "', () => {\n")
	sb.WriteString("    expect(typeof " + exp.Name + ").toBe('" + getTypeofValue(exp.Type) + "');\n")
	sb.WriteString("  });\n")

	sb.WriteString("});\n")
	return sb.String()
}

// generateTestCase generates a single test case
func generateTestCase(exp ExportedFunction, scenario TestScenario, framework string) string {
	var sb strings.Builder

	testName := scenario.Name
	if scenario.EdgeCase {
		testName = "✓ [EDGE CASE] " + testName
	}

	sb.WriteString("  it('" + testName + "', ")

	if exp.IsAsync {
		sb.WriteString("async ")
	}

	sb.WriteString("() => {\n")

	// Setup inputs
	if len(scenario.Inputs) > 0 {
		sb.WriteString("    // Arrange\n")
		for key, val := range scenario.Inputs {
			sb.WriteString("    const " + key + " = " + formatValue(val) + ";\n")
		}
		sb.WriteString("\n")
	}

	// Act
	sb.WriteString("    // Act\n")
	if exp.IsAsync {
		sb.WriteString("    const result = await " + exp.Name + "(")
	} else {
		sb.WriteString("    const result = " + exp.Name + "(")
	}

	// Add parameters
	for i, param := range exp.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.Name)
	}
	sb.WriteString(");\n\n")

	// Assert
	sb.WriteString("    // Assert\n")
	sb.WriteString("    expect(result).toBeDefined();\n")
	sb.WriteString("    // TODO: Add specific assertions based on expected behavior\n")

	sb.WriteString("  });\n")

	return sb.String()
}

// getTypeofValue returns the typeof value for a type string
func getTypeofValue(typeStr string) string {
	typeStr = strings.ToLower(typeStr)
	switch {
	case typeStr == "function":
		return "function"
	case typeStr == "class":
		return "function" // classes are functions in JS
	default:
		return "object"
	}
}

// formatValue formats a value for code generation
func formatValue(val interface{}) string {
	if val == nil {
		return "null"
	}

	switch v := val.(type) {
	case string:
		return "'" + v + "'"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		return "[]"
	case map[string]interface{}:
		return "{}"
	default:
		return "null"
	}
}
