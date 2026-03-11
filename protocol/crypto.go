package protocol

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

const nonceSize = 24

// PublicKey is a 32-byte X25519 public key.
type PublicKey [32]byte

// PrivateKey is a 32-byte X25519 private key.
type PrivateKey [32]byte

// GenerateKeyPair generates a new X25519 key pair for NaCl box.
func GenerateKeyPair() (PublicKey, PrivateKey, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return PublicKey{}, PrivateKey{}, err
	}
	return PublicKey(*pub), PrivateKey(*priv), nil
}

// Encrypt encrypts plaintext using NaCl box (X25519 + XSalsa20-Poly1305).
// Returns: nonce (24 bytes) || ciphertext.
func Encrypt(plaintext []byte, recipientPub PublicKey, senderPriv PrivateKey) ([]byte, error) {
	var nonce [nonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	pub := [32]byte(recipientPub)
	priv := [32]byte(senderPriv)

	sealed := box.Seal(nonce[:], plaintext, &nonce, &pub, &priv)
	return sealed, nil
}

// Decrypt decrypts a message produced by Encrypt.
// Expects: nonce (24 bytes) || ciphertext.
func Decrypt(message []byte, senderPub PublicKey, recipientPriv PrivateKey) ([]byte, error) {
	if len(message) < nonceSize+box.Overhead {
		return nil, errors.New("message too short")
	}

	var nonce [nonceSize]byte
	copy(nonce[:], message[:nonceSize])

	pub := [32]byte(senderPub)
	priv := [32]byte(recipientPriv)

	plaintext, ok := box.Open(nil, message[nonceSize:], &nonce, &pub, &priv)
	if !ok {
		return nil, errors.New("decryption failed: invalid key or corrupted message")
	}

	return plaintext, nil
}
