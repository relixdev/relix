# Relix Revenue Projections

**Date:** March 14, 2026
**Product:** Relix (relix.sh)
**Company:** ShadowScale (shadowscale.dev)

---

## Pricing Tiers

| Tier | Monthly | Annual (2 mo free) | Machines | Sessions | History |
|------|---------|---------------------|----------|----------|---------|
| **Free** | $0 | — | 1 | 2 | 7 days |
| **Plus** | $4.99 | $49.90/yr | 3 | 5 | 30 days |
| **Pro** | $14.99 | $149.90/yr | 10 | Unlimited | 90 days |
| **Team** | $24.99/user | Custom | Unlimited | Unlimited | 90 days |

---

## Cost Structure

### Fixed Monthly Costs

| Item | Cost | Notes |
|------|------|-------|
| Fly.io (relay) | $15 | 1 shared-cpu instance, scales with demand |
| Fly.io (cloud API) | $15 | 1 shared-cpu instance |
| Fly.io (PostgreSQL) | $15 | Managed Postgres, smallest tier |
| Domain (relix.sh) | ~$2.50 | $30/year amortized |
| Apple Developer Program | ~$8.25 | $99/year amortized |
| Google Play Developer | ~$0 | $25 one-time (already paid) |
| **Total fixed** | **~$56/month** | |

### Variable Costs Per Transaction

| Cost | Rate | Notes |
|------|------|-------|
| Stripe processing | 2.9% + $0.30 per charge | Standard Stripe pricing |
| Apple App Store cut | 15% (first $1M/yr) | Small Business Program rate |
| Google Play cut | 15% (first $1M/yr) | Standard small developer rate |

### Effective Revenue Per Subscriber

After payment processing fees:

| Tier | Price | Stripe Fee | Net (Direct) | Net (iOS/Android, 15% cut) |
|------|-------|------------|-------------|---------------------------|
| Plus (monthly) | $4.99 | $0.44 | $4.55 | $3.81 |
| Plus (annual) | $49.90 | $1.75 | $48.15 | $40.42 |
| Pro (monthly) | $14.99 | $0.73 | $14.26 | $12.01 |
| Pro (annual) | $149.90 | $4.65 | $145.25 | $122.77 |
| Team (monthly) | $24.99 | $1.02 | $23.97 | N/A (web only) |

**Assumption:** 60% of consumer subscriptions come through app stores (15% cut), 40% through web (Stripe only). Team tier is 100% web.

**Blended net revenue per Plus subscriber:** ~$4.11/month
**Blended net revenue per Pro subscriber:** ~$12.91/month
**Net revenue per Team subscriber:** ~$23.97/month

---

## Month 1-12 Projections (Conservative)

### Assumptions

- Month 1 launch drives 100 free signups
- Free user growth: 50% month-over-month for months 1-6, 30% for months 7-12
- Paid conversion rate: 10% of active free users convert within 30 days
- Tier mix of paid: 70% Plus, 25% Pro, 5% Team
- Monthly churn: 5% for Plus, 3% for Pro, 2% for Team
- 50% of users on annual plans by month 6

### User Growth

| Month | New Free | Total Free | New Paid | Total Paid | Plus | Pro | Team |
|-------|----------|------------|----------|------------|------|-----|------|
| 1 | 100 | 100 | 10 | 10 | 7 | 2 | 1 |
| 2 | 150 | 240 | 14 | 23 | 16 | 6 | 1 |
| 3 | 225 | 445 | 20 | 40 | 28 | 10 | 2 |
| 4 | 338 | 748 | 30 | 65 | 45 | 16 | 4 |
| 5 | 506 | 1,197 | 45 | 103 | 72 | 26 | 5 |
| 6 | 759 | 1,862 | 67 | 159 | 111 | 40 | 8 |
| 7 | 987 | 2,698 | 80 | 225 | 157 | 56 | 12 |
| 8 | 1,283 | 3,762 | 100 | 307 | 215 | 77 | 15 |
| 9 | 1,668 | 5,119 | 120 | 404 | 283 | 101 | 20 |
| 10 | 2,168 | 6,862 | 150 | 524 | 367 | 131 | 26 |
| 11 | 2,819 | 9,106 | 185 | 671 | 470 | 168 | 33 |
| 12 | 3,664 | 12,004 | 230 | 852 | 596 | 213 | 43 |

*Note: Total Free = cumulative signups minus those who converted or churned. Total Paid accounts for churn.*

### Monthly Recurring Revenue (MRR)

| Month | Plus MRR | Pro MRR | Team MRR | Total MRR | Costs | Net |
|-------|----------|---------|----------|-----------|-------|-----|
| 1 | $29 | $26 | $24 | $79 | $56 | $23 |
| 2 | $66 | $77 | $24 | $167 | $56 | $111 |
| 3 | $115 | $129 | $48 | $292 | $60 | $232 |
| 4 | $185 | $207 | $96 | $488 | $65 | $423 |
| 5 | $296 | $336 | $120 | $752 | $75 | $677 |
| 6 | $456 | $517 | $192 | $1,165 | $90 | $1,075 |
| 7 | $645 | $723 | $288 | $1,656 | $100 | $1,556 |
| 8 | $884 | $994 | $360 | $2,238 | $115 | $2,123 |
| 9 | $1,163 | $1,304 | $480 | $2,947 | $130 | $2,817 |
| 10 | $1,508 | $1,691 | $623 | $3,822 | $150 | $3,672 |
| 11 | $1,932 | $2,170 | $791 | $4,893 | $175 | $4,718 |
| 12 | $2,450 | $2,750 | $1,030 | $6,230 | $200 | $6,030 |

