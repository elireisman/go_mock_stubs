#### The Idea
An experiment to play with `go/ast` functionality. The tool generates unit test mocks from Golang source files where the public API is exposed as pointers to structs rather than interfaces. Given a Golang source code file, the tool will extract any public structs and their public methods, and generate a matching interface along with a package-private mock struct implementing no-ops for all methods.

#### Example

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

