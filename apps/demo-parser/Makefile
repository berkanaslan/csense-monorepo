test:
	go test -v -cover ./...

.PHONY: test

build:
	GOARCH=amd64 GOOS=linux go build -o output/bootstrap

.PHONY: build