*MRR figures are blended net (after Stripe + app store fees). Costs increase as infra scales.*

---

## Break-Even Analysis

### Monthly Break-Even

With fixed costs of ~$56/month at launch:

- **At 100% Plus subscribers:** 14 paying users ($56 / $4.11 per user)
- **At current tier mix (70/25/5):** ~8 paying users
- **Expected break-even:** Month 1 (10 paid users covers $56 fixed costs)

### Real Break-Even (Including Founder Time)

If we value Shadow's operating time at $0 (AI operator), the break-even is purely infrastructure costs. Relix becomes cash-flow positive in month 1 at even modest conversion rates.

If we impute $5,000/month for human equivalent costs (support, marketing, development):

- **Break-even at imputed costs:** ~$5,056/month MRR → approximately month 10-11

---

## LTV and CAC Estimates

### Customer Lifetime Value (LTV)

| Tier | ARPU/month (net) | Monthly Churn | Avg Lifetime | LTV |
|------|-------------------|---------------|--------------|-----|
| Plus | $4.11 | 5% | 20 months | $82 |
| Pro | $12.91 | 3% | 33 months | $426 |
| Team | $23.97 | 2% | 50 months | $1,199 |

**Blended LTV (at tier mix 70/25/5):** ~$176

### Customer Acquisition Cost (CAC)

For the first 90 days (organic only):

- **CAC = $0** (no paid acquisition)
- True cost is Shadow's time, but as an AI operator this is ~$0 marginal cost

After organic phase, target CAC:

- **Plus:** < $20 (LTV:CAC ratio > 4:1)
- **Pro:** < $100 (LTV:CAC ratio > 4:1)
- **Team:** < $300 (LTV:CAC ratio > 4:1)

### LTV:CAC Target

Minimum 3:1 ratio before scaling paid acquisition. At organic-only CAC of ~$0, this ratio is effectively infinite — strong signal to delay paid acquisition until organic growth slows.

---

## Annual Revenue Targets

| Period | Target MRR | Target ARR | Paid Users | Key Milestone |
|--------|-----------|------------|------------|---------------|
| Month 3 | $292 | $3,504 | 40 | Product-market fit validated |
| Month 6 | $1,165 | $13,980 | 159 | Organic growth engine proven |
| Month 12 | $6,230 | $74,760 | 852 | Sustainable business |
| Month 18 | ~$15,000 | ~$180,000 | ~2,000 | Consider first hire (part-time support) |
| Month 24 | ~$35,000 | ~$420,000 | ~4,500 | Evaluate paid acquisition channels |

---

## Sensitivity Analysis

### Optimistic Scenario (2x growth, 15% conversion)

| Month | Total Free | Total Paid | MRR |
|-------|------------|------------|-----|
| 6 | 3,700 | 400 | $2,900 |
| 12 | 24,000 | 2,100 | $15,500 |

### Pessimistic Scenario (0.5x growth, 5% conversion)

| Month | Total Free | Total Paid | MRR |
|-------|------------|------------|-----|
| 6 | 450 | 40 | $290 |
| 12 | 1,500 | 105 | $780 |

Even in the pessimistic scenario, infrastructure costs are covered by month 2, and the business remains cash-flow positive (excluding imputed labor). The risk is not bankruptcy — it's slow growth requiring patience.

---

## Key Risks to Revenue

| Risk | Impact | Mitigation |
|------|--------|------------|
| Anthropic improves Remote Control to match Relix | Reduces differentiation for Claude Code-only users | Multi-tool and E2E encryption remain unique; accelerate Aider/Cline adapters |
| Low conversion from free to paid | Revenue grows slowly | Free tier is deliberately generous to drive word-of-mouth; focus on making the paid upgrade trigger (machine limit) feel natural |
| High churn | LTV drops, MRR plateaus | Focus on D7/D30 retention before scaling; push notifications are a sticky feature |
| App Store rejection or delays | Delays launch | Submit early, follow guidelines strictly, have web fallback |
| Apple/Google increase take rate | Reduces margins | Push users to web checkout where possible; most revenue is marginal on take rate |

---

## Summary

Relix has a clear path to profitability:

- **Month 1:** Cash-flow positive on infrastructure costs
- **Month 6:** ~$1,165 MRR, 159 paid users
- **Month 12:** ~$6,230 MRR, 852 paid users, ~$75K ARR
- **Blended LTV:** ~$176 per paid user
- **CAC:** $0 for first 90 days (organic)

The unit economics are strong because fixed costs are low (AI operator, commodity cloud), variable costs are minimal (Stripe + app store fees), and the product has natural retention drivers (push notifications create daily habits). The primary risk is growth speed, not profitability.
