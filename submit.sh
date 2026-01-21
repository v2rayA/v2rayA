#!/bin/bash

# submit.sh
# Handles Git operations for both 'develop' and 'dev/xxx' feature branches.
# Supports separating Private and Public commits/pushes.
# Implements anchor-based incremental sync for Public commits.
# Smartly handles multiple remotes and visibility detection.
# Automatically ensures execution from Git root directory.

set -e

# ==========================================
# 1. Configuration
# ==========================================

DIFF_LINE_LIMIT=500
PRIVATE_DIFF_FILE="private.diff"
PUBLIC_DIFF_FILE="public.diff"

# Define Private Patterns List
# Add any file or directory path that should stay PRIVATE on develop branch
# These are regex patterns anchored to the root
PRIVATE_PATTERNS=(
    "^memory-bank/"
    "^doc/"
    "^\.clinerules/"
    "^\.claude/"
    "^\.vscode/"
    "^\.clineignore"
    "^AGENTS\.md"
    "^openspec/"
    "^scripts/"
    "^images/"
    "^test\.sh"
    "^submit\.sh"
    "^\.continue/"
)

# Construct Regex from list
PRIVATE_REGEX=""
for pattern in "${PRIVATE_PATTERNS[@]}"; do
    if [ -z "$PRIVATE_REGEX" ]; then
        PRIVATE_REGEX="$pattern"
    else
        PRIVATE_REGEX="$PRIVATE_REGEX|$pattern"
    fi
done

# Get script's own path relative to git root (for self-restoration logic)
# This is computed lazily when needed, but the function is defined here
get_script_rel_path() {
    local script_abs
    local git_root
    local script_dir
    
    # Get script directory without using cd (avoid failure when directory doesn't exist)
    script_dir="$(dirname "${BASH_SOURCE[0]}")"
    if [ -d "$script_dir" ]; then
        script_abs="$(cd "$script_dir" && pwd)/$(basename "${BASH_SOURCE[0]}")"
    else
        # Directory doesn't exist, use realpath or original path as fallback
        script_abs="$(realpath "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")"
    fi
    
    git_root="$(git rev-parse --show-toplevel 2>/dev/null || echo "")"
    if [ -n "$git_root" ]; then
        # Convert to relative path from git root
        echo "${script_abs#$git_root/}"
    else
        # Fallback: use the path as invoked
        echo "${BASH_SOURCE[0]}"
    fi
}

# ==========================================
# 2. Helper Functions
# ==========================================

# --- Logging Helpers ---
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Description: Print info message in green
# Usage: log_info "message"
# Params:
#   $1 - Message content
# Return: Formatted string to stdout
# Exit: 0
log_info() {
    echo -e "${GREEN}[INFO] $1${NC}"
}

# Description: Print warning message in yellow
# Usage: log_warn "message"
# Params:
#   $1 - Message content
# Return: Formatted string to stdout
# Exit: 0
log_warn() {
    echo -e "${YELLOW}[WARN] $1${NC}"
}

# Description: Print error message in red
# Usage: log_error "message"
# Params:
#   $1 - Message content
# Return: Formatted string to stdout
# Exit: 0
log_error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

STATUS_ORDER=("modified" "added" "deleted" "renamed" "copied" "unmerged" "untracked" "clean")
declare -A FILE_STATUS_LABELS
declare -A SELECTED_DIRS
declare -A DIR_SIZE_CACHE

# --- Git Root Validation ---

# Description: Ensure script is running from the git repository root.
#              Uses Inode comparison for robust path validation.
# Usage: ensure_git_root
# Params: None
# Return: None (Logs to stderr on error)
# Exit: Terminates script with exit code 1 if not in root
ensure_git_root() {
    local git_root
    local root_inode
    local pwd_inode

    git_root=$(git rev-parse --show-toplevel 2>/dev/null || echo "")

    if [ -z "$git_root" ]; then
        log_error "Not a git repository."
        log_info "To initialize this directory as a git repository, run:"
        echo "    $0 --init"
        exit 1
    fi

    # Compare Inode numbers to handle path format differences (e.g. /c/ vs C:/)
    # This is a robust way to check if PWD is physically the same as GIT_ROOT
    root_inode=$(stat -c '%i' "$git_root" 2>/dev/null || echo "0")
    pwd_inode=$(stat -c '%i' "$PWD" 2>/dev/null || echo "1")

    if [ "$root_inode" != "$pwd_inode" ]; then
        log_error "Script must be run from the repository root."
        log_error "Current directory: $PWD"
        log_error "Repository root:   $git_root"
        exit 1
    fi
}

# --- Repository Initialization ---

