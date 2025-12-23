.PHONY: build test clean fmt vet

build:
	go build -o websearch-mcp ./cmd/websearch-mcp

test:
	go test -v ./...

clean:
	rm -f websearch-mcp

fmt:
	go fmt ./...

vet:
	go vet ./...
