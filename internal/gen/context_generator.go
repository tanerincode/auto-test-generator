package gen

import (
	"fmt"
	"strings"
)

// ContextAwareTestGenerator generates tests using project context
type ContextAwareTestGenerator struct {
	ContextEngine *AugmentContextEngine
	Framework     string
}

// NewContextAwareTestGenerator creates a new context-aware generator
func NewContextAwareTestGenerator(contextEngine *AugmentContextEngine, framework string) *ContextAwareTestGenerator {
	return &ContextAwareTestGenerator{
		ContextEngine: contextEngine,
		Framework:     framework,
	}
}

// GenerateTestWithProjectContext generates a test file using full project context
func (ctg *ContextAwareTestGenerator) GenerateTestWithProjectContext(filePath string, code string) (string, error) {
	// Get comprehensive context for the file
	fileContext := ctg.ContextEngine.GetFileContext(filePath)

	// Extract exports
	exportsInterface, ok := fileContext["exports"]
	if !ok || exportsInterface == nil {
		// Fallback to basic generation if context not found
		return GenerateTest("", code, ctg.Framework)
	}

	exports, ok := exportsInterface.([]ExportedFunction)
	if !ok || len(exports) == 0 {
		return "", fmt.Errorf("no exported symbols found in %s", filePath)
	}

	// Get related code for better understanding
	relatedFiles := fileContext["related_files"].(map[string]string)

	// Generate test code
	testCode := ctg.generateTestCodeWithContext(filePath, code, exports, relatedFiles)
	return testCode, nil
}

// generateTestCodeWithContext creates test code with full project context
func (ctg *ContextAwareTestGenerator) generateTestCodeWithContext(
	filePath string,
	code string,
	exports []ExportedFunction,
	relatedFiles map[string]string,
) string {
	var sb strings.Builder

	// Header with context info
	sb.WriteString("/**\n")
	sb.WriteString(" * Auto-generated test file (Augment Context Engine)\n")
	sb.WriteString(" * Source: " + filePath + "\n")
	sb.WriteString(" * Generated with project-wide context analysis\n")
	sb.WriteString(" */\n\n")

	// Imports from source file
	sb.WriteString(ctg.generateImports(filePath, exports))
	sb.WriteString("\n")

	// Test framework imports
	sb.WriteString(ctg.generateFrameworkImports())
	sb.WriteString("\n")

	// Mock setup if there are related files
	if len(relatedFiles) > 0 {
		sb.WriteString(ctg.generateMockSetup(relatedFiles))
		sb.WriteString("\n")
	}

	// Generate describe blocks for each export
	for _, exp := range exports {
		sb.WriteString(ctg.generateDescribeBlock(exp, code))
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateImports creates import statements
func (ctg *ContextAwareTestGenerator) generateImports(filePath string, exports []ExportedFunction) string {
	var sb strings.Builder

	sb.WriteString("import { ")
	for i, exp := range exports {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(exp.Name)
	}
	sb.WriteString(" } from '../" + strings.TrimSuffix(filePath, ".ts") + "';\n")

	return sb.String()
}

// generateFrameworkImports creates test framework imports
func (ctg *ContextAwareTestGenerator) generateFrameworkImports() string {
	if ctg.Framework == "vitest" {
		return "import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';"
	}
	return "import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';"
}

// generateMockSetup creates mock setup code
func (ctg *ContextAwareTestGenerator) generateMockSetup(relatedFiles map[string]string) string {
	var sb strings.Builder

	sb.WriteString("// Mock setup for related dependencies\n")
	for filePath := range relatedFiles {
		mockName := strings.ReplaceAll(filePath, "/", "_")
		mockName = strings.ReplaceAll(mockName, ".", "_")
		sb.WriteString("// jest.mock('../" + strings.TrimSuffix(filePath, ".ts") + "');\n")
	}

	return sb.String()
}

// generateDescribeBlock creates a describe block for an export
func (ctg *ContextAwareTestGenerator) generateDescribeBlock(exp ExportedFunction, sourceCode string) string {
	var sb strings.Builder

	sb.WriteString("describe('" + exp.Name + "', () => {\n")

	// Setup
	sb.WriteString("  let result: any;\n\n")

	// Test: existence
	sb.WriteString("  it('should be defined', () => {\n")
	sb.WriteString("    expect(" + exp.Name + ").toBeDefined();\n")
	sb.WriteString("  });\n\n")

	// Test: type
	sb.WriteString("  it('should be a " + exp.Type + "', () => {\n")
	sb.WriteString("    expect(typeof " + exp.Name + ").toBe('" + getTypeofValue(exp.Type) + "');\n")
	sb.WriteString("  });\n\n")

	// Test: happy path
	sb.WriteString(ctg.generateHappyPathTest(exp))
	sb.WriteString("\n")

	// Test: edge cases
	sb.WriteString(ctg.generateEdgeCaseTests(exp))
	sb.WriteString("\n")

	// Test: async if applicable
	if exp.IsAsync {
		sb.WriteString(ctg.generateAsyncTest(exp))
		sb.WriteString("\n")
	}

	// Test: error handling
	sb.WriteString(ctg.generateErrorHandlingTest(exp))

	sb.WriteString("});\n")

	return sb.String()
}

// generateHappyPathTest creates a happy path test
func (ctg *ContextAwareTestGenerator) generateHappyPathTest(exp ExportedFunction) string {
	var sb strings.Builder

	sb.WriteString("  it('should work with valid inputs', () => {\n")
	sb.WriteString("    // Arrange\n")

	// Generate sample inputs
	for _, param := range exp.Parameters {
		sb.WriteString("    const " + param.Name + " = " + ctg.generateSampleInput(param.Type) + ";\n")
	}

	sb.WriteString("\n    // Act\n")
	sb.WriteString("    const result = " + exp.Name + "(")
	for i, param := range exp.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.Name)
	}
	sb.WriteString(");\n\n")

	sb.WriteString("    // Assert\n")
	sb.WriteString("    expect(result).toBeDefined();\n")
	sb.WriteString("    // TODO: Add specific assertions\n")
	sb.WriteString("  });\n")

	return sb.String()
}

