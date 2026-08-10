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

MAX_EMOJI=0              # 19 -> 0
MAX_TITLECASE_HEADING=0  # 43 -> 0
MAX_CLAUDISM=0           # 14 -> 0
MAX_PRODUCT_SUBJECT=0    # 3 -> 0
MAX_ROYAL_WE=0           # 4 -> 0
MAX_META_COMMENTARY=0    # 3 -> 0
MAX_HYPE=0               # 1 -> 0
MAX_TRIADIC_NEGATION=1   # 5 -> 1. The one left is the landing's hx-post/onClick/route
                         # list, where each item names a real thing you'd otherwise write.
MAX_EXCLAMATION=3        # 11 -> 3. All three are quoted UI copy ("Changes saved!")
MAX_PASSIVE=34           # 114 -> 34. What remains is mostly legitimate: HTML attributes
                         # ("is required"), Go terms ("are exported"), quoted strings,
                         # and adjectives the regex can't tell from verbs.

# Blank fenced code while preserving line numbers, so counts are prose-only.
strip() { awk '/^```/{c=!c; print ""; next} c{print ""; next} {print}' "$1"; }

# Same, but also blanks table rows. Used only for the emoji count: ✅/⚠️/❌ in a
# comparison matrix are scan markers doing real work (see the "does this scale?"
# table in recipes/counter/index.md), not decoration. In prose they are.
strip_tables() { strip "$1" | awk '/^[[:space:]]*\|/{print ""; next} {print}'; }

native() {
  local f r
  while IFS= read -r f; do
    r=$(grep -m1 'source_repo:' "$f" 2>/dev/null | sed 's/.*source_repo: *//;s/"//g')
    case "$r" in *livetemplate/docs|"") echo "$f";; esac
  done < <(find content -name '*.md' | sort)
}

FILES=$(native)
fail=0

count() { # count <label> <max> <regex> [grep-flag=-oE] [stripper=strip]
  local label=$1 max=$2 re=$3 flag=${4:--oE} pre=${5:-strip}
  local n=0 f hits
  while IFS= read -r f; do
    hits=$($pre "$f" | grep "$flag" "$re" 2>/dev/null | wc -l)
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

count emoji            "$MAX_EMOJI"             '[\x{2705}\x{2728}\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]' -oP strip_tables
count exclamation      "$MAX_EXCLAMATION"       '!'
count product-subject  "$MAX_PRODUCT_SUBJECT"   '^LiveTemplate (is|builds|provides|offers|supports|handles|makes|lets|gives|uses)'
count royal-we         "$MAX_ROYAL_WE"          "(^|[^a-z])(we'll|we've|we're|let's)"
# "the rest of the page" is often literal ("without blocking the rest of the
# page"). Only the document-describing-itself sense counts, so require a verb.
count meta-commentary  "$MAX_META_COMMENTARY"   "(the rest of (this|the) (page|section) (elaborates|explains|covers|walks|goes|shows)|worth stating|as you can see|it's important to note|in this (guide|section), we)"
count hype             "$MAX_HYPE"              '\b(seamless|effortless|blazing|delightful|best-in-class|empower|battle-tested|out of the box)\b'
count triadic-negation "$MAX_TRIADIC_NEGATION"  '\bno [^.,;]{2,40}, no [^.,;]{2,40}(,| and) no '
# Pinned in Phase 1. Do not change this regex without restating every baseline.
count passive          "$MAX_PASSIVE"           '\b(is|are|was|were|be|been)\s+(not\s+)?[a-z]+(ed|en)\b'
# "reach for" is deliberately absent: 24 files, house idiom, not drift. See VOICE.md.
count claudism         "$MAX_CLAUDISM"          '\b(load-bearing|spine|crisp|delve|testament to|nuanced|at its core|fundamentally|north star|unpack|double-click on|orthogonal|heavy lifting|footgun|batteries included|sane defaults|first-class|opinionated|ergonomic|primitives|earns its keep|tapestry|genuinely|meaningfully)\b'

# Title Case headings: two capitalised words in a row after a lowercase one.
# H1 is exempt — it carries the page name, which matches front-matter `title:`
# and feeds the nav and breadcrumbs that docs_ia_test.go and breadcrumb_test.go
# assert on. "Your First App" is a page's name, not prose.
n=0
while IFS= read -r f; do
  n=$((n + $(strip "$f" | grep -cE '^#{2,4} .*[a-z] [A-Z][a-z]+ [A-Z]')))
done <<< "$FILES"
if [ "$n" -gt "$MAX_TITLECASE_HEADING" ]; then
  printf '  FAIL  %-24s %4s  (ceiling %s)\n' titlecase-heading "$n" "$MAX_TITLECASE_HEADING"; fail=1
else
  printf '  ok    %-24s %4s  (ceiling %s)\n' titlecase-heading "$n" "$MAX_TITLECASE_HEADING"
fi

# H1 must be the front-matter title verbatim, or "Title — descriptive clause".
# The H1 feeds nav and breadcrumbs; sentence-casing it renames the page, which
# breadcrumb_test.go asserts against. This check exists because that happened.
drift=""
while IFS= read -r f; do
  t=$(grep -m1 '^title:' "$f" | sed 's/^title: *//; s/^"//; s/"$//')
  # Backticks are markup the front-matter title can't carry; compare without them.
  h=$(grep -m1 '^# ' "$f" | sed 's/^# //; s/`//g')
  [ -z "$t" ] || [ -z "$h" ] && continue
  case "$h" in "$t"|"$t "[—-]*) ;; *) drift="$drift        $f: title=\"$t\" h1=\"$h\"\n";; esac
done <<< "$FILES"
# chat.md's H1 is a descriptive sentence that predates this check.
drift=$(printf "$drift" | grep -v 'recipes/apps/chat.md')
if [ -n "$drift" ]; then
  echo "  FAIL  h1-vs-title"; echo "$drift"; fail=1
else
  printf '  ok    %-24s\n' h1-vs-title
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
