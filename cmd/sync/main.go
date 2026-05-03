// Command sync mirrors documentation from a source repo into the
// LiveTemplate docs site, using content/_meta/source-of-truth.yaml as
// the contract for which files map to which docs URLs.
//
// Invoked by .github/workflows/sync.yml on a repository_dispatch event
// from a source repo's release, or manually via workflow_dispatch.
//
// Usage:
//
//	sync --source-repo=https://github.com/livetemplate/livetemplate \
//	     --ref=v0.8.23 \
//	     --site-root=. \
//	     [--dry-run]
//
// Exit codes:
//
//	0 — sync succeeded (whether or not anything actually changed)
//	1 — invocation error (bad args, missing config)
//	2 — source-repo clone or read error
//	3 — write error in --site-root
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	srcRepo := flag.String("source-repo", "", "URL of the source GitHub repo to sync from (required)")
	ref := flag.String("ref", "main", "Git ref (tag or branch) to sync from")
	siteRoot := flag.String("site-root", ".", "Root of the docs site repo (containing content/)")
	dryRun := flag.Bool("dry-run", false, "List files that would change without writing them")
	flag.Parse()

	if *srcRepo == "" {
		log.Println("--source-repo is required")
		flag.Usage()
		os.Exit(1)
	}

	res, err := Run(Options{
		SourceRepo: *srcRepo,
		Ref:        *ref,
		SiteRoot:   *siteRoot,
		DryRun:     *dryRun,
	})
	if err != nil {
		log.Printf("sync failed: %v", err)
		// Map error categories to exit codes; the workflow uses these
		// to decide whether to file an issue vs. open a PR.
		var ec exitCoder
		if asErrCoder(err, &ec) {
			os.Exit(ec.exitCode())
		}
		os.Exit(1)
	}

	fmt.Printf("sync complete: %d updated, %d unchanged, %d skipped\n",
		res.Updated, res.Unchanged, res.Skipped)
	for _, p := range res.UpdatedPaths {
		fmt.Println("  updated  ", p)
	}
	for _, p := range res.SkippedReasons {
		fmt.Println("  skipped  ", p)
	}
}

// exitCoder lets internal errors carry a desired exit code.
type exitCoder interface {
	exitCode() int
}

func asErrCoder(err error, target *exitCoder) bool {
	if ec, ok := err.(exitCoder); ok {
		*target = ec
		return true
	}
	return false
}
