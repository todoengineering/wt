# wt - Git Worktree CLI

A Go-based CLI tool for managing Git worktrees with enhanced features including automatic tmux session management and editor integration.

## Installation

### Build from source
```bash
go build -o wt .
```

### Install globally
```bash
go install .
```

## Configuration

### Configuration Hierarchy

`wt` uses a hierarchical configuration system with the following priority (highest to lowest):

1. **Environment Variables** - Override all other settings
2. **Local Project Config** - `.wt.toml` in repository root
3. **Global User Config** - `~/.config/wt/config.toml` (XDG compliant)
4. **Built-in Defaults**

### Configuration Options

#### `worktrees_location`
**Type:** String
**Default:** `~/projects/worktrees`
**Description:** Base directory where all worktrees are created

#### `copy_files`
**Type:** Array of strings
**Default:** `[]` (empty)
**Description:** List of files/patterns to copy from main repository to new worktrees. Supports glob patterns.

#### `tmux_windows`
**Type:** Array of tables (`{ name = "<window-name>", command = "<optional shell command>" }`)
**Default:** `[]` (empty)
**Description:** Defines the tmux windows to create when a new session is started. Leave `command` empty to open a plain shell. Commands run in the worktree directory, so you can start servers, test runners, or editors automatically.

### Environment Variables

#### `WORKTREE_BASE_DIR`
Overrides the `worktrees_location` setting from configuration files.

```bash
export WORKTREE_BASE_DIR="/custom/worktree/path"
```

#### `XDG_CONFIG_HOME`
Changes the location of the global config file (follows XDG specification).

```bash
export XDG_CONFIG_HOME="/custom/config/path"
# Global config will be at: /custom/config/path/wt/config.toml
```

### Configuration Examples

#### Basic Global Configuration
**File:** `~/.config/wt/config.toml`
```toml
# Set custom worktree location
worktrees_location = "~/dev/worktrees"

# Copy common development files to new worktrees
copy_files = [
    ".env.example",
    ".vscode/settings.json",
    "docker-compose.yml"
]
```

#### Project-Specific Configuration
**File:** `.wt.toml` (in repository root)
```toml
# Override worktree location for this project only
worktrees_location = "/tmp/project-worktrees"

# Additional files to copy (merged with global)
copy_files = [
    "config/local.json",
    "certificates/*.pem"
]
```

#### Development Environment Setup
**Global config** for shared settings:
```toml
worktrees_location = "~/projects/worktrees"
copy_files = [
    ".env.example",
    ".gitignore",
    "README.md"
]
```

**Project config** for specific needs:
```toml
# Add project-specific files to copy
copy_files = [
    "docker-compose.dev.yml",
    "config/development.json",
    "scripts/setup.sh"
]

# Predefine tmux windows for this project
tmux_windows = [
    { name = "editor", command = "nvim ." },
    { name = "server", command = "pnpm dev" },
    { name = "tests", command = "pnpm test --watch" },
    { name = "terminal", command = "" }
]
```

#### File Copy Patterns

The `copy_files` configuration supports:

- **Exact filenames:** `.env.example`, `package.json`
- **Glob patterns:** `config/*.json`, `certificates/*.pem`, `scripts/*`
- **Nested paths:** `.vscode/settings.json`, `config/environments/dev.yml`

**Example patterns:**
```toml
copy_files = [
    # Environment files
    ".env*",

    # Configuration directories
    "config/**/*.json",

    # Development tools
    ".vscode/",
    ".idea/",

    # Scripts and utilities
    "scripts/setup.sh",
    "Makefile",

    # Docker files
    "docker-compose*.yml",
    "Dockerfile*"
]
```

### Configuration Merging

- **Global + Local:** `copy_files` arrays are merged (local appends to global)
- **Global + Local:** `tmux_windows` arrays are merged (global windows first, then local additions; duplicates are preserved)
- **Overrides:** `worktrees_location` in local config overrides global
- **Deduplication:** Duplicate entries in `copy_files` are automatically removed

## Commands

### Global Flags

All commands that open editors or create tmux sessions support these flags:

- `--no-editor` - Don't open the editor
- `--no-tmux` - Don't create/switch tmux sessions

### List worktrees
```bash
wt list
```

Lists all worktrees for the current repository, showing name, branch, and path.

### Create new worktree
```bash
# Create worktree for new branch
wt new <branch-name>

# Create from existing branch
wt new --from <branch-name>

# Interactive mode (prompts for branch name)
wt new

# Create without opening editor
wt new <branch-name> --no-editor

# Create without tmux integration
wt new <branch-name> --no-tmux
```

Creates a new Git branch and worktree, copies configured files, opens in editor, and creates tmux session.

### Open worktree
```bash
wt open

# Show all projects even when in a repo
wt open --all

# Filter to specific project
wt open --project <project-name>
```

Default behavior:
- In a repository: lists only that repo’s worktrees (no project selection).
- Outside a repository: choose project, then choose a worktree within it.
- Use `--all` to browse across all projects even when in a repository.

### Create worktree from existing branch
```bash
# Create worktree for specific branch, derive name from branch
wt new --from origin/release-1.2

# Create worktree for specific branch with explicit name
wt new release-1.2 --from origin/release-1.2

# Interactive branch selection (uses fzf)
wt new --from :pick
```

If a worktree already exists for the branch, wt offers to switch to it instead of creating a duplicate.

### Delete worktree
```bash
# Interactive selection with confirmation
wt delete

# Delete specific worktree with confirmation
wt delete <worktree-name>

# Force deletion without confirmation
wt delete <worktree-name> --force
```

Interactive deletion with safety checks:
- Shows branch information and path
- Prevents deletion of main repository
- Requires explicit confirmation (type "yes")
- Automatically kills associated tmux sessions
- `--force` flag skips confirmation prompt

### Command Examples

#### Typical Workflow
```bash
# List existing worktrees
wt list

# Create new feature branch worktree
wt new feature/user-auth

# Open an existing worktree
wt open

# Create worktree from existing branch
wt new --from origin/release-1.2

# Clean up when done
wt delete feature/user-auth
```

#### Multi-Project Development
```bash
# Open any worktree across all projects
wt open --all

# Work on specific project
wt open --project my-api

# Create worktree in current project
wt new hotfix/critical-bug
```

## Development

### Dependencies
```bash
go mod download
```

### Run without building
```bash
go run . <command>
```

### Run tests
```bash
go test ./...
```

### GitHub PR Workflow
1. Create a feature branch from `main` (`git checkout -b feature/my-change`).
2. Run `go fmt ./...` or `gofmt -w .` to keep formatting consistent.
3. Execute `go test ./...` and any relevant integration checks before pushing.
4. `git status --short` to verify only intentional changes are staged, then commit with a concise message.
5. Push the branch, open a PR against `main`, and include context, test results, and any follow-up notes.

## Requirements

- Git
- Go 1.21+
- fzf (for interactive commands)
- tmux (optional, for session management)
- $EDITOR environment variable (optional, for editor integration)
