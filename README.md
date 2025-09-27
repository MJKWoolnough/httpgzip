# httpgzip

[![CI](https://github.com/MJKWoolnough/httpgzip/actions/workflows/go-checks.yml/badge.svg)](https://github.com/MJKWoolnough/httpgzip/actions)
[![Go Reference](https://pkg.go.dev/badge/vimagination.zapto.org/httpgzip.svg)](https://pkg.go.dev/vimagination.zapto.org/httpgzip)
[![Go Report Card](https://goreportcard.com/badge/vimagination.zapto.org/httpgzip)](https://goreportcard.com/report/vimagination.zapto.org/httpgzip)

--
    import "vimagination.zapto.org/httpgzip"

Package httpgzip is a simple wrapper around http.FileServer that looks for a compressed version of a file and serves that if the client requested compressed content.

## Highlights

 - Can merge multiple `http.FileSystem` objects into a single `http.Handler`.
 - Will look for pre-compressed version of a requested file (`.gz`, `.br`, `.fl`, `.zst`).
 - Will serve appropriate compression based on `Accept-Encoding` header.

## Usage

```go
package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"vimagination.zapto.org/httpgzip"
)

func main() {
	// Prepare example filesystems

	var fs, gfs bytes.Buffer

	zw := zip.NewWriter(&fs)
	f, _ := zw.Create("file.html")

	io.WriteString(f, "<HTML>")
	zw.Close()

	zw = zip.NewWriter(&gfs)
	f, _ = zw.Create("file.html.gz")

	gf := gzip.NewWriter(f)

	io.WriteString(gf, "<HEAD>")

	gf.Close()
	zw.Close()

	zfs, _ := zip.NewReader(bytes.NewReader(fs.Bytes()), int64(fs.Len()))
	zgfs, _ := zip.NewReader(bytes.NewReader(gfs.Bytes()), int64(gfs.Len()))

	// Example

	handler := httpgzip.FileServer(http.FS(zfs), http.FS(zgfs))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/file.html", nil)
	r.Header.Set("Accept-encoding", "identity")

	handler.ServeHTTP(w, r)

	fmt.Println("Uncompressed:", w.Body)

	r.Header.Set("Accept-encoding", "gzip")

	w = httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	gr, _ := gzip.NewReader(w.Body)

	fmt.Print("Compressed: ")
	io.Copy(os.Stdout, gr)

	// Output:
	// Uncompressed: <HTML>
	// Compressed: <HEAD>
}
```

## Documentation

Full API docs can be found at:

https://pkg.go.dev/vimagination.zapto.org/httpgzip
