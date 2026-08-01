// Command clientversion prints the @livetemplate/client version this build of
// the docs site serves — livetemplate.ClientVersion, the constant behind
// {{lvtClientScriptURL}}.
//
// It exists so CI can provision the exact bundle the site ships rather than
// @latest. A browser test run against a different client than the docs serve
// proves nothing about the docs, and pinning by hand in a workflow re-rots on
// the next dependency bump — which is how the client README spent nineteen
// minor versions advertising 0.1.0.
//
// Usage:
//
//	npm pack "@livetemplate/client@$(go run ./cmd/clientversion)"
package main

import (
	"fmt"

	"github.com/livetemplate/livetemplate"
)

func main() {
	fmt.Println(livetemplate.ClientVersion)
}
