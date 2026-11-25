.PHONY: run build test test-http vet
run:
	go run ./cmd/dataagent
build:
	go build -o bin/dataagent ./cmd/dataagent
test:
	go test ./...
test-http:
	ODB_HTTP_TEST=1 go test ./... -count=1
vet:
	go vet ./...
