// The calm theme's acceptance criteria (issue #133), as assertions.
//
// These are deliberately separate from responsive_test.go. That file owns
// 1280/768/393 and a different question — whether the page survives a phone.
// This one owns the two widths the design was drawn against, 924px and 1440px,
// and the three rules the design states as absolutes:
//
//   - the page never scrolls sideways; a wide code block scrolls inside itself
//   - no text below 11.5px
//   - no text at or below 13px lighter than #6B6862
//
// The last two are measured, not eyeballed, because the shipped theme sizes
// most chrome in rem against a root that tinkerdown pins to 87.5% — three
// elements were already under the floor before this theme landed, and nothing
// but a computed-style read would have caught them.
//
// Run locally against a container or a tinkerdown serve of content/:
//
//	E2E_BASE_URL=http://127.0.0.1:8080 go test ./e2e -run TestCalmTheme
package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

// calmPages covers the landing plus one page of each archetype: A (docs
// article), B (section index) and C (reference entry).
var calmPages = []string{
	"/",
	"/getting-started/introduction",
	"/recipes/",
	"/reference/api",
}

// calmWidths are the two the design specifies. 924 is the interesting one: it
// is below the landing's 1100px rail breakpoint and narrow enough that a code
// block sized for a wider measure pushes the document wide.
var calmWidths = []int64{924, 1440}

// calmScopes are the containers the design owns. Scoping matters: a
// document-wide scan also picks up chrome this issue does not govern, and a
// test that cannot be made to pass gets weakened rather than fixed.
const calmScopes = `['.content-wrapper','nav.breadcrumbs','#tinkerdown-sidebar','.page-nav',
                     '.page-edit-link','.page-source-meta','.doc','.rail','header.site','footer.site']`

type calmOffender struct {
	ID    string  `json:"id"`
	Size  float64 `json:"size"`
	Color string  `json:"color"`
}

type calmReport struct {
	ScrollWidth float64        `json:"scrollWidth"`
	ClientWidth float64        `json:"clientWidth"`
	TooSmall    []calmOffender `json:"tooSmall"`
	TooLight    []calmOffender `json:"tooLight"`
	PreOverflow []string       `json:"preOverflow"`
}

// calmProbe walks every visible text-bearing element in the design-owned
// chrome and reports the ones that break a rule, plus any <pre> wide enough to
// overflow that is not set to scroll.
const calmProbe = `(() => {
  // WCAG relative luminance: "lighter" has to be measured, not compared as a
  // hex string — #5C9169 looks darker than #6B6862 and is not.
  const lum = c => {
    const m = c.match(/\d+(\.\d+)?/g); if (!m) return 1;
    const [r,g,b] = m.slice(0,3).map(v => { v = v/255; return v <= 0.03928 ? v/12.92 : Math.pow((v+0.055)/1.055, 2.4) });
    return 0.2126*r + 0.7152*g + 0.0722*b;
  };
  const FLOOR = lum('rgb(107, 104, 98)');   // #6B6862
  const tooSmall = [], tooLight = [];
  for (const sel of ` + calmScopes + `) {
    const root = document.querySelector(sel); if (!root) continue;
    for (const e of [root, ...root.querySelectorAll('*')]) {
      const cs = getComputedStyle(e);
      if (cs.display === 'none' || cs.visibility === 'hidden') continue;
      if (!e.getClientRects().length) continue;
      // Only elements holding their own text: an element's font-size is
      // irrelevant if every character in it belongs to a child.
      if (![...e.childNodes].some(n => n.nodeType === 3 && n.textContent.trim())) continue;
      const size = parseFloat(cs.fontSize);
      const cls = (e.className && typeof e.className === 'string') ? '.' + e.className.trim().split(/\s+/).join('.') : '';
      const id = e.tagName.toLowerCase() + cls;
      if (size < 11.5) tooSmall.push({ id, size, color: cs.color });
      if (size <= 13 && lum(cs.color) > FLOOR + 1e-9) tooLight.push({ id, size, color: cs.color });
    }
  }
  const preOverflow = [...document.querySelectorAll('pre')].filter(p => {
    const ox = getComputedStyle(p).overflowX;
    return p.scrollWidth > p.clientWidth + 1 && ox !== 'auto' && ox !== 'scroll';
  }).map(p => p.className || 'pre');
  const uniq = a => [...new Map(a.map(x => [x.id + ':' + x.size, x])).values()];
  return JSON.stringify({
    scrollWidth: document.documentElement.scrollWidth,
    clientWidth: document.documentElement.clientWidth,
    tooSmall: uniq(tooSmall), tooLight: uniq(tooLight), preOverflow,
  });
})()`

