# httpgzip
--
    import "github.com/MJKWoolnough/httpgzip"

Package httpgzip is a simple wrapper around http.FileServer that looks for
a gzip compressed version of a file and serves that if the client requested
gzip content

## Usage

#### func  FileServer

```go
func FileServer(root http.FileSystem) http.Handler
```
FileServer creates a wrapper around http.FileServer using the given
http.FileSystem
