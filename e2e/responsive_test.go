// Browser-driven responsive layout tests. These exist so that the class
// of CSS breakage we hit across v0.1.10–v0.1.14 (mobile horizontal
// overflow, desktop body width math, mobile body collapsing to 33px)
// fails BEFORE the next deploy, not after.
//
// They're driven from livetemplate/docs because the bugs they catch
// only manifest as cross-page integration: a tinkerdown CSS rule looks
// fine in unit tests against a one-page render, then breaks once you
// browse a real site at multiple viewports. Running here means every
// docs deploy is gated on the same checks an angry Daisy would do
// manually with a phone.
package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

type responsiveCase struct {
	name        string
	width       int64
	height      int64
	mobile      bool
	minBodyW    int64 // body must be at least this wide (catches the 33px collapse)
	maxBodyW    int64 // body must be at most this wide (catches over-expansion)
}

// representativePages lists URLs to sweep at every viewport. Picked to
// cover the bug surface exposed by past regressions: long inline code
// (/cli/), wide tables (/reference/template-support-matrix), pre blocks
// with negative-margin bleed (most pages), prerendered SVG diagrams
// (/recipes/architecture-flow), and a thin landing page (/).
var representativePages = []string{
	"/",
	"/cli/",
	"/reference/template-support-matrix",
	"/recipes/architecture-flow",
	"/getting-started/install",
}

// TestResponsiveLayoutAcrossViewports walks each representativePage at
// each viewport and asserts:
//
//   - documentElement.scrollWidth ≤ clientWidth+1 (no horizontal overflow,
//     the document-level signal — body.scrollWidth alone misses overflow
//     caused by body's own margin)
//   - body width is within plausible bounds (catches both the 33px
//     mobile collapse and the 1640px desktop over-expansion)
//   - clientWidth matches the requested viewport (sanity check on emulation)
//
// Skipped if E2E_SKIP_RESPONSIVE is set — for fast-iteration runs.
func TestResponsiveLayoutAcrossViewports(t *testing.T) {
	cases := []responsiveCase{
		{name: "desktop", width: 1280, height: 800, mobile: false, minBodyW: 700, maxBodyW: 1300},
		{name: "tablet", width: 768, height: 1024, mobile: true, minBodyW: 400, maxBodyW: 800},
		{name: "iphone-14", width: 393, height: 852, mobile: true, minBodyW: 350, maxBodyW: 410},
	}

	for _, vc := range cases {
		for _, path := range representativePages {
			vc, path := vc, path
			t.Run(fmt.Sprintf("%s/%s", vc.name, slug(path)), func(t *testing.T) {
				ctx, cancel := newCtx(t)
				defer cancel()

				var (
					clientWidth int64
					htmlScroll  int64
					bodyScroll  int64
					bodyWidth   int64
				)

				err := chromedp.Run(ctx,
					emulation.SetDeviceMetricsOverride(vc.width, vc.height, 1.0, vc.mobile),
					chromedp.Navigate(baseURL()+path),
					chromedp.WaitReady("body", chromedp.ByQuery),
					chromedp.Sleep(800*time.Millisecond),
					chromedp.Evaluate(`document.documentElement.clientWidth`, &clientWidth),
					chromedp.Evaluate(`document.documentElement.scrollWidth`, &htmlScroll),
					chromedp.Evaluate(`document.body.scrollWidth`, &bodyScroll),
					chromedp.Evaluate(`document.body.getBoundingClientRect().width`, &bodyWidth),
				)
				if err != nil {
					t.Fatalf("navigate %s @ %s: %v", path, vc.name, err)
				}

				if clientWidth != vc.width {
					t.Errorf("clientWidth = %d, want %d (viewport emulation drifted)", clientWidth, vc.width)
				}
				if htmlScroll > clientWidth+1 {
					t.Errorf("documentElement.scrollWidth = %d > clientWidth %d (horizontal page overflow); body.scrollWidth = %d, body.width = %d",
						htmlScroll, clientWidth, bodyScroll, bodyWidth)
				}
				if bodyWidth < vc.minBodyW {
					t.Errorf("body.width = %d, expected ≥ %d — body has collapsed (likely a width: calc(...) inheriting wrong sidebar offset)",
						bodyWidth, vc.minBodyW)
				}
				if bodyWidth > vc.maxBodyW {
					t.Errorf("body.width = %d, expected ≤ %d — body has over-expanded past the viewport (likely a missing width override under a margin-left rule)",
						bodyWidth, vc.maxBodyW)
				}
			})
		}
	}
}

// inViewportContext applies viewport emulation BEFORE Navigate, which
// is the only way to get tinkerdown's media-query CSS to evaluate
// against the right width on first paint.
func inViewportContext(width, height int64, mobile bool) chromedp.Tasks {
	return chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(width, height, 1.0, mobile),
	}
}

func slug(p string) string {
	if p == "/" {
		return "root"
	}
	out := make([]byte, 0, len(p))
	for i := 0; i < len(p); i++ {
		c := p[i]
		if c == '/' {
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
			continue
		}
		out = append(out, c)
	}
	if len(out) > 0 && out[0] == '-' {
		out = out[1:]
	}
	if len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return string(out)
}

// Avoid "imported and not used: context" — context is used inside
// chromedp.Run via the closure but the linker sometimes flags it on
// unused-import scans during incremental compile.
var _ = context.Background
