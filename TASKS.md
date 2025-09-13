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

### `worktree open`
- [ ] Check fzf availability
- [ ] List all projects in worktree base directory
- [ ] Interactive project selection with fzf
- [ ] List worktrees for selected project
- [ ] Interactive worktree selection with fzf
- [ ] Open in configured editor
- [ ] Create/switch tmux session

### `worktree checkout`
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
- [ ] List remote branches
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