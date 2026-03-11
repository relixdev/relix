package machine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/relixdev/relix/cloud/internal/billing"
	"github.com/relixdev/relix/cloud/internal/idgen"
	"github.com/relixdev/relix/cloud/internal/user"
)

// Registry manages machine registration and enforces per-tier limits.
type Registry struct {
	mu       sync.RWMutex
	machines map[string]*user.Machine   // machineID → machine
	byUser   map[string][]string        // userID → []machineID
	store    user.Store
}

// NewRegistry creates a Registry backed by the given user store (used to look
// up tier information for limit enforcement).
func NewRegistry(store user.Store) *Registry {
	return &Registry{
		machines: make(map[string]*user.Machine),
		byUser:   make(map[string][]string),
		store:    store,
	}
}

// Register adds a new machine for the given user, returning an error if the
// user's tier limit has been reached.
func (r *Registry) Register(ctx context.Context, userID, name, publicKey string) (*user.Machine, error) {
	u, err := r.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("machine: get user: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	current := len(r.byUser[userID])
	if !billing.CheckLimit(u.Tier, billing.ResourceMachines, current) {
		plan := billing.GetPlan(u.Tier)
		return nil, fmt.Errorf("machine: limit reached: %s tier allows %d machines",
			plan.DisplayName, plan.MachineLimit)
	}

	m := &user.Machine{
		ID:        idgen.New("mch"),
		UserID:    userID,
		Name:      name,
		PublicKey: publicKey,
		CreatedAt: time.Now().UTC(),
	}
	r.machines[m.ID] = m
	r.byUser[userID] = append(r.byUser[userID], m.ID)
	return m, nil
}

// List returns all machines registered to a user.
func (r *Registry) List(ctx context.Context, userID string) ([]*user.Machine, error) {
	// Verify user exists.
	if _, err := r.store.GetUserByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("machine: get user: %w", err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byUser[userID]
	result := make([]*user.Machine, 0, len(ids))
	for _, id := range ids {
		if m, ok := r.machines[id]; ok {
			copy := *m
			result = append(result, &copy)
		}
	}
	return result, nil
}

// Delete removes a machine, verifying it belongs to the given user.
func (r *Registry) Delete(ctx context.Context, userID, machineID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.machines[machineID]
	if !ok {
		return fmt.Errorf("machine: %q not found", machineID)
	}
	if m.UserID != userID {
		return fmt.Errorf("machine: %q does not belong to user %q", machineID, userID)
	}

	delete(r.machines, machineID)

	ids := r.byUser[userID]
	filtered := ids[:0]
	for _, id := range ids {
		if id != machineID {
			filtered = append(filtered, id)
		}
	}
	r.byUser[userID] = filtered
	return nil
}
