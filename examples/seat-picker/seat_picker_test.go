package seatpicker

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	e2etest "github.com/livetemplate/lvt/testing"
)

// ---------------------------------------------------------------------------
// White-box logic tests — fast, deterministic, no browser. These pin the
// conflict rule (the thing a skeptic doubts: "can two users double-book?")
// without the cost and flakiness of a real browser. The e2e test below
// proves the same rule survives the full WebSocket round trip.
// ---------------------------------------------------------------------------

// hold is a test helper mirroring what an action does: lock, expire, hold.
// owner is a session id (the white-box tests use "Alice"/"Bob" as opaque
// stand-ins for ctx.GroupID()).
func hold(c *Controller, owner, id string) (bool, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expire()
	return c.tryHold(owner, id)
}

// statusFor returns the projected status of a seat from owner's perspective.
func statusFor(c *Controller, owner, id string) string {
	var st State
	c.mu.Lock()
	c.project(&st, owner)
	c.mu.Unlock()
	for _, row := range st.Rows {
		for _, s := range row.Seats {
			if s.ID == id {
				return s.Status
			}
		}
	}
	return ""
}

func TestDoubleBookingIsImpossible(t *testing.T) {
	c := newController([]string{"A"}, 4)

	// Alice holds A1.
	if ok, _ := hold(c, "Alice", "A1"); !ok {
		t.Fatal("Alice should be able to hold an available seat")
	}
	// Bob cannot take the same seat — this is the core invariant.
	ok, msg := hold(c, "Bob", "A1")
	if ok {
		t.Error("Bob must not be able to hold a seat Alice already holds")
	}
	if msg == "" {
		t.Error("a denied hold should explain why")
	}

	// The projection reflects ownership per viewer.
	if got := statusFor(c, "Alice", "A1"); got != "mine" {
		t.Errorf("Alice should see A1 as 'mine', got %q", got)
	}
	if got := statusFor(c, "Bob", "A1"); got != "held" {
		t.Errorf("Bob should see A1 as 'held', got %q", got)
	}
}

func TestReleaseFreesSeatForOthers(t *testing.T) {
	c := newController([]string{"A"}, 4)
	hold(c, "Alice", "A1")

	// Alice releases (the action's inner step).
	c.mu.Lock()
	c.seats["A1"].holder = ""
	c.mu.Unlock()

	if ok, _ := hold(c, "Bob", "A1"); !ok {
		t.Error("Bob should be able to hold A1 after Alice releases it")
	}
	if got := statusFor(c, "Bob", "A1"); got != "mine" {
		t.Errorf("Bob should now own A1, got %q", got)
	}
}

func TestExpiredHoldIsReclaimed(t *testing.T) {
	c := newController([]string{"A"}, 4)
	hold(c, "Alice", "A1")

	// Force the hold into the past.
	c.mu.Lock()
	c.seats["A1"].expires = time.Now().Add(-time.Second)
	c.mu.Unlock()

	// Bob can now take it — expire() runs inside hold().
	if ok, _ := hold(c, "Bob", "A1"); !ok {
		t.Error("an expired hold should be reclaimable by another viewer")
	}
}

func TestBookingPersistsAndBlocks(t *testing.T) {
	c := newController([]string{"A"}, 4)
	hold(c, "Alice", "A1")
	hold(c, "Alice", "A2")

	c.mu.Lock()
	booked := c.bookHeld("Alice")
	c.mu.Unlock()
	if booked != 2 {
		t.Fatalf("expected to book 2 held seats, booked %d", booked)
	}

	if got := statusFor(c, "Alice", "A1"); got != "booked-mine" {
		t.Errorf("Alice should see her booking as 'booked-mine', got %q", got)
	}
	if got := statusFor(c, "Bob", "A1"); got != "booked" {
		t.Errorf("Bob should see A1 as 'booked', got %q", got)
	}
	// A booked seat never expires and can never be re-held.
	c.mu.Lock()
	c.seats["A1"].expires = time.Now().Add(-time.Hour)
	c.mu.Unlock()
	if ok, _ := hold(c, "Bob", "A1"); ok {
		t.Error("a booked seat must never become re-holdable")
	}
}

