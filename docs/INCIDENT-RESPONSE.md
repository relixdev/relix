# Relix — Incident Response Playbook

## Severity Levels

### P0 — Critical (Immediate Action)
- Complete service outage (API + Relay down)
- Data loss or corruption
- Security breach
- Payment processing failure

**Response time:** < 15 minutes  
**Communication:** Immediate status page update, email all users

### P1 — High (Within 1 Hour)
- Partial service degradation
- Mobile app crashes
- WebSocket connection failures
- Stripe webhook failures

**Response time:** < 1 hour  
**Communication:** Status page update

### P2 — Medium (Within 24 Hours)
- Non-critical bugs
- Feature regressions
- Performance issues

**Response time:** < 24 hours  
**Communication:** GitHub issue, Discord announcement

### P3 — Low (Next Sprint)
- Minor bugs
- Feature requests
- UX improvements

**Response time:** Next sprint  
**Communication:** GitHub issue

---

## Common Incidents

### API Server Down

**Symptoms:**
- `curl https://relix-api.shadowscale.dev/health` returns error
- Users can't login
- Mobile app shows "connection failed"

**Diagnosis:**
```bash
# Check container status
docker ps --filter "name=relix-cloud"

# Check logs
docker logs relix-cloud --tail 100

# Check PostgreSQL connectivity
docker exec relix-cloud pg_isready -h relix-postgres
```

**Resolution:**
```bash
# Restart container
docker restart relix-cloud

# If that fails, rebuild and restart
cd /home/shadow/services && docker compose up -d --force-recreate relix-cloud
```

**Prevention:**
- Health check monitoring (every 5 minutes)
- Auto-restart enabled in Docker Compose
- PostgreSQL connection pooling

---

### Relay WebSocket Failures

**Symptoms:**
- Mobile app can't connect to relay
- Sessions show "disconnected"
- Messages not delivered

**Diagnosis:**
```bash
# Check container status
docker ps --filter "name=relix-relay"

# Check logs
docker logs relix-relay --tail 100

# Check port
nc -zv relix-relay.shadowscale.dev 9081
```

**Resolution:**
```bash
# Restart container
docker restart relix-relay

# If that fails, rebuild and restart
cd /home/shadow/services && docker compose up -d --force-recreate relix-relay
```

**Prevention:**
- Connection monitoring
- Auto-reconnect in mobile app
- WebSocket heartbeat (every 30 seconds)

---

### PostgreSQL Issues

**Symptoms:**
- API returns 500 errors
- Users can't login
- Sessions not saved

**Diagnosis:**
```bash
# Check container status
docker ps --filter "name=relix-postgres"

# Check logs
docker logs relix-postgres --tail 100

# Check disk space
df -h

# Test connection
docker exec relix-postgres pg_isready
```

**Resolution:**
```bash
# Restart container
docker restart relix-postgres

# If disk full, clean up old logs
docker exec relix-postgres rm -rf /var/lib/postgresql/data/pg_log/*

# If corrupted, restore from backup
# (See backup/restore procedure below)
```

**Prevention:**
- Daily backups to S3
- Disk space monitoring (alert at 80%)
- Connection pooling

---

### Stripe Webhook Failures

**Symptoms:**
- Subscriptions not activated
- Billing errors
- Users not upgraded

**Diagnosis:**
```bash
# Check webhook logs
docker logs relix-cloud | grep -i stripe

# Test webhook endpoint
curl -X POST https://relix-api.shadowscale.dev/billing/webhook \
  -H "Stripe-Signature: test" \
  -d '{"type":"customer.subscription.updated"}'

# Check Stripe dashboard for failed webhooks
```

**Resolution:**
```bash
# Retry failed webhooks from Stripe dashboard
# Or manually update user subscription
psql -h relix-postgres -U relix -d relix -c "UPDATE users SET tier='pro' WHERE email='user@example.com';"
```

**Prevention:**
- Webhook signature verification
- Idempotency keys
- Failed webhook alerting

---

## Backup & Restore

### PostgreSQL Backup

```bash
# Daily backup (cron job)
docker exec relix-postgres pg_dump -U relix relix > /backups/relix-$(date +%Y%m%d).sql

# Upload to S3
aws s3 cp /backups/relix-$(date +%Y%m%d).sql s3://relix-backups/

# Retention: 30 days
find /backups -name "relix-*.sql" -mtime +30 -delete
```

### PostgreSQL Restore

```bash
# Download backup
aws s3 cp s3://relix-backups/relix-20260314.sql /tmp/restore.sql

# Restore
docker exec -i relix-postgres psql -U relix relix < /tmp/restore.sql
```

---

## Escalation

1. **On-call:** Shadow (automated)
2. **Escalate to:** Zach (if P0 not resolved in 30 minutes)
3. **Contact:** zach@shadowscale.dev

---

## Post-Incident Review

After any P0 or P1 incident:

1. Document timeline (what happened, when, who)
2. Root cause analysis (5 whys)
3. Action items (prevent recurrence)
4. Update this playbook if needed
