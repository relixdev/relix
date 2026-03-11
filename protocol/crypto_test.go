package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	if len(pub) != 32 {
		t.Errorf("public key length: got %d, want 32", len(pub))
	}
	if len(priv) != 32 {
		t.Errorf("private key length: got %d, want 32", len(priv))
	}

	// Keys should be different
	if bytes.Equal(pub[:], priv[:]) {
		t.Error("public and private keys are identical")
	}

	// Two calls should produce different keys
	pub2, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair (2): %v", err)
	}
	if bytes.Equal(pub[:], pub2[:]) {
		t.Error("two key pairs produced identical public keys")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// Alice and Bob generate key pairs
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, bobPriv, _ := GenerateKeyPair()

	plaintext := []byte("secret session data with code and approvals")

	// Alice encrypts for Bob
	ciphertext, err := Encrypt(plaintext, bobPub, alicePriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Ciphertext should be different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext equals plaintext")
	}

	// Ciphertext should be longer than plaintext (nonce + overhead)
	if len(ciphertext) <= len(plaintext) {
		t.Errorf("ciphertext not longer than plaintext: %d <= %d", len(ciphertext), len(plaintext))
	}

	// Bob decrypts
	decrypted, err := Decrypt(ciphertext, alicePub, bobPriv)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, _, _ := GenerateKeyPair()
	_, evePriv, _ := GenerateKeyPair()

	plaintext := []byte("secret data")
	ciphertext, _ := Encrypt(plaintext, bobPub, alicePriv)

	// Eve tries to decrypt with her private key (should fail)
	_, err := Decrypt(ciphertext, alicePub, evePriv)
	if err == nil {
		t.Error("expected decrypt to fail with wrong key, but it succeeded")
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, _, _ := GenerateKeyPair()
	_ = alicePub

	plaintext := []byte("same message")

	ct1, _ := Encrypt(plaintext, bobPub, alicePriv)
	ct2, _ := Encrypt(plaintext, bobPub, alicePriv)

	// Each encryption should use a random nonce, producing different ciphertext
	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of the same plaintext produced identical ciphertext")
	}
}

func TestEncryptDecryptEmptyMessage(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, bobPriv, _ := GenerateKeyPair()

	ciphertext, err := Encrypt([]byte{}, bobPub, alicePriv)
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, alicePub, bobPriv)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty, got %d bytes", len(decrypted))
	}
}

func TestSealOpenPayload(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, bobPriv, _ := GenerateKeyPair()

	payload := Payload{
		Kind: PayloadAssistantMessage,
		Seq:  1,
		Data: json.RawMessage(`{"text":"Hello"}`),
	}

	sealed, err := SealPayload(payload, bobPub, alicePriv)
	if err != nil {
		t.Fatalf("SealPayload: %v", err)
	}

	var opened Payload
	if err := OpenPayload(sealed, alicePub, bobPriv, &opened); err != nil {
		t.Fatalf("OpenPayload: %v", err)
	}

	if opened.Kind != PayloadAssistantMessage {
		t.Errorf("kind: got %q", opened.Kind)
	}
	if opened.Seq != 1 {
		t.Errorf("seq: got %d", opened.Seq)
	}
}
