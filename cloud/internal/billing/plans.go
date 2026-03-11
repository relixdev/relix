package billing

import "github.com/relixdev/relix/cloud/internal/user"

// Resource names for limit checks.
const (
	ResourceMachines = "machines"
	ResourceSessions = "sessions"
)

// Unlimited signals no cap on a resource.
const Unlimited = -1

// Plan holds the limits and pricing for a subscription tier.
type Plan struct {
	Tier            user.Tier
	DisplayName     string
	MonthlyPriceCents int // 0 = free
	MachineLimit    int
	SessionLimit    int
}

var plans = map[user.Tier]Plan{
	user.TierFree: {
		Tier:              user.TierFree,
		DisplayName:       "Free",
		MonthlyPriceCents: 0,
		MachineLimit:      3,
		SessionLimit:      2,
	},
	user.TierPlus: {
		Tier:              user.TierPlus,
		DisplayName:       "Plus",
		MonthlyPriceCents: 499,
		MachineLimit:      10,
		SessionLimit:      5,
	},
	user.TierPro: {
		Tier:              user.TierPro,
		DisplayName:       "Pro",
		MonthlyPriceCents: 1499,
		MachineLimit:      Unlimited,
		SessionLimit:      Unlimited,
	},
	user.TierTeam: {
		Tier:              user.TierTeam,
		DisplayName:       "Team",
		MonthlyPriceCents: 2499,
		MachineLimit:      Unlimited,
		SessionLimit:      Unlimited,
	},
}

// GetPlan returns the Plan for a tier. Falls back to Free for unknown tiers.
func GetPlan(tier user.Tier) Plan {
	p, ok := plans[tier]
	if !ok {
		return plans[user.TierFree]
	}
	return p
}

// CheckLimit returns true if the given current count is within the tier's limit
// for the named resource.
func CheckLimit(tier user.Tier, resource string, current int) bool {
	p := GetPlan(tier)
	var limit int
	switch resource {
	case ResourceMachines:
		limit = p.MachineLimit
	case ResourceSessions:
		limit = p.SessionLimit
	default:
		return true
	}
	if limit == Unlimited {
		return true
	}
	return current < limit
}
