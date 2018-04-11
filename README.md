#### The Idea
An experiment to play with `go/ast` functionality. Given an input `blah.go` file, the tool will parse out public API object exposed as pointers to structs, and construct an actual (testable) interface for each,
along with stubbing out a mock struct suitable for use in unit testing. This is based on friction encounted while writing unit tests and mocks for otherwise terrific libraries in Golang that are written this way (names withheld to protect the guilty!)


#### Example
How to build and run the tool:
```bash
# from repo root dir, build the binary:
make

# print output mock code to stdout
./stubber --source-file=path/to/client.go --stdout

# drop file named thing_mock.go in same dir as source file, ready to compile
./stubber --source=file=path/to/awesome/thing.go
```


#### Status
Purpose is _not_ to reinvent other more robust mocking libs, just to get to know `go/ast` a bit better.

On the other hand, at this point it handles all the input Golang source I throw at it, including all primitive, pointer, and struct types, N-dimensional arrays, chans (uni- and bi-directional), maps, interfaces, packages, imports with aliases, and even correctly pulling in only the imports the stub file requires, not the whole set the input source includes.
I suspect there are likely still a few corner cases around return values in the mock methods for some interface types. If you decide to try this tool out and you run afoul of such a case, please file an issue!