func TestCalmThemeAcceptance(t *testing.T) {
	for _, width := range calmWidths {
		for _, path := range calmPages {
			width, path := width, path
			t.Run(fmt.Sprintf("%d/%s", width, slug(path)), func(t *testing.T) {
				ctx, cancel := newCtx(t)
				defer cancel()

				var raw string
				if err := chromedp.Run(ctx,
					emulation.SetDeviceMetricsOverride(width, 900, 1.0, false),
					chromedp.Navigate(baseURL()+path),
					chromedp.WaitReady("body", chromedp.ByQuery),
					// The landing's live embeds mount and grow after load; measuring
					// before they settle reads a page that never existed.
					chromedp.Sleep(2500*time.Millisecond),
					chromedp.Evaluate(calmProbe, &raw),
				); err != nil {
					t.Fatalf("probe %s @ %dpx: %v", path, width, err)
				}

				var r calmReport
				if err := json.Unmarshal([]byte(raw), &r); err != nil {
					t.Fatalf("decode probe result for %s @ %dpx: %v\n%s", path, width, err, raw)
				}

				if r.ClientWidth != float64(width) {
					t.Fatalf("clientWidth = %.0f, want %d (viewport emulation drifted)", r.ClientWidth, width)
				}
				// The +1 absorbs sub-pixel rounding, matching responsive_test.go.
				if r.ScrollWidth > r.ClientWidth+1 {
					t.Errorf("documentElement.scrollWidth = %.0f > clientWidth %.0f — the page scrolls sideways. "+
						"Usual cause: a grid track declared 1fr instead of minmax(0,1fr), so a long "+
						"monospace line sets the track's min-content width.", r.ScrollWidth, r.ClientWidth)
				}
				for _, o := range r.TooSmall {
					t.Errorf("%s is %.2fpx — below the 11.5px floor. Set it in px: the shipped theme "+
						"sizes chrome in rem against a 14px root, so a rem value that looks fine is not.", o.ID, o.Size)
				}
				for _, o := range r.TooLight {
					t.Errorf("%s is %.2fpx in %s — text at or below 13px must be #6B6862 or darker. "+
						"Do not lighten it back; the muted greys failed contrast at 12px in the first pass.", o.ID, o.Size, o.Color)
				}
				for _, p := range r.PreOverflow {
					t.Errorf("<pre class=%q> is wider than its box but does not scroll — "+
						"a code block has to scroll inside itself, not widen the page.", p)
				}
			})
		}
	}
}

// TestCalmThemeGopherIsVendored guards the one asset rule in the issue: the
// design prototype hotlinked the gopher from raw.githubusercontent.com, which
// CSP img-src 'self' would block and which makes the page depend on GitHub.
func TestCalmThemeGopherIsVendored(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var srcs []string
	var attribution string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(`[...document.images].map(i => i.getAttribute('src'))`, &srcs),
		chromedp.Evaluate(`(document.querySelector('footer.site .attrib')||{}).innerText || ''`, &attribution),
	); err != nil {
		t.Fatalf("load landing: %v", err)
	}

	var gopher string
	for _, s := range srcs {
		if strings.HasSuffix(s, ".svg") {
			gopher = s
		}
		if strings.Contains(s, "raw.githubusercontent.com") {
			t.Errorf("image %q is hotlinked from GitHub — vendor it into content/assets/", s)
		}
	}
	if gopher == "" {
		t.Fatalf("no SVG image on the landing; images = %v", srcs)
	}
	if !strings.Contains(gopher, "/assets/") {
		t.Errorf("gopher src = %q, want it served from /assets/", gopher)
	}
	// CC BY 3.0 requires attribution to travel with the artwork.
	for _, want := range []string{"Renée French", "Takuya Ueda", "CC BY 3.0"} {
		if !strings.Contains(attribution, want) {
			t.Errorf("footer attribution is missing %q (got %q)", want, attribution)
		}
	}
}
