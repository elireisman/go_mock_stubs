build: *.go
	goimports -w main.go && go build -o stubber -v main.go

