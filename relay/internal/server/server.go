package server

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/auth"
	"github.com/relixdev/relix/relay/internal/conn"
	"github.com/relixdev/relix/relay/internal/hub"
	"nhooyr.io/websocket"
)

// Options configures optional server behaviour.
type Options struct {
	// PairRateMaxAttempts is the max pairing attempts per window per IP.
	// Zero uses the default (5).
	PairRateMaxAttempts int
	// PairRateWindow is the sliding window for rate limiting.
	// Zero uses the default (3s).
	PairRateWindow time.Duration
}

// Server is the relay HTTP handler.
type Server struct {
	mux         *http.ServeMux
	hub         *hub.Hub
	jwtSecret   string
	pairLimiter *hub.RateLimiter
}

// New creates a Server with routes registered using default options.
func New(jwtSecret string, h *hub.Hub) *Server {
	return NewWithOptions(jwtSecret, h, Options{})
}

// NewWithOptions creates a Server with routes registered and custom options.
func NewWithOptions(jwtSecret string, h *hub.Hub, opts Options) *Server {
	maxAttempts := opts.PairRateMaxAttempts
	window := opts.PairRateWindow
	if maxAttempts <= 0 {
		maxAttempts = hub.DefaultRateLimitMaxAttempts
	}
	if window <= 0 {
		window = hub.DefaultRateLimitWindow
	}

	s := &Server{
		mux:         http.NewServeMux(),
		hub:         h,
		jwtSecret:   jwtSecret,
		pairLimiter: hub.NewRateLimiter(maxAttempts, window),
	}
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ws", s.handleWS)
	s.mux.HandleFunc("/pair", s.handlePair)
	s.mux.Handle("/metrics", promhttp.Handler())
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}

	c := conn.New(ws)
	ctx := r.Context()

	// Step 1: read the first message — must be MsgAuth.
	authEnv, err := c.ReadEnvelope(ctx)
	if err != nil {
		ws.Close(websocket.StatusPolicyViolation, "auth read failed")
		return
	}
	if authEnv.Type != protocol.MsgAuth {
		ws.Close(websocket.StatusPolicyViolation, "first message must be auth")
		return
	}

	// Step 2: parse the auth payload for the JWT token.
	var authMsg protocol.AuthMessage
	if err := json.Unmarshal(authEnv.Payload, &authMsg); err != nil || authMsg.Token == "" {
		ws.Close(websocket.StatusPolicyViolation, "invalid auth payload")
		return
	}

	// Step 3: validate JWT.
	claims, err := auth.ValidateToken(authMsg.Token, s.jwtSecret)
	if err != nil {
		ws.Close(websocket.StatusPolicyViolation, "invalid token")
		return
	}

	userID := claims.UserID
	role := string(claims.Role)
	machineID := authEnv.MachineID

	// Step 4: register connection in hub.
	switch claims.Role {
	case auth.RoleAgent:
		s.hub.RegisterAgent(userID, machineID, c)
		defer s.hub.UnregisterAgent(userID, machineID)
	case auth.RoleMobile:
		s.hub.RegisterMobile(userID, c)
		defer s.hub.UnregisterMobile(userID, c)
	}

	// Step 5: read loop — route envelopes.
	for {
		env, err := c.ReadEnvelope(ctx)
		if err != nil {
			break
		}
		_ = s.hub.RouteEnvelope(userID, role, env)
	}
}

// pairRequest is the JSON body for POST /pair.
type pairRequest struct {
	Code string `json:"code"`
}

// pairResponse is the JSON body returned on successful pairing.
type pairResponse struct {
	UserID string `json:"user_id"`
}

func (s *Server) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate limit by client IP (strip port so all connections from same host share a bucket).
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr // fall back to raw value if parsing fails
	}
	if !s.pairLimiter.Allow(ip) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	var req pairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		http.Error(w, "bad request: code required", http.StatusBadRequest)
		return
	}

	userID, ok := s.hub.Pairing().ValidatePairing(req.Code)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pairResponse{UserID: userID})
}
