// Package httpgzip is a simple wrapper around http.FileServer that looks for
// a gzip compressed version of a file and serves that if the client requested
// gzip content
package httpgzip

import (
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
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
		if strings.TrimSpace(e) == "gzip" {
			p := path.Clean(r.URL.Path)
			nf, err := f.root.Open(p + ".gz")
			if err != nil && strings.HasSuffix(p, "/") {
				p += "index.html"
				nf, err = f.root.Open(p + ".gz")
			}
			if err == nil {
				w.Header().Set("Content-Encoding", "gzip")
				if ctype := mime.TypeByExtension(filepath.Ext(r.URL.Path)); ctype != "" {
					w.Header().Set("Content-Type", ctype)
					s, _ := nf.Stat()
					w.Header().Set("Content-Length", strconv.FormatInt(s.Size(), 10))
				}
				r.URL.Path = p + ".gz"
			}
			break
		}
	}
	f.h.ServeHTTP(w, r)
}
