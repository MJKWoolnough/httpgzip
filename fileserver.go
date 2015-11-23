// Package httpgzip is a simple wrapper around http.FileServer that looks for
// a gzip compressed version of a file and serves that if the client requested
// gzip content
package httpgzip

import (
	"net/http"
	"path"
	"strings"
)

type fileServer struct {
	root http.FileSystem
	h    http.Handler
}

// FileServer creates a wrapper around http.FileServer using the given
// http.FileSystem
func FileServer(root http.FileSystem) http.Handler {
	return fileServer{
		root,
		http.FileServer(root),
	}
}

// ServerHTTP implements the http.Handler interface
func (f fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, e := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.Trim(e) == "gzip" {
			if _, err := f.root.Open(path.Clean(r.URL.Path + ".gz")); err != nil {
				r.URL.Path += ".gz"
				w.Header().Set("Content-Encoding", "gzip")
				f.h.ServeHTTP(w, r)
			}
			break
		}
	}
	f.h.ServeHTTP(w, r)
}
