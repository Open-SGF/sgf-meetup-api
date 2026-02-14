.PHONY: test lint

test:
	go test ./cmd/... ./pkg/...

lint:
	golangci-lint run ./cmd/... ./pkg/...
