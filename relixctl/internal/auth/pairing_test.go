package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/auth"
	"github.com/relixdev/relix/relixctl/internal/crypto"
)

func TestPair(t *testing.T) {
	ownKP, err := generateTestKeyPair(t)
	if err != nil {
		t.Fatalf("generate own key pair: %v", err)
	}
	peerKP, err := generateTestKeyPair(t)
	if err != nil {
		t.Fatalf("generate peer key pair: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx := r.Context()

		var req auth.PairingRequest
		if err := wsjson.Read(ctx, conn, &req); err != nil {
			return
		}

		resp := auth.PairingResponse{
			PeerPublicKey: peerKP.PublicKey,
		}
		_ = wsjson.Write(ctx, conn, resp)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	peerDir := t.TempDir()
	peerPub, err := auth.Pair(wsURL, "auth-token", "PAIR-CODE", ownKP, peerDir)
	if err != nil {
		t.Fatalf("Pair returned error: %v", err)
	}
	if peerPub != peerKP.PublicKey {
		t.Errorf("got wrong peer public key")
	}
}

func TestGenerateSAS_Deterministic(t *testing.T) {
	pub1, _, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pub2, _, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	sas1 := auth.GenerateSAS(pub1, pub2)
	sas2 := auth.GenerateSAS(pub1, pub2)
	if sas1 != sas2 {
		t.Errorf("GenerateSAS not deterministic: %q vs %q", sas1, sas2)
	}
	if sas1 == "" {
		t.Error("GenerateSAS returned empty string")
	}
}

func TestGenerateSAS_Symmetric(t *testing.T) {
	pub1, _, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pub2, _, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	sasA := auth.GenerateSAS(pub1, pub2)
	sasB := auth.GenerateSAS(pub2, pub1)
	if sasA != sasB {
		t.Errorf("GenerateSAS not symmetric: %q vs %q", sasA, sasB)
	}
}

func TestGenerateSAS_FourEmojis(t *testing.T) {
	pub1, _, _ := protocol.GenerateKeyPair()
	pub2, _, _ := protocol.GenerateKeyPair()
	sas := auth.GenerateSAS(pub1, pub2)

	runes := []rune(sas)
	if len(runes) != 4 {
		t.Errorf("expected 4 emoji runes, got %d in %q", len(runes), sas)
	}
}

func generateTestKeyPair(t *testing.T) (*crypto.KeyPair, error) {
	t.Helper()
	dir := t.TempDir()
	return crypto.GenerateAndSave(dir)
}

// Ensure context import is used by helpers (websocket.Accept uses it internally).
var _ = context.Background
