package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/crypto"
)

// PairingRequest is the JSON message sent by the agent to initiate pairing.
type PairingRequest struct {
	Code      string             `json:"code"`
	PublicKey protocol.PublicKey `json:"public_key"`
}

// PairingResponse is the JSON message received from the relay during pairing.
type PairingResponse struct {
	PeerPublicKey protocol.PublicKey `json:"peer_public_key"`
}

// sasList is the emoji palette used for SAS generation.
// All entries are single Unicode codepoints in the supplementary planes.
var sasList = []rune{
	0x1F600, 0x1F601, 0x1F602, 0x1F603, 0x1F604, 0x1F605, 0x1F606, 0x1F607,
	0x1F608, 0x1F609, 0x1F60A, 0x1F60B, 0x1F60C, 0x1F60D, 0x1F60E, 0x1F60F,
	0x1F610, 0x1F611, 0x1F612, 0x1F613, 0x1F614, 0x1F615, 0x1F616, 0x1F617,
	0x1F618, 0x1F619, 0x1F61A, 0x1F61B, 0x1F61C, 0x1F61D, 0x1F61E, 0x1F61F,
	0x1F620, 0x1F621, 0x1F622, 0x1F623, 0x1F624, 0x1F625, 0x1F626, 0x1F627,
	0x1F628, 0x1F629, 0x1F62A, 0x1F62B, 0x1F62C, 0x1F62D, 0x1F62E, 0x1F62F,
	0x1F630, 0x1F631, 0x1F632, 0x1F633, 0x1F634, 0x1F635, 0x1F636, 0x1F637,
	0x1F638, 0x1F639, 0x1F63A, 0x1F63B, 0x1F63C, 0x1F63D, 0x1F63E, 0x1F63F,
	0x1F640, 0x1F641, 0x1F642, 0x1F643, 0x1F644, 0x1F645, 0x1F646, 0x1F647,
	0x1F648, 0x1F649, 0x1F64A, 0x1F64B, 0x1F64C, 0x1F64D, 0x1F64E, 0x1F64F,
	0x1F680, 0x1F681, 0x1F682, 0x1F683, 0x1F684, 0x1F685, 0x1F686, 0x1F687,
	0x1F688, 0x1F689, 0x1F68A, 0x1F68B, 0x1F68C, 0x1F68D, 0x1F68E, 0x1F68F,
	0x1F690, 0x1F691, 0x1F692, 0x1F693, 0x1F694, 0x1F695, 0x1F696, 0x1F697,
	0x1F698, 0x1F699, 0x1F69A, 0x1F69B, 0x1F69C, 0x1F69D, 0x1F69E, 0x1F69F,
	0x1F6A0, 0x1F6A1, 0x1F6A2, 0x1F6A3, 0x1F6A4, 0x1F6A5, 0x1F6A6, 0x1F6A7,
	0x1F6A8, 0x1F6A9, 0x1F6AA, 0x1F6AB, 0x1F6AC, 0x1F6AD, 0x1F6AE, 0x1F6AF,
	0x1F6B0, 0x1F6B1, 0x1F6B2, 0x1F6B3, 0x1F6B4, 0x1F6B5, 0x1F6B6, 0x1F6B7,
	0x1F6B8, 0x1F6B9, 0x1F6BA, 0x1F6BB, 0x1F6BC, 0x1F6BD, 0x1F6BE, 0x1F6BF,
	0x1F6C0, 0x1F6C1, 0x1F6C2, 0x1F6C3, 0x1F6C4, 0x1F6C5, 0x1F300, 0x1F301,
	0x1F302, 0x1F303, 0x1F304, 0x1F305, 0x1F306, 0x1F307, 0x1F308, 0x1F309,
	0x1F30A, 0x1F30B, 0x1F30C, 0x1F30D, 0x1F30E, 0x1F30F, 0x1F310, 0x1F311,
	0x1F312, 0x1F313, 0x1F314, 0x1F315, 0x1F316, 0x1F317, 0x1F318, 0x1F319,
	0x1F31A, 0x1F31B, 0x1F31C, 0x1F31D, 0x1F31E, 0x1F31F, 0x1F320, 0x1F330,
	0x1F331, 0x1F332, 0x1F333, 0x1F334, 0x1F335, 0x1F337, 0x1F338, 0x1F339,
	0x1F33A, 0x1F33B, 0x1F33C, 0x1F33D, 0x1F33E, 0x1F33F, 0x1F340, 0x1F341,
	0x1F342, 0x1F343, 0x1F344, 0x1F345, 0x1F346, 0x1F347, 0x1F348, 0x1F349,
	0x1F34A, 0x1F34B, 0x1F34C, 0x1F34D, 0x1F34E, 0x1F34F, 0x1F350, 0x1F351,
	0x1F352, 0x1F353, 0x1F354, 0x1F355, 0x1F356, 0x1F357, 0x1F358, 0x1F359,
	0x1F35A, 0x1F35B, 0x1F35C, 0x1F35D, 0x1F35E, 0x1F35F, 0x1F360, 0x1F361,
	0x1F362, 0x1F363, 0x1F364, 0x1F365, 0x1F366, 0x1F367, 0x1F368, 0x1F369,
	0x1F36A, 0x1F36B, 0x1F36C, 0x1F36D, 0x1F36E, 0x1F36F, 0x1F370, 0x1F371,
	0x1F372, 0x1F373, 0x1F374, 0x1F375, 0x1F376, 0x1F377, 0x1F378, 0x1F379,
}

