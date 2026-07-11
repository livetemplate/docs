// Package livedashboard is the docs recipe for out-of-band fan-out: refreshing
// many live sessions from a single background goroutine WITHOUT stashing a
// registry of Session handles.
//
// The shape it teaches (see docs/references/server-actions.md,
// "Fanning out to many sessions"):
//
//   - Every connection joins the shared "dashboard" developer topic in Mount
//     (ctx.Subscribe). Developer topics are deny-all by default, so Handler
//     authorizes it with WithTopicACL — the ACL is the security boundary a
//     cross-user topic requires.
//
//   - One process-wide time.Ticker (started in Handler) updates shared metrics
//     and calls handler.Publish("dashboard", "Refresh", nil). That out-of-band
//     Publish — no Context, safe from any goroutine — re-runs Refresh on every
//     subscriber, in every session group, and re-renders. No per-connection
//     timer, no OnConnect/OnDisconnect bookkeeping, no dead-handle pruning.
//
// Contrast with examples/patterns (per-connection session.TriggerAction timers)
// and examples/shared-notepad (in-band ctx.Publish peer sync triggered by a
// client action): this recipe is the ONE-goroutine-refreshes-EVERYONE case.
//
// There is no main() here. Production runs via the docs single-binary
// container, mounted by cmd/site at /apps/live-dashboard/. cmd/main.go wraps
// Handler in a standalone listener for local runs and the e2e suite.
package livedashboard

import (
	"embed"
	"net/http"
	"time"

	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"
)

//go:embed dashboard.tmpl
var templateFS embed.FS

// refreshInterval is how often the background goroutine pushes an update to
// every connected browser. One second keeps the demo (and its e2e test) snappy.
const refreshInterval = time.Second

// Handler returns the live-dashboard app as an http.Handler ready to mount, and
// starts its single background-refresh goroutine. AnonymousAuthenticator gives
// each browser its own session group — yet every group joins the same
// "dashboard" topic, so the out-of-band Publish reaches them all. Callers supply
// environment-specific options (origin allowlists, dev mode) via opts.
func Handler(opts ...livetemplate.Option) http.Handler {
	controller := &DashboardController{metrics: &metrics{updatedAt: "—"}}

	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFS(templateFS, "dashboard.tmpl"),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
		// Authorize the one shared topic this recipe uses. Developer topics are
		// deny-all by default; this ACL is the explicit boundary.
		livetemplate.WithTopicACL(func(topic, _ string, _ *http.Request) (bool, error) {
			return topic == dashboardTopic, nil
		}),
	}
	tmpl := livetemplate.Must(livetemplate.New("live-dashboard", append(baseOpts, opts...)...))

	live := tmpl.Handle(controller, livetemplate.AsState(&DashboardState{}))

	// One process-wide goroutine drives every connected session.
	go runBackgroundRefresh(live, controller.metrics)

	mux := http.NewServeMux()
	mux.Handle("/", live)
	mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
	mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)
	return mux
}

// >>> region:ticker
// runBackgroundRefresh is the single fan-out driver. It owns no connections and
// holds no Session handles — it just mutates shared state and Publishes to the
// topic. Every subscriber, in every session group and (with a PubSub
// broadcaster) on every instance, re-runs Refresh and re-renders.
func runBackgroundRefresh(handler livetemplate.LiveHandler, m *metrics) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	for range ticker.C {
		m.advance()
		// A transport failure to a remote instance is logged, not returned;
		// the only returned error is a malformed topic/action (a programmer
		// bug), so surfacing it would just crash a demo. Ignore deliberately.
		_ = handler.Publish(dashboardTopic, "Refresh", nil)
	}
}

// <<< region:ticker
