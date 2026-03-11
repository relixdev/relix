package user

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/relixdev/relix/cloud/internal/idgen"
)

// MemoryStore is an in-memory implementation of Store for development and testing.
type MemoryStore struct {
	mu            sync.RWMutex
	users         map[string]*User         // id → user
	byEmail       map[string]string        // email → id
	byGitHubID    map[string]string        // githubID → id
	subscriptions map[string]*Subscription // userID → subscription
}

// NewMemoryStore returns an initialised MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:         make(map[string]*User),
		byEmail:       make(map[string]string),
		byGitHubID:    make(map[string]string),
		subscriptions: make(map[string]*Subscription),
	}
}

func (s *MemoryStore) CreateUser(_ context.Context, u *User) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u.Email != "" {
		if _, exists := s.byEmail[u.Email]; exists {
			return nil, fmt.Errorf("user with email %q already exists", u.Email)
		}
	}

	copy := *u
	copy.ID = idgen.New("usr")
	copy.CreatedAt = time.Now().UTC()
	if copy.Tier == "" {
		copy.Tier = TierFree
	}

	s.users[copy.ID] = &copy
	if copy.Email != "" {
		s.byEmail[copy.Email] = copy.ID
	}
	if copy.GitHubID != "" {
		s.byGitHubID[copy.GitHubID] = copy.ID
	}
	return &copy, nil
}

func (s *MemoryStore) GetUserByID(_ context.Context, id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user %q not found", id)
	}
	copy := *u
	return &copy, nil
}

func (s *MemoryStore) GetUserByEmail(_ context.Context, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byEmail[email]
	if !ok {
		return nil, fmt.Errorf("user with email %q not found", email)
	}
	copy := *s.users[id]
	return &copy, nil
}

func (s *MemoryStore) GetUserByGitHubID(_ context.Context, githubID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byGitHubID[githubID]
	if !ok {
		return nil, fmt.Errorf("user with GitHub ID %q not found", githubID)
	}
	copy := *s.users[id]
	return &copy, nil
}

func (s *MemoryStore) UpdateUser(_ context.Context, u *User) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.users[u.ID]
	if !ok {
		return nil, fmt.Errorf("user %q not found", u.ID)
	}

	// Update email index if changed
	if existing.Email != u.Email {
		delete(s.byEmail, existing.Email)
		if u.Email != "" {
			s.byEmail[u.Email] = u.ID
		}
	}
	// Update GitHub index if changed
	if existing.GitHubID != u.GitHubID {
		delete(s.byGitHubID, existing.GitHubID)
		if u.GitHubID != "" {
			s.byGitHubID[u.GitHubID] = u.ID
		}
	}

	copy := *u
	s.users[u.ID] = &copy
	return &copy, nil
}

func (s *MemoryStore) UpsertSubscription(_ context.Context, sub *Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := *sub
	s.subscriptions[sub.UserID] = &copy
	// Also update the user's tier field for convenience
	if u, ok := s.users[sub.UserID]; ok {
		u.Tier = sub.Tier
	}
	return nil
}

func (s *MemoryStore) GetSubscription(_ context.Context, userID string) (*Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sub, ok := s.subscriptions[userID]
	if !ok {
		return nil, fmt.Errorf("subscription for user %q not found", userID)
	}
	copy := *sub
	return &copy, nil
}
