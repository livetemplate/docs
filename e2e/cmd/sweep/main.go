// Sweep tool: fetch /sitemap.xml from a docs site, visit every URL at
// desktop + iphone-14 viewports, capture screenshots, and report any
// page that has horizontal overflow or missing-content symptoms.
//
// Programmatic detection runs first; visually inspect only the pages
// flagged below the report. This keeps the manual review loop tight
// across ~50 URLs.
//
// Usage:
//
//	go run ./cmd/sweep https://livetemplate.fly.dev
//
// Outputs:
//   - /tmp/lvt-sweep/<viewport>/<slug>.png
//   - /tmp/lvt-sweep/report.txt
package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type sitemap struct {
	URLs []sitemapURL `xml:"url"`
}

type viewport struct {
	name   string
	width  int64
	height int64
	mobile bool
}

type pageReport struct {
	url            string
	viewport       string
	clientWidth    int64
	scrollWidth    int64
	bodyText       int64
	hasHorizontal  bool
	missingContent bool
	loadError      string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: sweep <base-url>")
		os.Exit(1)
	}
	base := strings.TrimRight(os.Args[1], "/")

	urls, err := fetchSitemap(base + "/sitemap.xml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch sitemap:", err)
		os.Exit(1)
	}
	sort.Strings(urls)
	fmt.Fprintf(os.Stderr, "discovered %d URLs in sitemap\n", len(urls))

	outDir := "/tmp/lvt-sweep"
	if err := os.RemoveAll(outDir); err != nil {
		fmt.Fprintln(os.Stderr, "wipe outdir:", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}

	viewports := []viewport{
		{name: "desktop", width: 1280, height: 800, mobile: false},
		{name: "iphone-14", width: 393, height: 852, mobile: true},
	}

	for _, v := range viewports {
		if err := os.MkdirAll(filepath.Join(outDir, v.name), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "mkdir:", err)
			os.Exit(1)
		}
	}

	allCtx, cancelAll := chromedp.NewContext(context.Background())
	defer cancelAll()
	allCtx, cancelAll = context.WithTimeout(allCtx, 10*time.Minute)
	defer cancelAll()

	var reports []pageReport
	for _, u := range urls {
		for _, v := range viewports {
			r := visit(allCtx, u, v, outDir)
			reports = append(reports, r)
			status := "ok"
			if r.loadError != "" {
				status = "ERR " + r.loadError
			} else if r.hasHorizontal {
				status = fmt.Sprintf("OVERFLOW scroll=%d view=%d", r.scrollWidth, r.clientWidth)
			} else if r.missingContent {
				status = fmt.Sprintf("THIN body=%d chars", r.bodyText)
			}
			fmt.Fprintf(os.Stderr, "  [%s] %s -> %s\n", v.name, u, status)
		}
	}

	writeReport(filepath.Join(outDir, "report.txt"), base, reports)
	fmt.Fprintf(os.Stderr, "\nReport written to %s/report.txt\n", outDir)
}

func fetchSitemap(u string) ([]string, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sm sitemap
	if err := xml.Unmarshal(body, &sm); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(sm.URLs))
	for _, e := range sm.URLs {
		out = append(out, e.Loc)
	}
	return out, nil
}

func visit(parent context.Context, target string, v viewport, outDir string) pageReport {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	r := pageReport{url: target, viewport: v.name}
	var clientW, scrollW, bodyChars int64
	var buf []byte

	err := chromedp.Run(ctx,
		emulation.SetDeviceMetricsOverride(v.width, v.height, 1.0, v.mobile),
		chromedp.Navigate(target),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(900*time.Millisecond),
		chromedp.Evaluate(`document.documentElement.clientWidth`, &clientW),
		chromedp.Evaluate(`document.body.scrollWidth`, &scrollW),
		chromedp.Evaluate(`document.body.innerText.length`, &bodyChars),
		chromedp.ActionFunc(func(ctx context.Context) error {
			b, err := page.CaptureScreenshot().WithCaptureBeyondViewport(true).Do(ctx)
			if err != nil {
				return err
			}
			buf = b
			return nil
		}),
	)
	if err != nil {
		r.loadError = err.Error()
		return r
	}
	r.clientWidth = clientW
	r.scrollWidth = scrollW
	r.bodyText = bodyChars
	// Allow 1px tolerance for sub-pixel rounding.
	r.hasHorizontal = scrollW > clientW+1
	// Pages with very thin body text usually mean a render failure or
	// a 404 served as 200. Top-level index pages can legitimately be short
	// (~600 chars) so we set the threshold low.
	r.missingContent = bodyChars < 200

	out := filepath.Join(outDir, v.name, slugFromURL(target)+".png")
	if err := os.WriteFile(out, buf, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write screenshot:", err)
	}
	return r
}

func slugFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "unknown"
	}
	p := strings.Trim(u.Path, "/")
	if p == "" {
		return "root"
	}
	return strings.ReplaceAll(p, "/", "-")
}

func writeReport(path, base string, reports []pageReport) {
	var b strings.Builder
	fmt.Fprintf(&b, "Sweep report for %s\n", base)
	fmt.Fprintf(&b, "Generated: %s\n\n", time.Now().Format(time.RFC3339))

	var problems []pageReport
	for _, r := range reports {
		if r.loadError != "" || r.hasHorizontal || r.missingContent {
			problems = append(problems, r)
		}
	}
	fmt.Fprintf(&b, "Total page-viewport visits: %d\n", len(reports))
	fmt.Fprintf(&b, "Flagged: %d\n\n", len(problems))

	if len(problems) == 0 {
		b.WriteString("No issues detected.\n")
	} else {
		b.WriteString("=== Flagged ===\n")
		for _, r := range problems {
			fmt.Fprintf(&b, "[%s] %s\n", r.viewport, r.url)
			if r.loadError != "" {
				fmt.Fprintf(&b, "    LOAD ERROR: %s\n", r.loadError)
			}
			if r.hasHorizontal {
				fmt.Fprintf(&b, "    OVERFLOW: scrollWidth=%d clientWidth=%d (delta=%d)\n",
					r.scrollWidth, r.clientWidth, r.scrollWidth-r.clientWidth)
			}
			if r.missingContent {
				fmt.Fprintf(&b, "    THIN: body innerText length = %d\n", r.bodyText)
			}
		}
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
	}
}
