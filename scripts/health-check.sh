#!/bin/bash
# Relix health check — run periodically (cron or systemd timer)
# Reports: API status, relay status, PostgreSQL status, disk usage, memory usage
set -e

API_URL="http://relix-api.shadowscale.dev:9082"
RELAY_URL="ws://relix-relay.shadowscale.dev:9081"
POSTGRES_CONTAINER="relix-postgres"

echo "🐉 Relix Health Check — $(date)"
echo ""

# API health
echo "1. API Health..."
HEALTH=$(curl -sf "$API_URL/health" 2>&1)
if [ $? -eq 0 ]; then
  echo "   ✅ API healthy: $HEALTH"
else
  echo "   ❌ API unhealthy"
  exit 1
fi

# Relay connectivity
echo "2. Relay Connectivity..."
if nc -zv relix-relay.shadowscale.dev 9081 2>&1 | grep -q "succeeded"; then
  echo "   ✅ Relay reachable"
else
  echo "   ❌ Relay unreachable"
  exit 1
fi

# PostgreSQL status
echo "3. PostgreSQL Status..."
if docker exec "$POSTGRES_CONTAINER" pg_isready -q 2>&1 | grep -q "accepting connections"; then
  echo "   ✅ PostgreSQL accepting connections"
else
  echo "   ❌ PostgreSQL not accepting connections"
  exit 1
fi

# Disk usage
echo "4. Disk Usage..."
DISK=$(df -h / | tail -1 | awk '{print $5}')
if [ "${DISK%\%}" -lt 80 ]; then
  echo "   ✅ Disk usage: $DISK"
else
  echo "   ⚠️  Disk usage: $DISK (high)"
fi

# Memory usage
echo "5. Memory Usage..."
MEM=$(free -h | grep Mem | awk '{print $3 "/" $2}')
echo "   ℹ️  Memory: $MEM"

# Container status
echo "6. Container Status..."
docker ps --filter "name=relix" --format "table {{.Names}}\t{{.Status}}" 2>&1 | tail -n +2 | while read line; do
  if echo "$line" | grep -q "Up"; then
    echo "   ✅ $line"
  else
    echo "   ❌ $line"
  fi
done

echo ""
echo "✅ All health checks passed"