// ---------------------------------------------------------------------------
// End-to-end cross-user test — two independent browser sessions (Alice and
// Bob) joined under different names. This is the proof the other recipes
// cannot offer: a selection by one *user* appears live in another user's
// browser, over a developer-defined shared topic, with no client code.
// ---------------------------------------------------------------------------

// join navigates a fresh browser session to the app and joins under name.
func join(t *testing.T, browserCtx context.Context, url, name string) {
	t.Helper()
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`[data-lvt-id]`, chromedp.ByQuery),
		e2etest.WaitFor(`typeof window.liveTemplateClient !== 'undefined'`, 5*time.Second),
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.SetValue(`input[name="name"]`, name, chromedp.ByQuery),
		chromedp.Click(`form[name="join"] button[type="submit"]`, chromedp.ByQuery),
		e2etest.WaitFor(`document.querySelector('button.seat') !== null`, 5*time.Second),
	)
	if err != nil {
		t.Fatalf("join as %q failed: %v", name, err)
	}
}

// seatClass is a JS expression for the live CSS class of a seat, by its
// value attribute. Parenthesised so callers can append .includes(...)
// without the || binding more loosely than the method call.
func seatClass(id string) string {
	return fmt.Sprintf(`(document.querySelector('button.seat[value="%s"]')?.className || '')`, id)
}

