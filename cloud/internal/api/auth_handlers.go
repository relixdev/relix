package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/relixdev/relix/cloud/internal/auth"
	"github.com/relixdev/relix/cloud/internal/user"
)

type authResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}

func (s *Server) handleAuthGitHub(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	ghUser, err := s.githubOAuth.ExchangeCode(r.Context(), req.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("github oauth: %v", err))
		return
	}

	githubIDStr := fmt.Sprintf("%d", ghUser.ID)
	u, err := s.userStore.GetUserByGitHubID(r.Context(), githubIDStr)
	if err != nil {
		// First login — create the user.
		u, err = s.userStore.CreateUser(r.Context(), &user.User{
			Email:    ghUser.Email,
			GitHubID: githubIDStr,
			Tier:     user.TierFree,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create user failed")
			return
		}
	}

	token, err := s.tokens.IssueToken(u.ID, auth.RoleMobile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token issue failed")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: u})
}

func (s *Server) handleEmailRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := s.emailAuth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := s.tokens.IssueToken(u.ID, auth.RoleMobile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token issue failed")
		return
	}
	writeJSON(w, http.StatusCreated, authResponse{Token: token, User: u})
}

func (s *Server) handleEmailLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := s.emailAuth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := s.tokens.IssueToken(u.ID, auth.RoleMobile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token issue failed")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: u})
}

func (s *Server) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	role := auth.RoleFromContext(r.Context())

	u, err := s.userStore.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	token, err := s.tokens.IssueToken(userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token issue failed")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: u})
}
