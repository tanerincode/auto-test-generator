package gen

import (
	"fmt"
)

// EnsureCursorCLIInstalled checks if Cursor CLI is available and offers to install if not
func EnsureCursorCLIInstalled() error {
	// Note: We don't check cursor --version because it opens the IDE
	// Cursor doesn't have a headless CLI API like Auggie

	fmt.Println("\n⚠️  Cursor Provider Limitation:")
	fmt.Println("Cursor is an IDE and doesn't support headless CLI operations.")
	fmt.Println("The tool will use basic test generation instead.")
	fmt.Println("\n💡 For AI-powered test generation, please use:")
	fmt.Println("   ./autotest -root ./your-project -provider auggie -allow-dirty")
	fmt.Println()

	return nil
}

// GenerateTestWithCursorCLI generates tests using Cursor CLI
func GenerateTestWithCursorCLI(filePath string, code string, framework string, projectContext string) (string, error) {
	// Ensure Cursor CLI is available
	if err := EnsureCursorCLIInstalled(); err != nil {
		return "", err
	}

	fmt.Printf("  ⏳ Generating tests for %s with Cursor AI...\n", filePath)

	// NOTE: Cursor CLI doesn't have a direct API for AI operations like Auggie does.
	// Cursor is meant to be an IDE, not a headless AI service.
	// The 'cursor' command is for opening files in the editor, not AI processing.

	fmt.Printf("  ℹ️  Cursor requires IDE interaction - falling back to basic generation\n")
	fmt.Printf("  💡 For AI-powered tests, use: -provider auggie\n")

	// Fallback to basic generation
	return GenerateTest(filePath, code, framework)
}