func TestSeatPickerE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	serverPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("free port (server): %v", err)
	}

	serverCmd := e2etest.StartTestServer(t, "./cmd", serverPort)
	defer func() {
		if serverCmd != nil && serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
	}()
	t.Logf("✅ server on :%d", serverPort)

	// Alice and Bob are two *separate* Chrome containers, not two tabs of
	// one. Separate browsers means separate cookies, separate session
	// groups — genuinely two different users. The cross-user update then
	// travels across session-group boundaries on the shared developer
	// topic, which is the exact behaviour SelfTopic recipes cannot show.
	newUser := func(name string) (context.Context, func()) {
		t.Helper()
		debugPort, perr := e2etest.GetFreePort()
		if perr != nil {
			t.Fatalf("free port (chrome %s): %v", name, perr)
		}
		chromeCmd := e2etest.StartDockerChrome(t, debugPort)
		_ = chromeCmd
		allocCtx, allocCancel := chromedp.NewRemoteAllocator(
			context.Background(), fmt.Sprintf("http://localhost:%d", debugPort))
		browserCtx, cancelBrowser := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
		timedCtx, cancelTimeout := context.WithTimeout(browserCtx, 90*time.Second)
		cleanup := func() {
			cancelTimeout()
			cancelBrowser()
			allocCancel()
			e2etest.StopDockerChrome(t, debugPort)
		}
		return timedCtx, cleanup
	}

	aliceCtx, closeAlice := newUser("Alice")
	defer closeAlice()
	bobCtx, closeBob := newUser("Bob")
	defer closeBob()

	url := e2etest.GetChromeTestURL(serverPort)

	t.Run("Initial_Load_And_UI_Standards", func(t *testing.T) {
		var html, violations string
		err := chromedp.Run(aliceCtx,
			chromedp.Navigate(url),
			chromedp.WaitVisible(`[data-lvt-id]`, chromedp.ByQuery),
			e2etest.WaitFor(`typeof window.liveTemplateClient !== 'undefined'`, 5*time.Second),
			chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
			chromedp.OuterHTML(`body`, &html, chromedp.ByQuery),
			chromedp.Evaluate(`(() => {
				const v = [];
				['onclick','onchange','oninput','onsubmit','onkeydown','onkeyup'].forEach(h => {
					document.querySelectorAll('[' + h + ']').forEach(el => v.push('inline ' + h + ' on <' + el.tagName.toLowerCase() + '>'));
				});
				document.querySelectorAll('[style]').forEach(el => v.push('inline style on <' + el.tagName.toLowerCase() + '>'));
				if (!document.querySelector('meta[name="color-scheme"]')) v.push('missing color-scheme meta');
				if (document.documentElement.lang !== 'en') v.push('missing lang=en');
				return v.join('; ');
			})()`, &violations),
		)
		if err != nil {
			t.Fatalf("initial load: %v", err)
		}
		if strings.Contains(html, "{{") {
			t.Error("rendered HTML leaked template expressions")
		}
		if !strings.Contains(html, `name="name"`) {
			t.Error("join form should ask for a name")
		}
		if violations != "" {
			t.Errorf("UI standard violations (no framework attributes / inline handlers allowed): %s", violations)
		}
	})

	t.Run("Selection_Is_Visible_To_Other_User", func(t *testing.T) {
		join(t, aliceCtx, url, "Alice")
		join(t, bobCtx, url, "Bob")

		// Alice selects A1.
		if err := chromedp.Run(aliceCtx,
			chromedp.Click(`button.seat[value="A1"]`, chromedp.ByQuery),
			e2etest.WaitFor(seatClass("A1")+`.includes('mine')`, 5*time.Second),
		); err != nil {
			t.Fatalf("Alice select A1: %v", err)
		}

		// Bob must see A1 become held — live, with no action on his part.
		// This is the cross-user broadcast; it is what SelfTopic recipes
		// cannot demonstrate.
		if err := chromedp.Run(bobCtx,
			e2etest.WaitFor(seatClass("A1")+`.includes('held')`, 5*time.Second),
		); err != nil {
			var bobClass string
			chromedp.Run(bobCtx, chromedp.Evaluate(seatClass("A1"), &bobClass))
			t.Fatalf("Bob should see A1 as held after Alice selects it; class was %q: %v", bobClass, err)
		}

		// And it is disabled for Bob — he cannot double-book.
		var disabled bool
		chromedp.Run(bobCtx, chromedp.Evaluate(`document.querySelector('button.seat[value="A1"]').disabled`, &disabled))
		if !disabled {
			t.Error("A1 should be disabled for Bob (held by Alice)")
		}
	})

	t.Run("Bidirectional_Broadcast", func(t *testing.T) {
		// Bob selects B2; Alice should see it held.
		if err := chromedp.Run(bobCtx,
			chromedp.Click(`button.seat[value="B2"]`, chromedp.ByQuery),
			e2etest.WaitFor(seatClass("B2")+`.includes('mine')`, 5*time.Second),
		); err != nil {
			t.Fatalf("Bob select B2: %v", err)
		}
		if err := chromedp.Run(aliceCtx,
			e2etest.WaitFor(seatClass("B2")+`.includes('held')`, 5*time.Second),
		); err != nil {
			t.Fatalf("Alice should see B2 held after Bob selects it: %v", err)
		}
	})

	t.Run("Booking_Broadcasts_To_Other_User", func(t *testing.T) {
		// Alice confirms her hold on A1.
		if err := chromedp.Run(aliceCtx,
			chromedp.Click(`button[name="confirm"]`, chromedp.ByQuery),
			e2etest.WaitFor(seatClass("A1")+`.includes('booked-mine')`, 5*time.Second),
		); err != nil {
			t.Fatalf("Alice confirm: %v", err)
		}
		// Bob sees A1 as booked (someone else's confirmed seat).
		if err := chromedp.Run(bobCtx,
			e2etest.WaitFor(seatClass("A1")+`.includes('booked') && !`+seatClass("A1")+`.includes('booked-mine')`, 5*time.Second),
		); err != nil {
			t.Fatalf("Bob should see A1 booked after Alice confirms: %v", err)
		}
	})

	t.Logf("🎉 cross-user seat-picker E2E passed")
}
