all: build

clean:
	rm stubber

setup: 
	go get -u golang.org/x/tools/cmd/goimports

build: *.go
	find . -type f -name '*.go' -exec goimports -w {} \;
	go build -o stubber -v main.go

