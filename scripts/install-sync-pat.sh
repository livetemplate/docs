#!/usr/bin/env bash
# install-sync-pat.sh — install the docs-sync PAT as repo secrets across
# the docs site and its 4 source repos. Reads the token from the env
# var DOCS_SYNC_PAT so it never lands in shell history or logs.
#
# Usage:
#   DOCS_SYNC_PAT='github_pat_...' ./scripts/install-sync-pat.sh
#
# Same PAT is installed under two names because the receiving and
# sending sides reference different secret names:
#   - DOCS_BOT_TOKEN      on livetemplate/docs           (PR creation)
#   - DOCS_DISPATCH_TOKEN on each of the 4 source repos  (dispatch fire)
#
# The PAT must be a fine-grained token scoped to livetemplate/docs only,
# with permissions Contents: Read+Write and Pull requests: Read+Write.
# See livetemplate/docs/CONTRIBUTING.md → Sync workflow → "PAT setup"
# for the click-by-click in the GitHub UI.
set -euo pipefail

if [ -z "${DOCS_SYNC_PAT:-}" ]; then
    echo "ERROR: DOCS_SYNC_PAT is not set." >&2
    echo "" >&2
    echo "Usage: DOCS_SYNC_PAT='github_pat_...' $0" >&2
    exit 1
fi

# Sanity-check format without echoing the value. Fine-grained PATs
# start with github_pat_; classic PATs start with ghp_. Either works
# for our purposes; warn (don't fail) on unrecognized prefixes.
case "$DOCS_SYNC_PAT" in
    github_pat_*) : ;;
    ghp_*)        echo "note: classic PAT detected — fine-grained recommended" ;;
    *)            echo "WARN: token doesn't look like a GitHub PAT (no github_pat_/ghp_ prefix)" >&2 ;;
esac

# Verify the PAT is actually valid before spending time on 5 sets.
if ! token=$DOCS_SYNC_PAT gh auth status --hostname github.com >/dev/null 2>&1 \
        && ! GH_TOKEN=$DOCS_SYNC_PAT gh api /user --jq .login >/dev/null 2>&1; then
    user=$(GH_TOKEN=$DOCS_SYNC_PAT gh api /user --jq .login 2>&1 || true)
    if [ -z "$user" ]; then
        echo "ERROR: GitHub rejected the PAT (api /user failed)" >&2
        echo "Check that the token hasn't expired and has Contents: R+W on livetemplate/docs" >&2
        exit 2
    fi
fi
authed_user=$(GH_TOKEN=$DOCS_SYNC_PAT gh api /user --jq .login)
echo "PAT authenticates as: $authed_user"
echo ""

set_secret() {
    local repo="$1"
    local secret_name="$2"
    if gh secret set "$secret_name" --repo "$repo" --body "$DOCS_SYNC_PAT" >/dev/null 2>&1; then
        echo "  ✓ set $secret_name on $repo"
    else
        echo "  ✗ FAILED to set $secret_name on $repo" >&2
        return 1
    fi
}

failed=0

echo "Installing DOCS_BOT_TOKEN on the docs repo (lets sync workflow open PRs):"
set_secret livetemplate/docs DOCS_BOT_TOKEN || failed=1
echo ""

echo "Installing DOCS_DISPATCH_TOKEN on the 4 source repos (lets release tags fire docs sync):"
for repo in livetemplate client lvt examples; do
    set_secret "livetemplate/$repo" DOCS_DISPATCH_TOKEN || failed=1
done

echo ""
if [ "$failed" -eq 0 ]; then
    echo "Done. Verify by triggering a sync:"
    echo ""
    echo "  gh workflow run sync.yml --repo livetemplate/docs \\"
    echo "    -f source_repo=https://github.com/livetemplate/livetemplate \\"
    echo "    -f ref=main"
    echo ""
    echo "After ~30s a sync PR should appear in livetemplate/docs."
else
    echo "Some secrets failed to set; see errors above." >&2
    exit 3
fi
