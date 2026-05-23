#!/bin/bash
set -euo pipefail

# LiveTemplate Docs Release Script
# Usage: ./release.sh <version>
# Example: ./release.sh 0.1.0
#
# Creates a new release by tagging the repository. Unlike tinkerdown
# (which has cli-release.yml building binaries on tag push), docs has
# no tag-triggered workflow — deploy.yml fires on every push to main,
# independent of tags. A docs tag is purely a milestone marker + the
# pre-condition for `gh release create` so the history is browseable.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 <version>"
    echo ""
    echo "Creates a new docs release by tagging the repository."
    echo "After tagging, run:"
    echo "  gh release create v<version> --generate-notes --title \"<title>\""
    echo ""
    echo "Examples:"
    echo "  $0 0.1.0    # Creates tag v0.1.0"
    echo "  $0 0.1.1    # Creates tag v0.1.1"
    exit 1
}

if [ $# -ne 1 ]; then
    usage
fi

VERSION="$1"

# Validate version format (semver without v prefix)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
    echo -e "${RED}Error: Invalid version format. Expected: X.Y.Z or X.Y.Z-suffix${NC}"
    echo "Examples: 0.1.0, 0.1.1, 1.0.0-rc.1"
    exit 1
fi

TAG="v${VERSION}"

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}Error: You have uncommitted changes. Please commit or stash them first.${NC}"
    git status --short
    exit 1
fi

# Check if we're on main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${YELLOW}Warning: You are on branch '$CURRENT_BRANCH', not 'main'.${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if tag already exists
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag $TAG already exists.${NC}"
    exit 1
fi

# Fetch latest from remote
echo -e "${YELLOW}Fetching latest from origin...${NC}"
git fetch origin

# Check if local is up to date with remote
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main 2>/dev/null || echo "")
if [ -n "$REMOTE" ] && [ "$LOCAL" != "$REMOTE" ]; then
    echo -e "${YELLOW}Warning: Local branch differs from origin/main.${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Show what we're about to do
echo ""
echo -e "${GREEN}Release Summary:${NC}"
echo "  Version: $VERSION"
echo "  Tag:     $TAG"
echo "  Commit:  $(git rev-parse --short HEAD)"
echo "  Message: $(git log -1 --pretty=%s)"
echo ""

read -p "Create and push tag $TAG? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create and push tag
echo -e "${YELLOW}Creating tag $TAG...${NC}"
git tag -a "$TAG" -m "Release $VERSION"

echo -e "${YELLOW}Pushing tag to origin...${NC}"
git push origin "$TAG"

echo ""
echo -e "${GREEN}Success! Tag $TAG has been pushed.${NC}"
echo ""
echo "Next step — create the GitHub release with auto-generated notes:"
echo "  gh release create $TAG --generate-notes --title \"<title>\""
echo ""
echo "Production deploy already fired on the main-branch merge that this tag points at;"
echo "this tag is the browseable milestone marker, not a deploy trigger."
