package network

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
)

type staticServer struct {
	fs       http.FileSystem
	tryFiles []string
}

func newStaticServer(fs http.FileSystem) *staticServer {
	return &staticServer{fs: fs, tryFiles: []string{"index.html", "index.htm", "/index.html", "/index.htm"}}
}

func (s *staticServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, POST, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlPath := r.URL.Path
	name := path.Clean("/" + urlPath)

	f, info, err := s.Open(name)
	if err == nil {
		if info.IsDir() {
			f.Close()
			if !strings.HasSuffix(urlPath, "/") || strings.HasSuffix(urlPath, "//") {
				redirectToSingleSlash(w, r)
				return
			}
			// URL already ends with a single "/": we never serve
			// directory listings, so treat this as 404 and let
			// openFallback decide whether to surface an SPA shell.
			s.serveNotFound(w, r, name)
			return
		}
		defer f.Close()
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		return
	}

	if !errors.Is(err, os.ErrNotExist) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	s.serveNotFound(w, r, name)
}

// Opens name via the wrapped filesystem and returns its FileInfo
// alongside the handle. Directory entries are returned as-is;
// callers decide how to handle them based on the URL shape.
func (s *staticServer) Open(name string) (http.File, os.FileInfo, error) {
	f, err := s.fs.Open(name)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, info, nil
}

// Walks tryFiles and returns the first regular file that can be
// opened. Entries starting with "/" are resolved against the static
// root; other entries are joined with the directory portion of
// name (or name itself when it already refers to a directory).
func (s *staticServer) openFallback(name string) (http.File, os.FileInfo, error) {
	for _, tf := range s.tryFiles {
		p := tf
		if !path.IsAbs(tf) {
			p = path.Join(name, tf)
		}
		f, err := s.fs.Open(p)
		if err != nil {
			continue
		}
		info, err := f.Stat()
		if err != nil || info.IsDir() {
			f.Close()
			continue
		}
		return f, info, nil
	}
	return nil, nil, os.ErrNotExist
}

func (s *staticServer) serveNotFound(w http.ResponseWriter, r *http.Request, name string) {
	f, info, err := s.openFallback(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

// Rewrites the URL so its path ends with exactly one "/" and issues
// a 301. Used when a directory is requested without a trailing slash,
// or with more than one consecutive trailing slash.
func redirectToSingleSlash(w http.ResponseWriter, r *http.Request) {
	u := *r.URL
	u.Path = strings.TrimRight(u.Path, "/") + "/"
	u.RawPath = ""
	http.Redirect(w, r, u.RequestURI(), http.StatusMovedPermanently)
}
