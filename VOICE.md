# Voice

How the prose on this site is written. This file exists because the landing copy
has been rewritten four times (#65, #83, #135, #136) and drifted back to the same
flat register every time. Nothing recorded the target, so every rewrite guessed.

Applies to `content/**` and to the writing around it — PR descriptions, this
file, `README.md`. The register leaks into the meta-writing first.

## Do

**Argue by concession.** Claim, then the objection, then concede it, then
resolve. Four short sentences beat one balanced one:

> A process enforced by a human performs a little better than documents. But a
> human is a human. They get sick, go on a vacation or quit.

**Vary sentence length hard.** Four words, then twenty. Fragments are fine.
Start sentences with But, So and And.

**Ask, then answer in one word.**

> But doesn't the import fetch the module over the network? Yes. That could be a
> problem.

**Understate.** "quite powerful", "fairly simple", "a little better", "for the
most part". Never a superlative.

**Name the catch.** "The catch here is to use `T.Errorf`…"

**Rank what matters.** "This snippet is the most interesting part."

**Admit limits without defending them.** "the abstraction is not complete."

**Position honestly against neighbours, conceding where they win.** htmx, templ,
Alpine and LiveView are all reasonable choices. Say where they're better.

**Stop.** No call to action, no recap, no "happy coding". When the page is done,
end it.

**Bare imperative for instructions.** "Use it in a test." "Run it." The reader is
`you`. Nothing is `we`.

## Don't

**Product as the subject.** Not "LiveTemplate builds reactive web UIs" — write
what the reader does, or what the mechanism does.

**Triadic negation as cadence.** "no client-side framework, no second state
model, no build step" is a drumbeat, not an argument. One negation, doing work.

**Meta-commentary about the document.** "This is the idea the rest of the page
elaborates" tells the reader nothing about the software.

**Exclamation marks, emoji, Title Case headings.** Headings are sentence case.
"The magic:" and "That's it!" are banned outright.

**Claudisms.** The vocabulary that signals a language model wrote it. It comes
back on every rewrite because it reads as thoughtful:

> load-bearing · spine · crisp · surfaces (as a verb) · delve · testament to ·
> underscore · nuanced · at its core · fundamentally · it's worth noting · that
> said · crucially · importantly · genuinely · meaningfully · north star ·
> unpack · double-click on · orthogonal · non-trivial · heavy lifting · footgun ·
> batteries included · sane defaults · first-class · opinionated · ergonomic ·
> primitives · the shape of · earns its keep · seam · tapestry · isn't just X,
> it's Y

Say the plain thing. "load-bearing" is "this line is what makes it work".
"surfaces an error" is "shows the error". "primitives" is usually "functions".

**The rhetorical reversal.** A contrast whose second half exists only to make the
sentence land:

> so you can check the claim rather than take it

The test: **does the contrast tell the reader something?** "sends a frame instead
of a form POST" does — both halves are concrete and you learn what the
alternative was. "check the claim rather than take it" doesn't. This is not a ban
on `rather than`; most uses on this site are the good kind. It's a ban on the
ornamental kind.

## No first person

The docs stay impersonal. That's a deliberate choice, and it means rhythm,
concession and understatement have to carry the voice on their own — there's no
`I` to lean on. Where a blog post would say "too costly for me", a page here says
"that cost is the thing this avoids".

## Four things that are not drift

**Emoji in a comparison table.** ✅ / ⚠️ / ❌ down a column are scan markers doing
real work — see the "does this scale?" matrix in `recipes/counter/index.md`. The
checker skips table rows for that reason. In prose they're decoration, and the
checker still counts them there.

**Title Case in an H1.** The H1 carries the page *name*, and it must be either
the front-matter `title:` verbatim or `Title — a sentence-case clause`. It feeds
the nav and breadcrumbs that `docs_ia_test.go` and `breadcrumb_test.go` assert
on, so downcasing it is a rename, not a copy edit — `breadcrumb_test.go` caught
exactly that during this pass. Rename a page deliberately or not at all. H2–H6
are prose and are sentence case.

**`reach for`** appears on 24 pages and is now house idiom, not a slip. Thin it
where the choosing isn't the point. It's deliberately not in the checker.

**The `ui-patterns/` heading skeleton** — `Template` / `Handler & state` / `When
to use` — is a catalog template. Fix the prose under it, not the shape.

## Every heading that ends in a period has to earn one

A heading punctuated as a sentence must *be* one: it needs a finite verb and
something doing the verb. `content/index.md` shipped four that weren't —
"A shared greeting wall, running on this page.", "The parts of that worth a
second look.", "Other things in here.", "Getting started." Those are labels. A
determiner, a noun, maybe a participle, and a full stop pretending.

This isn't passive voice, but it reads as passive for the same reason: no finite
verb, so nothing is doing anything. The repo owner called it out as passive, and
that instinct was right about the effect even though the term doesn't fit.

The tell is *where* it happens. Headings that make a technical claim stayed
sentences — "The button's name is the action", "The HTML rule runs again in Go",
"Two calls sync your tabs". The four fragments were all headings that orient the
reader: the demo, the section intro, the catch-all, the CTA. The register breaks
where the page navigates rather than argues, so check the connective headings
hardest.

No checker rule for this — "lacks a finite verb" needs a parser, and a fragile
approximation gets switched off. Read each heading and ask what the verb is.

Keep the rail label, the eyebrow and the h2 of a section saying the same thing.
When they drifted, all three were vague at once ("What's going on" / "What's
going on" / "The parts of that worth a second look.") and none named the
contents.

## A comparison table must concede

Every row of the landing's comparison sat under a neighbour's name and described
LiveTemplate instead — so the Alpine.js row read as a description of Alpine. Five
rows, five wins, no losses, from a library whose own eyebrow says alpha. A
comparison that loses to nobody gets discounted whole.

Describe the neighbour, name what it does better, then say where the line falls.
Don't make every row concede-then-win in the same shape either; that cadence is
the same sales pitch in a new suit. Some rows concede and stop.

## What the checker does

`scripts/voice-check.sh` counts the tells that can be counted: emoji,
exclamation marks, Title Case headings, product-as-subject openers, royal we,
claudisms, passive constructions. It runs on docs-native paths only.

It will not find a page that is merely flat. Most of this file is not
checkable — that's why it's written down.

## Pages this file can't fix

Anything with an upstream `source_repo` in its front matter is mirrored, and
`cmd/sync` overwrites the whole file on the next release. Fix those at the
source repo. See `content/_meta/source-of-truth.yaml` for the mapping.
