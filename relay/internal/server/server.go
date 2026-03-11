package server

import (
	"encoding/json"
	"net/http"

	"nhooyr.io/websocket"
)

// Server is the relay HTTP handler.
type Server struct {
	mux *http.ServeMux
}

// New creates a Server with routes registered.
func New() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ws", s.handleWS)
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
	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		// Accept already wrote the error response.
		return
	}
	// Placeholder: close immediately until full handler is wired.
	ws.Close(websocket.StatusNormalClosure, "not yet implemented")
}
