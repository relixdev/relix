package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// New returns a prefixed random ID, e.g. "usr_a3f9c2b1".
func New(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("idgen: rand.Read failed: %v", err))
	}
	return prefix + "_" + hex.EncodeToString(b)
}
