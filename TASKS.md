# Worktree CLI Development Tasks

## Completed ✅

- [x] Initialize Go module structure
- [x] Set up Cobra CLI framework
- [x] Create project directory structure
- [x] Implement basic `worktree list` command
  - [x] Check if in Git repository
  - [x] Get repository name
  - [x] List worktrees from standard directory
  - [x] Sort worktrees alphabetically

## Core Commands 🚧

### `worktree new <name>`
- [ ] Validate Git repository context
- [ ] Create new Git branch
- [ ] Create worktree directory structure
- [ ] Run `git worktree add` command
- [ ] Open in configured editor ($EDITOR)
- [ ] Create/switch tmux session
- [ ] Handle interactive mode (prompt for name)

### `worktree open`
- [ ] Check fzf availability
- [ ] List all projects in worktree base directory
- [ ] Interactive project selection with fzf
- [ ] List worktrees for selected project
- [ ] Interactive worktree selection with fzf
- [ ] Open in configured editor
- [ ] Create/switch tmux session

### `worktree branch`
- [ ] Check fzf availability
- [ ] Fetch remote branches
- [ ] List all branches (local and remote)
- [ ] Filter out current branch and HEAD
- [ ] Interactive branch selection with fzf
- [ ] Check if worktree already exists
- [ ] Create worktree for selected branch
- [ ] Handle special characters in branch names
- [ ] Open in configured editor
- [ ] Create/switch tmux session

### `worktree delete`
- [ ] Check fzf availability
- [ ] List all worktrees for current repo
- [ ] Interactive worktree selection with fzf
- [ ] Show confirmation prompt
- [ ] Validate confirmation input
- [ ] Prevent deletion of main worktree
- [ ] Run `git worktree remove --force`
- [ ] Clean up orphaned directories

## Integration Features 🔧

### Editor Integration
- [ ] Read $EDITOR environment variable
- [ ] Parse editor command with flags
- [ ] Validate editor availability
- [ ] Handle editor launch errors
- [ ] Support common editors (code, vim, emacs, etc.)

### Tmux Integration
- [ ] Check if tmux is installed
- [ ] Detect if running inside tmux
- [ ] Create detached tmux sessions
- [ ] Switch between tmux sessions
- [ ] Attach to tmux sessions from outside
- [ ] Handle session name conflicts

## Utility Functions 📦

### Git Operations
- [ ] Check Git installation
- [ ] Validate Git repository
- [ ] Get current branch
- [ ] List local branches
- [ ] List remote branches
- [ ] Create new branches
- [ ] Run worktree commands

### File System
- [ ] Create directory structures
- [ ] Clean up orphaned directories
- [ ] Validate paths
- [ ] Handle permissions errors

### Interactive UI
- [ ] Check fzf installation
- [ ] Create fzf selection interfaces
- [ ] Handle user input prompts
- [ ] Display progress indicators
- [ ] Show error messages

## Configuration 🔧

- [ ] Support WORKTREE_BASE_DIR environment variable
- [ ] Default to ~/projects/worktrees
- [ ] Validate configuration values
- [ ] Handle missing configuration gracefully

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

- [ ] Write README with installation instructions
- [ ] Document all commands with examples
- [ ] Create man page
- [ ] Add inline help text
- [ ] Document configuration options

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
- [ ] Config file support (YAML/TOML)
- [ ] Parallel worktree operations
- [ ] Cleanup command for orphaned worktrees
- [ ] Integration hooks for IDEs
- [ ] Worktree templates