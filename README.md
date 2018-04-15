#### The Idea
An experiment to play with `go/ast` functionality. The tool generates unit test mocks from Golang source files where the public API is exposed as pointers to structs rather than interfaces. Given a Golang source code file, the tool will extract any public structs and their public methods, and generate a matching interface along with a mock struct implementing no-ops for all methods.


#### Usage Example
After generating the method stubs, end users should:
1. Replace calls to the public API exposed as a struct pointer `*example.Thing` with calls to the generated interface `example.ThingIface`
2. Create a new struct embedding the generated `elastic.MockThing`, and implement only methods from the mock you'd like to test
3. Write tests against the methods you converted in step 1, supplying the mock implementation from step 2


#### Code Generation Example
```bash
# from repo root dir, build the binary:
make

# print output mock code to stdout
./gen_stubs --source-dir=examples --stdout

# writes output files based on input file paths, as: example/*_mock.go
./gen_stubs --source=dir=examples

# better example
mkdir tmp
pushd tmp && wget -q https://raw.githubusercontent.com/olivere/elastic/release-branch.v6/client.go && popd
./gen_stubs --source-dir=tmp --stdout
```

