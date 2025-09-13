# Worktree CLI Development Tasks

## Recent Enhancements 🎉
- Module renamed to `github.com/todoengineering/wt`
- Fixed bug where editor didn't open in tmux session context
- Enhanced repository name detection for worktrees
- Implemented TOML configuration system with XDG compliance
- Added file copying feature for untracked files (.env, certificates, etc.)

## Completed ✅

- [x] Initialize Go module structure
- [x] Set up Cobra CLI framework
- [x] Create project directory structure
- [x] Implement basic `worktree list` command
  - [x] Check if in Git repository
  - [x] Get repository name
  - [x] List worktrees from standard directory
  - [x] Sort worktrees alphabetically

## Core Commands ✅

### `worktree new <name>`
- [x] Validate Git repository context
- [x] Create new Git branch
- [x] Create worktree directory structure
- [x] Run `git worktree add` command
- [x] Open in configured editor ($EDITOR)
- [x] Create/switch tmux session
- [x] Handle interactive mode (prompt for name)
- [x] Copy configured files to new worktree

### `worktree switch [name]`
- [x] Interactive worktree selection with fzf
- [x] Direct selection by name
- [x] Open in configured editor
- [x] Create/switch tmux session
- [x] Handle editor in tmux session context

### `worktree open` ✅
- [x] Check fzf availability
- [x] List all projects in worktree base directory
- [x] Interactive project selection with fzf
- [x] List worktrees for selected project
- [x] Interactive worktree selection with fzf
- [x] Open in configured editor
- [x] Create/switch tmux session

### `worktree checkout` ✅
- [x] Check fzf availability
- [x] Fetch remote branches
- [x] List all branches (local and remote)
- [x] Filter out current branch and HEAD
- [x] Interactive branch selection with fzf
- [x] Check if worktree already exists
- [x] Create worktree for selected branch
- [x] Handle special characters in branch names
- [x] Open in configured editor
- [x] Create/switch tmux session

### `worktree delete` ✅
- [x] Check fzf availability
- [x] List all worktrees for current repo
- [x] Interactive worktree selection with fzf
- [x] Show confirmation prompt
- [x] Validate confirmation input
- [x] Prevent deletion of main worktree
- [x] Run `git worktree remove --force`
- [x] Clean up orphaned directories

## CLI Simplification & UX 🧭

Goal: Make the tool simpler and more intuitive by consolidating overlapping commands, introducing consistent flags, and clarifying defaults. This plan captures the intended changes with examples and acceptance criteria.

### Command Consolidation

- [ ] Collapse `wt switch` into `wt open`
  - Behavior: `wt open` opens an existing worktree and switches/creates a tmux session, and opens the editor (current behavior of both `open` and `switch`).
  - Scope default: If inside a Git repo, list that repo’s worktrees; if outside, list across projects.
  - [ ] Keep `wt switch` as an alias temporarily with a deprecation notice printed once per run.
  - Acceptance:
    - Running `wt open` in a repo shows only that repo’s worktrees.
    - Running `wt open --all` shows a project → worktree picker across all projects.
    - Running `wt switch` behaves identically to `wt open` and prints a deprecation notice.

- [ ] Rename `wt checkout` to `wt branch` OR merge into `wt new --from`
  - Option A (recommended): Merge into `wt new` with a `--from <branch>` flag to attach a new worktree to an existing branch.
    - Example: `wt new feature-login --from origin/feature-login`
    - Benefit: Fewer top-level commands; a single verb to “add a worktree”.
  - Option B: Keep a distinct `wt branch [branch]` command (rename from `checkout`) for clarity when working from existing branches.
    - Example: `wt branch origin/release-1.2`
  - [ ] Decide and implement Option A or B; if B, add deprecation notice for `checkout`.
  - Acceptance:
    - Users can create a worktree from an existing branch either via `new --from` (A) or via `branch` (B).
    - Existing “worktree already exists” detection still offers to switch instead of duplicating.

### Common Flags (consistent behavior)

- [x] `--no-editor`: Skip launching the editor after create/open/switch
  - Applies to: `open`, `new`, `checkout` (current), `switch` (alias planned)
  - Example: `wt open --no-editor`

- [x] `--no-tmux`: Skip creating/switching tmux sessions
  - Applies to: `open`, `new`, `checkout` (current), `switch` (alias planned)
  - Example: `wt new ticket-42 --no-tmux`

- [x] `--force` on delete: Skip confirmation prompt
  - Applies to: `delete`
  - Example: `wt delete ticket-42 --force`

- [ ] Scope flags
- [x] `--all` for cross-project operations
    - Applies to: `open`, `list`
    - Examples:
      - `wt open --all` → project → worktree picker across all projects
      - `wt list --all` → list all projects and their worktrees
- [ ] `--project <name>` to filter to a specific repository (when outside any repo)
    - Applies to: `open`, `list`
    - Status: Implemented for `open` (flag `--project`); pending for `list`.
    - Example: `wt open --project my-repo`

- [ ] Output flags for scripting on `list`
- [x] `--json` → machine-readable output
- [x] `--path-only` → just the filesystem paths (one per line)
  - Examples:
    - `wt list --json | jq '.'`
    - `wt list --path-only | xargs -I{} du -sh {}`

### Default Behavior Adjustments