// generateEdgeCaseTests creates edge case tests
func (ctg *ContextAwareTestGenerator) generateEdgeCaseTests(exp ExportedFunction) string {
	var sb strings.Builder

	if len(exp.Parameters) > 0 {
		sb.WriteString("  it('[EDGE CASE] should handle null/undefined inputs', () => {\n")
		sb.WriteString("    // TODO: Test with null/undefined values\n")
		sb.WriteString("    expect(true).toBe(true);\n")
		sb.WriteString("  });\n\n")
	}

	sb.WriteString("  it('[EDGE CASE] should handle empty inputs', () => {\n")
	sb.WriteString("    // TODO: Test with empty values\n")
	sb.WriteString("    expect(true).toBe(true);\n")
	sb.WriteString("  });\n")

	return sb.String()
}

// generateAsyncTest creates an async test
func (ctg *ContextAwareTestGenerator) generateAsyncTest(exp ExportedFunction) string {
	var sb strings.Builder

	sb.WriteString("  it('should handle async operations', async () => {\n")
	sb.WriteString("    // Arrange\n")

	for _, param := range exp.Parameters {
		sb.WriteString("    const " + param.Name + " = " + ctg.generateSampleInput(param.Type) + ";\n")
	}

	sb.WriteString("\n    // Act\n")
	sb.WriteString("    const result = await " + exp.Name + "(")
	for i, param := range exp.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.Name)
	}
	sb.WriteString(");\n\n")

	sb.WriteString("    // Assert\n")
	sb.WriteString("    expect(result).toBeDefined();\n")
	sb.WriteString("  });\n")

	return sb.String()
}

// generateErrorHandlingTest creates an error handling test
func (ctg *ContextAwareTestGenerator) generateErrorHandlingTest(exp ExportedFunction) string {
	var sb strings.Builder

	sb.WriteString("  it('should handle errors gracefully', () => {\n")
	sb.WriteString("    // TODO: Test error scenarios\n")
	sb.WriteString("    expect(true).toBe(true);\n")
	sb.WriteString("  });\n")

	return sb.String()
}

// generateSampleInput generates a sample input for a type
func (ctg *ContextAwareTestGenerator) generateSampleInput(typeStr string) string {
	typeStr = strings.ToLower(typeStr)

	switch {
	case strings.Contains(typeStr, "string"):
		return "'sample'"
	case strings.Contains(typeStr, "number"):
		return "42"
	case strings.Contains(typeStr, "boolean"):
		return "true"
	case strings.Contains(typeStr, "array"):
		return "[]"
	case strings.Contains(typeStr, "object"):
		return "{}"
	default:
		return "null"
	}
}
