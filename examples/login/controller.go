package loginrecipe

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/livetemplate/livetemplate"
)

// >>> region:state
// AuthController holds shared state and dependencies.
// This is a singleton that persists across sessions.
type AuthController struct {
	// For server-initiated updates (per-session)
	sessions map[string]livetemplate.Session
	mu       sync.Mutex

	// mountPath is the absolute URL prefix where this recipe is mounted
	// (e.g. "/apps/login/" in prod, "/" in the e2e suite). It is the
	// redirect target after Login/Logout: livetemplate.Context.Redirect
	// requires an absolute path, and http.StripPrefix strips the prefix
	// from r.URL.Path before the recipe sees it — so the recipe can't
	// reconstruct its own mount from the request alone. The caller (the
	// mount site in cmd/site or the e2e test-server) supplies it.
	mountPath string
}

// AuthState is pure data, cloned per session.
// Contains only serializable fields for the auth UI.
type AuthState struct {
	Username      string    `lvt:"persist"`
	IsLoggedIn    bool      `lvt:"persist"`
	ServerMessage string
	LoginTime     time.Time `lvt:"persist"`
}

// <<< region:state

// >>> region:login
// Login handles the "login" action.
func (c *AuthController) Login(state AuthState, ctx *livetemplate.Context) (AuthState, error) {
	username := ctx.GetString("username")
	password := ctx.GetString("password")

	// Field-level validation
	if username == "" {
		return state, livetemplate.NewFieldError("username", fmt.Errorf("username is required"))
	}
	if password == "" {
		return state, livetemplate.NewFieldError("password", fmt.Errorf("password is required"))
	}

	// Demo: accept any username with password "secret"
	if password != "secret" {
		ctx.SetFlash("error", "Invalid credentials")
		return state, nil
	}
	state.Username = username
	state.IsLoggedIn = true
	state.LoginTime = time.Now()
	state.ServerMessage = "" // Will be set when WebSocket connects

	// Set HttpOnly session cookie
	err := ctx.SetCookie(&http.Cookie{
		Name:     "session_token",
		Value:    fmt.Sprintf("session_%s_%d", username, time.Now().Unix()),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600, // 1 hour
	})
	if err != nil {
		return state, fmt.Errorf("failed to set cookie: %w", err)
	}

	// Redirect to the recipe's mount path (POST-Redirect-GET). The new
	// GET re-renders the template; IsLoggedIn=true is now persisted, so
	// the dashboard branch renders and the WebSocket then connects.
	return state, ctx.Redirect(c.mountPath, http.StatusSeeOther)
}

// <<< region:login

// >>> region:logout
// Logout handles the "logout" action.
func (c *AuthController) Logout(state AuthState, ctx *livetemplate.Context) (AuthState, error) {
	state.Username = ""
	state.IsLoggedIn = false
	state.ServerMessage = ""

	// Delete session cookie
	err := ctx.DeleteCookie("session_token")
	if err != nil {
		return state, fmt.Errorf("failed to delete cookie: %w", err)
	}

	// Redirect back to the login form at the recipe's mount path.
	return state, ctx.Redirect(c.mountPath, http.StatusSeeOther)
}

// <<< region:logout

// ServerWelcome handles the "serverWelcome" action (server-initiated welcome messages).
// This is triggered by TriggerAction from sendWelcomeMessage.
func (c *AuthController) ServerWelcome(state AuthState, ctx *livetemplate.Context) (AuthState, error) {
	message := ctx.GetString("message")
	state.ServerMessage = message
	return state, nil
}

// >>> region:onconnect
// OnConnect is called when a WebSocket connection is established.
// This is a lifecycle method on the controller.
func (c *AuthController) OnConnect(state AuthState, ctx *livetemplate.Context) (AuthState, error) {
	session := ctx.Session()

	log.Printf("WebSocket connected (user: %s, logged_in: %v)", state.Username, state.IsLoggedIn)

	// Store session for server-initiated updates
	if state.IsLoggedIn && session != nil {
		c.mu.Lock()
		c.sessions[state.Username] = session
		c.mu.Unlock()

		// Send a welcome message from the server.
		// This demonstrates server-initiated updates after WebSocket connects.
		go c.sendWelcomeMessage(state.Username, session)
	}

	return state, nil
}

// OnDisconnect is called when a WebSocket connection is closed.
func (c *AuthController) OnDisconnect() {
	log.Printf("WebSocket disconnected")
}

// <<< region:onconnect

// >>> region:serverpush
// sendWelcomeMessage sends a server-initiated welcome message via WebSocket.
// This demonstrates pushing updates from server to client without user action.
func (c *AuthController) sendWelcomeMessage(username string, session livetemplate.Session) {
	// Small delay so the page fully renders first
	time.Sleep(500 * time.Millisecond)

	// Trigger server-initiated action that returns modified state with the welcome data.
	// This updates the state and sends the update to all user's connections.
	if err := session.TriggerAction("serverWelcome", map[string]interface{}{
		"message": fmt.Sprintf("Welcome %s! This message was pushed from the server at %s",
			username, time.Now().Format("15:04:05")),
	}); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
	} else {
		log.Printf("Server-initiated welcome message sent to %s", username)
	}
}

// <<< region:serverpush
