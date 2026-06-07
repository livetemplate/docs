package e2e

import (
	"context"
	"strings"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// captureConsoleErrors streams browser console errors so a failed assertion
// can report what the page logged. Returns a closure that yields the captured
// list at call time. Shared by the embed/spine/IA tests.
func captureConsoleErrors(ctx context.Context) func() []string {
	var errs []string
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*runtime.EventConsoleAPICalled); ok && e.Type == runtime.APITypeError {
			parts := make([]string, 0, len(e.Args))
			for _, a := range e.Args {
				parts = append(parts, string(a.Value))
			}
			errs = append(errs, strings.Join(parts, " "))
		}
	})
	return func() []string { return errs }
}
