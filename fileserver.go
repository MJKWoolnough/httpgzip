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

	"github.com/MJKWoolnough/httpencoding"
)

const (
	contentEncoding = "Content-Encoding"
	contentType     = "Content-Type"
	contentLength   = "Content-Length"
	indexPage       = "index.html"
)

var encodings = map[string]string{
	"gzip":   ".gz",
	"x-gzip": ".gz",
	"br":     ".br",
}

type fileServer struct {
	root http.FileSystem
	h    http.Handler
}

// FileServer creates a wrapper around http.FileServer using the given
// http.FileSystem
func FileServer(root http.FileSystem) http.Handler {
	return &fileServer{
		root,
		http.FileServer(root),
	}
}

func (f *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !httpencoding.HandleEncoding(w, r, f) {
		httpencoding.InvalidEncoding(w)
	}
}

func (f *fileServer) Handle(w http.ResponseWriter, r *http.Request, encoding string) bool {
	if encoding == "" {
		f.h.ServeHTTP(w, r)
		return true
	}
	ext, ok := encodings[encoding]
	if !ok {
		return false
	}
	p := path.Clean(r.URL.Path)
	m := p
	nf, err := f.root.Open(p + ext)
	if strings.HasSuffix(p, "/") {
		m += indexPage
		if err != nil {
			nf, err = f.root.Open(p + ext)
			p += indexPage
		}
	}
	if err == nil {
		if ctype := mime.TypeByExtension(filepath.Ext(m)); ctype != "" {
			s, err := nf.Stat()
			if err == nil {
				w.Header().Set(contentType, ctype)
				w.Header().Set(contentLength, strconv.FormatInt(s.Size(), 10))
				w.Header().Set(contentEncoding, encoding)
				r.URL.Path = p + ext
				f.h.ServeHTTP(w, r)
				return true
			}
		}
	}
	return false
}
