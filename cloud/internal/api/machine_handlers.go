package api

import (
	"encoding/json"
	"net/http"

	"github.com/relixdev/relix/cloud/internal/auth"
	"github.com/relixdev/relix/cloud/internal/user"
)

func (s *Server) handleListMachines(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	machines, err := s.registry.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if machines == nil {
		machines = []*user.Machine{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"machines": machines})
}

func (s *Server) handleRegisterMachine(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		Name      string `json:"name"`
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.PublicKey == "" {
		writeError(w, http.StatusBadRequest, "name and public_key are required")
		return
	}

	m, err := s.registry.Register(r.Context(), userID, req.Name, req.PublicKey)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (s *Server) handleDeleteMachine(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	machineID := r.PathValue("id")
	if machineID == "" {
		writeError(w, http.StatusBadRequest, "machine id is required")
		return
	}

	if err := s.registry.Delete(r.Context(), userID, machineID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
