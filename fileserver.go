// Package httpgzip is a simple wrapper around http.FileServer that looks for
// a compressed version of a file and serves that if the client requested
// compressed content
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
	"gzip":    ".gz",
	"x-gzip":  ".gz",
	"br":      ".br",
	"deflate": ".fl",
}

type overlay []http.FileSystem

func (o overlay) Open(name string) (f http.File, err error) {
	for _, fs := range o {
		f, err = fs.Open(name)
		if err == nil {
			return f, nil
		}
	}
	return nil, err
}

type fileServer struct {
	root http.FileSystem
	h    http.Handler
}

// FileServer creates a wrapper around http.FileServer using the given
// http.FileSystem
//
// Additional http.FileSystem's can be specified and will be turned into a
// Handler that checks each in order, stopping at the first
func FileServer(root http.FileSystem, roots ...http.FileSystem) http.Handler {
	if len(roots) > 0 {
		overlays := make(overlay, 1, len(roots)+1)
		overlays[0] = root
		overlays = append(overlays, roots...)
		root = overlays
	}
	return &fileServer{
		root,
		http.FileServer(root),
	}
}

func (f *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fsh := fileserverHandler{
		fileServer: f,
		w:          w,
		r:          r,
	}
	if !httpencoding.HandleEncoding(r, &fsh) {
		httpencoding.InvalidEncoding(w)
	}
}

type fileserverHandler struct {
	*fileServer
	w http.ResponseWriter
	r *http.Request
}

func (f *fileserverHandler) Handle(encoding string) bool {
	if encoding == "" {
		httpencoding.ClearEncoding(f.r)
		f.h.ServeHTTP(f.w, f.r)
		return true
	}
	ext, ok := encodings[encoding]
	if !ok {
		return false
	}
	p := path.Clean(f.r.URL.Path)
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
				f.w.Header().Set(contentType, ctype)
				f.w.Header().Set(contentLength, strconv.FormatInt(s.Size(), 10))
				f.w.Header().Set(contentEncoding, encoding)
				f.r.URL.Path = p + ext
				httpencoding.ClearEncoding(f.r)
				f.h.ServeHTTP(f.w, f.r)
				return true
			}
		}
	}
	return false
}
