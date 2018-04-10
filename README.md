#### The Idea
An experiment to play with `go/ast` functionality. Given an input `blah.go` file, the tool will parse out public API object exposed as pointers to structs, and construct an actual (testable) interface for each,
along with stubbing out a mock struct suitable for use in unit testing. This is based on friction encounted while writing unit tests and mocks for otherwise terrific libraries in Golang that are written this way (names withheld to protect the guilty!)

#### Example
```bash
# from repo root dir:
make

# print output mock code to stdout
./stubber --source-file=path/to/client.go --stdout

# drop file named thing_mock.go in same dir as source file, ready to compile
./stubber --source=file=path/to/awesome/thing.go
```

#### Status
WIP, handles the major stuff, thin on the corner cases. Again: purpose is _not_ to reinvent other more robust mocking libs, just to get to know `go/ast` a bit better, and to see how tricky it would be to service this particulr use case. 

