all: build

BINARY_NAME = gen_stubs

clean:
	@find examples -type f -name '*_mock.go' -delete
	@rm $(BINARY_NAME) &>/dev/null

format:
	go get -u golang.org/x/tools/cmd/goimports
	find . -type f -name '*.go' -exec goimports -w {} \;

build: *.go
	go build -o $(BINARY_NAME) -v main.go

