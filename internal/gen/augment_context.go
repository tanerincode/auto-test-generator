package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AugmentContextEngine manages project indexing and context retrieval
type AugmentContextEngine struct {
	ProjectRoot  string
	IndexedCode  map[string]string // filepath -> code content
	Exports      map[string][]ExportedFunction
	Dependencies map[string][]string
	Initialized  bool
}

// NewAugmentContextEngine creates a new context engine for the project
func NewAugmentContextEngine(projectRoot string) *AugmentContextEngine {
	return &AugmentContextEngine{
		ProjectRoot:  projectRoot,
		IndexedCode:  make(map[string]string),
		Exports:      make(map[string][]ExportedFunction),
		Dependencies: make(map[string][]string),
		Initialized:  false,
	}
}

// IndexProject scans and indexes all TypeScript files in the project
func (ace *AugmentContextEngine) IndexProject() error {
	fmt.Println("🔍 Indexing project with Augment context engine...")

	// Find all TypeScript files
	err := filepath.Walk(ace.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-TS files
		if info.IsDir() {
			// Skip node_modules, .git, build, dist
			if strings.Contains(path, "node_modules") ||
				strings.Contains(path, ".git") ||
				strings.Contains(path, "/build/") ||
				strings.Contains(path, "/dist/") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .ts and .tsx files
		if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
			return nil
		}

		// Skip .d.ts files
		if strings.HasSuffix(path, ".d.ts") {
			return nil
		}

		// Skip test files
		if strings.Contains(path, ".test.") || strings.Contains(path, ".spec.") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("⚠️  Failed to read %s: %v\n", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(ace.ProjectRoot, path)
		ace.IndexedCode[relPath] = string(content)

		// Extract exports and dependencies
		symbols := extractExports(string(content))
		deps := extractDependencies(string(content))

		// Convert exportedSymbol to ExportedFunction
		exports := make([]ExportedFunction, len(symbols))
		for i, sym := range symbols {
			exports[i] = ExportedFunction{
				Name:    sym.name,
				Type:    sym.kind,
				IsAsync: sym.isAsync,
			}
		}

		ace.Exports[relPath] = exports
		ace.Dependencies[relPath] = deps

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to index project: %w", err)
	}

	ace.Initialized = true
	fmt.Printf("✓ Indexed %d TypeScript files\n", len(ace.IndexedCode))
	return nil
}

// GetRelatedCode finds related code files that might help understand the target file
func (ace *AugmentContextEngine) GetRelatedCode(targetFile string) map[string]string {
	related := make(map[string]string)

	if !ace.Initialized {
		return related
	}

	// Get dependencies of target file
	targetDeps := ace.Dependencies[targetFile]
	if len(targetDeps) == 0 {
		return related
	}

	// Find files that match the dependencies
	for _, dep := range targetDeps {
		// Skip external packages
		if strings.HasPrefix(dep, ".") {
			// Resolve relative import
			dir := filepath.Dir(targetFile)
			resolvedPath := filepath.Join(dir, dep)
			resolvedPath = filepath.Clean(resolvedPath)

			// Try with .ts and .tsx extensions
			for _, ext := range []string{".ts", ".tsx", "/index.ts", "/index.tsx"} {
				searchPath := resolvedPath + ext
				if code, exists := ace.IndexedCode[searchPath]; exists {
					related[searchPath] = code
					break
				}
			}
		}
	}

	return related
}

// GetExportContext returns information about exported symbols in related files
func (ace *AugmentContextEngine) GetExportContext(targetFile string) map[string][]ExportedFunction {
	context := make(map[string][]ExportedFunction)

	if !ace.Initialized {
		return context
	}

	// Get related files
	related := ace.GetRelatedCode(targetFile)

	// Add exports from related files
	for relPath := range related {
		if exports, exists := ace.Exports[relPath]; exists {
			context[relPath] = exports
		}
	}

	return context
}

// GenerateContextualPrompt creates a detailed prompt for test generation with project context
func (ace *AugmentContextEngine) GenerateContextualPrompt(targetFile string, code string) string {
	var sb strings.Builder

	sb.WriteString("# Test Generation Context\n\n")
	sb.WriteString("## Target File\n")
	sb.WriteString("File: " + targetFile + "\n\n")

	// Add target file exports
	if exports, exists := ace.Exports[targetFile]; exists && len(exports) > 0 {
		sb.WriteString("### Exported Symbols\n")
		for _, exp := range exports {
			sb.WriteString(fmt.Sprintf("- **%s** (%s)", exp.Name, exp.Type))
			if exp.IsAsync {
				sb.WriteString(" [async]")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Add related files context
	relatedContext := ace.GetExportContext(targetFile)
	if len(relatedContext) > 0 {
		sb.WriteString("## Related Dependencies\n")
		for relPath, exports := range relatedContext {
			sb.WriteString(fmt.Sprintf("### %s\n", relPath))
			for _, exp := range exports {
				sb.WriteString(fmt.Sprintf("- %s (%s)\n", exp.Name, exp.Type))
			}
			sb.WriteString("\n")
		}
	}

	// Add code snippet
	sb.WriteString("## Source Code\n")
	sb.WriteString("```typescript\n")
	sb.WriteString(code)
	sb.WriteString("\n```\n\n")

	sb.WriteString("## Test Generation Requirements\n")
	sb.WriteString("1. Generate comprehensive test cases covering happy paths and edge cases\n")
	sb.WriteString("2. Use the project's testing framework (Jest/Vitest)\n")
	sb.WriteString("3. Include tests for:\n")
	sb.WriteString("   - Basic functionality\n")
	sb.WriteString("   - Async operations (if applicable)\n")
	sb.WriteString("   - Error handling\n")
	sb.WriteString("   - Edge cases (null, undefined, empty inputs)\n")
	sb.WriteString("4. Mock external dependencies appropriately\n")
	sb.WriteString("5. Use descriptive test names\n")

	return sb.String()
}

// GetProjectStats returns statistics about the indexed project
func (ace *AugmentContextEngine) GetProjectStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if !ace.Initialized {
		return stats
	}

	totalExports := 0
	totalDependencies := 0
	fileCount := len(ace.IndexedCode)

	for _, exports := range ace.Exports {
		totalExports += len(exports)
	}

	for _, deps := range ace.Dependencies {
		totalDependencies += len(deps)
	}

	stats["files_indexed"] = fileCount
	stats["total_exports"] = totalExports
	stats["total_dependencies"] = totalDependencies
	stats["avg_exports_per_file"] = float64(totalExports) / float64(fileCount)

	return stats
}

// FindSimilarFiles finds files with similar exports or patterns
func (ace *AugmentContextEngine) FindSimilarFiles(targetFile string, limit int) []string {
	if !ace.Initialized || limit <= 0 {
		return nil
	}

	targetExports := ace.Exports[targetFile]
	if len(targetExports) == 0 {
		return nil
	}

	// Build a map of export names to files
	exportToFiles := make(map[string][]string)
	for filePath, exports := range ace.Exports {
		if filePath == targetFile {
			continue
		}
		for _, exp := range exports {
			exportToFiles[exp.Name] = append(exportToFiles[exp.Name], filePath)
		}
	}

	// Find files with similar exports
	similarFiles := make(map[string]int)
	for _, targetExp := range targetExports {
		if files, exists := exportToFiles[targetExp.Name]; exists {
			for _, file := range files {
				similarFiles[file]++
			}
		}
	}

	// Sort and return top matches
	var result []string
	for file := range similarFiles {
		result = append(result, file)
		if len(result) >= limit {
			break
		}
	}

	return result
}

// GetFileContext returns comprehensive context for a file
func (ace *AugmentContextEngine) GetFileContext(targetFile string) map[string]interface{} {
	context := make(map[string]interface{})

	if !ace.Initialized {
		return context
	}

	context["file"] = targetFile
	context["exports"] = ace.Exports[targetFile]
	context["dependencies"] = ace.Dependencies[targetFile]
	context["related_files"] = ace.GetRelatedCode(targetFile)
	context["similar_files"] = ace.FindSimilarFiles(targetFile, 3)

	return context
}

// PrintIndexSummary prints a summary of the indexed project
func (ace *AugmentContextEngine) PrintIndexSummary() {
	if !ace.Initialized {
		fmt.Println("Project not indexed yet")
		return
	}

	stats := ace.GetProjectStats()
	fmt.Println("\n📊 Project Index Summary")
	fmt.Println("========================")
	fmt.Printf("Files indexed: %d\n", stats["files_indexed"])
	fmt.Printf("Total exports: %d\n", stats["total_exports"])
	fmt.Printf("Total dependencies: %d\n", stats["total_dependencies"])
	fmt.Printf("Avg exports per file: %.2f\n", stats["avg_exports_per_file"])
	fmt.Println()
}
