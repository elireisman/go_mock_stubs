all: build

clean:
	rm stubber

setup: 
	go get -u golang.org/x/tools/cmd/goimports

build: *.go
	goimports -w main.go && go build -o stubber -v main.go

