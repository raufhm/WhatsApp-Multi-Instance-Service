package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
)

//go:embed all:dist
var distFS embed.FS

// dashboardStaticFS returns the embedded frontend build rooted at dist/.
// The dist directory always contains at least a .gitkeep placeholder (committed)
// so the binary compiles even when the frontend has not been built. When the
// real build is absent, the dashboard serves a helpful error page.
func dashboardStaticFS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Printf("[dashboard] frontend not built (run: cd frontend && npm run build): %v", err)
		return emptyFS{}
	}

	// Log the number of embedded files to aid debugging.
	if entries, err := fs.ReadDir(sub, "."); err == nil {
		log.Printf("[dashboard] embedded %d dist files", len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			log.Printf("[dashboard] embedded: %s", e.Name())
		}
	}

	return sub
}

// emptyFS is a minimal empty filesystem used when the frontend build is absent.
type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) { return nil, fs.ErrNotExist }

var _ = os.ErrNotExist
