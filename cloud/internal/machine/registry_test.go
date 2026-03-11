package machine_test

import (
	"context"
	"testing"

	"github.com/relixdev/relix/cloud/internal/machine"
	"github.com/relixdev/relix/cloud/internal/user"
)

func setup(t *testing.T) (context.Context, *machine.Registry, *user.MemoryStore) {
	t.Helper()
	ctx := context.Background()
	store := user.NewMemoryStore()
	reg := machine.NewRegistry(store)
	return ctx, reg, store
}

func createUser(t *testing.T, ctx context.Context, store *user.MemoryStore, tier user.Tier) *user.User {
	t.Helper()
	u, err := store.CreateUser(ctx, &user.User{Email: "test@example.com", Tier: tier})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func TestRegister_Basic(t *testing.T) {
	ctx, reg, store := setup(t)
	u := createUser(t, ctx, store, user.TierFree)

	m, err := reg.Register(ctx, u.ID, "laptop", "pubkey123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if m.ID == "" {
		t.Error("machine ID should be set")
	}
	if m.UserID != u.ID {
		t.Errorf("machine UserID = %q, want %q", m.UserID, u.ID)
	}
	if m.Name != "laptop" {
		t.Errorf("machine Name = %q, want laptop", m.Name)
	}
}

func TestRegister_FreeTierLimit(t *testing.T) {
	ctx, reg, store := setup(t)
	u := createUser(t, ctx, store, user.TierFree)

	// Free tier allows 3 machines
	for i := 0; i < 3; i++ {
		if _, err := reg.Register(ctx, u.ID, "machine", "key"); err != nil {
			t.Fatalf("register %d: %v", i, err)
		}
	}

	// 4th should fail
	_, err := reg.Register(ctx, u.ID, "extra", "key")
	if err == nil {
		t.Error("expected error registering 4th machine on free tier")
	}
}

func TestRegister_PlusTierLimit(t *testing.T) {
	ctx, reg, store := setup(t)
	u := createUser(t, ctx, store, user.TierPlus)

	for i := 0; i < 10; i++ {
		if _, err := reg.Register(ctx, u.ID, "m", "k"); err != nil {
			t.Fatalf("register %d: %v", i, err)
		}
	}
	_, err := reg.Register(ctx, u.ID, "extra", "k")
	if err == nil {
		t.Error("expected error registering 11th machine on plus tier")
	}
}

func TestRegister_ProTierUnlimited(t *testing.T) {
	ctx, reg, store := setup(t)
	u := createUser(t, ctx, store, user.TierPro)

	for i := 0; i < 20; i++ {
		if _, err := reg.Register(ctx, u.ID, "m", "k"); err != nil {
			t.Fatalf("register %d on pro tier: %v", i, err)
		}
	}
}

func TestList(t *testing.T) {
	ctx, reg, store := setup(t)
	u := createUser(t, ctx, store, user.TierFree)

	reg.Register(ctx, u.ID, "laptop", "key1")
	reg.Register(ctx, u.ID, "server", "key2")

	machines, err := reg.List(ctx, u.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(machines) != 2 {
		t.Errorf("list returned %d machines, want 2", len(machines))
	}
}

func TestDelete(t *testing.T) {
	ctx, reg, store := setup(t)
	u := createUser(t, ctx, store, user.TierFree)

	m, _ := reg.Register(ctx, u.ID, "laptop", "key1")

	if err := reg.Delete(ctx, u.ID, m.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	machines, _ := reg.List(ctx, u.ID)
	if len(machines) != 0 {
		t.Errorf("expected 0 machines after delete, got %d", len(machines))
	}
}

func TestDelete_WrongUser(t *testing.T) {
	ctx, reg, store := setup(t)
	u1 := createUser(t, ctx, store, user.TierFree)
	u2, _ := store.CreateUser(ctx, &user.User{Email: "other@example.com"})

	m, _ := reg.Register(ctx, u1.ID, "laptop", "key")

	err := reg.Delete(ctx, u2.ID, m.ID)
	if err == nil {
		t.Error("expected error deleting another user's machine")
	}
}

func TestUpgradeTierAllowsMoreMachines(t *testing.T) {
	ctx, reg, store := setup(t)
	u := createUser(t, ctx, store, user.TierFree)

	for i := 0; i < 3; i++ {
		reg.Register(ctx, u.ID, "m", "k")
	}

	// Hit free limit
	_, err := reg.Register(ctx, u.ID, "extra", "k")
	if err == nil {
		t.Fatal("expected limit error")
	}

	// Upgrade to Plus
	u.Tier = user.TierPlus
	store.UpdateUser(ctx, u)

	// Now should work
	_, err = reg.Register(ctx, u.ID, "extra", "k")
	if err != nil {
		t.Errorf("after upgrade to Plus, register failed: %v", err)
	}
}
