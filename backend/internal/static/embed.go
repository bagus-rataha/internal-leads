// Package static embeds the built frontend single-page application so the
// Go binary can serve it directly, with no separate web server needed.
//
// dist holds the frontend's production build output (frontend/dist, copied
// in during the Docker image build). Only a dist/.gitkeep placeholder is
// committed to source control - the real build output stays gitignored - so
// this package, and anything that imports it, keeps compiling for
// backend-only development even when no frontend build has been produced
// locally.
package static

import (
	"embed"
	"io/fs"
)

// distFS holds everything under dist, including dotfiles. The "all:" prefix
// is required so dist/.gitkeep itself is not silently excluded by embed's
// default dotfile skip - without it, an empty dist directory would make the
// embed pattern match zero files and fail the build.
//
//go:embed all:dist
var distFS embed.FS

// DistFS returns the embedded frontend build rooted at dist, so paths
// resolve from "/" (e.g. "/index.html") instead of "/dist/index.html".
func DistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

// MustDistFS is DistFS but panics on error. Safe to call at startup: dist is
// guaranteed to exist in the embedded tree because dist/.gitkeep is always
// committed, so fs.Sub can only fail here if that invariant is broken.
func MustDistFS() fs.FS {
	sub, err := DistFS()
	if err != nil {
		panic(err)
	}
	return sub
}
