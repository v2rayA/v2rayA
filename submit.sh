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
PRIVATE_DIFF_FILE="private_diff.txt"
PUBLIC_DIFF_FILE="public_diff.txt"

# Define Private Patterns List
# Add any file or directory path that should stay PRIVATE on develop branch
# These are regex patterns anchored to the root
PRIVATE_PATTERNS=(
    "^memory-bank/"
    "^doc/"
    "^\.clinerules"
    "^\.claude"
    "^\.vscode"
    "^\.clineignore"
    "^AGENTS\.md"
    "^openspec"
    "^submit\.sh"
    "^\.continue"
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
        log_error "Not a git repository. Please run this script inside a git repository."
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
    parent_commit=$(git rev-parse HEAD^ 2>/dev/null || echo "")
    
    if [ -z "$parent_commit" ]; then
        # Initial commit or orphan branch, skip check
        return 0
    fi

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
        log_error "Use './scripts/submit.sh --restore-private' to recover private files if needed."
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
    ensure_git_root

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

# Description: List remote public branches (exclude develop/dev/*)
# Usage: list_remote_public_branches
# Params: None
# Return: List of remote public branches
list_remote_public_branches() {
    local branches
    branches=$(git branch -r --list | sed 's/^ *//' | grep -vE '/dev/|/develop|/HEAD' || true)

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
        echo "Usage: ./submit.sh --new-dev-branch <new-feature> <base>"
        exit 1
    fi

    if [ -z "$base_input" ]; then
        log_error "Missing base branch."
        log_info "Available remote public branches:"
        git fetch --all --prune >/dev/null 2>&1 || log_warn "Fetch failed. Listing cached remotes..."
        list_remote_public_branches
        echo ""
        echo "Usage: ./submit.sh --new-dev-branch <new-feature> <base>"
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

    log_info "Creating $new_branch from $base_remote..."
    git checkout -b "$new_branch" "$base_remote"
    log_info "New dev branch created: $new_branch"
    log_info "Next steps: ./submit.sh --restore-private && initialize private commit."
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
    
    # Candidates: ALL remote branches EXCEPT dev/* and HEAD
    # This ensures we can detect parents like 'joint_module_display' or 'master'
    local candidates=$(git branch -r --list "$remote/*" | grep -vE "$remote/dev/|$remote/HEAD")
    
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
# Return: List of filenames (Stdout)
# Exit: 0
get_changed_files() {
    git status --porcelain | sed 's/^...//'
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
    local files=$(get_changed_files)
    local private_files=""
    
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $files; do
        if is_private_file "$file"; then
            private_files+="$file"$'\n'
        fi
    done
    IFS="$OLD_IFS"
    echo "$private_files"
}

# Description: Get list of changed public files
# Usage: get_public_changes
# Params: None
# Return: List of public filenames (Stdout)
# Exit: 0
get_public_changes() {
    local files=$(get_changed_files)
    local public_files=""
    
    OLD_IFS="$IFS"
    IFS=$'\n'
    for file in $files; do
        if ! is_private_file "$file"; then
            public_files+="$file"$'\n'
        fi
    done
    IFS="$OLD_IFS"
    echo "$public_files"
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

    while [[ "$dir" != "." && "$dir" != "/" ]]; do
        if [[ -n "${SELECTED_DIRS["$dir"]}" ]]; then
            return 0
        fi
        dir=$(dirname "$dir")
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

    for entry in "${file_entries[@]}"; do
        IFS=$'\t' read -r type class path labels <<< "$entry"
        if [ "$class" = "$target_class" ]; then
            local parent
            parent=$(dirname "$path")
            if ! has_selected_ancestor "$parent"; then
                echo "${path} [${labels}]"
            fi
        fi
    done
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
    
    local raw_diff=$(echo "$changes" | xargs git diff --color=never --)
    local line_count=$(echo "$raw_diff" | wc -l)
    
    if [ "$line_count" -gt "$DIFF_LINE_LIMIT" ]; then
        echo "$raw_diff" > "$diff_file"
        log_warn "Diff output is too large ($line_count lines)."
        log_info "Diff content has been written to: $diff_file"
    else
        echo "$changes" | xargs git diff --color=always --
    fi
}

# --- Restore Private Files ---

# Description: Restore private files from a specified source or auto-detect from parent history.
# Usage: restore_private_files [source_ref]
# Params:
#   $1 - (Optional) Source ref (branch name or commit hash). If empty, tries auto-detection.
# Return: None
# Exit: 0 on success, 1 on failure
restore_private_files() {
    local source_ref="$1"
    local pathspecs=""
    
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

    # 2. Restore
    # Convert regex patterns to pathspecs (remove anchors)
    for pattern in "${PRIVATE_PATTERNS[@]}"; do
        clean_path="${pattern#^}"
        pathspecs="$pathspecs $clean_path"
    done
    
    log_info "Restoring private files from '$source_ref'..."
    log_info "Targets: $pathspecs"
    
    if git checkout "$source_ref" -- $pathspecs; then
        log_info "Successfully restored private files."
    else
        log_error "Failed to restore private files. Check if source '$source_ref' exists."
        exit 1
    fi
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

# 3.1 Early Exit for Help (Bypass Git Root Check)
if [[ "$1" == "--help" ]]; then
    show_help
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
        get_private_changes
        ;;
        
    --print-public-status)
        get_public_changes
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
        echo "=== Private Files List ==="
        get_private_changes
        echo ""
        echo "=== Private Files Diff ==="
        changes=$(get_private_changes)
        if [ -n "$changes" ]; then
             print_diff_safe "$changes" "$PRIVATE_DIFF_FILE"
        fi
        ;;

    --print-public-show)
        echo "=== Public Files List ==="
        get_public_changes
        echo ""
        echo "=== Public Files Diff ==="
        changes=$(get_public_changes)
        if [ -n "$changes" ]; then
             print_diff_safe "$changes" "$PUBLIC_DIFF_FILE"
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
        changes=$(get_private_changes)
        if [ -z "$changes" ]; then
            log_info "No private changes to commit."
            exit 0
        fi
        
        msg="${COMMIT_MSG_PRIVATE}"
        if [ -z "$msg" ]; then
            log_error "Missing commit message for Private changes."
            echo ""
            echo "The following Private files have changes waiting to be committed:"
            echo "$changes" | sed 's/^/- /'
            echo ""
            echo "Please set the 'COMMIT_MSG_PRIVATE' environment variable with a descriptive message before running this command."
            exit 1
        fi
        
        log_info "Committing Private changes..."
        echo "$changes" | xargs git add
        git commit -m "$msg"
        ;;
        
    --commit-public)
        changes=$(get_public_changes)
        if [ -z "$changes" ]; then
            log_info "No public changes to commit."
            exit 0
        fi
        
        msg="${COMMIT_MSG_PUBLIC}"
        if [ -z "$msg" ]; then
            log_error "Missing commit message for Public changes."
            echo ""
            echo "The following Public files have changes waiting to be committed:"
            echo "$changes" | sed 's/^/- /'
            echo ""
            echo "Please set the 'COMMIT_MSG_PUBLIC' environment variable with a descriptive message before running this command."
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

        # Setup Public Branch Locally
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
