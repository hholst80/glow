# AGENTS.md

Development guidelines for the Glow project - a terminal-based markdown reader.

## Project Overview

- **Language:** Go 1.24+
- **Build:** `go build`
- **Test:** `go test ./...` or `task test`
- **Lint:** `golangci-lint run` or `task lint`

## Code Style

- Use `gofumpt` for formatting (stricter than `gofmt`)
- Use `goimports` for import organization
- Tab indentation for Go files
- Follow existing patterns in the codebase
- Keep functions focused and modular

## Issue-First Workflow

**All work must start with an Issue.** Do not implement features or fixes without a corresponding GitHub Issue.

### Creating an Issue
1. Check existing issues to avoid duplicates
2. Use clear, descriptive titles
3. Provide context: what, why, and acceptance criteria
4. Apply appropriate labels (see Label Definitions below)

### When to Implement
Only implement an Issue when it has the `ready-for-pr` label. This label signals:
- The issue has been triaged and accepted
- The approach has been discussed (if non-trivial)
- Someone can start implementation

## Label Definitions

### Type Labels (required)
| Label | Description |
|-------|-------------|
| `type:bug` | Something isn't working correctly |
| `type:feature` | New functionality request |
| `type:docs` | Documentation improvements |
| `type:chore` | Maintenance, dependencies, CI |

### Status Labels (required)
| Label | Description |
|-------|-------------|
| `status:triage` | Needs review and categorization |
| `status:accepted` | Issue is valid and will be addressed |
| `status:blocked` | Waiting on external dependency or decision |
| `status:wontfix` | Will not be addressed (with explanation) |

### Implementation Signal
| Label | Description |
|-------|-------------|
| `ready-for-pr` | Ready for implementation - PRs welcome |

### Priority Labels (optional)
| Label | Description |
|-------|-------------|
| `priority:high` | Should be addressed soon |
| `priority:low` | Nice to have, no urgency |

### Label Colors
```
type:bug        - #d73a4a (red)
type:feature    - #a2eeef (cyan)
type:docs       - #0075ca (blue)
type:chore      - #fef2c0 (yellow)
status:triage   - #fbca04 (yellow)
status:accepted - #0e8a16 (green)
status:blocked  - #b60205 (dark red)
status:wontfix  - #ffffff (white)
ready-for-pr    - #7057ff (purple)
priority:high   - #ff6600 (orange)
priority:low    - #c5def5 (light blue)
```

### Syncing Labels
Run this task to create/update all labels in the repository:
```bash
task labels:sync
```

### Re-labeling Issues
To have an AI agent review and re-label all open issues based on current conventions:
```bash
task labels:audit
```

This will output a report of suggested label changes. Review before applying.

## Process for Bug Fixes

### 0. Ensure Issue Exists
- Find or create an Issue describing the bug
- Issue must have `type:bug` and `ready-for-pr` labels before starting work

### 1. Reproduce and Understand
- Reproduce the bug locally before making changes
- Read relevant code thoroughly to understand the root cause
- Check if there are existing tests covering the affected area

### 2. Write a Failing Test (When Applicable)
- Add a test case that reproduces the bug
- Place tests in the appropriate `*_test.go` file
- For network-dependent tests, use `t.Skip()` when network unavailable

### 3. Implement the Fix
- Make minimal, focused changes to fix the issue
- Avoid refactoring unrelated code in the same PR
- Follow existing code patterns and conventions

### 4. Validate
```bash
task lint          # Check for linting issues
task test          # Run all tests
go build           # Ensure it compiles
```

### 5. Commit
Use conventional commit format:
```
fix: brief description of the fix

Longer explanation if needed.
Fixes #123
```

## Process for New Features

### 0. Ensure Issue Exists
- Find or create an Issue describing the feature
- Issue must have `type:feature` and `ready-for-pr` labels before starting work
- For non-trivial features, discuss approach in the Issue first

### 1. Plan the Implementation
- Review existing architecture in relevant packages:
  - `main.go` - CLI commands and source handling
  - `ui/` - TUI components (pager, stash, styles)
  - `utils/` - Helper utilities
- Identify which files need modification
- Consider cross-platform implications (see `*_windows.go`, `*_darwin.go`)

### 2. Implement Incrementally
- Start with core functionality
- Add UI components if needed (using Bubble Tea patterns)
- Follow existing patterns for:
  - CLI flags: Use Cobra command structure in `main.go`
  - Configuration: Use Viper patterns in `config_cmd.go`
  - Styling: Use Lip Gloss patterns in `ui/styles.go`

### 3. Add Tests
- Write unit tests for new functionality
- Test edge cases and error conditions
- Ensure tests pass locally before pushing

### 4. Update Documentation
- Update README.md if adding user-facing features
- Add usage examples for new CLI flags
- Document configuration options if applicable

### 5. Validate
```bash
task lint          # No new linting issues
task test          # All tests pass
go build && ./glow # Manual testing
```

### 6. Commit
Use conventional commit format:
```
feat: add new feature description

Explanation of what the feature does and why.
```

## Process for Creating a PR

