build:
	go build -o relay ./cmd/relay/

test:
	go test -race ./...

docker:
	docker build -t relix-relay .

clean:
	rm -f relay
