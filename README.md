# httpgzip
--
    import "vimagination.zapto.org/httpgzip"

Package httpgzip is a simple wrapper around http.FileServer that looks for a
compressed version of a file and serves that if the client requested compressed
### content

## Usage

#### func  FileServer

```go
func FileServer(root http.FileSystem, roots ...http.FileSystem) http.Handler
```
FileServer creates a wrapper around http.FileServer using the given
http.FileSystem

Additional http.FileSystem's can be specified and will be turned into a Handler
that checks each in order, stopping at the first

#### func  FileServerWithHandler

```go
func FileServerWithHandler(root http.FileSystem, handler http.Handler) http.Handler
```
FileServerWithHandler acts like FileServer, but allows a custom Handler instead
of the http.FileSystem wrapped http.FileServer
