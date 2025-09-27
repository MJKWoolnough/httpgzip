package httpgzip_test

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

func Example() {
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
