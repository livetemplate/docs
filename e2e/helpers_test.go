package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
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

type wsFrame struct {
	Direction string
	Data      string
	Parsed    map[string]any
}

type wsRecorder struct {
	mu     sync.Mutex
	frames []wsFrame
}

func recordWSFrames(ctx context.Context) *wsRecorder {
	r := &wsRecorder{}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventWebSocketFrameReceived:
			r.mu.Lock()
			defer r.mu.Unlock()
			f := wsFrame{Direction: "received", Data: ev.Response.PayloadData}
			if json.Valid([]byte(f.Data)) {
				_ = json.Unmarshal([]byte(f.Data), &f.Parsed)
			}
			r.frames = append(r.frames, f)
		case *network.EventWebSocketFrameSent:
			r.mu.Lock()
			defer r.mu.Unlock()
			f := wsFrame{Direction: "sent", Data: ev.Response.PayloadData}
			if json.Valid([]byte(f.Data)) {
				_ = json.Unmarshal([]byte(f.Data), &f.Parsed)
			}
			r.frames = append(r.frames, f)
		}
	})
	return r
}

func (r *wsRecorder) received() []wsFrame {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []wsFrame
	for _, f := range r.frames {
		if f.Direction == "received" {
			out = append(out, f)
		}
	}
	return out
}

func (r *wsRecorder) receivedWithTree() []wsFrame {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []wsFrame
	for _, f := range r.frames {
		if f.Direction == "received" {
			if _, ok := f.Parsed["tree"]; ok {
				out = append(out, f)
			}
		}
	}
	return out
}

func (r *wsRecorder) receivedSince(n int) []wsFrame {
	all := r.received()
	if n >= len(all) {
		return nil
	}
	return all[n:]
}

func (r *wsRecorder) receivedWithTreeSince(n int) []wsFrame {
	all := r.receivedWithTree()
	if n >= len(all) {
		return nil
	}
	return all[n:]
}

func (r *wsRecorder) waitForReceivedCount(want int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(r.received()) >= want {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout: got %d received frames, want >= %d", len(r.received()), want)
}

func (r *wsRecorder) waitForReceivedWithTreeCount(want int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(r.receivedWithTree()) >= want {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout: got %d tree frames, want >= %d", len(r.receivedWithTree()), want)
}

func (r *wsRecorder) dump() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var b strings.Builder
	for i, f := range r.frames {
		data := f.Data
		if len(data) > 200 {
			data = data[:200] + "..."
		}
		fmt.Fprintf(&b, "[%d] %s: %s\n", i, f.Direction, data)
	}
	return b.String()
}
