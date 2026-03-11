.PHONY: build test lint clean run

build:
	go build ./...

test:
	go test -race -count=1 -timeout 60s ./...

lint:
	go vet ./...

clean:
	rm -f bin/cloud

run:
	go run ./cmd/cloud
