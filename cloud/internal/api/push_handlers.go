package api

import (
	"encoding/json"
	"net/http"

	"github.com/relixdev/relix/cloud/internal/push"
)

func (s *Server) handlePushRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceToken string `json:"device_token"`
		Platform    string `json:"platform"` // "apns" or "fcm"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DeviceToken == "" {
		writeError(w, http.StatusBadRequest, "device_token is required")
		return
	}
	// TODO: persist device token to DB with userID association.
	writeJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

func (s *Server) handlePushSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceToken string            `json:"device_token"`
		Title       string            `json:"title"`
		Body        string            `json:"body"`
		Data        map[string]string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DeviceToken == "" || req.Title == "" {
		writeError(w, http.StatusBadRequest, "device_token and title are required")
		return
	}

	n := push.Notification{
		DeviceToken: req.DeviceToken,
		Title:       req.Title,
		Body:        req.Body,
		Data:        req.Data,
	}
	if err := s.push.Send(r.Context(), n); err != nil {
		writeError(w, http.StatusInternalServerError, "push send failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}
