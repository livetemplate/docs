#!/usr/bin/env bash
# build-changelog.sh — concatenate the four source-repo CHANGELOGs into
# a single docs-site /changelog page. Each repo gets its own H2 section
# pointing at the canonical CHANGELOG on GitHub for full history.
set -euo pipefail

cd "$(dirname "$0")/.."

OUT=content/changelog.md

cat > "$OUT" <<'HEADER'
---
title: "Changelog"
---

# Changelog

The full release history across the four LiveTemplate ecosystem repos.
Per-repo CHANGELOGs remain canonical in their source repos; this page
mirrors them in one place for convenience. The Phase 3 sync action will
keep each section in step with its source on every release.

> Each section below is the **full** CHANGELOG of the corresponding repo.
> Use Ctrl-F to find a specific version.

HEADER

emit_section() {
    local repo_dir="$1"
    local repo_name="$2"
    local pretty_name="$3"
    local source_path="${repo_dir}/CHANGELOG.md"
    if [ ! -f "$source_path" ]; then
        echo "WARN: no CHANGELOG at $source_path" >&2
        return
    fi
    {
        echo ""
        echo "---"
        echo ""
        echo "## $pretty_name"
        echo ""
        echo "_Canonical source: [livetemplate/${repo_name}/CHANGELOG.md](https://github.com/livetemplate/${repo_name}/blob/main/CHANGELOG.md)_"
        echo ""
        # Skip the source's own H1 + boilerplate header. Start emitting
        # at the first H2 line (versioned release).
        awk '/^## /{p=1} p' "$source_path"
    } >> "$OUT"
}

emit_section "../livetemplate" "livetemplate" "livetemplate (Go framework)"
emit_section "../client"       "client"       "@livetemplate/client (TypeScript client)"
emit_section "../lvt"          "lvt"          "lvt (CLI)"
emit_section "../examples"     "examples"     "examples (apps)"

echo "wrote $OUT ($(wc -l < "$OUT") lines)"
