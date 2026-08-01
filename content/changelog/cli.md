---
title: "Changelog"
source_repo: "https://github.com/livetemplate/lvt"
source_path: "CHANGELOG.md"
source_ref: "v0.2.0"
source_commit: "2a88feb18e38e6e1b76b12952a1dc68cce094c63"
---

# Changelog

All notable changes to the LVT CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2026-07-18

### Changes

- chore(deps): adopt livetemplate v0.19.0 + land release-gated C8 goldens and e2e (#337) (bc1659c)
- test(e2e): migrate older WS-ready guards to WaitForWebSocketReady (closes #335) (#336) (f3a6dff)
- test(e2e): chromedp regression for livetemplate#414 + bump to v0.11.1 (#334) (2a7b37a)
- docs(roadmap): clarify lvt transport broadcast vs livetemplate topic publish (#333) (8cfe136)



## [v0.1.8] - 2026-05-21

### Changes

- Phase 6: lvt#331 fix + GOWORK=off + PollUntil helper + V14 key-assertion (v0.1.8-prep) (#332) (e3b78fa)



## [v0.1.7] - 2026-05-20

### Changes

- feat(broadcast)!: Phase 5 — pin livetemplate v0.10.0; rename Broadcast→ReloadClients; integrate V14 e2e (#330) (92b0848)
- chore: clear pre-existing lint baseline + add .golangci.yml config (#329) (3bd71f2)
- chore(scripts): fix pre-commit lint invocation for golangci-lint v2 (#328) (d2ffa33)
- test(e2e): chromedp coverage for lifecycle ergonomics (livetemplate#408) (#323) (e9f6628)
- test(e2e): pin explicit-submitter round-trip via chromedp (#321) (2ead2a9)
- chore: bump livetemplate to v0.9.0 (#320) (9cdf846)
- docs(README): cross-link to https://livetemplate.fly.dev docs site (#319) (f488d41)
- ci: dispatch livetemplate/docs sync on release tag (#318) (9c2715b)



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
