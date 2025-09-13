# Product Requirements Document: Worktree CLI

## Overview
A Go-based CLI tool for managing Git worktrees with enhanced features including automatic tmux session management and editor integration.

## Goals
- Simplify Git worktree management with an intuitive interface
- Automate common workflows around worktrees
- Integrate seamlessly with development environments (tmux, editors)
- Provide a consistent worktree organization structure

## Core Features

### 1. Worktree Organization
- **Standard Directory Structure**: All worktrees stored in `~/projects/worktrees/{repository-name}/`
- **Automatic Directory Creation**: Creates necessary directories if they don't exist
- **Repository Context Awareness**: Commands work within the context of the current Git repository

### 2. Commands

#### `worktree new <name>`
- Creates a new Git branch with the specified name
- Creates a worktree for that branch in the standard location
- Opens the worktree in the configured editor
- Creates and switches to a tmux session named after the worktree

**Interactive Mode**: If name not provided, prompts user for input

#### `worktree open`
- Interactive two-step selection process using fzf:
  1. Select project from available repositories
  2. Select specific worktree within that project
- Opens selected worktree in configured editor
- Creates/switches to corresponding tmux session

#### `worktree list`
- Lists all on-disk worktrees for the current repository
- Shows worktrees from the standard directory structure
- Displays in sorted order

#### `worktree branch`
- Lists all available branches (local and remote) using fzf
- Excludes current branch and HEAD from selection
- Fetches remote branches if not available locally
- Creates worktree for selected branch
- Handles special characters in branch names (converts to safe directory names)
- Detects and opens existing worktrees instead of creating duplicates
- Opens in editor and creates/switches tmux session

#### `worktree delete`
- Interactive selection of worktree to delete using fzf
- Shows worktree name and associated branch
- Requires explicit confirmation ("yes")
- Prevents deletion of main repository worktree
- Force removes worktree and cleans up references

### 3. Editor Integration

#### Behavior
- Respects `$EDITOR` environment variable
- Supports editor commands with flags (e.g., `code -n`)
- Opens worktree directory in editor after creation/selection
- Shows clear error if EDITOR not configured

### 4. Tmux Integration

#### Session Management
- **Session Naming**: Uses worktree name as tmux session name
- **Automatic Creation**: Creates detached tmux session in worktree directory
- **Smart Switching**: 
  - If inside tmux: uses `tmux switch-client`
  - If outside tmux: uses `tmux attach-session`
- **Existing Session Handling**: Switches to existing session rather than creating duplicate
- **Graceful Degradation**: Works without tmux (skips session creation)

## Technical Requirements

### Dependencies
- Git (required)
- fzf (required for interactive commands: open, branch, delete)
- tmux (optional, for session management)
- Configured $EDITOR (required for opening worktrees)

### Error Handling
- Validate Git repository context before operations
- Check for required dependencies before using them
- Provide clear error messages with actionable guidance
- Handle edge cases (no worktrees, missing directories, etc.)

### Platform Support
- Primary target: Unix-like systems (Linux, macOS)
- Shell-agnostic (should work from any shell, not just zsh)

## User Experience

### Command Structure
```
worktree [command] [options]

Commands:
  new <name>    Create new worktree with new branch
  open          Open existing worktree (interactive)
  list          List worktrees for current repo
  branch        Create worktree from existing branch
  delete        Delete a worktree (interactive)
  help          Show usage information
```

### Interactive Elements
- Use fzf for all selection interfaces
- Provide clear prompts for user input
- Show contextual information during selection (paths, branches)
- Confirm destructive operations

### Feedback
- Clear status messages for all operations
- Progress indicators for long-running operations (fetch, worktree creation)
- Success/failure messages with next steps

## Configuration

### Environment Variables
- `EDITOR`: Editor command to use for opening worktrees
- `WORKTREE_BASE_DIR`: Override default `~/projects/worktrees` location (optional)

### Future Configuration Options
- Custom tmux session prefix
- Editor-specific configurations
- Branch naming conventions
- Auto-cleanup of orphaned worktrees

## Success Metrics
- Reduces time to switch between feature branches
- Eliminates manual worktree path management
- Prevents common worktree errors (duplicate names, orphaned directories)
- Seamless integration with existing development workflow

## Future Enhancements
1. **Worktree Templates**: Pre-configured setups for different project types
2. **Status Command**: Show status of all worktrees (dirty, ahead/behind, etc.)
3. **Sync Command**: Update all worktrees with latest changes
4. **Config File**: Per-project and global configuration options
5. **Shell Completions**: Bash/Zsh/Fish completion support
6. **Parallel Operations**: Create multiple worktrees simultaneously
7. **Cleanup Command**: Remove orphaned worktrees and prune references
8. **Integration APIs**: Hooks for IDE plugins and other tools
