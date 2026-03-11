package crypto

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relixdev/protocol"
)

const (
	publicKeyFile  = "public.key"
	privateKeyFile = "private.key"
)

// KeyPair holds a loaded X25519 key pair.
type KeyPair struct {
	PublicKey  protocol.PublicKey
	PrivateKey protocol.PrivateKey
}

// GenerateAndSave generates a new X25519 key pair, writes the hex-encoded keys
// to dir/public.key and dir/private.key, and returns the pair.
// The private key file is written with 0600 permissions.
func GenerateAndSave(dir string) (*KeyPair, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create key dir: %w", err)
	}

	pub, priv, err := protocol.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate key pair: %w", err)
	}

	pubHex := hex.EncodeToString(pub[:])
	privHex := hex.EncodeToString(priv[:])

	if err := os.WriteFile(filepath.Join(dir, publicKeyFile), []byte(pubHex), 0644); err != nil {
		return nil, fmt.Errorf("write public key: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, privateKeyFile), []byte(privHex), 0600); err != nil {
		return nil, fmt.Errorf("write private key: %w", err)
	}

	return &KeyPair{PublicKey: pub, PrivateKey: priv}, nil
}

// Load reads an existing key pair from dir/public.key and dir/private.key.
func Load(dir string) (*KeyPair, error) {
	pub, err := readKey(filepath.Join(dir, publicKeyFile))
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	priv, err := readKey(filepath.Join(dir, privateKeyFile))
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	if len(pub) != 32 || len(priv) != 32 {
		return nil, fmt.Errorf("invalid key length: pub=%d priv=%d", len(pub), len(priv))
	}

	var kp KeyPair
	copy(kp.PublicKey[:], pub)
	copy(kp.PrivateKey[:], priv)
	return &kp, nil
}

func readKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(data))
}
