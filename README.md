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

Configuration uses TOML format and follows XDG specifications:

- Global config: `~/.config/wt/config.toml`
- Local project config: `.wt.toml` (in repository root)
- Environment variable: `WORKTREE_BASE_DIR` (overrides config)

Example configuration:
```toml
worktrees_location = "~/projects/worktrees"
```

## Commands

### List worktrees
```bash
wt list
```

### Create new worktree
```bash
wt new <branch-name>
# or interactive mode:
wt new
```

### Open worktree (interactive selection)
```bash
wt open
```

### Switch to worktree
```bash
wt switch
```

### Create worktree from existing branch
```bash
wt checkout
```

### Delete worktree
```bash
wt delete
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

## Requirements

- Git
- Go 1.21+
- fzf (for interactive commands)
- tmux (optional, for session management)
- $EDITOR environment variable (optional, for editor integration)