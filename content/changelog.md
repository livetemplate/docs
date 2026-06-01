---
title: "Changelog"
description: "Release history across the LiveTemplate Go framework, TypeScript client, lvt CLI, examples, and docs site."
---

# Changelog

The full release history across the four LiveTemplate ecosystem repos.
Per-repo CHANGELOGs remain canonical in their source repos; this page
mirrors them in one place for convenience. The Phase 3 sync action will
keep each section in step with its source on every release.

> Each section below is the **full** CHANGELOG of the corresponding repo.
> Use Ctrl-F to find a specific version.


---

## livetemplate (Go framework)

_Canonical source: [livetemplate/livetemplate/CHANGELOG.md](https://github.com/livetemplate/livetemplate/blob/main/CHANGELOG.md)_

## [v0.8.23] - 2026-05-02

### Changes

- refactor: streaming range Phase 8.5 — remove dead keyGen plumbing (#370) (2ef24d11)
- perf: streaming range Phase 7 — type-direct hash + parallel build (#369) (24950bf9)
- feat: streaming range Phase 6 — recursive transition + LargeTable demo (#368) (900d1da8)
- feat: streaming range Phase 5 — benchmark gate + measured §7 numbers (#366) (73cff639)
- feat: streaming range Phase 4 — cleanup + spec update (#365) (45756b72)
- feat: streaming range Phase 3 — caller integration (cutover) (#364) (6d46c35e)
- feat: streaming range Phase 2 — diff entry point (callable, unwired) (#363) (c07977d9)
- feat: streaming range Phase 1 — foundational types (no-op) (#362) (2a28b70d)
- docs(proposals): Phase 0 audit for streaming range rendering (#361) (f075d7b5)
- docs(proposals): streaming range rendering (#360) (89688276)
- docs(proposals): record lvt-scroll-away top edge ship + Pattern #10 status (a5a5b4bc)
- docs(proposals): tick Session 7 boxes + Implementation Notes (895865f1)
- docs(proposals): tick Session 6 boxes + Session 6 implementation notes (83226ab2)
- docs(proposals): patterns Session 5 complete + 3 implementation notes (d6efdacc)



## [v0.8.22] - 2026-04-25

### Changes

- chore: ignore .claude/scheduled_tasks.lock (Claude Code transient state) (b6cb4f52)
- fix: prune expired flash before render (not after sendUpdate) (#359) (ad5f1071)
- docs(proposals): patterns Session 4 complete + 6 implementation notes (747aedce)




<a name="v0.8.21"></a>
## [v0.8.21] - 2026-04-22

### Bug Fixes

- eliminate race in Redis pub/sub init and add subscription retry ([#355](https://github.com/livefir/livetemplate/issues/355))

### Documentation

- scroll effect targeting, lvt-scroll-away, and chat recipe ([#356](https://github.com/livefir/livetemplate/issues/356))
- **proposals:** update patterns for v0.8.19 + v0.8.33 ([#358](https://github.com/livefir/livetemplate/issues/358))


<a name="v0.8.20"></a>
## [v0.8.20] - 2026-04-21

### Documentation

- update scroll-sentinel to lvt-scroll-sentinel attribute ([#352](https://github.com/livefir/livetemplate/issues/352))
- automatic client-side state preservation ([#351](https://github.com/livefir/livetemplate/issues/351))


<a name="v0.8.19"></a>
## [v0.8.19] - 2026-04-18

### Documentation

- **proposals:** Session 3 complete + server-push pattern lessons ([#338](https://github.com/livefir/livetemplate/issues/338))

### Features

- __navigate__ action + flash persist-until-cleared lifecycle ([#344](https://github.com/livefir/livetemplate/issues/344))


<a name="v0.8.18"></a>
## [v0.8.18] - 2026-04-14

### Bug Fixes

- wire Session.TriggerAction into lifecycle contexts ([#336](https://github.com/livefir/livetemplate/issues/336))
- **ci:** update [@livetemplate](https://github.com/livetemplate)/client to latest in cross-repo tests ([#328](https://github.com/livefir/livetemplate/issues/328))

### Documentation

- patterns example proposal ([#333](https://github.com/livefir/livetemplate/issues/333))
- update dialog routing with polyfill context ([#331](https://github.com/livefir/livetemplate/issues/331))
- README rewrite proposal ([#268](https://github.com/livefir/livetemplate/issues/268)) ([#332](https://github.com/livefir/livetemplate/issues/332))
- comprehensive documentation overhaul ([#329](https://github.com/livefir/livetemplate/issues/329))
- **proposals:** patterns session 2 tracker ([#335](https://github.com/livefir/livetemplate/issues/335))
- **proposals:** patterns session 1 tracker + implementation notes ([#334](https://github.com/livefir/livetemplate/issues/334))


<a name="v0.8.17"></a>
## [v0.8.17] - 2026-04-10

### Bug Fixes

- parse individual form fields in multipart submissions ([#326](https://github.com/livefir/livetemplate/issues/326))

### Documentation

- update attribute-reduction proposal with Phase 2 completion status ([#324](https://github.com/livefir/livetemplate/issues/324))


<a name="v0.8.16"></a>
## [v0.8.16] - 2026-04-04

### Documentation

- mark Phase 2E complete in attribute-reduction proposal

### Features

- Tier 1 file uploads — HTTP multipart with progress tracking


<a name="v0.8.15"></a>
## [v0.8.15] - 2026-04-04

### Bug Fixes

- unreserve action field, update tests to use lvt-action ([#321](https://github.com/livefir/livetemplate/issues/321))

### Documentation

- mark Phase 1B as complete in progress tracker ([#322](https://github.com/livefir/livetemplate/issues/322))
- update client-attributes reference for action-fix changes
- add lvt-form:action, lvt-nav: group, lvt-on:change to proposal
- mark Phase 1A complete in attribute-reduction proposal
- attribute reduction proposal — design + implementation plan ([#288](https://github.com/livefir/livetemplate/issues/288))


<a name="v0.8.14"></a>
## [v0.8.14] - 2026-04-02

### Features

- add AriaDisabled and FlashTag template helpers ([#318](https://github.com/livefir/livetemplate/issues/318))


<a name="v0.8.13"></a>
## [v0.8.13] - 2026-04-02

### Documentation

- add ephemeral-components guide ([#316](https://github.com/livefir/livetemplate/issues/316))


<a name="v0.8.12"></a>
## [v0.8.12] - 2026-04-01

### Documentation

- add attribute reduction proposal ([#288](https://github.com/livefir/livetemplate/issues/288)) ([#292](https://github.com/livefir/livetemplate/issues/292))

### Features

- selective state persistence via lvt:"persist" tag ([#308](https://github.com/livefir/livetemplate/issues/308))
- simplify error rendering with ErrorTag and AriaInvalid helpers ([#307](https://github.com/livefir/livetemplate/issues/307))


<a name="v0.8.11"></a>
## [v0.8.11] - 2026-04-01

### Features

- add WithEphemeralState() to opt out of state persistence ([#301](https://github.com/livefir/livetemplate/issues/301))


<a name="v0.8.10"></a>
## [v0.8.10] - 2026-03-31

### Bug Fixes

- skip HTTP POST persistence on action error + add multi-tab dedup logging ([#296](https://github.com/livefir/livetemplate/issues/296))
- only create .uploads directory when uploads are configured ([#287](https://github.com/livefir/livetemplate/issues/287))

### Documentation

- add Tier 1 file uploads proposal ([#271](https://github.com/livefir/livetemplate/issues/271)) ([#291](https://github.com/livefir/livetemplate/issues/291))
- state safety, current limitations, and progressive enhancement ([#284](https://github.com/livefir/livetemplate/issues/284))

### Features

- simplify state management and persistence defaults ([#298](https://github.com/livefir/livetemplate/issues/298))
- make state persistence opt-in via WithStatePersistence() ([#295](https://github.com/livefir/livetemplate/issues/295))
- per-connection state persists to session store for page refresh ([#290](https://github.com/livefir/livetemplate/issues/290))


<a name="v0.8.9"></a>
## [v0.8.9] - 2026-03-30

### Bug Fixes

- flash messages not rendered in WebSocket tree-diff mode ([#283](https://github.com/livefir/livetemplate/issues/283))
- pull latest from remote before starting release ([#281](https://github.com/livefir/livetemplate/issues/281))


<a name="v0.8.8"></a>
## [v0.8.8] - 2026-03-29

### Bug Fixes

- AsState panics if state contains dependency types ([#273](https://github.com/livefir/livetemplate/issues/273))

### Features

- per-connection state scoping (LiveView-style socket assigns) ([#275](https://github.com/livefir/livetemplate/issues/275))

### Breaking change


actions no longer auto-broadcast state or persist to SessionStore.

Key changes:
- Remove auto-broadcast and SessionStore persist from WebSocket action loop
- Add ctx.BroadcastAction() API for explicit cross-connection dispatch
- Restructure WS message loop to select-based event loop (readPump + DispatchChan)
- Add GroupActionMessage type and Redis PubSub support for cross-instance broadcast
- Handle BroadcastAction from both WebSocket and HTTP POST paths


<a name="v0.8.7"></a>
## [v0.8.7] - 2026-03-27

### Features

- formless standalone buttons — remove hidden form ([#263](https://github.com/livefir/livetemplate/issues/263))


<a name="v0.8.6"></a>
## [v0.8.6] - 2026-03-26

### Bug Fixes

- use current branch name in release script instead of hardcoded main/master
- preserve struct methods in template data map ([#254](https://github.com/livefir/livetemplate/issues/254))

### Features

- communicate Change() capability to client via initial render metadata ([#253](https://github.com/livefir/livetemplate/issues/253))


<a name="v0.8.5"></a>
## [v0.8.5] - 2026-03-25

### Bug Fixes

- session benchmarks fail with 'client too slow' ([#209](https://github.com/livefir/livetemplate/issues/209))
- track dynamic pubsub subscriptions for reconnect and wire into mount ([#213](https://github.com/livefir/livetemplate/issues/213))
- check X-Forwarded-Proto in WebSocket origin checker ([#190](https://github.com/livefir/livetemplate/issues/190))

### Code Refactoring

- move progressive complexity examples to examples repo ([#248](https://github.com/livefir/livetemplate/issues/248))
- deduplicate generateItemHash into shared keys package ([#208](https://github.com/livefir/livetemplate/issues/208))
- rewrite parse package with custom AST evaluator ([#199](https://github.com/livefir/livetemplate/issues/199))

### Documentation

- update perf docs with TreeNode pooling investigation results ([#228](https://github.com/livefir/livetemplate/issues/228))
- update performance docs and baseline for recent optimizations ([#227](https://github.com/livefir/livetemplate/issues/227))
- update performance docs and baseline for recent optimizations ([#217](https://github.com/livefir/livetemplate/issues/217))

### Features

- progressive complexity model for form handling ([#233](https://github.com/livefir/livetemplate/issues/233))
- enhance ValidationToMultiError with friendly names and new tags ([#218](https://github.com/livefir/livetemplate/issues/218))
- add WithTrustForwardedHeaders config option ([#211](https://github.com/livefir/livetemplate/issues/211))

### Performance Improvements

- system card benchmark and per-session memory optimization ([#235](https://github.com/livefir/livetemplate/issues/235))
- replace encoding/json with json-iterator in hot paths ([#229](https://github.com/livefir/livetemplate/issues/229))
- reduce allocations with shared statics, buffer pool, and reflection dedup ([#224](https://github.com/livefir/livetemplate/issues/224))
- replace TreeNode Dynamics map with slice for ~20% speedup ([#220](https://github.com/livefir/livetemplate/issues/220))
- reduce template parsing allocations by 50-57% per render ([#219](https://github.com/livefir/livetemplate/issues/219))
- optimize range diffing with pre-computed context ([#212](https://github.com/livefir/livetemplate/issues/212))
- switch fingerprint hash to FNV-1a; add stress tests ([#205](https://github.com/livefir/livetemplate/issues/205))


<a name="v0.8.4"></a>
## [v0.8.4] - 2026-03-14

### Bug Fixes

- unify divergent expression evaluation paths ([#176](https://github.com/livefir/livetemplate/issues/176)) ([#179](https://github.com/livefir/livetemplate/issues/179))
- use cookie-based flash messages instead of URL query params ([#136](https://github.com/livefir/livetemplate/issues/136))

### Documentation

- refresh benchmark baseline and remove stale references ([#185](https://github.com/livefir/livetemplate/issues/185))
- batch address 9 documentation follow-up issues ([#178](https://github.com/livefir/livetemplate/issues/178))
- update performance docs to reflect current codebase ([#175](https://github.com/livefir/livetemplate/issues/175))
- audit and reorganize proposals directory ([#173](https://github.com/livefir/livetemplate/issues/173))
- replace api-reference.md with Go library API reference ([#164](https://github.com/livefir/livetemplate/issues/164))
- rewrite uploads.md for Controller+State pattern ([#163](https://github.com/livefir/livetemplate/issues/163))
- fix broken links in CONFIGURATION.md and client-attributes.md ([#162](https://github.com/livefir/livetemplate/issues/162))
- fix session.md interface signatures and add missing features ([#161](https://github.com/livefir/livetemplate/issues/161))
- expand server-actions.md with pubsub package details ([#160](https://github.com/livefir/livetemplate/issues/160))
- fix controller-pattern.md phantom methods, add missing APIs ([#159](https://github.com/livefir/livetemplate/issues/159))
- fix authentication.md phantom methods and broken link ([#158](https://github.com/livefir/livetemplate/issues/158))
- update template-support-matrix.md with current codebase state ([#157](https://github.com/livefir/livetemplate/issues/157))
- fix spec inaccuracies found during implementation verification ([#156](https://github.com/livefir/livetemplate/issues/156))
- move lvt-specific guides to lvt repo ([#153](https://github.com/livefir/livetemplate/issues/153))
- improve new contributor walkthrough guide ([#152](https://github.com/livefir/livetemplate/issues/152))
- audit specs, design, performance, and CLAUDE.md (Batch 5) ([#149](https://github.com/livefir/livetemplate/issues/149))
- update configuration and reference docs (Batch 4) ([#147](https://github.com/livefir/livetemplate/issues/147))
- audit and fix guide documentation (Batch 3) ([#146](https://github.com/livefir/livetemplate/issues/146))
- regenerate core architecture docs (Batch 2) ([#145](https://github.com/livefir/livetemplate/issues/145))
- update component import paths in doc comments
- archive 22 completed planning artifacts (Batch 1) ([#144](https://github.com/livefir/livetemplate/issues/144))
- Add comprehensive documentation overhaul plan and update README index ([#138](https://github.com/livefir/livetemplate/issues/138))
- fix metric names to match prometheus.go output
- fix internal/observe imports and document TraceMiddleware removal ([#137](https://github.com/livefir/livetemplate/issues/137))

### Features

- integrate LVT_WS_BUFFER_SIZE into EnvConfig system ([#151](https://github.com/livefir/livetemplate/issues/151))
- support template variable declarations ($c := .) in parser ([#150](https://github.com/livefir/livetemplate/issues/150))
- support template variable declarations ($c := .) in parser


<a name="v0.8.3"></a>
## [v0.8.3] - 2026-02-27

### Bug Fixes

- skip npm tests in pre-commit when client/ directory is absent
- cache HTTP templates per session to enable diff optimization ([#134](https://github.com/livefir/livetemplate/issues/134))
- add component attribute to all remaining slog calls ([#132](https://github.com/livefir/livetemplate/issues/132))
- slog cleanup — error handling, formatting, and component attributes ([#130](https://github.com/livefir/livetemplate/issues/130))
- enable burst mutation fuzz tests and fix KeyStability invariant ([#118](https://github.com/livefir/livetemplate/issues/118))
- handle complex insertion patterns in range differential operations ([#113](https://github.com/livefir/livetemplate/issues/113))

### Code Refactoring

- migrate log.Printf to structured slog logging ([#100](https://github.com/livefir/livetemplate/issues/100)) ([#123](https://github.com/livefir/livetemplate/issues/123))

### Documentation

- document auto-key behavioral change in release notes ([#121](https://github.com/livefir/livetemplate/issues/121))
- document fingerprint-based diff architecture ([#120](https://github.com/livefir/livetemplate/issues/120))


<a name="v0.8.2"></a>
## [v0.8.2] - 2026-02-02

### Features

- comprehensive fuzz testing framework with TypeScript oracle ([#110](https://github.com/livefir/livetemplate/issues/110))


<a name="v0.8.1"></a>
## [v0.8.1] - 2026-01-26

### Bug Fixes

- skip Redis tests gracefully when Docker is unavailable ([#109](https://github.com/livefir/livetemplate/issues/109))
- address Copilot review comments on API accuracy
- correct API references and range operation format in walkthrough

### Features

- auto-generated keys for range items without explicit key attribute ([#108](https://github.com/livefir/livetemplate/issues/108))
- progressive enhancement support for non-JS form submissions ([#102](https://github.com/livefir/livetemplate/issues/102))


<a name="v0.8.0"></a>
## [v0.8.0] - 2026-01-18


<a name="v0.7.12"></a>
## [v0.7.12] - 2026-01-10

### Bug Fixes

- preserve statics for conditional blocks in tree updates ([#84](https://github.com/livefir/livetemplate/issues/84))


<a name="v0.7.11"></a>
## [v0.7.11] - 2026-01-06

### Bug Fixes

- recognize append/prepend patterns to prevent statics resend on load_more ([#83](https://github.com/livefir/livetemplate/issues/83))


<a name="v0.7.10"></a>
## [v0.7.10] - 2026-01-04

### Bug Fixes

- handle range→else transitions in top-level range handling


<a name="v0.7.9"></a>
## [v0.7.9] - 2026-01-03

### Bug Fixes

- invalidate registry when conditional becomes empty ([#81](https://github.com/livefir/livetemplate/issues/81))


<a name="v0.7.8"></a>
## [v0.7.8] - 2025-12-27

### Bug Fixes

- **diff:** detect tree node changes when statics differ
- **mount:** enable flash messages on HTTP redirects with query params


<a name="v0.7.7"></a>
## [v0.7.7] - 2025-12-26

### Features

- add per-connection flash messages ([#79](https://github.com/livefir/livetemplate/issues/79))


<a name="v0.7.6"></a>
## [v0.7.6] - 2025-12-25

### Features

- add query parameter support for Mount and action handlers ([#78](https://github.com/livefir/livetemplate/issues/78))


<a name="v0.7.5"></a>
## [v0.7.5] - 2025-12-24

### Bug Fixes

- handle non-TreeNode to TreeNode transitions in range updates ([#77](https://github.com/livefir/livetemplate/issues/77))
- handle non-TreeNode to TreeNode transitions in range updates


<a name="v0.7.4"></a>
## [v0.7.4] - 2025-12-23

### Bug Fixes

- ensure Range.Statics populated for empty→items transitions ([#76](https://github.com/livefir/livetemplate/issues/76))


<a name="v0.7.3"></a>
## [v0.7.3] - 2025-12-22

### Bug Fixes

- support heterogeneous range items with per-item statics ([#75](https://github.com/livefir/livetemplate/issues/75))


<a name="v0.7.2"></a>
## [v0.7.2] - 2025-12-20

### Bug Fixes

- add type guard in SetDynamic to prevent raw structs in tree dynamics ([#74](https://github.com/livefir/livetemplate/issues/74))

### Features

- action.go updates for livepage ([#73](https://github.com/livefir/livetemplate/issues/73))


<a name="v0.7.1"></a>
## [v0.7.1] - 2025-12-14

### Bug Fixes

- mark range statics path in registry for proper caching ([#72](https://github.com/livefir/livetemplate/issues/72))


<a name="v0.7.0"></a>
## [v0.7.0] - 2025-12-10

### Documentation

- update all documentation for Controller+State API (v0.7.0) ([#70](https://github.com/livefir/livetemplate/issues/70))

### Features

- add component template registration support ([#71](https://github.com/livefir/livetemplate/issues/71))


<a name="v0.6.0"></a>
## [v0.6.0] - 2025-12-04


<a name="v0.5.2"></a>
## [v0.5.2] - 2025-12-03

### Documentation

- update client-attributes reference with reactive attributes and more ([#65](https://github.com/livefir/livetemplate/issues/65))
- add reactive attributes proposal ([#64](https://github.com/livefir/livetemplate/issues/64))

### Features

- store pattern redesign with automatic method dispatch ([#66](https://github.com/livefir/livetemplate/issues/66))


<a name="v0.5.1"></a>
## [v0.5.1] - 2025-11-30

### Documentation

- add authentication and session reference documentation ([#63](https://github.com/livefir/livetemplate/issues/63))


<a name="v0.5.0"></a>
## [v0.5.0] - 2025-11-30

### Documentation

- update documentation for Session API ([#62](https://github.com/livefir/livetemplate/issues/62))
- improve README structure and narrative flow ([#59](https://github.com/livefir/livetemplate/issues/59))

### Features

- add Session API for server-initiated actions ([#61](https://github.com/livefir/livetemplate/issues/61))
- add HTTP methods to ActionContext for authentication (v0.5) ([#60](https://github.com/livefir/livetemplate/issues/60))
- add coverage targets to Makefile ([#57](https://github.com/livefir/livetemplate/issues/57))


<a name="v0.4.2-debug.2"></a>
## [v0.4.2-debug.2] - 2025-11-22

### Bug Fixes

- add log package import for debug logging

### Documentation

- update investigation with breakthrough findings from timing instrumentation


<a name="v0.4.2-debug.1"></a>
## [v0.4.2-debug.1] - 2025-11-22


<a name="v0.4.1"></a>
## [v0.4.1] - 2025-11-22

### Bug Fixes

- use async WebSocket Send() instead of blocking WriteMessage() ([#56](https://github.com/livefir/livetemplate/issues/56))


<a name="v0.4.0"></a>
## [v0.4.0] - 2025-11-22

### Code Refactoring

- **registry:** achieve Grade A code quality for async WebSocket ([#55](https://github.com/livefir/livetemplate/issues/55))


<a name="v0.3.2"></a>
## [v0.3.2] - 2025-11-20

### Bug Fixes

- convert validation error field names to lowercase


<a name="v0.3.1"></a>
## [v0.3.1] - 2025-11-19

### Bug Fixes

- send live tree update after upload completion ([#54](https://github.com/livefir/livetemplate/issues/54))
- send live tree update after upload completion ([#53](https://github.com/livefir/livetemplate/issues/53))

### Features

- Phoenix LiveView-inspired file upload system v0.3.0 ([#52](https://github.com/livefir/livetemplate/issues/52))


<a name="v0.3.0"></a>
## [v0.3.0] - 2025-11-12

### Bug Fixes

- use GOWORK=off in release script to avoid workspace issues
- address minor code review issues
- address code review feedback

### Code Refactoring

- make New() fail-fast on template parsing errors ([#51](https://github.com/livefir/livetemplate/issues/51))

### Documentation

- add optimization task list to performance bottlenecks
- add performance section to README
- add performance characteristics analysis
- add comprehensive benchmarking guide
- document performance bottlenecks from profiling
- add design and implementation plan

### Performance Improvements

- address code review recommendations
- establish performance baseline
- add end-to-end user journey benchmarks
- add end-to-end template benchmarks
- add Phase 4 (Render) and Phase 5 (Send) benchmarks
- add Phase 3 (Diff) benchmarks
- add Phase 2 (Build) benchmarks
- add Phase 1 (Parse) benchmarks


<a name="v0.2.1"></a>
## [v0.2.1] - 2025-11-11

### Bug Fixes

- allow template discovery in internal directories for multi kit support
- template auto-discovery for go run and lvt serve ([#49](https://github.com/livefir/livetemplate/issues/49))
- improve template auto-discovery robustness ([#47](https://github.com/livefir/livetemplate/issues/47))

### Documentation

- remove version-specific references from contributor walkthrough
- create comprehensive contributor walkthrough for 5-phase architecture
- simplify README to focus on core value proposition ([#48](https://github.com/livefir/livetemplate/issues/48))


<a name="v0.2.0"></a>
## [v0.2.0] - 2025-11-09

### Code Refactoring

- improve key generation and fingerprinting robustness
- complete Phase 2 - move 4 functions to internal packages ([#44](https://github.com/livefir/livetemplate/issues/44))
- align template.go with 5-phase architecture ([#43](https://github.com/livefir/livetemplate/issues/43))
- reduce public API surface area from 11 to 7 files ([#46](https://github.com/livefir/livetemplate/issues/46))
- **conditional:** eliminate duplication and improve error handling ([#40](https://github.com/livefir/livetemplate/issues/40))
- **context:** achieve Grade A code quality ([#31](https://github.com/livefir/livetemplate/issues/31))
- **field:** achieve Grade A code quality ([#36](https://github.com/livefir/livetemplate/issues/36))
- **fingerprint:** fix circular detection and improve robustness
- **helpers:** achieve Grade A code quality ([#35](https://github.com/livefir/livetemplate/issues/35))
- **parse:** achieve Grade A code quality ([#38](https://github.com/livefir/livetemplate/issues/38))
- **parse:** achieve Grade A code quality ([#41](https://github.com/livefir/livetemplate/issues/41))
- **prepare:** achieve Grade A code quality ([#34](https://github.com/livefir/livetemplate/issues/34))
- **range:** achieve Grade A code quality ([#37](https://github.com/livefir/livetemplate/issues/37))
- **range_ops:** achieve Grade A code quality ([#33](https://github.com/livefir/livetemplate/issues/33))
- **render:** achieve Grade A code quality ([#42](https://github.com/livefir/livetemplate/issues/42))
- **render:** performance, security, and quality improvements ([#27](https://github.com/livefir/livetemplate/issues/27))
- **template:** achieve Grade A- code quality with 5-phase architecture ([#45](https://github.com/livefir/livetemplate/issues/45))
- **tree_compare:** achieve Grade A code quality ([#32](https://github.com/livefir/livetemplate/issues/32))
- **types:** achieve Grade A quality with comprehensive tests and documentation
- **var_context:** achieve Grade A code quality ([#39](https://github.com/livefir/livetemplate/issues/39))
- **wrapper:** improve security, correctness, and robustness - Grade A ([#29](https://github.com/livefir/livetemplate/issues/29))


<a name="v0.1.3"></a>
## [v0.1.3] - 2025-11-07


<a name="ls"></a>
## [ls] - 2025-11-07

### Bug Fixes

- update release script for Go-only releases
- use absolute paths for replace directives in cross-repo tests
- resolve race conditions in RedisBroadcaster

### Code Refactoring

- API reduction for v0.2.0 - reduce public API surface area ([#23](https://github.com/livefir/livetemplate/issues/23))

### Documentation

- update RELEASE.md for Go-only releases

### Features

- Code review backlog implementation - Issues [#12](https://github.com/livefir/livetemplate/issues/12)-52 ([#24](https://github.com/livefir/livetemplate/issues/24))
- add comprehensive unit tests for internal packages ([#22](https://github.com/livefir/livetemplate/issues/22))

### BREAKING CHANGE


SessionStore methods now require context.Context parameter

This change adds proper context propagation throughout the session store
layer, enabling timeout control, cancellation, and tracing for all Redis
and session operations.

Changes to SessionStore interface:
- Get(ctx context.Context, groupID string) Stores
- Set(ctx context.Context, groupID string, stores Stores)
- Delete(ctx context.Context, groupID string)
- List(ctx context.Context) []string

Implementation updates:

MemorySessionStore:
- Accepts context parameter for interface compliance
- Operations are in-memory so context not used internally

RedisSessionStore:
- Uses provided context for all Redis operations
- getWithRetry and execPipelineWithRetry now respect context
- Context-aware sleep during retry backoff
- Checks for context cancellation before each retry attempt

Benefits:
- Redis operations can be cancelled mid-flight
- Timeouts are properly respected across retry logic
- Trace IDs and request metadata can be propagated
- Better observability in distributed systems
- Prevents resource leaks from hung operations

Migration guide:
- All SessionStore method calls must now pass context
- Use r.Context() in HTTP handlers for request-scoped context
- Use context.Background() for background operations
- Consider using context.WithTimeout() for bounded operations

### Breaking Change


No - added field to struct, backward compatible.

Note: Only one pre-existing test failure (TestTemplateGenerateTreeWithFuncMap)

🤖 Generated with [Claude Code](https://claude.com/claude-code)


<a name="v0.1.2"></a>
## [v0.1.2] - 2025-11-03

### Bug Fixes

- exclude extracted components from test workflow

### Features

- add cross-repository testing and local development workflows


<a name="v0.1.1"></a>
## [v0.1.1] - 2025-11-03


<a name="v0.1.0"></a>
## v0.1.0 - 2025-11-03

### Bug Fixes

- improve binary build and archive naming in release script
- increase test timeout in release script from 30s to 120s
- remove t.Parallel() from e2e tests to prevent timeout deadlocks
- resolve flaky TestConnectionLimits_ConcurrentAccess test
- add LVT_DEV_MODE to todos e2e test and update hardcoded client paths
- set LVT_DEV_MODE=true in test server startup
- correct observability API usage in example
- prevent accidental .golangci.yml restoration
- resolve all golangci-lint issues and enhance CI validation
- **lvt:** prevent auth tests from generating files in commands/internal ([#19](https://github.com/livefir/livetemplate/issues/19))
- **lvt:** move auth command under lvt gen subcommands ([#17](https://github.com/livefir/livetemplate/issues/17))

### Code Refactoring

- Phase 4 - Extract large functions into internal/diff package
- move remaining build functions to internal/build (Phase 3.2)
- move fingerprinting functions to internal/build
- integrate internal/parse package and remove tree_ast.go
- move tree types to internal/build package
- convert TDD tests to maintainable table-driven format

### Documentation

- Update documentation for repository restructuring
- Complete Milestone 2 - Horizontal Scaling Documentation & Implementation ([#20](https://github.com/livefir/livetemplate/issues/20))
- add first principles document and fix pre-commit hook ([#18](https://github.com/livefir/livetemplate/issues/18))
- update all docs to reflect v1.0 internal package architecture
- Phase 5 - Migration guide, observability example, and test fixtures
- mark refactoring as complete and ready to merge
- update REFACTORING_PROGRESS.md for Phase 3 completion
- update REFACTORING_PROGRESS.md for Phase 3.1 completion
- update REFACTORING_PROGRESS.md - Phase 2 complete
- add comprehensive observability guide
- comprehensive documentation audit and API accuracy fixes ([#4](https://github.com/livefir/livetemplate/issues/4))

### Features

- update release script to use GitHub CLI and publish npm package
- add testcontainers for Redis testing
- Add deployment stack generation (lvt gen stack) ([#21](https://github.com/livefir/livetemplate/issues/21))
- create internal/parse package for template parsing
- observability and architecture documentation
- add comprehensive TDD tests for all Go template actions
- implement comprehensive granular fragment support for all template actions
- implement granular range fragment system with CRUD operations
- **lvt:** add lvt gen auth command - Complete (Phases 1-6) ([#15](https://github.com/livefir/livetemplate/issues/15))


[Unreleased]: https://github.com/livefir/livetemplate/compare/v0.8.21...HEAD
[v0.8.21]: https://github.com/livefir/livetemplate/compare/v0.8.20...v0.8.21
[v0.8.20]: https://github.com/livefir/livetemplate/compare/v0.8.19...v0.8.20
[v0.8.19]: https://github.com/livefir/livetemplate/compare/v0.8.18...v0.8.19
[v0.8.18]: https://github.com/livefir/livetemplate/compare/v0.8.17...v0.8.18
[v0.8.17]: https://github.com/livefir/livetemplate/compare/v0.8.16...v0.8.17
[v0.8.16]: https://github.com/livefir/livetemplate/compare/v0.8.15...v0.8.16
[v0.8.15]: https://github.com/livefir/livetemplate/compare/v0.8.14...v0.8.15
[v0.8.14]: https://github.com/livefir/livetemplate/compare/v0.8.13...v0.8.14
[v0.8.13]: https://github.com/livefir/livetemplate/compare/v0.8.12...v0.8.13
[v0.8.12]: https://github.com/livefir/livetemplate/compare/v0.8.11...v0.8.12
[v0.8.11]: https://github.com/livefir/livetemplate/compare/v0.8.10...v0.8.11
[v0.8.10]: https://github.com/livefir/livetemplate/compare/v0.8.9...v0.8.10
[v0.8.9]: https://github.com/livefir/livetemplate/compare/v0.8.8...v0.8.9
[v0.8.8]: https://github.com/livefir/livetemplate/compare/v0.8.7...v0.8.8
[v0.8.7]: https://github.com/livefir/livetemplate/compare/v0.8.6...v0.8.7
[v0.8.6]: https://github.com/livefir/livetemplate/compare/v0.8.5...v0.8.6
[v0.8.5]: https://github.com/livefir/livetemplate/compare/v0.8.4...v0.8.5
[v0.8.4]: https://github.com/livefir/livetemplate/compare/v0.8.3...v0.8.4
[v0.8.3]: https://github.com/livefir/livetemplate/compare/v0.8.2...v0.8.3
[v0.8.2]: https://github.com/livefir/livetemplate/compare/v0.8.1...v0.8.2
[v0.8.1]: https://github.com/livefir/livetemplate/compare/v0.8.0...v0.8.1
[v0.8.0]: https://github.com/livefir/livetemplate/compare/v0.7.12...v0.8.0
[v0.7.12]: https://github.com/livefir/livetemplate/compare/v0.7.11...v0.7.12
[v0.7.11]: https://github.com/livefir/livetemplate/compare/v0.7.10...v0.7.11
[v0.7.10]: https://github.com/livefir/livetemplate/compare/v0.7.9...v0.7.10
[v0.7.9]: https://github.com/livefir/livetemplate/compare/v0.7.8...v0.7.9
[v0.7.8]: https://github.com/livefir/livetemplate/compare/v0.7.7...v0.7.8
[v0.7.7]: https://github.com/livefir/livetemplate/compare/v0.7.6...v0.7.7
[v0.7.6]: https://github.com/livefir/livetemplate/compare/v0.7.5...v0.7.6
[v0.7.5]: https://github.com/livefir/livetemplate/compare/v0.7.4...v0.7.5
[v0.7.4]: https://github.com/livefir/livetemplate/compare/v0.7.3...v0.7.4
[v0.7.3]: https://github.com/livefir/livetemplate/compare/v0.7.2...v0.7.3
[v0.7.2]: https://github.com/livefir/livetemplate/compare/v0.7.1...v0.7.2
[v0.7.1]: https://github.com/livefir/livetemplate/compare/v0.7.0...v0.7.1
[v0.7.0]: https://github.com/livefir/livetemplate/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/livefir/livetemplate/compare/v0.5.2...v0.6.0
[v0.5.2]: https://github.com/livefir/livetemplate/compare/v0.5.1...v0.5.2
[v0.5.1]: https://github.com/livefir/livetemplate/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/livefir/livetemplate/compare/v0.4.2-debug.2...v0.5.0
[v0.4.2-debug.2]: https://github.com/livefir/livetemplate/compare/v0.4.2-debug.1...v0.4.2-debug.2
[v0.4.2-debug.1]: https://github.com/livefir/livetemplate/compare/v0.4.1...v0.4.2-debug.1
[v0.4.1]: https://github.com/livefir/livetemplate/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/livefir/livetemplate/compare/v0.3.2...v0.4.0
[v0.3.2]: https://github.com/livefir/livetemplate/compare/v0.3.1...v0.3.2
[v0.3.1]: https://github.com/livefir/livetemplate/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/livefir/livetemplate/compare/v0.2.1...v0.3.0
[v0.2.1]: https://github.com/livefir/livetemplate/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/livefir/livetemplate/compare/v0.1.3...v0.2.0
[v0.1.3]: https://github.com/livefir/livetemplate/compare/ls...v0.1.3
[ls]: https://github.com/livefir/livetemplate/compare/v0.1.2...ls
[v0.1.2]: https://github.com/livefir/livetemplate/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/livefir/livetemplate/compare/v0.1.0...v0.1.1

---

## @livetemplate/client (TypeScript client)

_Canonical source: [livetemplate/client/CHANGELOG.md](https://github.com/livetemplate/client/blob/main/CHANGELOG.md)_

## [v0.8.40] - 2026-05-02

### Changes

- fix: always run fire-on-change directive scans (#107) (#114) (dff1765)



## [v0.8.39] - 2026-05-02

### Changes

- feat: per-op targeted DOM mutation for range diff ops (#107) (#108) (8f34384)



## [v0.8.38] - 2026-04-28

### Changes

- feat: HTML5 drag-and-drop event support (#101) (#106) (54ebdec)



## [v0.8.37] - 2026-04-28

### Changes

- fix(directives): remove empty style attr after highlight cleanup (#105) (187de33)



## [v0.8.36] - 2026-04-28

### Changes

- feat: lvt-scroll-away top edge for scroll-to-top buttons (#103) (661b8c2)



## [v0.8.35] - 2026-04-27

### Changes

- feat: reconnect WebSocket on visibility change (iOS background fix) (#99) (ef57b41)



## [v0.8.34] - 2026-04-22

### Changes

- fix(release): use explicit refspec to update tracking ref before sync check (#98) (51b5510)
- feat: data-lvt-target for scroll effects + lvt-scroll-away visibility toggle (#94) (860861b)



## [v0.8.33] - 2026-04-20

### Changes

- fix(morphdom): allow child updates inside open dialogs (#93) (ae78517)
- refactor(observer): replace scroll-sentinel id with lvt-scroll-sentinel attribute (#92) (e8666db)



## [v0.8.32] - 2026-04-20

### Changes

- fix(ws): detach handlers before closing socket on disconnect (#91) (f38891d)



## [v0.8.31] - 2026-04-20

### Changes

- fix(morphdom): preserve datalist while connected input is focused (#85) (ef9edea)



## [v0.8.30] - 2026-04-19

### Changes

- feat: hash-driven element activation for deep-linking (#86) (c85d36c)



## [v0.8.29] - 2026-04-18

### Changes

- fix(release): prompt before releasing with un-pushed local commits (54e5b08)
- fix(release): auto-switch to main before releasing (81d3c75)
- fix(morphdom): preserve checkbox/radio checked state across updates (#81) (ab879f7)
- Revert "fix(morphdom): preserve checkbox/radio checked state across updates" (adc9e55)
- fix(morphdom): preserve checkbox/radio checked state across updates (0d791e0)



## [v0.8.28] - 2026-04-18

### Changes

- fix(checkbox): send array of values for multiple same-name checkboxes (#78) (2bca20e)
- chore(release): v0.8.27 (72a925f)
- fix(link-interceptor): fix popstate back/forward navigation regression (053a6b7)



## [v0.8.27] - 2026-04-17

### Changes

- fix(link-interceptor): fix popstate back/forward navigation regression (053a6b7)



## [v0.8.26] - 2026-04-17

### Changes

- feat: lvt-ignore attributes, __navigate__ SPA nav, DOMParser script fix (#72) (966d65d)



## [Unreleased]

### Added

- `lvt-ignore` attribute: morphdom escape hatch that skips an element and its entire subtree during diff (equivalent to Phoenix LiveView's `phx-update="ignore"`). Checked on `fromEl` (live DOM) so both server templates and client JS can use it. Use `data-lvt-force-update` on the server's version to bypass and resume diffing.
- `lvt-ignore-attrs` attribute: morphdom escape hatch that preserves user-managed attributes (e.g. `open` on `<details>`) while still diffing children. Checked on `fromEl` for consistency with `lvt-ignore`. Use `data-lvt-force-update` to bypass.
- In-band `__navigate__` SPA navigation: same-pathname link clicks send `{action:"__navigate__", data:<params>}` over the existing WebSocket instead of fetching new HTML. Requires server-side support (livetemplate/livetemplate#344).
- DOMParser fallback in `updateDOM`: HTML containing `<script>` tags is now parsed via `DOMParser` to avoid a Chrome `innerHTML` bug that creates phantom duplicate DOM nodes after script tags.

### Breaking Changes

- **Cross-pathname same-handler navigation now always reconnects.** Previously, if two routes shared the same `data-lvt-id`, navigating between them would do an in-place DOM swap without reconnecting. This fast path has been removed; all cross-pathname navigations (regardless of handler ID) now trigger a full WebSocket reconnect. This is the correct behavior — same-ID across paths means two distinct routes, and `sendNavigate` cannot express a path change. **If your app shares a `data-lvt-id` across routes, expect a reconnect flash where there was none before.**

### Deployment note

The `__navigate__` in-band action is a no-op on server versions before livetemplate/livetemplate#344. Deploy the server update before or simultaneously with this client version to avoid same-pathname link clicks sending an unrecognized WebSocket action.

## [v0.8.25] - 2026-04-15

### Changes

- fix(ci): upgrade npm in publish workflow for OIDC trusted publishing (74bd7c5)



## [v0.8.24] - 2026-04-15

### Changes

- chore: gitignore .claude/scheduled_tasks.lock (28e30e7)
- ci: publish to npm via OIDC trusted publishing (#71) (9053e38)



## [v0.8.23] - 2026-04-14

### Changes

- fix: cross-handler SPA nav, infinite scroll race, animation cleanup (#69) (e40f4b1)



## [v0.8.22] - 2026-04-13

### Changes

- fix: cross-handler SPA navigation, navigation edge cases, and Tier 1 file uploads (#58) (798ca90)



## [v0.8.21] - 2026-04-11

### Changes

- feat: polyfill command/commandfor for cross-browser dialog support (#57) (565176c)



## [v0.8.20] - 2026-04-10

### Changes

- feat: extend livetemplate.css with shared utilities and chat styles (#54) (c12e1e8)



## [v0.8.19] - 2026-04-05

### Changes

- feat: add data-lvt-target for cross-element targeting in lvt-el: methods (#53) (89aa203)



## [Unreleased]

### Added

- feat: `data-lvt-target` attribute for cross-element targeting — `lvt-el:` methods can now operate on a different element via `#id` or `closest:selector`

## [v0.8.18] - 2026-04-05

### Changes

- chore(release): v0.8.17 (d6b41a4)
- feat: extend lvt-el: to support native DOM event triggers (#49) (ddf92c2)
- fix: form.name DOM shadowing + skip File objects in FormData parsing (58cf0c2)



## [Unreleased]

### Added

- feat: `lvt-el:{method}:on:{event}` now supports any native DOM event as trigger (click, focusin, focusout, mouseenter, mouseleave, keydown, etc.) — no server round-trip, CSP-safe
- feat: `lvt-fx:{effect}:on:{event}` supports DOM event triggers (e.g., `lvt-fx:highlight:on:click="flash"`) and lifecycle triggers (e.g., `lvt-fx:highlight:on:success="flash"`)

## [v0.8.17] - 2026-04-05

### Changes

- fix: form.name DOM shadowing + skip File objects in FormData parsing (58cf0c2)


## [v0.8.16] - 2026-04-04

### Changes




## [v0.8.15] - 2026-04-04

### Changes

- feat: Tier 1 file uploads — HTTP fetch for forms with file inputs (387e2fe)



## [v0.8.14] - 2026-04-04

### Changes

- fix: lvt-form:action routing, lvt-nav:no-intercept, unreserve action field (#45) (6598832)



## [v0.8.13] - 2026-04-04

### Changes

- Phase 1A: Client attribute reduction — generic event router + removals (#44) (5328b85)



## [v0.8.12] - 2026-04-02

### Changes

- feat: add client-side toast directive (#42) (5f3a1e2)



## [v0.8.10] - 2026-03-30

### Changes

- feat: auto-wire Change() for <select> elements & fix cursor reset (#40) (b2ddc56)



## [v0.8.9] - 2026-03-30

### Changes

- fix: harden release script with clean build and verification (#37) (b2b59e4)
- fix: pull latest from remote before releasing (#36) (74d6f74)



## [v0.8.8] - 2026-03-29

### Changes




## [v0.8.7] - 2026-03-27

### Changes

- feat: formless standalone button support (#29) (55c4dca)
- feat: implement Change() auto-inference client support (#25) (64476eb)
- feat: auto-intercept forms for progressive complexity (#23) (b4b9672)
- fix: use wss:// for WebSocket on HTTPS pages (#22) (398f752)



## [v0.8.6] - 2026-03-26

### Changes

- fix: use current branch name in release script instead of hardcoded main (97c79ec)
- feat: implement Change() auto-inference client support (#25) (64476eb)
- feat: auto-intercept forms for progressive complexity (#23) (b4b9672)
- fix: use wss:// for WebSocket on HTTPS pages (#22) (398f752)



## [v0.8.5] - 2026-03-25

### Changes

- feat: auto-intercept forms for progressive complexity (#23) (b4b9672)
- fix: use wss:// for WebSocket on HTTPS pages (#22) (398f752)



## [v0.8.4] - 2026-03-14

### Changes




## [v0.8.3] - 2026-02-27

### Changes

- chore: upgrade Go to 1.26 in CI workflows (#21) (dcf15f7)



## [v0.8.2] - 2026-02-02

### Changes

- fix: use deep merge for update operations to preserve statics (#20) (a71c25d)
- fix: preserve large integers as strings to prevent precision loss (#19) (8c13758)
- feat: support auto-generated _k keys in range item matching (#18) (5d5f727)



## [v0.8.0] - 2026-01-18

### Changes




## [v0.7.12] - 2026-01-10

### Changes

- fix(event-delegation): debounce captures latest input value (3d5b5e9)
- fix(client): skip debounce for search event (clear button) (35adeb7)
- fix(client): handle search event for input type="search" clear button (9afc00e)



## [v0.7.11] - 2026-01-05

### Changes

- fix(tree-renderer): handle range→non-range transitions in deepMergeTreeNodes (#16) (f95a08b)



## [v0.7.10] - 2026-01-04

### Changes

- feat(modal): add data-modal-close-action attribute support (#15) (c8321b3)
- fix(ci): increase max-turns and simplify review prompt (bba2fe7)
- fix(ci): use stable claude-code-action v1 with correct inputs (eb2d2e4)
- feat(modal): add data-modal-close-action attribute support (8bf64f4)
- fix(ci): use correct input parameter for claude-code-action (50c8a1f)



## [v0.7.9] - 2026-01-03

### Changes

- fix(release): sync with full core library version (9d5be47)
- fix(modal): simplify modal close button handling (#14) (404b210)



## [v0.7.7] - 2025-12-26

### Changes

- fix: query params in WebSocket URL + password field handling (#13) (42604a6)



## [v0.7.4] - 2025-12-23

### Changes

- add .npmrc (0e1ef6e)



## [v0.7.3] - 2025-12-22

### Changes

- fix: support heterogeneous range items with per-item statics (#12) (badad08)
- fix: handle plain data objects gracefully in tree renderer (#11) (c64fb24)
- feat: client updates for livepage features (#10) (cb6af54)
- fix: apply differential ops to existing range structures (#9) (50a3ebc)
- fix: handle objects with only numeric keys in renderValue (#8) (b1c7827)
- feat: add lvt-focus-trap and lvt-autofocus attributes (#7) (7b14402)
- feat: add reactive attributes for action lifecycle events (#6) (46e2065)



## [v0.7.2] - 2025-12-20

### Changes

- fix: support heterogeneous range items with per-item statics (#12)
- fix: handle plain data objects gracefully in tree renderer (#11) (c64fb24)
- feat: client updates for livepage features (#10) (cb6af54)
- fix: apply differential ops to existing range structures (#9) (50a3ebc)
- fix: handle objects with only numeric keys in renderValue (#8) (b1c7827)
- feat: add lvt-focus-trap and lvt-autofocus attributes (#7) (7b14402)
- feat: add reactive attributes for action lifecycle events (#6) (46e2065)



## [v0.7.1] - 2025-12-14

### Changes

- fix: apply differential ops to existing range structures (#9) (50a3ebc)
- fix: handle objects with only numeric keys in renderValue (#8) (b1c7827)
- feat: add lvt-focus-trap and lvt-autofocus attributes (#7) (7b14402)
- feat: add reactive attributes for action lifecycle events (#6) (46e2065)



## [v0.7.0] - 2025-12-10

### Changes




## [v0.7.0] - 2025-12-10

### Changes

- feat: add lvt-focus-trap and lvt-autofocus attributes (#7) (7b14402)
- feat: add reactive attributes for action lifecycle events (#6) (46e2065)



## [v0.4.1] - 2025-11-27

### Changes

- feat: improve test coverage from 38% to 60% (#4) (9755643)
- Add Claude Code GitHub Workflow (#5) (79e3d0b)



## [v0.4.0] - 2025-11-22

### Changes

- fix: use numeric constant instead of WebSocket.OPEN (#3) (6462ccb)
- fix(upload): clear file input after successful upload to prevent duplicate uploads (af6f7aa)
- feat(upload): implement AutoUpload config and form submit trigger (b77e1ff)


Initial release of @livetemplate/client as a standalone package.

### Features

- TypeScript client for LiveTemplate tree-based updates
- WebSocket transport for real-time updates
- DOM morphing with morphdom
- Focus management and form lifecycle
- Event delegation
- Modal management

---

## lvt (CLI)

_Canonical source: [livetemplate/lvt/CHANGELOG.md](https://github.com/livetemplate/lvt/blob/main/CHANGELOG.md)_

## [v0.1.6] - 2026-05-02

### Changes

- feat(testing): add RecordWSFrames as canonical WS-frame capture helper (#317) (093d95f)
- test(e2e): regenerate todos goldens for streaming range Phase 4 (#316) (8f27b68)



## [v0.1.5] - 2026-04-27

### Changes

- fix(testing): respect test PORT/LVT_DEV_MODE over inherited env (#315) (7275c8f)



## [v0.1.4] - 2026-04-26

### Changes

- fix(testing): un-throttle Chrome container so server-pushed renders apply at 1Hz (#314) (e187813)
- refactor: replace scroll-sentinel id with lvt-scroll-sentinel attribute (#313) (03bca0d)
- feat: migrate add modal to native dialog + command/commandfor (#312) (1f01b84)
- fix: gracefully handle missing CSS instead of panicking (2f54313)



## [v0.1.3] - 2026-04-10

### Changes

- fix: exclude .worktrees/ from nested module tag discovery in release script (0e77a6d)
- fix: save screenshot to working directory for Claude CLI sandbox access (2ac2581)
- refactor: use claude CLI instead of Anthropic SDK for visual checks (ae46a2f)
- feat: add ValidateScreenshotWithLLM() for LLM-powered visual UI review (03fb143)
- feat: add ValidatePicoCSS() chromedp action for Pico CSS convention checking (82dce8c)
- fix(ci): fix YAML parse error in components-independence workflow (#310) (e7c8582)
- feat(testing): add ServeCSS for shared LiveTemplate CSS (1e9976a)



## [v0.1.2] - 2026-04-05

### Changes

- feat!: attribute reduction — client-side open/close + template migration (#292) (eb9cd54)
- fix(release): auto-tag nested Go modules on release (9843894)



## [v0.1.0] - 2025-11-03

Initial release of LVT CLI as a standalone package extracted from the LiveTemplate monorepo.

### Features

- **Code Generation**
  - Generate CRUD resources with models, handlers, and views
  - Generate standalone views
  - Generate complete applications
  - Field parsing and validation

- **Kit System**
  - Built-in kits: Tailwind CSS, Bulma, Pico CSS, None
  - Kit cascade: Project → User → System
  - ~60 CSS helper methods per kit
  - Component templates
  - Generator templates
  - Kit creation and customization tools
  - Kit validation

- **Development Server**
  - Hot reload via WebSocket
  - File watching with fsnotify
  - Automatic browser refresh
  - Serves static assets
  - Live template rendering

- **Database Tools**
  - Migration creation and management
  - Seeder creation and execution
  - SQLite and modernc.org/sqlite support
  - Migration status tracking

- **Stack Generators**
  - Docker configurations
  - Systemd service files
  - Deployment scripts

- **Interactive UI**
  - Terminal UI for app creation
  - Resource generator wizard
  - View generator wizard
  - Built with Bubble Tea and Lipgloss

- **Testing Utilities**
  - E2E test helpers
  - Chromedp integration for browser testing
  - Test server utilities
  - Golden file testing

### Infrastructure

- **Release Automation**: Automated release script with version synchronization
- **CI/CD**: GitHub Actions workflows
- **Pre-commit Hooks**: Go formatting, linting, and testing
- **GoReleaser**: Multi-platform binary builds
- **Version Tracking**: VERSION file for release management

### Documentation

- Complete README with examples
- Contributing guidelines
- Version synchronization strategy with core library

### Related Versions

- Core Library: v0.1.0
- Client Library: v0.1.0
- Examples: v0.1.0

---

## Version Synchronization

LVT follows the LiveTemplate core library's major.minor version (X.Y):

- Patch versions (X.Y.Z) are independent
- Minor/major versions must match core library
- See README.md for details

---

## examples (apps)

_Canonical source: [livetemplate/examples/CHANGELOG.md](https://github.com/livetemplate/examples/blob/main/CHANGELOG.md)_

## [v0.1.0] - 2025-11-03

Initial release of LiveTemplate Examples as a standalone repository.

### Examples Included

1. **Counter** - Simple state management with reactive updates
2. **Chat** - Multi-user chat application with WebSocket
3. **Todos** - Full CRUD application with SQLite database
4. **Graceful Shutdown** - Proper server shutdown handling
5. **Observability** - Logging, metrics, and tracing
6. **Testing** - E2E testing patterns with Chromedp
7. **Production** - Production deployment configuration
8. **Trace Correlation** - Request tracing and correlation IDs

### Features

- **Self-contained Examples**: Each example has its own go.mod
- **E2E Testing**: Chromedp-based browser tests
- **CDN Integration**: Examples use CDN version of client library
- **Documentation**: README for each example with setup instructions
- **Production Patterns**: Real-world deployment examples

### Infrastructure

- Go module configuration for all examples
- Import paths updated for extracted repositories
- .gitignore for build artifacts
- VERSION file for release tracking

### Documentation

- Complete main README with all examples
- Contributing guidelines for new examples
- Individual README files per example

### Related Versions

- Core Library: v0.1.0
- Client Library: v0.1.0
- LVT CLI: v0.1.0

---

## Version Synchronization

Examples follow the LiveTemplate core library's major.minor version (X.Y):

- Patch versions (X.Y.Z) are independent
- Minor/major versions must match core library
- See README.md for details
