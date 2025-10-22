package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/tanerincode/auto-test-generator/internal/exec"
	"github.com/tanerincode/auto-test-generator/internal/gen"
	"github.com/tanerincode/auto-test-generator/internal/scan"
)

func main() {
	root := flag.String("root", ".", "Root directory of the project")
	fw := flag.String("fw", "auto", "Framework: auto, jest, or vitest")
	out := flag.String("out", "", "Optional output directory for tests (mirrors structure)")
	dryRun := flag.Bool("dry-run", false, "Print plan and diffs without writing")
	changedOnly := flag.Bool("changed-only", false, "Limit to git diff against origin/main")
	maxWorkers := flag.Int("max-workers", runtime.NumCPU(), "Maximum concurrent workers")
	minCoverage := flag.Float64("min-coverage", 0, "Minimum coverage threshold (0-100); fail if below")
	allowDirty := flag.Bool("allow-dirty", false, "Allow running with dirty working tree")

	flag.Parse()

	// Handle login command
	if len(flag.Args()) > 0 {
		cmd := flag.Args()[0]
		if cmd == "login" {
			if err := gen.LoginToAuggie(); err != nil {
				log.Fatalf("Login failed: %v", err)
			}
			return
		}
		if cmd == "help" || cmd == "-h" || cmd == "--help" {
			fmt.Println("autotest - Auto-generate Jest/Vitest tests for TypeScript files")
			fmt.Println("\nUsage:")
			fmt.Println("  autotest login                Login to Augment Code (one time setup)")
			fmt.Println("  autotest -root <path> [flags] Generate tests for project")
			fmt.Println("\nExamples:")
			fmt.Println("  ./autotest login")
			fmt.Println("  ./autotest -root ./my-project -allow-dirty")
			fmt.Println("  ./autotest -root ./my-project -allow-dirty -dry-run")
			fmt.Println("\nFlags:")
			flag.PrintDefaults()
			return
		}
	}

	// Check if -root was explicitly provided
	rootProvided := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "root" {
			rootProvided = true
		}
	})

	if !rootProvided {
		fmt.Println("❌ Error: -root flag is required")
		fmt.Println("\nUsage:")
		fmt.Println("  ./autotest -root <project-path> -allow-dirty")
		fmt.Println("\nExample:")
		fmt.Println("  ./autotest -root ./my-project -allow-dirty")
		fmt.Println("\nFor more help:")
		fmt.Println("  ./autotest help")
		os.Exit(1)
	}

	// Validate flags
	if *fw != "auto" && *fw != "jest" && *fw != "vitest" {
		log.Fatalf("invalid framework: %s (must be auto, jest, or vitest)", *fw)
	}
	if *minCoverage < 0 || *minCoverage > 100 {
		log.Fatalf("min-coverage must be between 0 and 100")
	}
	if *maxWorkers < 1 {
		log.Fatalf("max-workers must be at least 1")
	}

	// Check git status unless --allow-dirty
	if !*allowDirty {
		dirty, err := scan.IsWorkingTreeDirty(*root)
		if err != nil {
			log.Fatalf("failed to check git status: %v", err)
		}
		if dirty {
			log.Fatalf("working tree is dirty; commit changes or use --allow-dirty")
		}
	}

	// Detect framework
	framework := *fw
	if framework == "auto" {
		detected, err := exec.DetectFramework(*root)
		if err != nil {
			log.Fatalf("failed to detect framework: %v", err)
		}
		framework = detected
		fmt.Printf("Detected framework: %s\n", framework)
	}

	// Setup Auggie CLI
	fmt.Println("🤖 Using Auggie CLI for AI-powered test generation...")
	if err := gen.EnsureAuggieCLIInstalled(); err != nil {
		log.Fatalf("failed to setup Auggie CLI: %v", err)
	}

	// Scan for files needing tests
	candidates, err := scan.FindCandidates(*root, *changedOnly)
	if err != nil {
		log.Fatalf("failed to scan files: %v", err)
	}

	if len(candidates) == 0 {
		fmt.Println("No files need tests.")
		return
	}

	fmt.Printf("Found %d file(s) needing tests\n", len(candidates))

	// Build work queue
	type workItem struct {
		path string
		code string
	}
	workQueue := make([]workItem, 0, len(candidates))

	for _, candidate := range candidates {
		code, err := os.ReadFile(candidate)
		if err != nil {
			log.Printf("warning: failed to read %s: %v", candidate, err)
			continue
		}
		workQueue = append(workQueue, workItem{path: candidate, code: string(code)})
	}

	// Process with worker pool
	results := make(chan gen.TestResult, len(workQueue))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *maxWorkers)

	for _, item := range workQueue {
		wg.Add(1)
		go func(wi workItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			testPath := scan.DefaultTestPath(wi.path, framework, *out)
			relPath, _ := filepath.Rel(*root, wi.path)

			// Generate test with Auggie CLI
			testCode, err := gen.GenerateTestWithAugmentCLI(relPath, wi.code, framework, "")

			if err != nil {
				results <- gen.TestResult{
					SourcePath: wi.path,
					TestPath:   testPath,
					Error:      fmt.Errorf("generation failed: %w", err),
				}
				return
			}

			results <- gen.TestResult{
				SourcePath: wi.path,
				TestPath:   testPath,
				TestCode:   testCode,
			}
		}(item)
	}

	wg.Wait()
	close(results)

	// Collect results
	var testResults []gen.TestResult
	var failedCount int
	for result := range results {
		if result.Error != nil {
			log.Printf("error: %s: %v", result.SourcePath, result.Error)
			failedCount++
		} else {
			testResults = append(testResults, result)
		}
	}

	if len(testResults) == 0 {
		if failedCount > 0 {
			log.Fatalf("all generations failed")
		}
		fmt.Println("No tests generated.")
		return
	}

	fmt.Printf("Generated %d test(s)\n", len(testResults))

	// Dry-run: print plan
	if *dryRun {
		fmt.Println("\n=== DRY RUN: Test Generation Plan ===")
		for _, result := range testResults {
			fmt.Printf("\nSource: %s\n", result.SourcePath)
			fmt.Printf("Test:   %s\n", result.TestPath)
			fmt.Printf("Lines:  %d\n", len(result.TestCode))
		}
		fmt.Println("\n(No files written in dry-run mode)")
		return
	}

	// Write tests
	var written int
	for _, result := range testResults {
		dir := filepath.Dir(result.TestPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("error: failed to create directory %s: %v", dir, err)
			continue
		}

		if err := os.WriteFile(result.TestPath, []byte(result.TestCode), 0644); err != nil {
			log.Printf("error: failed to write %s: %v", result.TestPath, err)
			continue
		}

		fmt.Printf("✓ %s\n", result.TestPath)
		written++
	}

	fmt.Printf("\nWrote %d test file(s)\n", written)

	// Run tests on affected scope
	if written > 0 {
		testPaths := make([]string, len(testResults))
		for i, result := range testResults {
			testPaths[i] = result.TestPath
		}

		fmt.Println("\nRunning tests...")
		if err := exec.RunTests(testPaths, framework, *root); err != nil {
			log.Printf("warning: test run failed: %v", err)
		}
	}

	// Check coverage if requested
	if *minCoverage > 0 {
		fmt.Printf("\nChecking coverage (minimum: %.1f%%)\n", *minCoverage)
		coverage, err := exec.GetCoverage(*root, framework)
		if err != nil {
			log.Printf("warning: failed to get coverage: %v", err)
		} else if coverage < *minCoverage {
			log.Fatalf("coverage %.1f%% is below minimum %.1f%%", coverage, *minCoverage)
		}
		fmt.Printf("Coverage: %.1f%% ✓\n", coverage)
	}

	fmt.Println("\nDone!")
}