### 1. Create a Feature Branch
```bash
git checkout master
git pull origin master
git checkout -b feature/short-description
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring

### 2. Make Changes and Commit
- Make focused, atomic commits
- Follow conventional commit format
- Keep commits logically grouped

### 3. Validate Before Pushing
```bash
task lint          # Fix any linting issues
task test          # Ensure all tests pass
go build           # Verify it compiles
```

### 4. Push and Create PR
```bash
git push -u origin feature/short-description
gh pr create --title "feat: description" --body "## Summary
- What this PR does

## Test Plan
- How it was tested"
```

### 5. Address CI Failures
If CI fails after pushing:
- Check the failing workflow in GitHub Actions
- Fix issues locally
- Push additional commits to the same branch
- CI will re-run automatically

### 6. Respond to Review Feedback
- Address all reviewer comments
- Push fixes as new commits (don't force-push during review)
- Re-request review when ready

## Process for PR Reviews

### Before Reviewing

Ensure CI checks pass:
- **build.yml** - Build succeeds
- **lint.yml** - No new linting issues
- **coverage.yml** - Tests pass with coverage

### Review Checklist

#### Code Quality
- [ ] Code follows existing patterns and conventions
- [ ] No unnecessary complexity or over-engineering
- [ ] Functions are focused and appropriately sized
- [ ] Error handling is appropriate

#### Security
- [ ] No hardcoded credentials or secrets
- [ ] User input is validated where applicable
- [ ] No obvious security vulnerabilities (OWASP top 10)

#### Testing
- [ ] New functionality has test coverage
- [ ] Edge cases are tested
- [ ] Tests are meaningful (not just for coverage)

#### Documentation
- [ ] README updated for user-facing changes
- [ ] Code comments where logic isn't self-evident
- [ ] Commit messages follow conventional format

#### Compatibility
- [ ] Cross-platform considerations addressed
- [ ] No breaking changes to existing behavior (unless intentional)

### Providing Feedback
- Be specific and actionable
- Explain the "why" behind suggestions
- Distinguish between blocking issues and nitpicks
- Approve when ready, request changes if blocking issues exist

## Linting Rules

The project uses extensive linting via `.golangci.yml`. Key enabled linters:
- `gosec` - Security checks
- `goconst` - Repeated strings that should be constants
- `misspell` - Spelling errors
- `revive` - Style and correctness
- `wrapcheck` - Error wrapping
- `unparam` - Unused parameters

Run `task lint` before committing to catch issues early.

## Testing Guidelines

```bash
# Run all tests
task test

# Run with coverage
go test -race -covermode atomic -coverprofile=profile.cov ./...

# Run specific test
go test -run TestName ./...
```

- Use table-driven tests for multiple cases
- Mark network-dependent tests with appropriate skip logic
- Keep tests focused and fast

## Commit Message Format

Follow conventional commits:

```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `style` - Formatting, no code change
- `refactor` - Code change that neither fixes bug nor adds feature
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

Examples:
```
feat: add support for GitLab URLs
fix: handle empty markdown files gracefully
chore(deps): bump golang.org/x/sys
```

## Directory Structure

```
/root/glow/
├── main.go              # Entry point, CLI commands
├── ui/                  # TUI components
│   ├── ui.go           # Main UI model
│   ├── pager.go        # Markdown viewport
│   ├── stash.go        # File browser
│   └── styles.go       # Styling definitions
├── utils/              # Helper utilities
├── Taskfile.yaml       # Task runner config
├── .golangci.yml       # Linter config
└── .github/workflows/  # CI/CD
```

## Dependencies

Key libraries to be familiar with:
- **cobra** - CLI framework
- **viper** - Configuration
- **bubbletea** - TUI framework
- **glamour** - Markdown rendering
- **lipgloss** - Terminal styling

## CI/CD

GitHub Actions runs on every PR:
1. **Build** - Compiles for multiple platforms
2. **Lint** - golangci-lint checks
3. **Coverage** - Tests with race detector and coverage reporting
4. **Security** - govulncheck, semgrep, ruleguard

All checks should pass before merging.

## Release Process

This fork uses **date-based versioning** instead of semantic versioning.

### Version Format

```
YYYY.MM.DD      - Primary release for that date
YYYY.MM.DD.N    - Additional releases on the same date (N = 1, 2, 3, ...)
```

Examples:
- `2025.12.30` - First release on December 30, 2025
- `2025.12.30.1` - Second release on the same day

### Creating a Release

1. **Ensure master is ready**
   ```bash
   git checkout master
   git pull origin master
   task lint && task test && go build
   ```

2. **Create a date-based tag**
   ```bash
   # Format: YYYY.MM.DD
   git tag 2025.12.30
   git push origin 2025.12.30
   ```

3. **CI/CD handles the rest**
   - Goreleaser builds static binaries for all platforms
   - Creates GitHub Release with artifacts
   - Generates checksums

### What Gets Built

The release includes static binaries for:
- **Operating Systems:** Linux, macOS, Windows, FreeBSD, OpenBSD, NetBSD
- **Architectures:** amd64 (x86_64), arm64

Each release includes:
- Binary archives (`.tar.gz`, `.zip` for Windows)
- Shell completions (bash, zsh, fish)
- Man pages
- Checksums file

### Distribution

**GitHub Releases only.** No package managers (Homebrew, apt, etc.).

Users download binaries directly from:
```
https://github.com/hholst80/glow/releases
```

### Hotfix Releases

For same-day fixes:
```bash
git tag 2025.12.30.1
git push origin 2025.12.30.1
```