# Description: Initialize a non-git directory as a git repository with proper branch structure.
#              Creates main/master with public files, then develop with all files.
#              Automatically detects nested git repositories and adds them as submodules.
# Usage: init_repository
# Params: None
# Return: None
# Exit: 0 on success, 1 on failure
init_repository() {
    # 1. Check if already a git repository
    if git rev-parse --show-toplevel &>/dev/null; then
        log_error "This directory is already a git repository."
        log_info "Use other commands to manage existing repository."
        exit 1
    fi

    # 2. Scan for nested git repositories BEFORE initializing
    log_info "Scanning for nested git repositories..."
    local nested_repos=""
    local public_submodules=""
    local private_submodules=""
    
    # Find all .git directories (excluding future top-level .git)
    # mindepth 2 ensures we skip ./.git (which doesn't exist yet anyway)
    while IFS= read -r git_dir; do
        [ -z "$git_dir" ] && continue
        local repo_path="${git_dir%/.git}"
        repo_path="${repo_path#./}"
        
        # Check if this nested repo has a remote configured
        local remote_url
        remote_url=$(git -C "$repo_path" remote get-url origin 2>/dev/null || echo "")
        
        if [ -z "$remote_url" ]; then
            log_error "Nested git repository '$repo_path' has no remote configured."
            log_error "Please configure a remote for it before running --init:"
            echo ""
            echo "    cd $repo_path && git remote add origin <your-repo-url>"
            echo ""
            exit 1
        fi
        
        nested_repos+="$repo_path"$'\n'
        
        # Classify as public or private based on PRIVATE_PATTERNS
        if is_private_file "$repo_path/"; then
            private_submodules+="$repo_path|$remote_url"$'\n'
            log_info "  Found private submodule: $repo_path"
        else
            public_submodules+="$repo_path|$remote_url"$'\n'
            log_info "  Found public submodule: $repo_path"
        fi
    done < <(find . -mindepth 2 -name ".git" -type d 2>/dev/null)
    
    # Remove trailing newlines
    nested_repos="${nested_repos%$'\n'}"
    public_submodules="${public_submodules%$'\n'}"
    private_submodules="${private_submodules%$'\n'}"
    
    if [ -z "$nested_repos" ]; then
        log_info "No nested git repositories found."
    fi

    # 3. Initialize the repository
    log_info "Initializing git repository..."
    git init

    # 4. Get default branch name from git config (main or master)
    local default_branch
    default_branch=$(git config --get init.defaultBranch 2>/dev/null || echo "main")
    
    log_info "Default branch will be: $default_branch"

    # 5. Collect public files (exclude private patterns, .git directories, and nested .git contents)
    log_info "Scanning for public files..."
    local public_files
    public_files=$(find . -type f -not -path './.git/*' -not -path '*/.git/*' 2>/dev/null | sed 's|^\./||' | grep -vE "$PRIVATE_REGEX" || true)

    if [ -z "$public_files" ] && [ -z "$public_submodules" ]; then
        log_warn "No public files found. Creating empty initial commit on $default_branch."
        git commit --allow-empty -m "Initial empty commit"
    else
        if [ -n "$public_files" ]; then
            log_info "Adding public files to initial commit..."
            # Use xargs instead of while loop to avoid subshell issues with set -e
            # Use --ignore-errors to suppress warnings about .gitignore'd files
            echo "$public_files" | xargs -d '\n' git add --ignore-errors 2>/dev/null || true
        fi
        
        # Add public submodules
        if [ -n "$public_submodules" ]; then
            log_info "Adding public submodules..."
            while IFS='|' read -r repo_path remote_url; do
                [ -z "$repo_path" ] && continue
                log_info "  Adding submodule: $repo_path -> $remote_url"
                git submodule add "$remote_url" "$repo_path"
            done <<< "$public_submodules"
        fi
        
        git commit -m "Initial public commit"
    fi

    log_info "Branch '$default_branch' created with public files."

    # 6. Create develop branch (based on main/master)
    log_info "Creating develop branch based on $default_branch..."
    git checkout -b develop

    # 7. Add private submodules and private files
    local has_private_changes=0
    
    # Add private submodules first
    if [ -n "$private_submodules" ]; then
        log_info "Adding private submodules..."
        while IFS='|' read -r repo_path remote_url; do
            [ -z "$repo_path" ] && continue
            log_info "  Adding submodule: $repo_path -> $remote_url"
            git submodule add "$remote_url" "$repo_path"
            has_private_changes=1
        done <<< "$private_submodules"
    fi
    
    # Add private files
    log_info "Scanning for private files..."
    local private_files
    private_files=$(find . -type f -not -path './.git/*' -not -path '*/.git/*' 2>/dev/null | sed 's|^\./||' | grep -E "$PRIVATE_REGEX" || true)

    if [ -n "$private_files" ]; then
        log_info "Adding private files..."
        # Use xargs instead of while loop to avoid subshell issues with set -e
        # Use --ignore-errors to suppress warnings about .gitignore'd files
        echo "$private_files" | xargs -d '\n' git add --ignore-errors 2>/dev/null || true
        has_private_changes=1
    fi
    
    if [ "$has_private_changes" -eq 1 ]; then
        git commit -m "Add private development files"
    else
        log_info "No private files or submodules found."
    fi

    echo ""
    log_info "Repository initialized successfully!"
    echo ""
    echo "Branch structure:"
    # Calculate max branch name length for alignment (develop=7 chars)
    local max_len=${#default_branch}
    [ 7 -gt $max_len ] && max_len=7
    printf "  %-${max_len}s  <- Public files only (for open source)\n" "$default_branch"
    printf "  %-${max_len}s  <- All files (public + private, current branch)\n" "develop"
    
    if [ -n "$public_submodules" ] || [ -n "$private_submodules" ]; then
        echo ""
        echo "Submodules:"
        if [ -n "$public_submodules" ]; then
            while IFS='|' read -r repo_path remote_url; do
                [ -z "$repo_path" ] && continue
                echo "  $repo_path (public)"
            done <<< "$public_submodules"
        fi
        if [ -n "$private_submodules" ]; then
            while IFS='|' read -r repo_path remote_url; do
                [ -z "$repo_path" ] && continue
                echo "  $repo_path (private)"
            done <<< "$private_submodules"
        fi
    fi
    
    echo ""
    echo "Next steps:"
    echo "  1. Add remote:  git remote add origin <your-repo-url>"
    echo "  2. Push develop: $0 --push-private"
    echo "  3. Push public:  $0 --push-public"
}

# --- Help Message ---

# Description: Display script usage and options
# Usage: show_help
# Params: None
# Return: Help text to stdout
# Exit: 0
show_help() {
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  --init                    Initialize a non-git directory with proper branch structure."
    echo "                            Creates main/master with public files, develop with all files."
    echo "  --print-private-patterns  Print the list of regex patterns for private files."
    echo "  --print-public-patterns   Print the list of regex patterns for public files (inverse of private)."
    echo "  --print-private-files     Print all private files (aggregated by folder)."
    echo "  --print-public-files      Print all public files (aggregated by folder)."
    echo "  --print-private-status    Show status of private modified/untracked files."
    echo "  --print-public-status     Show status of public modified/untracked files."
    echo "  --print-private-diff      Show git diff of private modified files."
    echo "  --print-public-diff       Show git diff of public modified files."
    echo "  --print-private-show      Show status and git diff of private modified files."
    echo "  --print-public-show       Show status and git diff of public modified files."
    echo "  --check-prerequisites     Check branch naming and base purity."
    echo "  --new-dev-branch          Create a new dev branch. Usage: --new-dev-branch <new-feature> <base>"
    echo "  --commit-private          Commit private changes. Requires 'COMMIT_MSG_PRIVATE' env var."
    echo "  --commit-public           Commit public changes. Requires 'COMMIT_MSG_PUBLIC' env var."
    echo "  --push-private            Push current private branch (develop or dev/xxx) to PRIVATE remotes only."
    echo "  --push-public             Sync public commits to public branch (main/feature...), then push to ALL remotes."
    echo "  --restore-private [ref]   Restore private files. Auto-detects source if ref is omitted."
    echo "  --help                    Show this help message."
    echo ""
    echo "Configuration:"
    echo "  DIFF_LINE_LIMIT           $DIFF_LINE_LIMIT (Diffs longer than this are written to file)"
    echo "  PRIVATE_DIFF_FILE         $PRIVATE_DIFF_FILE"
    echo "  PUBLIC_DIFF_FILE          $PUBLIC_DIFF_FILE"
    echo ""
    echo "Environment Variables:"
    echo "  COMMIT_MSG_PRIVATE        Commit message for private changes."
    echo "  COMMIT_MSG_PUBLIC         Commit message for public changes."
}

# --- Cleanup ---

# Description: Trap handler to restore original branch on exit
# Usage: Trap implicitly calls this
# Params: None
# Return: None
# Exit: Preserves the exit code of the last command
cleanup() {
    local exit_code=$?
    local current_branch=$(git branch --show-current)
    if [ -n "$BRANCH_PRIVATE" ] && [ "$current_branch" != "$BRANCH_PRIVATE" ]; then
        log_info "Switching back to $BRANCH_PRIVATE..."
        git checkout "$BRANCH_PRIVATE"
    fi
    exit $exit_code
}

# --- Branch Detection Helpers ---

# Description: Detect default main branch name (main or master)
# Usage: get_main_branch_name "origin"
# Params:
#   $1 - Remote name (optional, defaults to checking local/origin)
# Return: "main" or "master" (Stdout)
# Exit: 0
get_main_branch_name() {
    local remote="${1:-origin}"
    
    if git show-ref --verify --quiet "refs/heads/main" || git show-ref --verify --quiet "refs/remotes/$remote/main"; then
        echo "main"
    elif git show-ref --verify --quiet "refs/heads/master" || git show-ref --verify --quiet "refs/remotes/$remote/master"; then
        echo "master"
    else
        # Default fallback if neither found (unlikely in valid repo)
        echo "main"
    fi
}

# --- Sanity Checks ---

# Description: Check if current branch is based on a private branch (develop or dev/*).
#              Fails if parent commit belongs ONLY to other private branches.
# Usage: check_private_base
# Params: None
# Return: None
# Exit: 1 if invalid base detected
check_private_base() {
    local parent_commit
    
    # First verify HEAD^ is a valid commit reference
    # --verify: requires the input to be a valid object name
    # --quiet: do not output anything, just return exit code
    if ! git rev-parse --verify --quiet HEAD^ >/dev/null 2>&1; then
        # Initial commit or orphan branch, skip check
        return 0
    fi
    
    parent_commit=$(git rev-parse HEAD^)

    # 1. Immunity Check: Is parent part of ANY public branch?
    # grep -vE "dev/|develop" filters out private branches. If output is non-empty, it's public.
    # Check both local and remote public branches.
    if git branch --contains "$parent_commit" | grep -vE "dev/|develop" >/dev/null 2>&1; then
        return 0
    fi
    if git branch -r --contains "$parent_commit" | grep -vE "dev/|develop" >/dev/null 2>&1; then
        return 0
    fi

    # 2. Guilt Check: Is parent part of ANY OTHER private branch?
    # If it's not public, but belongs to another dev branch, it's a private base.
    local current_branch
    current_branch=$(git branch --show-current)
    local private_refs
    private_refs=$(git branch --contains "$parent_commit" | grep -E "dev/|develop" | grep -v "$current_branch" || true)

    if [ -n "$private_refs" ]; then
        log_error "Sanity Check Failed: Current branch seems to be based on private history."
        log_error "Parent commit $parent_commit belongs to private branch(es):"
        echo "$private_refs"
        log_error "Please rebase your branch onto a public branch (main, master, or feature/*)."
        log_error "Use '$0 --restore-private' to recover private files if needed."
        exit 1
    fi
}

# --- Prerequisites Check ---

# Description: Run branch naming and base purity checks explicitly.
# Usage: check_prerequisites
# Params: None
# Return: None
# Exit: 1 if checks fail
check_prerequisites() {
    local current_branch
    current_branch=$(git branch --show-current)

    if [ "$current_branch" = "develop" ]; then
        log_info "Branch check passed: develop"
    elif [[ "$current_branch" == dev/* ]]; then
        log_info "Branch check passed: $current_branch"
    else
        log_error "Current branch '$current_branch' is not a valid development branch."
        log_error "Must be 'develop' or start with 'dev/'."
        exit 1
    fi

    check_private_base
    log_info "Prerequisites check passed."
}

# --- New Dev Branch Creation Helpers ---

# Description: List remote public branches (exclude develop/dev/*/dev and HEAD)
# Usage: list_remote_public_branches
# Params: None
# Return: List of remote public branches
list_remote_public_branches() {
    local branches
    branches=$(git branch -r --list | sed 's/^ *//' | grep -vE '/dev/|/develop$|/dev$|/HEAD' || true)

    if [ -z "$branches" ]; then
        log_warn "No remote public branches found."
        return
    fi

    echo "$branches"
}

# Description: Create a new dev branch from a public base branch
# Usage: create_new_dev_branch <new-feature> <base>
# Params:
#   $1 - New feature name
#   $2 - Base branch (main/master/feature/* or origin/*)
# Exit: 1 on failure
create_new_dev_branch() {
    local feature_name="$1"
    local base_input="$2"

    if [ -z "$feature_name" ]; then
        log_error "Missing new feature name."
        echo "Usage: $0 --new-dev-branch <new-feature> <base>"
        exit 1
    fi

    if [ -z "$base_input" ]; then
        log_error "Missing base branch."
        log_info "Available remote public branches:"
        git fetch --all --prune >/dev/null 2>&1 || log_warn "Fetch failed. Listing cached remotes..."
        list_remote_public_branches
        echo ""
        echo "Usage: $0 --new-dev-branch <new-feature> <base>"
        exit 1
    fi

    log_info "Fetching remotes..."
    git fetch --all --prune || log_warn "Fetch failed. Proceeding with cached remotes..."

    local base_branch="$base_input"
    local base_remote=""
    local local_base=""

    if [[ "$base_input" == */* ]]; then
        base_branch="${base_input#*/}"
        base_remote="$base_input"
        local_base="$base_branch"
    else
        base_branch="$base_input"
        base_remote="origin/$base_input"
        local_base="$base_input"
    fi

    if [[ "$base_branch" == "develop" || "$base_branch" == dev/* ]]; then
        log_error "Base branch '$base_branch' is private. Use a public branch (main/master/feature/*)."
        exit 1
    fi

    if [[ "$base_branch" != "main" && "$base_branch" != "master" && "$base_branch" != feature/* ]]; then
        log_error "Base branch '$base_branch' is not a supported public branch."
        log_error "Allowed: main, master, feature/*"
        exit 1
    fi

    if ! git show-ref --verify --quiet "refs/remotes/$base_remote"; then
        log_error "Remote base branch '$base_remote' not found."
        exit 1
    fi

    if git show-ref --verify --quiet "refs/heads/$local_base"; then
        local counts
        counts=$(git rev-list --left-right --count "$local_base...$base_remote")
        local ahead
        local behind
        ahead=$(echo "$counts" | awk '{print $1}')
        behind=$(echo "$counts" | awk '{print $2}')

        if [ "$ahead" -ne 0 ] || [ "$behind" -ne 0 ]; then
            log_error "Local branch '$local_base' is not in sync with '$base_remote'."
            log_error "Ahead: $ahead, Behind: $behind. Please sync before creating a new dev branch."
            exit 1
        fi
    fi

    local new_branch="dev/$feature_name"
    if git show-ref --verify --quiet "refs/heads/$new_branch"; then
        log_error "Branch '$new_branch' already exists."
        exit 1
    fi

    # Remember current branch to copy script from
    local original_branch="$CURRENT_BRANCH"
    
    log_info "Creating $new_branch from $base_remote..."
    git checkout -b "$new_branch" "$base_remote"
    log_info "New dev branch created: $new_branch"
    
    # Copy this script from original branch to enable --restore-private
    local script_path
    script_path=$(get_script_rel_path)
    if [ -n "$original_branch" ] && git show "$original_branch:$script_path" &>/dev/null; then
        log_info "Copying $script_path from $original_branch..."
        local script_dir
        script_dir=$(dirname "$script_path")
        if [ "$script_dir" != "." ]; then
            mkdir -p "$script_dir"
        fi
        git show "$original_branch:$script_path" > "$script_path"
        chmod +x "$script_path"
    fi
    
    echo ""
    echo "Next steps:"
    echo "  1. Restore private files:  $0 --restore-private"
    echo "  2. Initialize private commit:  COMMIT_MSG_PRIVATE='chore: initialize private context' $0 --commit-private"
}

# --- Smart Parent Detection ---

# Description: Find the best parent branch for creating a new feature branch based on merge-base history
# Usage: find_best_parent "origin" "dev/feature-a"
# Params:
#   $1 - Remote name
#   $2 - Current private branch name
# Return: Commit hash of the best merge base (Stdout), or empty string
# Exit: 0
find_best_parent() {
    local remote="$1"
    local private_branch="$2"
    
    # Candidates: ALL remote branches EXCEPT private branches (dev/*, develop, dev) and HEAD
    # This ensures we only consider public branches like 'main', 'master', 'feature/*'
    local candidates=$(git branch -r --list "$remote/*" | grep -vE "$remote/dev/|$remote/develop$|$remote/dev$|$remote/HEAD")
    
    local best_base=""
    local best_candidate=""
    local latest_time=0
    
    OLD_IFS="$IFS"
    IFS=$'\n'
    for cand in $candidates; do
        # Strip whitespace
        cand=$(echo "$cand" | xargs)
        
        # Calculate merge base
        # Redirect stderr to suppress errors if unrelated histories
        local base=$(git merge-base "$private_branch" "$cand" 2>/dev/null || true)
        
        if [ -n "$base" ]; then
            # Get commit timestamp of the base
            local ts=$(git show -s --format=%ct "$base")
            
            if [ "$ts" -gt "$latest_time" ]; then
                latest_time=$ts
                best_base=$base
                best_candidate=$cand
            fi
        fi
    done
    IFS="$OLD_IFS"
    
    if [ -n "$best_base" ]; then
        echo "$best_base" # Return the commit hash
        log_info "Detected parent: $best_candidate (Base: $best_base)" >&2
    else
        echo ""
    fi
}

# --- File Classification Helpers ---

# Description: Get list of changed (modified/untracked) files
# Usage: get_changed_files
# Params: None
# Return: List of filenames (Stdout), with directories expanded to individual files
# Exit: 0
get_changed_files() {
    # Use git status with -uall to expand directories automatically
    git status --porcelain -uall | sed 's/^...//'
}

# Description: Check if a file matches private patterns
# Usage: is_private_file "filename"
# Params:
#   $1 - Filename to check
# Return: None
# Exit: 0 if private, 1 if public
is_private_file() {
    echo "$1" | grep -qE "$PRIVATE_REGEX"
}

# Description: Get list of changed private files
# Usage: get_private_changes
# Params: None
# Return: List of private filenames (Stdout)
# Exit: 0
get_private_changes() {
    # Use grep to batch filter private files (much faster than per-file shell loop)
    get_changed_files | grep -E "$PRIVATE_REGEX" || true
}

# Description: Get list of changed public files
# Usage: get_public_changes
# Params: None
# Return: List of public filenames (Stdout)
# Exit: 0
get_public_changes() {
    # Use grep -v to batch filter public files (much faster than per-file shell loop)
    get_changed_files | grep -vE "$PRIVATE_REGEX" || true
}

# Description: Get directory size with caching
# Usage: get_dir_size "path"
get_dir_size() {
    local dir="$1"

    if [ -n "${DIR_SIZE_CACHE["$dir"]}" ]; then
        echo "${DIR_SIZE_CACHE["$dir"]}"
        return
    fi

    if [ -d "$dir" ]; then
        local size
        size=$(du -hs "$dir" 2>/dev/null | awk '{print $1}')
        DIR_SIZE_CACHE["$dir"]="${size:-0B}"
    else
        DIR_SIZE_CACHE["$dir"]="0B"
    fi

    echo "${DIR_SIZE_CACHE["$dir"]}"
}

# Description: Add a status label to a file
# Usage: add_status_label "path" "label"
add_status_label() {
    local file="$1"
    local label="$2"
    FILE_STATUS_LABELS["$file|$label"]=1
}

# Description: Build status label map from git status
# Usage: build_status_map
build_status_map() {
    FILE_STATUS_LABELS=()

    while IFS= read -r -d '' entry; do
        local status="${entry:0:2}"
        local path="${entry:3}"

        if [[ "$status" == R* || "$status" == C* ]]; then
            path="${entry#* -> }"
        fi

        if [[ "$status" == "??" ]]; then
            add_status_label "$path" "untracked"
            continue
        fi

        if [[ "$status" == "!!" ]]; then
            add_status_label "$path" "ignored"
            continue
        fi

        local x="${status:0:1}"
        local y="${status:1:1}"

        if [[ "$x" == "M" || "$y" == "M" || "$x" == "T" || "$y" == "T" ]]; then
            add_status_label "$path" "modified"
        fi
        if [[ "$x" == "A" || "$y" == "A" ]]; then
            add_status_label "$path" "added"
        fi
        if [[ "$x" == "D" || "$y" == "D" ]]; then
            add_status_label "$path" "deleted"
        fi
        if [[ "$x" == "R" || "$y" == "R" ]]; then
            add_status_label "$path" "renamed"
        fi
        if [[ "$x" == "C" || "$y" == "C" ]]; then
            add_status_label "$path" "copied"
        fi
        if [[ "$x" == "U" || "$y" == "U" ]]; then
            add_status_label "$path" "unmerged"
        fi
    done < <(git status --porcelain=v1 -z --ignored=matching -uall)
}


# Description: Check if a directory has a selected ancestor
# Usage: has_selected_ancestor "path"
has_selected_ancestor() {
    local dir="$1"

    while [[ "$dir" == */* ]]; do
        dir="${dir%/*}"  # Bash built-in: remove last path component (faster than dirname)
        if [[ -z "$dir" ]]; then
            break
        fi
        if [[ -n "${SELECTED_DIRS["$dir"]}" ]]; then
            return 0
        fi
    done

    return 1
}

# Description: Print all private/public files with folder aggregation and status labels
# Usage: print_classified_files "private"|"public"
print_classified_files() {
    local target_class="$1"
    local order_csv
    local priv_re_escaped
    local awk_output=()
    local dir_entries=()
    local file_entries=()
    local dirs_to_print=()

    SELECTED_DIRS=()
    DIR_SIZE_CACHE=()

    order_csv=$(IFS=,; echo "${STATUS_ORDER[*]}")
    priv_re_escaped="${PRIVATE_REGEX//\\/\\\\}"

    mapfile -t awk_output < <(
        gawk -v priv_re="$priv_re_escaped" -v order="$order_csv" '
            BEGIN {
                RS = "\0"
                ORS = "\n"
                OFS = "\t"
                split(order, order_arr, ",")
                order_len = length(order_arr)
            }

            function dirname(path,    n, a) {
                n = split(path, a, "/")
                if (n <= 1) {
                    return "."
                }
                return substr(path, 1, length(path) - length(a[n]) - 1)
            }

            function has_pure_ancestor(path, cls,    dir) {
                dir = dirname(path)
                while (dir != "." && dir != "/") {
                    if (dir_class_map[dir] == cls) {
                        return 1
                    }
                    dir = dirname(dir)
                }
                return 0
            }

            FNR == NR {
                entry = $0
                if (entry == "") {
                    next
                }
                status = substr(entry, 1, 2)
                path = substr(entry, 4)

                if (status ~ /^R/ || status ~ /^C/) {
                    split(entry, moved, " -> ")
                    path = moved[2]
                }

                if (status == "??") {
                    file_labels[path, "untracked"] = 1
                    next
                }

                if (status == "!!") {
                    file_labels[path, "ignored"] = 1
                    next
                }

                x = substr(status, 1, 1)
                y = substr(status, 2, 1)

                if (x == "M" || y == "M" || x == "T" || y == "T") {
                    file_labels[path, "modified"] = 1
                }
                if (x == "A" || y == "A") {
                    file_labels[path, "added"] = 1
                }
                if (x == "D" || y == "D") {
                    file_labels[path, "deleted"] = 1
                }
                if (x == "R" || y == "R") {
                    file_labels[path, "renamed"] = 1
                }
                if (x == "C" || y == "C") {
                    file_labels[path, "copied"] = 1
                }
                if (x == "U" || y == "U") {
                    file_labels[path, "unmerged"] = 1
                }

                next
            }

            {
                file = $0
                if (file == "") {
                    next
                }

                class = (file ~ priv_re) ? "private" : "public"
                file_class[file] = class

                if (!file_seen[file]++) {
                    files[++file_count] = file
                }

                label_str = ""
                label_count = 0
                for (i = 1; i <= order_len; i++) {
                    lab = order_arr[i]
                    if (file_labels[file, lab]) {
                        label_str = label_str ? label_str ", " lab : lab
                        label_count++
                    }
                }
                if (label_count == 0) {
                    label_str = "clean"
                }

                file_label_str[file] = label_str

                dir = dirname(file)
                while (dir != "." && dir != "/") {
                    dir_files[dir]++
                    if (class == "private") {
                        dir_private[dir]++
                    } else {
                        dir_public[dir]++
                    }

                    for (i = 1; i <= order_len; i++) {
                        lab = order_arr[i]
                        if (file_labels[file, lab]) {
                            dir_labels[dir, lab] = 1
                            if (lab != "clean") {
                                dir_has_nonclean[dir] = 1
                            }
                        }
                    }
                    if (label_count == 0) {
                        dir_labels[dir, "clean"] = 1
                    }
                    dir = dirname(dir)
                }
            }

            END {
                for (i = 1; i <= file_count; i++) {
                    file = files[i]
                    cls = file_class[file]
                    if (!has_pure_ancestor(file, cls)) {
                        print "FILE", cls, file, file_label_str[file]
                    }
                }

                for (dir in dir_files) {
                    priv = dir_private[dir] + 0
                    pub = dir_public[dir] + 0
                    if (priv > 0 && pub == 0) {
                        dir_class = "private"
                    } else if (pub > 0 && priv == 0) {
                        dir_class = "public"
                    } else {
                        dir_class = "mixed"
                    }
                    dir_class_map[dir] = dir_class

                    label_str = ""
                    for (i = 1; i <= order_len; i++) {
                        lab = order_arr[i]
                        if (lab == "clean" && dir_has_nonclean[dir]) {
                            continue
                        }
                        if (dir_labels[dir, lab]) {
                            label_str = label_str ? label_str ", " lab : lab
                        }
                    }
                    if (label_str == "") {
                        label_str = "clean"
                    }

                    print "DIR", dir_class, dir, dir_files[dir], label_str
                }
            }
        ' <(git status --porcelain=v1 -z -uall) <(
            {
                git ls-files -z
                git ls-files -z --others --exclude-standard
            }
        )
    )

    for entry in "${awk_output[@]}"; do
        IFS=$'\t' read -r type class path count labels <<< "$entry"
        if [ "$type" = "DIR" ]; then
            dir_entries+=("$entry")
        elif [ "$type" = "FILE" ]; then
            file_entries+=("$entry")
        fi
    done

    mapfile -t dir_entries < <(
        printf '%s\n' "${dir_entries[@]}" | awk -F'\t' '{print length($3) "\t" $0}' | sort -n | cut -f2-
    )

    # Pass 1: Collect directories that will be printed
    for entry in "${dir_entries[@]}"; do
        IFS=$'\t' read -r type class path count labels <<< "$entry"
        if [ "$class" = "$target_class" ]; then
            if ! has_selected_ancestor "$path"; then
                dirs_to_print+=("$path")
                SELECTED_DIRS["$path"]=1
            fi
        fi
    done

    # Batch du for all directories at once
    if [ ${#dirs_to_print[@]} -gt 0 ]; then
        while IFS=$'\t' read -r size dir; do
            DIR_SIZE_CACHE["$dir"]="$size"
        done < <(du -hs "${dirs_to_print[@]}" 2>/dev/null)
    fi

    # Pass 2: Print directories with cached sizes
    for entry in "${dir_entries[@]}"; do
        IFS=$'\t' read -r type class path count labels <<< "$entry"
        if [ "$class" = "$target_class" ]; then
            if [ -n "${SELECTED_DIRS["$path"]}" ]; then
                local size="${DIR_SIZE_CACHE["$path"]:-0B}"
                echo "${path}/ (${count} files, ${size}) [${labels}]"
            fi
        fi
    done

    # Build a complete set of all directories to skip (including all ancestors)
    local -A SKIP_DIRS
    for dir in "${!SELECTED_DIRS[@]}"; do
        SKIP_DIRS["$dir"]=1
    done

    for entry in "${file_entries[@]}"; do
        IFS=$'\t' read -r type class path labels <<< "$entry"
        if [ "$class" = "$target_class" ]; then
            # Fast path: check if any selected directory is a prefix of this file path
            local should_skip=0
            for dir in "${!SELECTED_DIRS[@]}"; do
                if [[ "$path" == "$dir/"* ]]; then
                    should_skip=1
                    break
                fi
            done
            if [ "$should_skip" -eq 1 ]; then
                continue
            fi
            echo "${path} [${labels}]"
        fi
    done
}

# Description: Check if a file is binary using hybrid detection strategy
# Usage: is_binary_file "filepath"
# Params:
#   $1 - File path to check
# Return: None
# Exit: 0 if binary, 1 if text
is_binary_file() {
    local file="$1"
    
    [ ! -f "$file" ] && return 1
    
    local ext="${file##*.}"
    local base="${file##*/}"
    
    # Strategy 1: Fast check by common binary extensions
    if [ "$ext" != "$base" ]; then
        ext="${ext,,}"  # lowercase
        case "$ext" in
            pdf|png|jpg|jpeg|gif|bmp|ico|svg|webp|tiff|psd|ai|eps|raw|cr2|nef|arw|\
            mp3|mp4|avi|mkv|mov|wmv|flv|wav|ogg|flac|aac|m4a|wma|\
            zip|tar|gz|bz2|xz|7z|rar|cab|iso|dmg|deb|rpm|\
            exe|dll|so|dylib|a|o|lib|obj|bin|elf|\
            dat|db|sqlite|sqlite3|mdb|accdb|\
            woff|woff2|ttf|otf|eot|\
            class|jar|pyc|pyo|elc|\
            doc|docx|xls|xlsx|ppt|pptx|odt|ods|odp)
                return 0
                ;;
        esac
    fi
    
    # Strategy 2: For truly extensionless files, check for NUL bytes in first 8KB
    if [ "$ext" = "$base" ]; then
        # No dot at all in filename (e.g., "Makefile", "README")
        if head -c 8192 "$file" 2>/dev/null | LC_ALL=C grep -q $'\x00'; then
            return 0
        fi
    fi
    
    return 1
}

# Description: Print git diff --stat style output for changed files
# Usage: print_stat_for_files "file_list" "output_file"
# Params:
#   $1 - List of files (newline separated)
#   $2 - Filename to write to if output is too large
# Return: Stat output (Stdout) or info message
# Exit: 0
print_stat_for_files() {
    local changes="$1"
    local diff_file="$2"
    
    if [ -z "$changes" ]; then
        return
    fi
    
    local stat_output=""
    local total_files=0
    local total_insertions=0
    local total_deletions=0
    local max_name_len=0
    local max_changes=0
    
    # Get terminal width for adaptive bar length (like git diff --stat)
    local term_width="${COLUMNS:-80}"
    if command -v tput &>/dev/null; then
        term_width=$(tput cols 2>/dev/null || echo 80)
    fi
    
    # Build a set of tracked files (single git call)
    declare -A tracked_files_set
    while IFS= read -r -d '' tracked_file; do
        tracked_files_set["$tracked_file"]=1
    done < <(git ls-files -z 2>/dev/null)
    
    # First pass: collect all file info and find max name length
    declare -a file_info_list=()
    
    # Separate tracked and untracked files for batch processing
    local tracked_files=""
    local untracked_files=""
    
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $changes; do
        [ -z "$file" ] && continue
        
        local name_len=${#file}
        if [ "$name_len" -gt "$max_name_len" ]; then
            max_name_len=$name_len
        fi
        
        if [ -n "${tracked_files_set["$file"]}" ]; then
            tracked_files+="$file"$'\n'
        else
            untracked_files+="$file"$'\n'
        fi
    done
    IFS="$OLD_IFS"
    
    # Process tracked files with batch git diff --numstat
    # Need to handle both staged (--cached) and unstaged changes
    if [ -n "$tracked_files" ]; then
        # Use associative array to merge staged and unstaged stats per file
        declare -A file_insertions
        declare -A file_deletions
        declare -A file_seen_in_diff
        
        # Get unstaged changes (working tree vs index)
        while IFS=$'\t' read -r insertions deletions file; do
            [ -z "$file" ] && continue
            [ "$insertions" = "-" ] && insertions=0
            [ "$deletions" = "-" ] && deletions=0
            file_insertions["$file"]=$((${file_insertions["$file"]:-0} + insertions))
            file_deletions["$file"]=$((${file_deletions["$file"]:-0} + deletions))
            file_seen_in_diff["$file"]=1
        done < <(printf "%s" "$tracked_files" | xargs -d '\n' git diff --numstat -- 2>/dev/null)
        
        # Get staged changes (index vs HEAD)
        while IFS=$'\t' read -r insertions deletions file; do
            [ -z "$file" ] && continue
            [ "$insertions" = "-" ] && insertions=0
            [ "$deletions" = "-" ] && deletions=0
            file_insertions["$file"]=$((${file_insertions["$file"]:-0} + insertions))
            file_deletions["$file"]=$((${file_deletions["$file"]:-0} + deletions))
            file_seen_in_diff["$file"]=1
        done < <(printf "%s" "$tracked_files" | xargs -d '\n' git diff --cached --numstat -- 2>/dev/null)
        
        # Build file_info_list from merged results
        for file in "${!file_seen_in_diff[@]}"; do
            local ins=${file_insertions["$file"]:-0}
            local del=${file_deletions["$file"]:-0}
            file_info_list+=("tracked|$file|$ins|$del")
            total_insertions=$((total_insertions + ins))
            total_deletions=$((total_deletions + del))
            total_files=$((total_files + 1))
        done
    fi
    
    # Process untracked files: use extension-based binary detection (fast)
    # Then batch wc -l for text files
    local text_files=""
    local binary_files=""
    
    # Build list of known binary extensions for fast grep-based detection
    local binary_ext_pattern='\.(pdf|png|jpg|jpeg|gif|bmp|ico|svg|webp|tiff|psd|ai|eps|raw|cr2|nef|arw|mp3|mp4|avi|mkv|mov|wmv|flv|wav|ogg|flac|aac|m4a|wma|zip|tar|gz|bz2|xz|7z|rar|cab|iso|dmg|deb|rpm|exe|dll|so|dylib|a|o|lib|obj|bin|elf|dat|db|sqlite|sqlite3|mdb|accdb|woff|woff2|ttf|otf|eot|class|jar|pyc|pyo|elc|doc|docx|xls|xlsx|ppt|pptx|odt|ods|odp)$'
    
    # Fast extension-based classification
    binary_files=$(echo "$untracked_files" | grep -iE "$binary_ext_pattern" || true)
    text_files=$(echo "$untracked_files" | grep -ivE "$binary_ext_pattern" || true)
    
    # Process binary files with batch stat
    if [ -n "$binary_files" ]; then
        while IFS=$'\t' read -r size file; do
            [ -z "$file" ] && continue
            file_info_list+=("binary|$file|$size|0")
            total_files=$((total_files + 1))
        done < <(echo "$binary_files" | xargs -d '\n' stat -c '%s	%n' 2>/dev/null || true)
    fi
    
    # Process text files with batch wc -l (using awk for fast parsing)
    if [ -n "$text_files" ]; then
        # Use a temporary file to avoid subshell variable scope issues
        local wc_tmp=$(mktemp)
        echo "$text_files" | xargs -d '\n' wc -l 2>/dev/null > "$wc_tmp" || true
        
        # Parse wc output with awk (much faster than shell while-read loop)
        # wc -l output format: "  123 filename" or "123 filename"
        # Last line is "total" when multiple files - skip it
        while IFS='|' read -r lines file; do
            [ -z "$file" ] && continue
            file_info_list+=("untracked|$file|$lines|0")
            total_insertions=$((total_insertions + lines))
            total_files=$((total_files + 1))
        done < <(awk '!/^ *[0-9]+ +total$/ && NF >= 2 {
            lines = $1
            # Reconstruct filename (handles spaces in names)
            $1 = ""
            file = $0
            sub(/^ +/, "", file)
            print lines "|" file
        }' "$wc_tmp")
        
        rm -f "$wc_tmp"
    fi
    
    # Second pass: format output with awk (much faster than shell loop)
    # Write file_info_list to temp file for awk processing
    local format_tmp=$(mktemp)
    printf '%s\n' "${file_info_list[@]}" > "$format_tmp"
    
    stat_output=$(awk -F'|' -v max_len="$max_name_len" '
    BEGIN {
        # Pre-generate + and - characters for bar construction
        for (i = 1; i <= 50; i++) {
            plus_chars = plus_chars "+"
            minus_chars = minus_chars "-"
        }
    }
    {
        status = $1
        file = $2
        value1 = $3 + 0
        value2 = $4 + 0
        
        if (status == "binary") {
            # Binary file: show size in human-readable format
            size_bytes = value1
            if (size_bytes >= 1048576) {
                size_human = int(size_bytes / 1048576) "M"
            } else if (size_bytes >= 1024) {
                size_human = int(size_bytes / 1024) "K"
            } else {
                size_human = size_bytes "B"
            }
            printf " %-" max_len "s | Bin 0 -> %s (untracked)\n", file, size_human
        } else {
            insertions = value1
            deletions = value2
            total_changes = insertions + deletions
            bar_len = (total_changes > 50) ? 50 : total_changes
            plus_len = (total_changes > 0) ? int(bar_len * insertions / total_changes) : 0
            minus_len = bar_len - plus_len
            
            bar = substr(plus_chars, 1, plus_len) substr(minus_chars, 1, minus_len)
            
            suffix = ""
            if (status == "untracked") {
                suffix = " (untracked)"
            }
            
            printf " %-" max_len "s | %4d %s%s\n", file, total_changes, bar, suffix
        }
    }
    ' "$format_tmp")
    
    rm -f "$format_tmp"
    
    # Add summary line
    if [ "$total_files" -gt 0 ]; then
        local summary=""
        summary=$(printf " %d file%s changed" "$total_files" "$([ $total_files -gt 1 ] && echo 's')")
        if [ "$total_insertions" -gt 0 ]; then
            summary+=", $total_insertions insertion"
            [ "$total_insertions" -gt 1 ] && summary+="s"
            summary+="(+)"
        fi
        if [ "$total_deletions" -gt 0 ]; then
            summary+=", $total_deletions deletion"
            [ "$total_deletions" -gt 1 ] && summary+="s"
            summary+="(-)"
        fi
        # Ensure stat_output ends with newline before appending summary
        if [ -n "$stat_output" ] && [[ "$stat_output" != *$'\n' ]]; then
            stat_output+=$'\n'
        fi
        stat_output+="$summary"$'\n'
    fi
    
    # Line limit check
    local line_count
    line_count=$(echo "$stat_output" | wc -l)
    
    if [ "$line_count" -gt "$DIFF_LINE_LIMIT" ]; then
        echo "$stat_output" > "$diff_file"
        log_warn "Status output is too large ($line_count lines)."
        log_info "Status content has been written to: $diff_file"
    else
        echo "$stat_output"
    fi
}

# Description: Generate stat output string (without line limit check)
# Usage: generate_stat_output "file_list"
# Params:
#   $1 - List of files (newline separated)
# Return: Stat output string (Stdout)
# Exit: 0
generate_stat_output() {
    local changes="$1"
    
    if [ -z "$changes" ]; then
        return
    fi
    
    local stat_output=""
    local total_files=0
    local total_insertions=0
    local total_deletions=0
    local max_name_len=0
    
    # Build a set of tracked files (single git call)
    declare -A tracked_files_set
    while IFS= read -r -d '' tracked_file; do
        tracked_files_set["$tracked_file"]=1
    done < <(git ls-files -z 2>/dev/null)
    
    # First pass: collect all file info and find max name length
    declare -a file_info_list=()
    
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $changes; do
        [ -z "$file" ] && continue
        
        local name_len=${#file}
        if [ "$name_len" -gt "$max_name_len" ]; then
            max_name_len=$name_len
        fi
        
        if [ -n "${tracked_files_set["$file"]}" ]; then
            # Tracked file: get diff stats
            local numstat
            numstat=$(git diff --numstat -- "$file" 2>/dev/null | head -1)
            if [ -n "$numstat" ]; then
                local insertions deletions
                insertions=$(echo "$numstat" | awk '{print $1}')
                deletions=$(echo "$numstat" | awk '{print $2}')
                [ "$insertions" = "-" ] && insertions=0
                [ "$deletions" = "-" ] && deletions=0
                file_info_list+=("tracked|$file|$insertions|$deletions")
                total_insertions=$((total_insertions + insertions))
                total_deletions=$((total_deletions + deletions))
                total_files=$((total_files + 1))
            fi
        else
            # Untracked file: detect binary using is_binary_file()
            if [ -f "$file" ]; then
                if is_binary_file "$file"; then
                    local size
                    size=$(stat -c%s "$file" 2>/dev/null || echo "0")
                    file_info_list+=("binary|$file|$size|0")
                    total_files=$((total_files + 1))
                else
                    local lines
                    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
                    lines=$(echo "$lines" | tr -d ' ')
                    file_info_list+=("untracked|$file|$lines|0")
                    total_insertions=$((total_insertions + lines))
                    total_files=$((total_files + 1))
                fi
            fi
        fi
    done
    IFS="$OLD_IFS"
    
    # Second pass: format output with aligned columns
    for info in "${file_info_list[@]}"; do
        IFS='|' read -r status file value1 value2 <<< "$info"
        
        if [ "$status" = "binary" ]; then
            local size_bytes=$value1
            local size_human
            if [ "$size_bytes" -ge 1048576 ]; then
                size_human="$((size_bytes / 1048576))M"
            elif [ "$size_bytes" -ge 1024 ]; then
                size_human="$((size_bytes / 1024))K"
            else
                size_human="${size_bytes}B"
            fi
            stat_output+=" $(printf "%-${max_name_len}s" "$file") | Bin 0 -> $size_human (untracked)"$'\n'
        else
            local insertions=$value1
            local deletions=$value2
            local total_changes=$((insertions + deletions))
            local bar_len=$((total_changes > 50 ? 50 : total_changes))
            local plus_len=$((bar_len * insertions / (total_changes > 0 ? total_changes : 1)))
            local minus_len=$((bar_len - plus_len))
            
            local bar=""
            for ((i=0; i<plus_len; i++)); do bar+="+"; done
            for ((i=0; i<minus_len; i++)); do bar+="-"; done
            
            local suffix=""
            if [ "$status" = "untracked" ]; then
                suffix=" (untracked)"
            fi
            
            stat_output+=" $(printf "%-${max_name_len}s" "$file") | $(printf "%4d" "$total_changes") ${bar}${suffix}"$'\n'
        fi
    done
    
    # Add summary line
    if [ "$total_files" -gt 0 ]; then
        local summary=""
        summary=$(printf " %d file%s changed" "$total_files" "$([ $total_files -gt 1 ] && echo 's')")
        if [ "$total_insertions" -gt 0 ]; then
            summary+=", $total_insertions insertion"
            [ "$total_insertions" -gt 1 ] && summary+="s"
            summary+="(+)"
        fi
        if [ "$total_deletions" -gt 0 ]; then
            summary+=", $total_deletions deletion"
            [ "$total_deletions" -gt 1 ] && summary+="s"
            summary+="(-)"
        fi
        stat_output+="$summary"$'\n'
    fi
    
    echo "$stat_output"
}

# Description: Generate diff output string (without line limit check)
# Usage: generate_diff_output "file_list"
# Params:
#   $1 - List of files to diff
# Return: Diff content string (Stdout)
# Exit: 0
generate_diff_output() {
    local changes="$1"
    
    if [ -z "$changes" ]; then
        return
    fi
    
    local raw_diff=""
    local tracked_files=""
    local untracked_text_files=""
    local untracked_binary_files=""
    
    # Classify files into tracked and untracked (text vs binary)
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $changes; do
        [ -z "$file" ] && continue
        if git ls-files --error-unmatch "$file" &>/dev/null; then
            tracked_files+="$file"$'\n'
        else
            if is_binary_file "$file"; then
                untracked_binary_files+="$file"$'\n'
            else
                untracked_text_files+="$file"$'\n'
            fi
        fi
    done
    IFS="$OLD_IFS"
    
    # Tracked files diff
    if [ -n "$tracked_files" ]; then
        local tracked_diff
        tracked_diff=$(echo "$tracked_files" | xargs git diff --color=never -- 2>/dev/null)
        if [ -n "$tracked_diff" ]; then
            raw_diff+="$tracked_diff"$'\n'
        fi
    fi
    
    # Untracked text files: use git diff --no-index
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $untracked_text_files; do
        [ -z "$file" ] && continue
        if [ -f "$file" ]; then
            local file_diff
            file_diff=$(git diff --no-index --color=never /dev/null "$file" 2>/dev/null || true)
            if [ -n "$file_diff" ]; then
                raw_diff+="$file_diff"$'\n'
            fi
        fi
    done
    IFS="$OLD_IFS"
    
    # Untracked binary files: show Git-style message
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $untracked_binary_files; do
        [ -z "$file" ] && continue
        if [ -f "$file" ]; then
            raw_diff+="diff --git a/dev/null b/$file"$'\n'
            raw_diff+="new file mode 100644"$'\n'
            raw_diff+="Binary files /dev/null and $file differ"$'\n'
        fi
    done
    IFS="$OLD_IFS"
    
    echo "$raw_diff"
}

# Description: Print combined stat + diff output (like git show)
# Usage: print_show_for_files "file_list" "output_file"
# Params:
#   $1 - List of files (newline separated)
#   $2 - Filename to write to if output is too large
# Return: Combined output (Stdout) or info message
# Exit: 0
print_show_for_files() {
    local changes="$1"
    local diff_file="$2"
    
    if [ -z "$changes" ]; then
        return
    fi
    
    # Generate both parts
    local stat_output
    local diff_output
    stat_output=$(generate_stat_output "$changes")
    diff_output=$(generate_diff_output "$changes")
    
    # Combine: stat + separator + diff
    local combined_output=""
    if [ -n "$stat_output" ]; then
        combined_output+="$stat_output"
        combined_output+=$'\n'
    fi
    if [ -n "$diff_output" ]; then
        if [ -n "$stat_output" ]; then
            combined_output+=$'\n'$'\n'
        fi
        combined_output+="$diff_output"
    fi
    
    # Check total line count
    local line_count
    line_count=$(echo "$combined_output" | wc -l)
    
    if [ "$line_count" -gt "$DIFF_LINE_LIMIT" ]; then
        echo "$combined_output" > "$diff_file"
        log_warn "Output is too large ($line_count lines)."
        log_info "Content has been written to: $diff_file"
    else
        # Re-output with colors for diff part
        if [ -n "$stat_output" ]; then
            echo "$stat_output"
        fi
        
        if [ -n "$diff_output" ] && [ -n "$stat_output" ]; then
            echo ""
            echo ""
        fi

        # Re-generate diff with color
        if [ -n "$changes" ]; then
            local tracked_files=""
            local untracked_text_files=""
            local untracked_binary_files=""
            
            OLD_IFS="$IFS"
            IFS=$'\n'
            for file in $changes; do
                [ -z "$file" ] && continue
                if git ls-files --error-unmatch "$file" &>/dev/null; then
                    tracked_files+="$file"$'\n'
                else
                    if is_binary_file "$file"; then
                        untracked_binary_files+="$file"$'\n'
                    else
                        untracked_text_files+="$file"$'\n'
                    fi
                fi
            done
            IFS="$OLD_IFS"
            
            if [ -n "$tracked_files" ]; then
                echo "$tracked_files" | xargs git diff --color=always -- 2>/dev/null
            fi
            
            OLD_IFS="$IFS"
            IFS=$'\n'
            for file in $untracked_text_files; do
                [ -z "$file" ] && continue
                if [ -f "$file" ]; then
                    git diff --no-index --color=always /dev/null "$file" 2>/dev/null || true
                fi
            done
            for file in $untracked_binary_files; do
                [ -z "$file" ] && continue
                if [ -f "$file" ]; then
                    echo "diff --git a/dev/null b/$file"
                    echo "new file mode 100644"
                    echo "Binary files /dev/null and $file differ"
                fi
            done
            IFS="$OLD_IFS"
        fi
    fi
}

# Description: Print git diff, writing to file if too large
# Usage: print_diff_safe "file_list" "output_file"
# Params:
#   $1 - List of files to diff
#   $2 - Filename to write to if diff is huge
# Return: Diff content (Stdout) or info message
# Exit: 0
print_diff_safe() {
    local changes="$1"
    local diff_file="$2"
    
    if [ -z "$changes" ]; then
        return
    fi
    
    local raw_diff=""
    local tracked_files=""
    local untracked_text_files=""
    local untracked_binary_files=""
    
    # Classify files into tracked and untracked (text vs binary)
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $changes; do
        [ -z "$file" ] && continue
        if git ls-files --error-unmatch "$file" &>/dev/null; then
            tracked_files+="$file"$'\n'
        else
            # Check if untracked file is binary
            if is_binary_file "$file"; then
                untracked_binary_files+="$file"$'\n'
            else
                untracked_text_files+="$file"$'\n'
            fi
        fi
    done
    IFS="$OLD_IFS"
    
    # Tracked files diff (git handles binary files automatically)
    if [ -n "$tracked_files" ]; then
        local tracked_diff
        tracked_diff=$(echo "$tracked_files" | xargs git diff --color=never -- 2>/dev/null)
        if [ -n "$tracked_diff" ]; then
            raw_diff+="$tracked_diff"$'\n'
        fi
    fi
    
    # Untracked text files: use git diff --no-index
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $untracked_text_files; do
        [ -z "$file" ] && continue
        if [ -f "$file" ]; then
            # git diff --no-index returns non-zero exit code which is normal
            local file_diff
            file_diff=$(git diff --no-index --color=never /dev/null "$file" 2>/dev/null || true)
            if [ -n "$file_diff" ]; then
                raw_diff+="$file_diff"$'\n'
            fi
        fi
    done
    IFS="$OLD_IFS"
    
    # Untracked binary files: show Git-style "Binary files differ" message
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $untracked_binary_files; do
        [ -z "$file" ] && continue
        if [ -f "$file" ]; then
            raw_diff+="diff --git a/dev/null b/$file"$'\n'
            raw_diff+="new file mode 100644"$'\n'
            raw_diff+="Binary files /dev/null and $file differ"$'\n'
        fi
    done
    IFS="$OLD_IFS"
    
    local line_count
    line_count=$(echo "$raw_diff" | wc -l)
    
    if [ "$line_count" -gt "$DIFF_LINE_LIMIT" ]; then
        echo "$raw_diff" > "$diff_file"
        log_warn "Diff output is too large ($line_count lines)."
        log_info "Diff content has been written to: $diff_file"
    else
        # Re-generate with color output
        if [ -n "$tracked_files" ]; then
            echo "$tracked_files" | xargs git diff --color=always -- 2>/dev/null
        fi
        OLD_IFS="$IFS"
        IFS=$'\n'
        for file in $untracked_text_files; do
            [ -z "$file" ] && continue
            if [ -f "$file" ]; then
                git diff --no-index --color=always /dev/null "$file" 2>/dev/null || true
            fi
        done
        # Binary files: print message (no color needed)
        for file in $untracked_binary_files; do
            [ -z "$file" ] && continue
            if [ -f "$file" ]; then
                echo "diff --git a/dev/null b/$file"
                echo "new file mode 100644"
                echo "Binary files /dev/null and $file differ"
            fi
        done
        IFS="$OLD_IFS"
    fi
}

# --- Restore Private Files ---

# Description: Restore private files from a specified source or auto-detect from parent history.
#              Uses git ls-tree + grep to find matching files, then restores them individually.
#              Excludes this script itself during restore (to avoid overwriting while running),
#              then restores the script at the end.
# Usage: restore_private_files [source_ref]
# Params:
#   $1 - (Optional) Source ref (branch name or commit hash). If empty, tries auto-detection.
# Return: None
# Exit: 0 on success, 1 on failure
restore_private_files() {
    local source_ref="$1"
    
    # Get this script's path relative to git root
    local script_path
    script_path=$(get_script_rel_path)
    
    # 1. Auto-detection if source_ref is missing
    if [ -z "$source_ref" ]; then
        log_info "No source reference provided. Attempting to auto-detect from parent branch history..."
        
        # Reuse find_best_parent logic to find the public parent we are based on.
        # Note: find_best_parent prints info to stderr, hash to stdout.
        # We need a remote. Assuming 'origin' is primary or using first available.
        local primary_remote="origin"
        if ! git remote | grep -q "^origin$"; then
            primary_remote=$(git remote | head -n 1)
        fi
        
        # Get the commit hash of the base
        local parent_base_hash
        parent_base_hash=$(find_best_parent "$primary_remote" "$CURRENT_BRANCH")
        
        if [ -z "$parent_base_hash" ]; then
            log_error "Could not detect a public parent branch to infer source."
            log_error "Please specify source branch manually: --restore-private <source>"
            exit 1
        fi
        
        # Search for cherry-pick signature in history starting from parent base
        # We look back 50 commits from the base
        local cherry_pick_line
        cherry_pick_line=$(git log -n 50 --grep="cherry picked from commit" --format="%b" "$parent_base_hash" | grep "cherry picked from commit" | head -n 1)
        
        if [ -n "$cherry_pick_line" ]; then
            source_ref=$(echo "$cherry_pick_line" | sed -E 's/.*cherry picked from commit ([0-9a-f]+).*/\1/')
            log_info "Auto-detected source commit: $source_ref (found in parent history)"
        else
            log_error "Could not find any cherry-pick traces in parent history."
            log_error "Please specify source branch manually: --restore-private <source>"
            exit 1
        fi
    fi

    # 2. Get list of ALL files in source commit, then filter by PRIVATE_PATTERNS
    log_info "Scanning private files in '$source_ref'..."
    local private_files
    private_files=$(git ls-tree -r --name-only "$source_ref" 2>/dev/null | grep -E "$PRIVATE_REGEX" || true)
    
    if [ -z "$private_files" ]; then
        log_warn "No private files found in source '$source_ref'."
        return 0
    fi
    
    local file_count
    file_count=$(echo "$private_files" | wc -l)
    log_info "Found $file_count private files to restore."
    
    # 3. Restore each matched file (excluding this script during restore)
    local restored_count=0
    local script_found=0
    
    while IFS= read -r file; do
        [ -z "$file" ] && continue
        
        # Skip this script, restore it last
        if [ "$file" = "$script_path" ]; then
            script_found=1
            continue
        fi
        
        # Create parent directory if needed
        local file_dir
        file_dir=$(dirname "$file")
        if [ "$file_dir" != "." ] && [ ! -d "$file_dir" ]; then
            mkdir -p "$file_dir"
        fi
        
        # Restore the file
        if git checkout "$source_ref" -- "$file" 2>/dev/null; then
            restored_count=$((restored_count + 1))
        else
            log_warn "Failed to restore: $file"
        fi
    done <<< "$private_files"
    
    # 4. Finally restore this script itself (if it was found in source)
    if [ "$script_found" -eq 1 ]; then
        local script_dir
        script_dir=$(dirname "$script_path")
        if [ "$script_dir" != "." ] && [ ! -d "$script_dir" ]; then
            mkdir -p "$script_dir"
        fi
        if git checkout "$source_ref" -- "$script_path" 2>/dev/null; then
            restored_count=$((restored_count + 1))
            log_info "Restored $script_path (self)."
        fi
    fi
    
    log_info "Successfully restored $restored_count private files."
}

# --- Visibility Detection ---

# Description: Detect if a remote repository is Public, Private, or Unreachable
# Usage: detect_repo_visibility "origin"
# Params:
#   $1 - Remote name
# Return: "public", "private", or "unreachable" (Stdout)
# Exit: 0
detect_repo_visibility() {
    local remote_name="$1"
    
    local remote_url=$(git remote get-url "$remote_name" 2>/dev/null || echo "")
    
    if [ -z "$remote_url" ]; then
        echo "public"
        return
    fi
    
    local http_url="$remote_url"
    
    if [[ "$http_url" =~ ^git@ ]]; then
        http_url=$(echo "$http_url" | sed -E 's|^git@([^:]+):|https://\1/|')
    elif [[ "$http_url" =~ ^ssh:// ]]; then
        http_url=$(echo "$http_url" | sed -E 's|^ssh://[^@]+@|https://|')
    fi
    
    http_url=${http_url%.git}
    
    # Check for git proxy configuration
    local git_proxy
    git_proxy=$(git config --get http.proxy || echo "")
    if [ -z "$git_proxy" ]; then
        git_proxy=$(git config --get https.proxy || echo "")
    fi
    
    # Strip potential carriage return from git config output
    git_proxy=$(echo "$git_proxy" | tr -d '\r')

    local curl_opts=""
    if [ -n "$git_proxy" ]; then
        export http_proxy="$git_proxy"
        export https_proxy="$git_proxy"
        # User suggested '-k' if proxy is used
        curl_opts="-k"
    fi

    # Check with curl (5s timeout)
    # "000" indicates connection failure.
    # shellcheck disable=SC2086
    local status_code=$(curl -I -s -o /dev/null -w "%{http_code}" --max-time 5 $curl_opts "$http_url")
    
    [ -n "$status_code" ] || status_code="000"

    case "$status_code" in
        000)
            echo "unreachable"
            ;;
        200|301)
            echo "public"
            ;;
        404|403|401|302)
            echo "private"
            ;;
        *)
            echo "public"
            ;;
    esac
}

# ==========================================
# 3. Main Logic
# ==========================================

# 3.1 Early Exit for Help and Init (Bypass Git Root Check)
if [[ "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# 3.1.1 Handle --init (Bypass Git Root Check - creates new repo)
if [[ "$1" == "--init" ]]; then
    init_repository
    exit 0
fi

# 3.2 Ensure Git Root (Strict Check)
ensure_git_root

# 3.3 Branch Detection
CURRENT_BRANCH=$(git branch --show-current)
BRANCH_PRIVATE=""
BRANCH_PUBLIC=""

if [ "$CURRENT_BRANCH" = "develop" ]; then
    BRANCH_PRIVATE="develop"
    # Dynamic detection of main branch name (main or master)
    BRANCH_PUBLIC=$(get_main_branch_name)
elif [[ "$CURRENT_BRANCH" == dev/* ]]; then
    BRANCH_PRIVATE="$CURRENT_BRANCH"
    SUFFIX="${CURRENT_BRANCH#dev/}"
    BRANCH_PUBLIC="feature/$SUFFIX"
else
    # Fail early for action commands if on invalid branch
    if [[ "$1" != "--print"* && "$1" != "--restore-private" && "$1" != "--new-dev-branch" ]]; then
        log_error "Current branch '$CURRENT_BRANCH' is not a valid development branch."
        log_error "Must be 'develop' or start with 'dev/'."
        exit 1
    fi
fi

# 3.4 Sanity Check (Private Base Detection)
# Only verify base for private branches (develop or dev/*)
# Skip check for read-only commands
if [[ "$1" != "--print"* && "$1" != "--help" && -n "$BRANCH_PRIVATE" ]]; then
    check_private_base
fi

# 3.5 Command Processing
cmd="$1"

case "$cmd" in
    --print-private-patterns)
        for pattern in "${PRIVATE_PATTERNS[@]}"; do
            echo "$pattern"
        done
        ;;
        
    --print-public-patterns)
        echo "Everything NOT matching private patterns:"
        for pattern in "${PRIVATE_PATTERNS[@]}"; do
            echo "  ! $pattern"
        done
        ;;
        
    --print-private-files)
        print_classified_files "private"
        ;;
        
    --print-public-files)
        print_classified_files "public"
        ;;
        
    --print-private-status)
        changes=$(get_private_changes)
        if [ -n "$changes" ]; then
            print_stat_for_files "$changes" "$PRIVATE_DIFF_FILE"
        fi
        ;;
        
    --print-public-status)
        changes=$(get_public_changes)
        if [ -n "$changes" ]; then
            print_stat_for_files "$changes" "$PUBLIC_DIFF_FILE"
        fi
        ;;

    --print-private-diff)
        changes=$(get_private_changes)
        if [ -n "$changes" ]; then
             print_diff_safe "$changes" "$PRIVATE_DIFF_FILE"
        fi
        ;;

    --print-public-diff)
        changes=$(get_public_changes)
        if [ -n "$changes" ]; then
             print_diff_safe "$changes" "$PUBLIC_DIFF_FILE"
        fi
        ;;

    --print-private-show)
        changes=$(get_private_changes)
        if [ -n "$changes" ]; then
            print_show_for_files "$changes" "$PRIVATE_DIFF_FILE"
        fi
        ;;

    --print-public-show)
        changes=$(get_public_changes)
        if [ -n "$changes" ]; then
            print_show_for_files "$changes" "$PUBLIC_DIFF_FILE"
        fi
        ;;
        
    --check-prerequisites)
        check_prerequisites
        ;;
        
    --new-dev-branch)
        create_new_dev_branch "$2" "$3"
        ;;
    
    --restore-private)
        restore_private_files "$2"
        ;;
        
    --commit-private)
        # Check for changes FIRST - if no changes, exit cleanly without requiring env var
        changes=$(get_private_changes)
        if [ -z "$changes" ]; then
            log_info "No private changes to commit."
            exit 0
        fi
        
        # Only require commit message if there are actual changes
        msg="${COMMIT_MSG_PRIVATE}"
        if [ -z "$msg" ]; then
            log_error "Missing commit message for Private changes."
            echo ""
            echo "Please set the 'COMMIT_MSG_PRIVATE' environment variable with a descriptive message before running this command."
            echo "Example: COMMIT_MSG_PRIVATE='Your message' $0 --commit-private"
            exit 1
        fi
        
        log_info "Committing Private changes..."
        echo "$changes" | xargs git add
        git commit -m "$msg"
        ;;
        
    --commit-public)
        # Check for changes FIRST - if no changes, exit cleanly without requiring env var
        changes=$(get_public_changes)
        if [ -z "$changes" ]; then
            log_info "No public changes to commit."
            exit 0
        fi
        
        # Only require commit message if there are actual changes
        msg="${COMMIT_MSG_PUBLIC}"
        if [ -z "$msg" ]; then
            log_error "Missing commit message for Public changes."
            echo ""
            echo "Please set the 'COMMIT_MSG_PUBLIC' environment variable with a descriptive message before running this command."
            echo "Example: COMMIT_MSG_PUBLIC='Your message' $0 --commit-public"
            exit 1
        fi
        
        log_info "Committing Public changes..."
        echo "$changes" | xargs git add
        git commit -m "$msg"
        ;;
        
    --push-private)
        remotes=$(git remote)
        if [ -z "$remotes" ]; then
            log_error "No remotes defined."
            exit 1
        fi

        for remote in $remotes; do
            log_info "Checking remote: $remote..."
            visibility=$(detect_repo_visibility "$remote")
            
            if [ "$visibility" = "private" ]; then
                log_info "Pushing $BRANCH_PRIVATE to $remote/$BRANCH_PRIVATE..."
                git push -u "$remote" "$BRANCH_PRIVATE"
            elif [ "$visibility" = "unreachable" ]; then
                log_warn "Remote '$remote' is unreachable. Skipping push."
            else
                log_info "Skipping push to $remote/$BRANCH_PRIVATE (Public Repo detected)."
            fi
        done
        
        log_info "Cleaning up temporary diff files..."
        rm -f "$PRIVATE_DIFF_FILE" "$PUBLIC_DIFF_FILE"
        ;;
        
    --push-public)
        log_info "Syncing public changes from $BRANCH_PRIVATE to $BRANCH_PUBLIC..."
        
        # Ensure we are on the expected private branch
        if [ "$CURRENT_BRANCH" != "$BRANCH_PRIVATE" ]; then
            log_error "Unexpected branch mismatch. Expected $BRANCH_PRIVATE."
            exit 1
        fi

        # Determine Primary Remote for fetching 'main' or public branch base
        primary_remote="origin"
        if ! git remote | grep -q "^origin$"; then
            primary_remote=$(git remote | head -n 1)
        fi
        
        if [ -z "$primary_remote" ]; then
            log_error "No remotes found. Cannot sync."
            exit 1
        fi
        
        log_info "Using primary remote: $primary_remote"

        # Setup cleanup trap
        trap cleanup EXIT
        
        # Ensure we have the latest info about public branch or main
        log_info "Fetching $primary_remote..."
        git fetch "$primary_remote" || log_warn "Fetch failed. Proceeding with local cache..."

        # Check if there are no remote PUBLIC branches (only private branches like develop/dev/* exist)
        # This handles the case of a new repository where only 'develop' has been pushed
        remote_public_branch_count=$(git branch -r --list "$primary_remote/*" | grep -vE "$primary_remote/dev/|$primary_remote/develop$|$primary_remote/dev$|$primary_remote/HEAD" | wc -l)
        if [ "$remote_public_branch_count" -eq 0 ]; then
            log_info "No remote public branches detected (only private branches exist on remote)."
            
            # Check if local public branch already exists (e.g., created by --init)
            if git show-ref --verify --quiet "refs/heads/$BRANCH_PUBLIC"; then
                log_info "Local public branch '$BRANCH_PUBLIC' already exists. Switching to it..."
                git checkout "$BRANCH_PUBLIC"
            else
                log_info "Creating orphan branch '$BRANCH_PUBLIC' for public-only content..."
                
                # Create orphan branch (no parent commit)
                git checkout --orphan "$BRANCH_PUBLIC"
                # Remove all files from index (orphan branch starts with staged files from previous HEAD)
                git rm -rf --cached . >/dev/null 2>&1 || true
                # Clean untracked files that might conflict
                git clean -fd >/dev/null 2>&1 || true
                # Create initial empty commit
                git commit --allow-empty -m "Initial empty commit for public branch"
                
                log_info "Orphan branch '$BRANCH_PUBLIC' created with empty initial commit."
            fi
        else
            # Setup Public Branch Locally (normal case: remote has public branches)
            # 1. Check if local public branch exists
            if git show-ref --verify --quiet "refs/heads/$BRANCH_PUBLIC"; then
                log_info "Public branch '$BRANCH_PUBLIC' exists locally."
            else
                log_info "Public branch '$BRANCH_PUBLIC' does not exist locally."
                
                # 2. Check if remote public branch exists
                if git show-ref --verify --quiet "refs/remotes/$primary_remote/$BRANCH_PUBLIC"; then
                    log_info "Tracking remote branch '$primary_remote/$BRANCH_PUBLIC'..."
                    git branch --track "$BRANCH_PUBLIC" "$primary_remote/$BRANCH_PUBLIC"
                else
                    log_info "Remote public branch not found. Detecting best parent..."
                    
                    # 3. Smart Parent Detection
                    best_base_hash=$(find_best_parent "$primary_remote" "$BRANCH_PRIVATE")
                    
                    if [ -n "$best_base_hash" ]; then
                        log_info "Creating '$BRANCH_PUBLIC' based on detected parent hash: $best_base_hash..."
                        git branch "$BRANCH_PUBLIC" "$best_base_hash"
                    else
                        # Fallback to main/master if detection fails
                        fallback_branch=$(get_main_branch_name "$primary_remote")
                        log_warn "No suitable parent detected. Fallback to '$primary_remote/$fallback_branch'..."
                        
                        if git show-ref --verify --quiet "refs/remotes/$primary_remote/$fallback_branch"; then
                             git branch "$BRANCH_PUBLIC" "$primary_remote/$fallback_branch"
                        else
                             log_error "Cannot find '$primary_remote/$fallback_branch' to base new feature branch on."
                             exit 1
                        fi
                    fi
                fi
            fi

            # Switch to Public Branch
            git checkout "$BRANCH_PUBLIC"
        fi

        # Pull latest changes if tracked
        if git rev-parse --verify @{u} >/dev/null 2>&1; then
            log_info "Pulling latest changes for $BRANCH_PUBLIC..."
            git pull --rebase || log_warn "Pull failed or no upstream."
        fi

        # Find Sync Range
        log_info "Finding last synced commit..."
        last_synced_line=$(git log -n 100 --grep="(cherry picked from commit" --format="%b" | grep "(cherry picked from commit" | head -n 1)
        
        if [ -n "$last_synced_line" ]; then
            last_synced_hash=$(echo "$last_synced_line" | sed -E 's/.*cherry picked from commit ([0-9a-f]+).*/\1/')
            range="${last_synced_hash}..$BRANCH_PRIVATE"
            log_info "Syncing from last anchor: $last_synced_hash"
            commits_to_sync=$(git log "$range" --reverse --format="%H" --no-merges)
        else
            log_info "No sync anchor found. Syncing all divergent commits..."
            commits_to_sync=$(git log "$BRANCH_PUBLIC".."$BRANCH_PRIVATE" --reverse --format="%H" --no-merges)
        fi
        
        if [ -n "$commits_to_sync" ]; then
            for commit in $commits_to_sync; do
                if git log -n 100 --grep="(cherry picked from commit $commit)" --format="%H" | grep -q .; then
                    continue
                fi

                # Check private files
                files=$(git show --name-only --format= "$commit")
                is_private=0
                
                OLD_IFS="$IFS"
                IFS=$'\n'
                for file in $files; do
                    if is_private_file "$file"; then
                        is_private=1
                        break
                    fi
                done
                IFS="$OLD_IFS"
                
                if [ "$is_private" -eq 1 ]; then
                    log_info "Skipping private commit: $commit"
                    continue
                fi
                
                log_info "Cherry-picking: $commit"
                if ! git cherry-pick -x --allow-empty --keep-redundant-commits "$commit"; then
                    log_error "Cherry-pick failed. Please resolve conflicts manually then continue."
                    exit 1
                fi
            done
        else
            log_info "No commits to sync."
        fi
        
        # Multi-Remote Push Logic
        remotes=$(git remote)
        for remote in $remotes; do
            echo ""
            log_info "--- Processing remote: $remote ---"
            
            # Connectivity Check
            visibility=$(detect_repo_visibility "$remote")
            if [ "$visibility" = "unreachable" ]; then
                log_warn "Remote '$remote' is unreachable. Skipping push to avoid hanging."
                continue
            fi

            log_info "Pushing $BRANCH_PUBLIC to $remote..."
            git push -u "$remote" "$BRANCH_PUBLIC"
        done
        
        rm -f "$PRIVATE_DIFF_FILE" "$PUBLIC_DIFF_FILE"
        echo ""
        log_info "Sync and Push completed successfully."
        ;;
        
    *)
        show_help
        exit 1
        ;;
esac
