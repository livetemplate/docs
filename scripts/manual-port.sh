#!/usr/bin/env bash
# manual-port.sh — Phase 2 helper: copy a source file into the docs site
# with provenance frontmatter so the "Edit this page on GitHub" link
# resolves to the source repo, not livetemplate/docs.
#
# Usage:
#   manual-port.sh <source_repo_dir> <source_path> <site_url> [<title>]
#
# Example:
#   manual-port.sh ../livetemplate docs/guides/progressive-complexity.md \
#       /guides/progressive-complexity "Progressive Complexity"
#
# This does NOT rewrite cross-repo GitHub URLs. Phase 3 sync action
# handles that systematically; for Phase 2 manual ports we leave the
# URLs intact and audit them separately.
set -euo pipefail

if [ $# -lt 3 ]; then
    echo "Usage: $0 <source_repo_dir> <source_path> <site_url> [<title>]"
    exit 1
fi

SRC_REPO="$1"
SRC_PATH="$2"
SITE_URL="$3"
TITLE="${4:-}"

SRC_FILE="$SRC_REPO/$SRC_PATH"
if [ ! -f "$SRC_FILE" ]; then
    echo "ERROR: source file not found: $SRC_FILE" >&2
    exit 1
fi

# Determine destination path. site_url maps to content/<...>.md;
# trailing-slash URLs become content/<...>/index.md.
DEST_REL="${SITE_URL#/}"
case "$DEST_REL" in
    */)  DEST="content/${DEST_REL}index.md" ;;
    "")  DEST="content/index.md" ;;
    *)   DEST="content/${DEST_REL}.md" ;;
esac

# Determine source repo URL from the directory structure (../<repo-name>).
REPO_NAME=$(basename "$(realpath "$SRC_REPO")")
case "$REPO_NAME" in
    livetemplate) REPO_URL="https://github.com/livetemplate/livetemplate" ;;
    client)       REPO_URL="https://github.com/livetemplate/client" ;;
    lvt)          REPO_URL="https://github.com/livetemplate/lvt" ;;
    examples)     REPO_URL="https://github.com/livetemplate/examples" ;;
    *)            REPO_URL="https://github.com/livetemplate/$REPO_NAME" ;;
esac

# Get source commit SHA for provenance traceability.
SRC_COMMIT=$(cd "$SRC_REPO" && git log -1 --format=%H -- "$SRC_PATH" 2>/dev/null || echo "unknown")

# Extract title from source file if not provided.
if [ -z "$TITLE" ]; then
    # Try YAML frontmatter title:
    TITLE=$(awk '/^---$/{n++; next} n==1 && /^title:/ {sub(/^title:[[:space:]]*"?/, ""); sub(/"?[[:space:]]*$/, ""); print; exit}' "$SRC_FILE")
    # Fall back to first H1.
    if [ -z "$TITLE" ]; then
        TITLE=$(awk '/^# /{sub(/^# /, ""); print; exit}' "$SRC_FILE")
    fi
    # Ultimate fallback: derive from filename.
    if [ -z "$TITLE" ]; then
        TITLE=$(basename "$SRC_PATH" .md | tr '-' ' ')
    fi
fi

mkdir -p "$(dirname "$DEST")"

# Strip any existing frontmatter from the source body, then write our
# own with the provenance fields the docs-site renderer reads.
{
    echo "---"
    echo "title: \"$TITLE\""
    echo "source_repo: \"$REPO_URL\""
    echo "source_path: \"$SRC_PATH\""
    echo "source_commit: \"$SRC_COMMIT\""
    echo "---"
    echo ""
    awk '
        BEGIN { in_fm = 0; fm_done = 0 }
        NR == 1 && /^---$/ { in_fm = 1; next }
        in_fm && /^---$/ { in_fm = 0; fm_done = 1; next }
        in_fm { next }
        { if (!fm_done && NR == 1) fm_done = 1; print }
    ' "$SRC_FILE"
} > "$DEST"

echo "ported  $SRC_PATH  →  $DEST  (commit $SRC_COMMIT)"
