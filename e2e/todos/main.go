// Test-server binary for docs/e2e/todos/todos_test.go. The upstream
// test framework (e2etest.StartTestServer) shells out to `go run .`
// from the test cwd, so the test directory must contain a main package.
//
// Tests navigate to bare root with BasicAuth in the URL:
//
//	http://alice:password@host.docker.internal:<port>/
//
// — todos is a single-page recipe (one mount, no internal routing), so
// a single mount at "/" is all the tests need. Production cmd/site
// instead mounts the same handler at "/apps/todos/" with StripPrefix;
// calls to todos.Handler return the same first-call handler via
// handlerOnce, so test and prod share a single logical app reachable
// at different URL prefixes.
//
// WithDevMode + WithPermissiveOriginCheck:
//   - DevMode enables verbose logging the test framework captures
//   - PermissiveOriginCheck waives the production allowlist; the e2e
//     framework spins servers on random ports the allowlist can't
//     anticipate. Permissive gates this binary to test use only.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	todos "github.com/livetemplate/docs/content/recipes/todos/_app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	h := todos.Handler(
		livetemplate.WithDevMode(true),
		livetemplate.WithPermissiveOriginCheck(),
	)
	log.Printf("todos test-server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, h))
}
