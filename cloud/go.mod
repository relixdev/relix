module github.com/relixdev/relix/cloud

go 1.22

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	golang.org/x/crypto v0.31.0
)

require github.com/lib/pq v1.11.2

require github.com/stripe/stripe-go/v76 v76.25.0

replace github.com/relixdev/protocol => ../protocol
