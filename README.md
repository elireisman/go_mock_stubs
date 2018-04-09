### The Idea
An experiment to play with `go/ast` functionality. Use it to parse public APIs based on pointers to structs, and construct an actual interface,
along with stubbing out a mock struct suitable for use in unit testing. This is based on friction encounted while writing tests for otherwise terrific libraries in Golang that are written this way (names withheld to protect the guilty!)

Status: WIP, handles the major stuff, thin on the corner cases. Again: purpose is _not_ to reinvent other more robust mocking libs, just to some real work with `go/ast`.

