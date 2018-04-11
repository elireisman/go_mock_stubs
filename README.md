#### The Idea
An experiment to play with `go/ast` functionality. The tool generates unit test mocks from Golang source files where the public API is exposed as pointers to structs rather than interfaces. Given a Golang source code file, the tool will extract any public structs and their public methods, and generate a matching interface along with a package-private mock struct implementing no-ops for all methods.

#### Example

```bash
# from repo root dir, build the binary:
make

# print output mock code to stdout
./stubber --source-file=example/one.go --stdout

# writes output as file example/one_mock.go
./stubber --source=file=example/one.go
```

