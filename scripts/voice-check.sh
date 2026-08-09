#!/usr/bin/env bash
# Counts the voice tells that can be counted. See VOICE.md for the ones that can't.
#
# This is a ratchet, not a gate: each MAX_* below is the count on the day the
# check was added or last tightened. The build fails if a number goes UP. When
# you bring one down, lower the ceiling in the same commit — that's what stops
# the next rewrite from undoing this one.
#
# Scope is docs-native pages only. Pages with an upstream source_repo are
# overwritten by cmd/sync, so failing the commit on them would punish the wrong
# person; fix those at the source repo instead.
set -uo pipefail
cd "$(dirname "$0")/.."

MAX_EMOJI=5              # was 19 before the Phase 1 rewrites. target 0
MAX_EXCLAMATION=6        # was 11. target 0 — the rest are in recipes/apps/
MAX_TITLECASE_HEADING=41 # was 43. target 0 — 36 of these are in recipes/apps/
MAX_PRODUCT_SUBJECT=2    # was 3. target 0
MAX_ROYAL_WE=1           # was 3, all in getting-started/your-first-app.md. target 0
MAX_META_COMMENTARY=1    # was 3. target 0
MAX_CLAUDISM=14          # target 0
MAX_TRIADIC_NEGATION=5   # target 1 (the install page's list is genuine)
MAX_HYPE=1               # target 0 — "seamless scrolling"
MAX_PASSIVE=114          # target 35 — the largest bucket, and the slowest to move

# Blank fenced code while preserving line numbers, so counts are prose-only.
strip() { awk '/^```/{c=!c; print ""; next} c{print ""; next} {print}' "$1"; }

native() {
  local f r
  while IFS= read -r f; do
    r=$(grep -m1 'source_repo:' "$f" 2>/dev/null | sed 's/.*source_repo: *//;s/"//g')
    case "$r" in *livetemplate/docs|"") echo "$f";; esac
  done < <(find content -name '*.md' | sort)
}

FILES=$(native)
fail=0

count() { # count <label> <max> <regex> [grep-flag, default -oE]
  local label=$1 max=$2 re=$3 flag=${4:--oE}
  local n=0 f hits
  while IFS= read -r f; do
    hits=$(strip "$f" | grep "$flag" "$re" 2>/dev/null | wc -l)
    n=$((n + hits))
  done <<< "$FILES"
  if [ "$n" -gt "$max" ]; then
    printf '  FAIL  %-24s %4s  (ceiling %s)\n' "$label" "$n" "$max"; fail=1
  else
    printf '  ok    %-24s %4s  (ceiling %s)\n' "$label" "$n" "$max"
  fi
  [ "$n" -lt "$max" ] && printf '        ^ below ceiling — lower MAX_%s to %s\n' \
    "$(echo "$label" | tr 'a-z-' 'A-Z_')" "$n"
  return 0
}

echo "voice-check: $(echo "$FILES" | wc -l) docs-native pages"

count emoji            "$MAX_EMOJI"             '[\x{2705}\x{2728}\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]' -oP
count exclamation      "$MAX_EXCLAMATION"       '!'
count product-subject  "$MAX_PRODUCT_SUBJECT"   '^LiveTemplate (is|builds|provides|offers|supports|handles|makes|lets|gives|uses)'
count royal-we         "$MAX_ROYAL_WE"          "(^|[^a-z])(we'll|we've|we're|let's)"
count meta-commentary  "$MAX_META_COMMENTARY"   "(the rest of (this|the) page|worth stating|as you can see|it's important to note|in this (guide|section), we)"
count hype             "$MAX_HYPE"              '\b(seamless|effortless|blazing|delightful|best-in-class|empower|battle-tested|out of the box)\b'
count triadic-negation "$MAX_TRIADIC_NEGATION"  '\bno [^.,;]{2,40}, no [^.,;]{2,40}(,| and) no '
# Pinned in Phase 1. Do not change this regex without restating every baseline.
count passive          "$MAX_PASSIVE"           '\b(is|are|was|were|be|been)\s+(not\s+)?[a-z]+(ed|en)\b'
# "reach for" is deliberately absent: 24 files, house idiom, not drift. See VOICE.md.
count claudism         "$MAX_CLAUDISM"          '\b(load-bearing|spine|crisp|delve|testament to|nuanced|at its core|fundamentally|north star|unpack|double-click on|orthogonal|heavy lifting|footgun|batteries included|sane defaults|first-class|opinionated|ergonomic|primitives|earns its keep|tapestry|genuinely|meaningfully)\b'

# Title Case headings: two capitalised words in a row after a lowercase one.
n=0
while IFS= read -r f; do
  n=$((n + $(strip "$f" | grep -cE '^#{1,4} .*[a-z] [A-Z][a-z]+ [A-Z]')))
done <<< "$FILES"
if [ "$n" -gt "$MAX_TITLECASE_HEADING" ]; then
  printf '  FAIL  %-24s %4s  (ceiling %s)\n' titlecase-heading "$n" "$MAX_TITLECASE_HEADING"; fail=1
else
  printf '  ok    %-24s %4s  (ceiling %s)\n' titlecase-heading "$n" "$MAX_TITLECASE_HEADING"
fi

# Phrases with no defensible use. Scoped to native pages like everything else —
# cli/index.md and contributing/livetemplate.md also say "That's it!", but those
# are mirrored and belong to the upstream PRs.
banned=$(echo "$FILES" | tr '\n' '\0' | xargs -0 grep -nE "The magic:|That's it!|Perfect for .*!|Why So Simple" 2>/dev/null)
if [ -n "$banned" ]; then
  echo "  FAIL  banned-phrase"
  echo "$banned" | sed 's/^/        /'
  fail=1
else
  printf '  ok    %-24s\n' banned-phrase
fi

echo
[ "$fail" -eq 0 ] && echo "voice-check: pass" || echo "voice-check: FAIL — see VOICE.md"
exit "$fail"
