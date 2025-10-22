# autotest

A production-grade Go CLI tool that auto-generates Jest/Vitest tests for TypeScript/TSX files in Node.js projects. Powered by [Auggie CLI](https://augmentcode.com/) for intelligent, AI-driven test generation.

## Overview

`autotest` automatically generates comprehensive test files for your TypeScript codebase by analyzing your source code, understanding dependencies, and creating realistic test scenarios. It uses Auggie CLI, an AI-powered code assistant, to generate high-quality tests that cover happy paths, edge cases, and error handling.

## Key Features

- **🤖 AI-Powered Generation**: Uses Auggie CLI for intelligent, context-aware test generation
- **⚡ Fast & Concurrent**: Multi-threaded processing with configurable worker pool
- **🎯 Framework Detection**: Automatically detects Jest or Vitest from `package.json`
- **🔍 Smart Scanning**: Finds TypeScript files without tests, respects exclusion patterns
- **📁 Flexible Output**: Place tests next to source or mirror structure under custom directory
- **👀 Dry-Run Mode**: Preview changes before writing files
- **🔀 Git Integration**: Limit to changed files with `--changed-only`
- **📊 Coverage Checks**: Enforce minimum coverage thresholds
- **🛡️ Safe by Design**: Only generates test files; never touches production code
- **🧠 Project Context**: Optionally indexes entire codebase for better test understanding

## Installation

```bash
# Clone or download the repository
cd auto-test-generator

# Build the binary
make build
```

The binary will be available at `bin/autotest`.

## Prerequisites

Before using `autotest`, you need:

1. **Node.js project** with TypeScript/TSX files
2. **Jest or Vitest** installed as a dev dependency
3. **Auggie CLI** (automatically installed on first run)
   - Requires authentication: run `./autotest login` first

## Quick Start

```bash
# 1. Login to Augment Code (one-time setup)
./autotest login

# 2. Generate tests for your project
./autotest -root ./your-project -allow-dirty

# 3. Preview what will be generated (dry-run)
./autotest -root ./your-project -allow-dirty -dry-run
```

## Usage

### Basic Usage

**Note:** The `-root` flag is **required** to specify the project directory.

```bash
# Generate tests for all TypeScript files in a project
./autotest -root ./my-project -allow-dirty

# Preview changes without writing files
./autotest -root ./my-project -allow-dirty -dry-run
```

### Flags

- **`-root string`** (required)
  - Root directory of the Node.js project to scan
  - Must contain `package.json` with Jest or Vitest

- **`-allow-dirty`** (default: `false`)
  - Allow running with uncommitted changes in the working tree
  - By default, refuses to run if working tree has uncommitted changes
  - Recommended for initial testing

- **`-fw string`** (default: `auto`)
  - Test framework: `auto`, `jest`, or `vitest`
  - `auto` detects from `package.json` (prefers Vitest if both present)

- **`-out string`** (default: empty)
  - Optional output directory for tests
  - If set, mirrors the source directory structure under this path
  - If empty, places tests next to source files

- **`-dry-run`** (default: `false`)
  - Print the generation plan without writing files
  - Useful for previewing what will be generated

- **`-changed-only`** (default: `false`)
  - Limit scanning to files changed against `origin/main` (or `origin/master`)
  - Requires a git repository with remote tracking

- **`-max-workers int`** (default: number of CPUs)
  - Maximum concurrent workers for test generation
  - Increase for faster processing on multi-core systems

- **`-min-coverage float`** (default: `0`)
  - Minimum coverage threshold (0-100)
  - If set, fails if coverage is below this percentage after generation
  - Runs test suite with coverage after generation

### Examples

#### Login (first time setup)

```bash
./autotest login
```

#### Generate tests for a specific project

```bash
./autotest -root ./my-project -allow-dirty
```

#### Preview changes without writing

```bash
./autotest -root ./example -dry-run -allow-dirty
```

#### Generate tests only for changed files

```bash
./autotest -root ./my-project -changed-only
```

#### Use Vitest and mirror tests under `tests/` directory

```bash
./autotest -root ./my-project -fw vitest -out tests -allow-dirty
```

#### Enforce minimum coverage threshold

```bash
./autotest -root ./my-project -min-coverage 80 -allow-dirty
```

#### Use all flags together

```bash
./autotest -root ./src -fw vitest -out tests -changed-only -dry-run -max-workers 8 -allow-dirty
```

## AI-Powered Test Generation

The tool uses **Auggie CLI** by default for intelligent, AI-powered test generation.

### How It Works

1. **Authentication**: One-time login to Augment Code
2. **Code Analysis**: Auggie analyzes your TypeScript source code
3. **Context Understanding**: Understands code semantics, dependencies, and patterns
4. **Test Generation**: Creates comprehensive tests with:
   - Happy path scenarios
   - Edge cases (null, undefined, empty inputs)
   - Error handling
   - Async operation handling
   - Proper mocking of dependencies
5. **Quality Assurance**: Generated tests are realistic and meaningful

### Benefits

- **High-Quality Tests**: AI understands code intent and generates appropriate tests
- **Comprehensive Coverage**: Includes edge cases and error scenarios
- **Time Savings**: Generates tests in seconds that would take minutes to write manually
- **Learning Tool**: See how to properly structure tests for your code
- **Consistency**: All tests follow best practices and patterns

### Authentication

```bash
# First time setup
./autotest login

# This will:
# 1. Install Auggie CLI (if not already installed)
# 2. Open browser for authentication
# 3. Save credentials for future use
```

After login, the tool will use Auggie CLI automatically for all test generation.

## Architecture & How It Works

### Workflow

1. **Scanning**: Discovers TypeScript/TSX files, excluding:
   - `node_modules/`
   - `.d.ts` declaration files
   - Existing test files (`.test.ts`, `.test.tsx`, `.spec.ts`, `.spec.tsx`)
   - `build/` and `dist/` directories
   - Files already covered by tests

2. **Framework Detection**: Reads `package.json` to detect Jest or Vitest
   - Prefers Vitest if both are present
   - Falls back to checking lockfiles (`pnpm-lock.yaml`, `yarn.lock`, `package-lock.json`)

3. **AI-Powered Generation**: For each file without tests:
   - Sends source code to Auggie CLI for analysis
   - Extracts exported functions, classes, and constants
   - Generates comprehensive test cases covering:
     - Basic functionality and happy paths
     - Async operations (if applicable)
     - Edge cases (null, undefined, empty inputs)
     - Error handling
     - Proper dependency mocking
   - Uses AI to understand code semantics and generate realistic tests

4. **Output**: Places tests according to framework convention:
   - Jest: `foo.test.ts` next to `foo.ts`
   - Vitest: `foo.spec.ts` next to `foo.ts`
   - With `-out`: mirrors structure under specified directory

5. **Verification**: Runs the test suite on generated tests to verify they pass

6. **Coverage**: Optionally checks coverage against minimum threshold

### Concurrent Processing

The tool uses a worker pool pattern for concurrent test generation:
- Configurable number of workers (default: number of CPU cores)
- Each worker processes one file at a time
- Results are collected and written sequentially
- Failures are logged but don't stop the overall process

## Project Structure

```
auto-test-generator/
├── bin/                    # Compiled binary
├── cmd/
│   └── autotest/
│       └── main.go        # CLI entry point
├── internal/
│   ├── scan/
│   │   └── scan.go        # File scanning and git integration
│   ├── gen/
│   │   ├── generate.go    # Basic test generation
│   │   ├── augment.go     # Auggie CLI integration
│   │   ├── augment_context.go    # Context engine
│   │   └── context_generator.go  # Context-aware generation
│   └── exec/
│       └── runner.go      # Framework detection and test execution
├── example/               # Example TypeScript project for testing
├── Makefile              # Build and development tasks
├── go.mod                # Go module dependencies
└── README.md             # This file
```

## Framework Detection

The tool automatically detects the test framework in this order:

1. Check `devDependencies` in `package.json` for `vitest` (preferred)
2. Check `devDependencies` in `package.json` for `jest`
3. Check `dependencies` in `package.json` for `vitest`
4. Check `dependencies` in `package.json` for `jest`
5. Check lockfiles (`pnpm-lock.yaml`, `yarn.lock`, `package-lock.json`)

To override automatic detection, use `-fw jest` or `-fw vitest`.

## Roadmap

We're actively working on exciting new features to make test generation even more powerful and flexible:

### 🚀 Planned Features

#### 1. **Cursor IDE Integration**
- Direct integration with Cursor IDE
- Generate tests from within the editor
- Real-time test generation as you code
- Inline test suggestions

#### 2. **Direct AI Provider Support**
Choose your preferred AI provider without needing external CLI tools:
- **OpenAI** (GPT-4, GPT-3.5)
- **Anthropic Claude** (Claude 3, Claude 2)
- **Google Gemini** (Gemini Pro)
- Configure API keys in config file
- Switch providers on-the-fly with flags

#### 3. **Enhanced Context Engine**
- Better project understanding
- Cross-file dependency analysis
- Test template learning from existing tests
- Custom test patterns and styles

#### 4. **Additional Test Frameworks**
- React Testing Library support
- Vue Test Utils support
- Playwright/Cypress E2E tests
- Custom test framework templates

### Contributing

Contributions are welcome! If you'd like to help with any of the planned features or have ideas for new ones, please:
1. Open an issue to discuss the feature
2. Fork the repository
3. Create a feature branch
4. Submit a pull request

## Development

### Prerequisites

- Go 1.21 or higher
- Node.js (for running the example project)
- Git

### Build

```bash
# Build the binary
make build

# The binary will be at bin/autotest
```

### Run Tests

```bash
# Run Go tests
make test
```

### Format Code

```bash
# Format Go code
make fmt
```

### Lint

```bash
# Run linter
make lint
```

### All Checks

```bash
# Run all checks (fmt, lint, test)
make check
```

### Clean

```bash
# Remove build artifacts
make clean
```

### Development Workflow

```bash
# 1. Make changes to Go code
# 2. Format and lint
make fmt lint

# 3. Build
make build

# 4. Test with the example project
./bin/autotest -root ./example -allow-dirty -dry-run

# 5. Run tests
make test
```

## Dependencies

### Go Dependencies

- **`github.com/bmatcuk/doublestar/v4`**: Fast glob pattern matching for file scanning
- **`github.com/go-git/go-git/v5`**: Pure Go Git implementation for repository operations

### Runtime Dependencies

- **Auggie CLI** (`@augmentcode/auggie`): AI-powered test generation (auto-installed)
- **Node.js**: Required for running Jest/Vitest tests
- **npm/pnpm/yarn**: Package manager for Node.js projects

## Troubleshooting

### Auggie CLI Installation Fails

If automatic installation of Auggie CLI fails:

```bash
# Install manually
npm install -g @augmentcode/auggie

# Or with yarn
yarn global add @augmentcode/auggie
```

### Authentication Issues

If you encounter authentication errors:

```bash
# Re-login to Augment Code
./autotest login

# Or use Auggie directly
auggie --login
```

### Git Working Tree Dirty Error

If you see "working tree is dirty":

```bash
# Option 1: Commit your changes
git add .
git commit -m "Your changes"

# Option 2: Use --allow-dirty flag
./autotest -root ./your-project -allow-dirty
```

### Framework Not Detected

If the tool can't detect your test framework:

```bash
# Explicitly specify the framework
./autotest -root ./your-project -fw jest -allow-dirty
# or
./autotest -root ./your-project -fw vitest -allow-dirty
```

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [Augment Code](https://augmentcode.com/) for powering AI test generation
- [go-git](https://github.com/go-git/go-git) for Git operations
- [doublestar](https://github.com/bmatcuk/doublestar) for glob pattern matching

## Support

For questions, issues, or feature requests:
- Open an issue on GitHub
- Check existing issues for solutions
- Contribute improvements via pull requests

---

**Made with ❤️ for developers who want better test coverage**

