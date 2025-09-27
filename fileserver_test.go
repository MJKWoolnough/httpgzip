package httpgzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stat int64

func (stat) Name() string {
	return ""
}

func (s stat) Size() int64 {
	return int64(s)
}

func (stat) Mode() fs.FileMode {
	return 0
}

func (stat) ModTime() time.Time {
	return time.Now()
}

func (stat) IsDir() bool {
	return false
}

func (stat) Sys() any {
	return nil
}

type file struct {
	*bytes.Reader
}

func (f *file) Close() error {
	return nil
}

func (f *file) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return stat(f.Size()), nil
}

type memFS map[string][]byte

func (m memFS) Open(name string) (http.File, error) {
	f, ok := m[name]
	if !ok {
		return nil, fs.ErrNotExist
	}

	return &file{Reader: bytes.NewReader(f)}, nil
}

func gzipContent(content string) []byte {
	var buf bytes.Buffer

	g := gzip.NewWriter(&buf)

	io.WriteString(g, content)

	g.Close()

	return buf.Bytes()
}

func TestFileserver(t *testing.T) {
	srv := httptest.NewServer(FileServer(memFS{
		"/file.txt":        []byte("content"),
		"/anotherFile.txt": []byte("Hello, World!"),
		"/some_file":       []byte("<HTML>"),
	}, memFS{
		"/anotherFile.txt.gz": gzipContent("Foo Bar"),
		"/some_file.gz":       gzipContent("<HTML>"),
	}))
	client := srv.Client()

	for n, test := range [...]struct {
		name            string
		identityContent string
		content         string
		contentType     string
	}{
		{
			name: "/unknown",
		},
		{
			name:            "/file.txt",
			identityContent: "content",
			content:         "content",
			contentType:     "text/plain; charset=utf-8",
		},
		{
			name:            "/anotherFile.txt",
			identityContent: "Hello, World!",
			content:         "Foo Bar",
			contentType:     "text/plain; charset=utf-8",
		},
		{
			name:            "/some_file",
			identityContent: "<HTML>",
			content:         "<HTML>",
			contentType:     "text/html; charset=utf-8",
		},
	} {
		testFile(t, client, n+1, 1, srv.URL+test.name, test.identityContent, test.contentType)
		testFile(t, client, n+1, 2, srv.URL+test.name, test.content, test.contentType)
	}
}

func testFile(t *testing.T, client *http.Client, test, part int, url, content, contentType string) {
	var resp *http.Response

	if part == 1 {
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Accept-Encoding", "identity")

		resp, _ = client.Do(req)
	} else {
		resp, _ = client.Get(url)
	}

	if content == "" {
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("test %d.%d: expecting status 404, got %d", test, part, resp.StatusCode)
		}

		return
	}

	if ct := resp.Header.Get("Content-Type"); ct != contentType {
		t.Errorf("test %d.%d.1: expecting content type %q, got %q", test, part, contentType, ct)
	} else if body, _ := io.ReadAll(resp.Body); string(body) != content {
		t.Errorf("test %d.%d.2: expecting content %q, got %q", test, part, content, body)
	}
}