// Pair connects to the relay WebSocket, sends a pairing request with the given
// code and own public key, receives the peer's public key, prints a 4-emoji
// SAS for visual verification, saves the peer public key to keysDir/peer.pub,
// and returns the peer public key.
func Pair(relayURL, authToken, code string, ks *crypto.KeyPair, keysDir string) (protocol.PublicKey, error) {
	ctx := context.Background()

	conn, _, err := websocket.Dial(ctx, relayURL, nil)
	if err != nil {
		return protocol.PublicKey{}, fmt.Errorf("pair: dial relay: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Send pairing request.
	req := PairingRequest{
		Code:      code,
		PublicKey: ks.PublicKey,
	}
	if err := wsjson.Write(ctx, conn, req); err != nil {
		return protocol.PublicKey{}, fmt.Errorf("pair: send request: %w", err)
	}

	// Receive peer public key.
	var resp PairingResponse
	if err := wsjson.Read(ctx, conn, &resp); err != nil {
		return protocol.PublicKey{}, fmt.Errorf("pair: read response: %w", err)
	}

	// Display SAS for visual verification.
	sas := GenerateSAS(ks.PublicKey, resp.PeerPublicKey)
	fmt.Printf("Verify these 4 emoji match on your mobile device: %s\n", sas)

	// Save peer public key.
	if err := savePeerPublicKey(keysDir, resp.PeerPublicKey); err != nil {
		return protocol.PublicKey{}, fmt.Errorf("pair: save peer key: %w", err)
	}

	return resp.PeerPublicKey, nil
}

// GenerateSAS derives a deterministic 4-emoji string from two public keys.
// The result is the same regardless of argument order (symmetric).
// Algorithm: SHA256 of (sorted concatenation of both keys), map first 4 bytes
// to the emoji list.
func GenerateSAS(pub1, pub2 protocol.PublicKey) string {
	// Sort the keys so the result is symmetric.
	a, b := pub1[:], pub2[:]
	if compareBytes(a, b) > 0 {
		a, b = b, a
	}

	h := sha256.New()
	h.Write(a)
	h.Write(b)
	digest := h.Sum(nil)

	emojis := make([]rune, 4)
	for i := 0; i < 4; i++ {
		idx := int(digest[i]) % len(sasList)
		emojis[i] = sasList[idx]
	}
	return string(emojis)
}

func compareBytes(a, b []byte) int {
	for i := range a {
		if i >= len(b) {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func savePeerPublicKey(dir string, pub protocol.PublicKey) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create keys dir: %w", err)
	}
	hexKey := hex.EncodeToString(pub[:])
	path := filepath.Join(dir, "peer.pub")
	return os.WriteFile(path, []byte(hexKey), 0644)
}
