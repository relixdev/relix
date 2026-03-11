module github.com/relixdev/relix/relay

go 1.26.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/relixdev/protocol v0.0.0
	nhooyr.io/websocket v1.8.17
)

require (
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)

replace github.com/relixdev/protocol => ../protocol
