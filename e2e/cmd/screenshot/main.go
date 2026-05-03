// Screenshot tool for visual diagnosis. Takes full-page screenshots
// of one or more URLs at multiple viewports and writes them to disk.
//
// On mobile viewports, also clicks the .tinkerdown-nav-toggle button
// (if present) and takes a second screenshot of the open sidebar — useful
// for verifying that the hamburger actually exposes the menu.
//
// Usage:
//
//	go run ./cmd/screenshot https://livetemplate.fly.dev/
//
// Writes screenshots to /tmp/lvt-shots/.
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type viewport struct {
	name   string
	width  int64
	height int64
	mobile bool
}

func main() {
	urls := os.Args[1:]
	if len(urls) == 0 {
		urls = []string{"https://livetemplate.fly.dev/"}
	}

	outDir := "/tmp/lvt-shots"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}

	viewports := []viewport{
		{name: "desktop", width: 1280, height: 800, mobile: false},
		{name: "tablet", width: 768, height: 1024, mobile: true},
		{name: "iphone-se", width: 375, height: 667, mobile: true},
		{name: "iphone-14", width: 393, height: 852, mobile: true},
		{name: "iphone-14-pro-max", width: 430, height: 932, mobile: true},
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	for _, u := range urls {
		for _, v := range viewports {
			path := pathSlug(u)
			out := filepath.Join(outDir, fmt.Sprintf("%s-%s.png", v.name, path))

			var buf []byte
			actions := []chromedp.Action{
				emulation.SetDeviceMetricsOverride(v.width, v.height, 1.0, v.mobile),
				chromedp.Navigate(u),
				chromedp.WaitVisible("body", chromedp.ByQuery),
				chromedp.Sleep(800 * time.Millisecond),
				fullPageScreenshot(&buf),
			}

			if err := chromedp.Run(ctx, actions...); err != nil {
				fmt.Fprintf(os.Stderr, "render %s @ %s: %v\n", u, v.name, err)
				continue
			}

			if err := os.WriteFile(out, buf, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", out, err)
				continue
			}
			fmt.Println(out)

			// On mobile, also click the toggle and capture the open state.
			if v.mobile {
				openOut := filepath.Join(outDir, fmt.Sprintf("%s-%s-open.png", v.name, path))
				var openBuf []byte
				openActions := []chromedp.Action{
					chromedp.Click(".tinkerdown-nav-toggle", chromedp.ByQuery),
					chromedp.Sleep(500 * time.Millisecond), // animation
					fullPageScreenshot(&openBuf),
				}
				if err := chromedp.Run(ctx, openActions...); err != nil {
					fmt.Fprintf(os.Stderr, "open render %s: %v\n", u, err)
					continue
				}
				if err := os.WriteFile(openOut, openBuf, 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "write open %s: %v\n", openOut, err)
					continue
				}
				fmt.Println(openOut)
			}
		}
	}
}

func fullPageScreenshot(out *[]byte) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		buf, err := page.CaptureScreenshot().WithCaptureBeyondViewport(true).Do(ctx)
		if err != nil {
			return err
		}
		*out = buf
		return nil
	})
}

func pathSlug(raw string) string {
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
