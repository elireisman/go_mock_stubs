build: *.go
	goimports -w main.go && go build -v main.go

