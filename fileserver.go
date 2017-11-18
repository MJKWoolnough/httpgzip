// Package httpgzip is a simple wrapper around http.FileServer that looks for
// a gzip compressed version of a file and serves that if the client requested
// gzip content
package httpgzip

import (
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	acceptEncoding   = "Accept-Encoding"
	contentEncoding  = "Content-Encoding"
	contentType      = "Content-Type"
	contentLength    = "Content-Length"
	anyEncoding      = "*"
	gzipEncoding     = "gzip"
	gzExt            = ".gz"
	identityEncoding = "identity"
	acceptSplit      = ","
	partSplit        = ";"
	weightPrefix     = "q="
	indexPage        = "index.html"
)

type encodings []encoding

func (e encodings) Len() int {
	return len(e)
}

func (e encodings) Less(i, j int) bool {
	return e[i].weight < e[j].weight
}

func (e encodings) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type encoding struct {
	encoding string
	weight   float32
}

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

func (f fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get(acceptEncoding)
	accepts := make(encodings, 0, strings.Count(acceptHeader, acceptSplit)+1)
Loop:
	for _, accept := range strings.Split(acceptHeader, acceptSplit) {
		parts := strings.Split(strings.TrimSpace(accept), partSplit)
		var (
			weight float64 = 1
			err    error
		)
		for _, part := range parts[1:] {
			if strings.HasPrefix(strings.TrimSpace(part), weightPrefix) {
				weight, err = strconv.ParseFloat(part[len(weightPrefix):], 32)
				if err != nil || weight < 0 || weight > 1 { // return an malformed header response?
					continue Loop
				}
				break
			}
		}
		accepts = append(accepts, encoding{
			encoding: strings.ToLower(strings.TrimSpace(parts[0])),
			weight:   float32(weight),
		})
	}
	sort.Sort(accepts)
	allowIdentity := true
	for _, accept := range accepts {
		switch accept.encoding {
		case gzipEncoding:
			if f.serveFile(w, r, gzExt, gzipEncoding) {
				return
			}
		case identityEncoding:
			allowIdentity = accept.weight != 0
			break
		case anyEncoding:
			allowIdentity = accept.weight != 0
		}
	}
	if allowIdentity {
		f.h.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func (f fileServer) serveFile(w http.ResponseWriter, r *http.Request, ext, encoding string) bool {
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
