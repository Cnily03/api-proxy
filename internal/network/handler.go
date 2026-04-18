package network

import (
	"net/http"
	"os"
	"path"

	"api-proxy/internal/network/api"
	"api-proxy/internal/service"
)

// Builds the admin panel's top-level HTTP handler: mounts the /api
// routes and, when staticDir is non-empty, serves static files from
// that directory without directory listings; otherwise non-API paths
// return 404.
func Setup(rules *service.RuleService, auth *service.AuthService, staticDir string) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/", api.Setup(rules, auth))

	if staticDir != "" {
		fs := http.FileServer(noDirList(http.Dir(staticDir)))
		mux.Handle("/", fs)
	} else {
		mux.HandleFunc("/", http.NotFound)
	}

	return mux
}

// Wraps an http.FileSystem to disable directory listings while still
// allowing directories that contain an index file to be served.
func noDirList(fs http.FileSystem) http.FileSystem {
	return noDirListFS{fs: fs, tryFiles: []string{"index.html", "index.htm"}}
}

type noDirListFS struct {
	fs       http.FileSystem
	tryFiles []string
}

// Opens the named path through the wrapped FileSystem; when the path
// resolves to a directory, returns os.ErrNotExist unless an index file
// (index.html / index.htm) is present, preventing auto-generated
// directory listings.
func (n noDirListFS) Open(name string) (http.File, error) {
	f, err := n.fs.Open(name)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if stat.IsDir() {
		found := false
		for _, tf := range n.tryFiles {
			index, err := n.fs.Open(path.Join(name, tf))
			if err == nil {
				index.Close()
				found = true
				break
			}
		}
		if !found {
			f.Close()
			return nil, os.ErrNotExist
		}
	}
	return f, nil
}