- [ ] `open` default scope:
  - In a repo → list only that repo’s worktrees (skip project selection).
  - Outside a repo → show project → worktree interactive selection.
  - Acceptance: Behavior matches above without requiring flags.

- [ ] `open <worktree-name>` shortcut when inside a repo
  - Example: `wt open feature-x` jumps directly if it exists; otherwise shows friendly error with suggestions.

- [ ] `open --all` accepts `<repo>/<worktree>` to jump directly
  - Example: `wt open --all my-repo/feature-x`

### Aliases (optional)

- [ ] `ls` → `list`
- [ ] `rm` → `delete`
- [ ] Keep `switch` alias to `open` during deprecation window

### Migration Guide (add to README)

- Before → After
  - `wt switch` → `wt open` (or `wt open --all` outside a repo)
  - `wt checkout` → `wt new --from <branch>` (Option A) or `wt branch` (Option B)
  - `wt list` → unchanged; add `--all`, `--json`, `--path-only`
  - `wt delete` → unchanged; add `--force`

Examples:

```bash
# Open an existing worktree in current repo
wt open
wt open payment-refactor

# Open across projects
wt open --all
wt open --all my-repo/payment-refactor

# Create worktree on a new branch
wt new feature/signup-flow

# Create worktree from an existing branch (Option A)
wt new release-1.2 --from origin/release-1.2

# OR if Option B is chosen
wt branch origin/release-1.2

# List worktrees
wt list
wt list --all
wt list --json
wt list --path-only

# Delete worktree
wt delete feature/signup-flow
wt delete feature/signup-flow --force

# Suppress integrations when desired
wt open --no-editor
wt new feature/foo --no-tmux
```

### Deprecation Strategy

- [ ] Print a one-line deprecation notice when `switch` or `checkout` are used, pointing to the replacement command.
- [ ] Keep aliases for at least one minor release cycle.
- [ ] Update `--help` for root and affected commands with new flags and examples.

### Implementation Checklist

- [ ] Wire `--no-editor` and `--no-tmux` through editor/tmux helpers without breaking defaults.
- [ ] Add `--all` and `--project` plumbing to `open`/`list` selection paths.
- [ ] Add JSON and path-only output modes to `list` with stable schema.
- [ ] Rename `checkout` → `branch` OR add `--from` to `new` and remove `checkout`.
- [ ] Implement aliasing + deprecation messages.
- [ ] Update README command docs and usage examples.
- [ ] Consider shell completions update once flags settle (bash/zsh/fish).

## Integration Features 🔧

### Editor Integration
- [x] Read $EDITOR environment variable
- [x] Parse editor command with flags
- [x] Validate editor availability
- [x] Handle editor launch errors
- [x] Support common editors (code, vim, emacs, etc.)

### Tmux Integration
- [x] Check if tmux is installed
- [x] Detect if running inside tmux
- [x] Create detached tmux sessions
- [x] Switch between tmux sessions
- [x] Attach to tmux sessions from outside
- [x] Handle session name conflicts

## Utility Functions 📦

### Git Operations
- [x] Check Git installation
- [x] Validate Git repository
- [x] Get current branch
- [x] List local branches
- [x] List remote branches
- [x] Create new branches
- [x] Run worktree commands

### File System
- [x] Create directory structures
- [ ] Clean up orphaned directories
- [x] Validate paths
- [x] Handle permissions errors

### Interactive UI
- [x] Check fzf installation
- [x] Create fzf selection interfaces
- [x] Handle user input prompts
- [ ] Display progress indicators
- [x] Show error messages

## Configuration 🔧

- [x] Support WORKTREE_BASE_DIR environment variable
- [x] Default to ~/projects/worktrees
- [x] Validate configuration values
- [x] Handle missing configuration gracefully
- [x] XDG-compliant config file support (TOML)
- [x] Global config at ~/.config/wt/config.toml
- [x] Local project config at .wt.toml
- [x] Config merging (local overrides/extends global)
- [x] Configurable worktrees location
- [x] Copy files to new worktrees (copy_files config)
- [x] Support glob patterns in copy_files

## Error Handling 🛡️

- [ ] Validate all dependencies on startup
- [ ] Provide clear error messages
- [ ] Handle edge cases gracefully
- [ ] Add recovery suggestions
- [ ] Log errors appropriately

## Testing 🧪

- [ ] Unit tests for Git operations
- [ ] Unit tests for worktree management
- [ ] Integration tests for commands
- [ ] Test tmux integration
- [ ] Test editor integration
- [ ] Test error scenarios

## Documentation 📚

- [x] Write README with installation instructions
- [x] Document all commands with examples
- [ ] Create man page
- [x] Add inline help text
- [x] Document configuration options

## Build & Release 🚀

- [ ] Set up build scripts
- [ ] Create Makefile
- [ ] Configure CI/CD pipeline
- [ ] Create release binaries for multiple platforms
- [ ] Set up versioning strategy
- [ ] Create installation script

## Future Enhancements 💡

- [ ] Shell completions (bash, zsh, fish)
- [ ] Worktree status command
- [ ] Sync command for updating worktrees
- [x] Config file support (TOML) - Completed
- [ ] Parallel worktree operations
- [ ] Cleanup command for orphaned worktrees
- [ ] Integration hooks for IDEs
- [ ] Worktree templates
