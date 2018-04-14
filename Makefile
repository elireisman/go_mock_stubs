all: build

BINARY_NAME = gen_stubs

clean:
	@find examples -type f -name '*_mock.go' -delete
	@rm $(BINARY_NAME)

setup: 
	go get -u golang.org/x/tools/cmd/goimports

build: *.go
	find . -type f -name '*.go' -exec goimports -w {} \;
	go build -o $(BINARY_NAME) -v main.go

