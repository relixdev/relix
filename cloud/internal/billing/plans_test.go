package billing_test

import (
	"testing"

	"github.com/relixdev/relix/cloud/internal/billing"
	"github.com/relixdev/relix/cloud/internal/user"
)

func TestCheckLimit_Free(t *testing.T) {
	// Free tier: 3 machines, 2 sessions
	cases := []struct {
		resource string
		current  int
		want     bool
	}{
		{billing.ResourceMachines, 0, true},
		{billing.ResourceMachines, 2, true},
		{billing.ResourceMachines, 3, false}, // at limit
		{billing.ResourceSessions, 1, true},
		{billing.ResourceSessions, 2, false},
	}
	for _, tc := range cases {
		got := billing.CheckLimit(user.TierFree, tc.resource, tc.current)
		if got != tc.want {
			t.Errorf("CheckLimit(free, %s, %d) = %v, want %v",
				tc.resource, tc.current, got, tc.want)
		}
	}
}

func TestCheckLimit_Plus(t *testing.T) {
	if !billing.CheckLimit(user.TierPlus, billing.ResourceMachines, 9) {
		t.Error("Plus should allow 9 machines")
	}
	if billing.CheckLimit(user.TierPlus, billing.ResourceMachines, 10) {
		t.Error("Plus should deny 10 machines (limit is 10)")
	}
}

func TestCheckLimit_Pro_Unlimited(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if !billing.CheckLimit(user.TierPro, billing.ResourceMachines, i) {
			t.Fatalf("Pro should allow unlimited machines, failed at %d", i)
		}
	}
}

func TestCheckLimit_UnknownResource(t *testing.T) {
	if !billing.CheckLimit(user.TierFree, "unknown_resource", 9999) {
		t.Error("unknown resource should always pass")
	}
}

func TestGetPlan(t *testing.T) {
	p := billing.GetPlan(user.TierPlus)
	if p.MonthlyPriceCents != 499 {
		t.Errorf("Plus price = %d, want 499", p.MonthlyPriceCents)
	}
	if p.MachineLimit != 10 {
		t.Errorf("Plus machine limit = %d, want 10", p.MachineLimit)
	}
}
