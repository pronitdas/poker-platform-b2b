# Section 12: Operations and Disaster Recovery

## 12.1 Observability Strategy

The B2B poker platform requires comprehensive observability to ensure operational excellence, rapid incident response, and continuous improvement. This section covers the observability stack, service level objectives, incident management, backup/restore procedures, and disaster recovery planning.

### 12.1.1 Observability Stack

**Three Pillars of Observability:**

| Pillar | Tool | Purpose | Retention |
|--------|------|---------|-----------|
| **Logs** | Loki + Promtail | Application logs, structured events | 30 days hot, 1 year cold |
| **Metrics** | Prometheus + Grafana | Time-series metrics, dashboards, alerts | 90 days hot, 1 year downsampled |
| **Traces** | Jaeger / OpenTelemetry | Distributed tracing, request flows | 7 days hot, 30 days warm |

**Log Architecture:**

```
┌─────────────────────────────────────────────────────────┐
│                    Services (Go/Node.js)                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │Game Svc  │  │Real-Time │  │Auth Svc  │             │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘             │
└───────┼────────────┼────────────┼────────────────────┘
        │            │            │
        ▼            ▼            ▼
┌─────────────────────────────────────────────────────────┐
│                   Promtail Agents                        │
│  (Log aggregation, labeling, filtering)                  │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                      Loki Cluster                        │
│  (Log storage, indexing, query engine)                  │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                    Grafana Dashboard                     │
│  (Log visualization, search, alerts)                    │
└─────────────────────────────────────────────────────────┘
```

**Structured Logging Format:**

```go
type LogEntry struct {
    Timestamp   time.Time              `json:"timestamp"`
    Level       string                 `json:"level"` // INFO, WARN, ERROR
    Service     string                 `json:"service"`
    Instance    string                 `json:"instance"`
    TraceID     string                 `json:"trace_id"`
    SpanID      string                 `json:"span_id"`
    UserID      string                 `json:"user_id,omitempty"`
    HandID      string                 `json:"hand_id,omitempty"`
    TableID     string                 `json:"table_id,omitempty"`
    ClubID      string                 `json:"club_id,omitempty"`
    Message     string                 `json:"message"`
    Fields      map[string]interface{} `json:"fields,omitempty"`
    Error       string                 `json:"error,omitempty"`
    Duration    int64                  `json:"duration_ms,omitempty"`
}
```

**Log Retention Policy:**

| Log Level | Hot Storage (SSD) | Cold Storage (S3) | Archive (Glacier) |
|-----------|-------------------|-------------------|-------------------|
| **ERROR** | 30 days | 1 year | 7 years |
| **WARN** | 30 days | 1 year | 1 year |
| **INFO** | 7 days | 90 days | - |
| **DEBUG** | 1 day | - | - |

### 12.1.2 Metrics and Dashboards

**Key Metrics Categories:**

**Business Metrics:**
- `poker_active_players_total` - Current active players
- `poker_tables_active_total` - Active game tables
- `poker_hands_completed_total` - Total hands completed (counter)
- `poker_rake_collected_total` - Total rake collected (gauge)
- `poker_disputes_created_total` - Disputes created (counter)

**Performance Metrics:**
- `poker_hand_duration_seconds` - Hand duration histogram
- `poker_action_latency_seconds` - Player action to state update latency
- `poker_websocket_message_duration_seconds` - WebSocket message latency
- `poker_database_query_duration_seconds` - Database query latency
- `go_goroutines` - Active goroutines (Go-specific)
- `nodejs_eventloop_lag_seconds` - Node.js event loop lag

**Infrastructure Metrics:**
- `process_cpu_seconds_total` - CPU usage
- `process_resident_memory_bytes` - Memory usage
- `go_memstats_alloc_bytes` - Allocated memory (Go)
- `nodejs_heap_size_total_bytes` - Heap size (Node.js)
- `redis_commands_processed_total` - Redis operations
- `postgres_connections_active` - Active database connections

**Alerting Metrics:**

| Metric | Condition | Severity | Alert |
|--------|-----------|----------|-------|
| **Error Rate** | `poker_errors_total / poker_requests_total > 0.01` | Critical | PagerDuty |
| **P99 Latency** | `histogram_quantile(0.99, poker_action_latency_seconds) > 0.5` | High | Slack |
| **Memory Usage** | `process_resident_memory_bytes > 2GB` | High | Slack |
| **Database Connections** | `postgres_connections_active / postgres_connections_max > 0.8` | Medium | Email |
| **Redis Latency** | `redis_command_duration_seconds_p99 > 0.1` | Medium | Slack |

**Critical Dashboards:**

1. **Platform Overview Dashboard**
   - Active players, tables, hands per minute
   - Error rate, latency percentiles
   - Infrastructure health (CPU, memory, disk)
   - Real-time alerts

2. **Game Engine Dashboard**
   - Hand duration distribution
   - Action latency by type (bet, fold, raise)
   - WebSocket message throughput
   - Game state consistency checks

3. **Database Dashboard**
   - Query latency by table
   - Connection pool utilization
   - Replication lag
   - Index hit rates, cache hit rates

4. **Anti-Cheat Dashboard**
   - ML model scores distribution
   - Review queue backlog
   - False positive/negative rates
   - Model drift indicators

### 12.1.3 Distributed Tracing

**Trace Propagation:**

```go
// Go service with OpenTelemetry
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func HandlePlayerAction(ctx context.Context, req PlayerActionRequest) error {
    tracer := otel.Tracer("game-engine")
    ctx, span := tracer.Start(ctx, "HandlePlayerAction")
    defer span.End()

    // Add span attributes
    span.SetAttributes(
        attribute.String("player_id", req.PlayerID),
        attribute.String("hand_id", req.HandID),
        attribute.String("action_type", req.ActionType),
    )

    // Trace child operation
    err := ValidateAction(ctx, req)
    if err != nil {
        span.RecordError(err)
        return err
    }

    // Update game state
    err = UpdateGameState(ctx, req)
    if err != nil {
        span.RecordError(err)
        return err
    }

    return nil
}
```

**Critical Traces:**

| Trace | Spans | Sampling |
|-------|-------|----------|
| **Player Join Table** | Auth → Validate → Allocate Seat → Broadcast | 10% |
| **Game Hand** | Deal → Actions → Evaluate → Payout | 1% |
| **Payment Deposit** | Payment Gateway → Validate → Update Balance | 5% |
| **Anti-Cheat Scoring** | Event Ingest → ML Model → Alert Generation | 10% |
| **Error Traces** | All error paths | 100% |

---

## 12.2 Service Level Objectives (SLOs) and Indicators (SLIs)

### 12.2.1 SLI Definitions

**Availability SLI:**

```
SLI = (Total Time - Downtime) / Total Time

Where Downtime = Time when:
- Error rate > 1% for > 5 minutes
- P99 latency > 1s for > 5 minutes
- Service is unreachable
```

**Latency SLI:**

```
SLI = Percentage of requests with latency < threshold

Thresholds by operation:
- Player action: P99 < 200ms
- WebSocket message: P99 < 100ms
- API request: P99 < 500ms
- Database query: P99 < 50ms
```

**Correctness SLI:**

```
SLI = (Total Transactions - Incorrect Transactions) / Total Transactions

Incorrect Transactions:
- Incorrect pot calculations
- Wrong hand winners
- Incorrect rake deductions
- Payment processing errors
```

**Freshness SLI:**

```
SLI = 1 - (Data Staleness / Reporting Interval)

For real-time game state:
- State updates visible within 200ms
- Leaderboards refreshed within 5s
- Analytics data available within 10m
```

### 12.2.2 SLO Targets

**Note:** Per Section 7 assumptions, contractual baseline is 99.5% availability (~3.65 hours/month downtime allowed). Higher targets below represent stretch goals for operational excellence.

| Service | Metric | Target | Measurement Window |
|---------|--------|--------|---------------------|
| **Game Engine** | Availability | 99.5% (baseline), 99.9% (stretch goal) | 30 days rolling |
| **Game Engine** | Latency (P99) | 200ms | 24 hours |
| **Real-Time Service** | Availability | 99.5% (baseline), 99.95% (stretch goal) | 30 days rolling |
| **Real-Time Service** | Message Delivery | 99.9% (stretch goal) | 24 hours |
| **API Gateway** | Availability | 99.5% (baseline), 99.9% (stretch goal) | 30 days rolling |
| **API Gateway** | Latency (P99) | 500ms | 24 hours |
| **Database** | Availability | 99.5% (baseline), 99.9% (stretch goal) | 30 days rolling |
| **Database** | Query Latency (P99) | 50ms | 24 hours |
| **Payment Service** | Availability | 99.9% (stretch goal) | 30 days rolling |
| **Payment Service** | Correctness | 100% | Transaction |

### 12.2.3 Error Budget Management

**Error Budget Calculation:**

```
Error Budget = (100% - SLO) × Measurement Window

Example for 99.5% baseline SLO over 30 days:
Error Budget = 0.5% × 30 days = 216 minutes downtime allowed

Example for 99.9% stretch goal over 30 days:
Error Budget = 0.1% × 30 days = 43.2 minutes downtime allowed
```

**Error Budget Policy:**

| Budget Remaining | Action |
|------------------|--------|
| **> 50%** | Normal operations, all releases allowed |
| **25% - 50%** | Slow down releases, extra monitoring |
| **10% - 25%** | Emergency-only releases, pause non-critical features |
| **< 10%** | Stop all releases, focus on reliability |

**Error Budget Dashboard:**

```yaml
error_budgets:
  game_engine:
    slo: 99.5
    period: 30d
    budget: 216m
    consumed: 75m
    remaining: 141m
    policy: "normal"
  game_engine_stretch:
    slo: 99.9
    period: 30d
    budget: 43.2m
    consumed: 15.3m
    remaining: 27.9m
    policy: "normal"

  real_time_service:
    slo: 99.5
    period: 30d
    budget: 216m
    consumed: 195m
    remaining: 21m
    policy: "normal"
  real_time_service_stretch:
    slo: 99.95
    period: 30d
    budget: 21.6m
    consumed: 18.9m
    remaining: 2.7m
    policy: "slow_down"
```

---

## 12.3 Incident Management

### 12.3.1 Incident Severity Levels

**Severity Classification:**

| Severity | Name | Description | Response Time | SLA |
|----------|------|-------------|---------------|-----|
| **SEV-1** | Critical | Complete service outage, data loss, incorrect payouts affecting users | 15 minutes | 4-hour resolution |
| **SEV-2** | High | Major feature degradation, 50%+ users affected, significant financial impact | 30 minutes | 8-hour resolution |
| **SEV-3** | Medium | Minor feature issues, < 10% users affected, low financial impact | 1 hour | 24-hour resolution |
| **SEV-4** | Low | Cosmetic issues, documentation errors, single-user impact | 4 hours | 48-hour resolution |

**Incident Examples by Severity:**

| Severity | Example Incident |
|----------|------------------|
| **SEV-1** | Database corruption causing incorrect hand winners |
| **SEV-1** | Payment service down, deposits/withdrawals failing |
| **SEV-1** | Game engine crash, no players can play |
| **SEV-2** | WebSocket message delivery failing for 50%+ players |
| **SEV-2** | Anti-cheat service down, monitoring disabled |
| **SEV-2** | Admin panel inaccessible for agents |
| **SEV-3** | Leaderboard not updating |
| **SEV-3** | Push notifications delayed > 5 minutes |
| **SEV-4** | Minor UI glitch in player app |

### 12.3.2 Incident Response Process

**On-Call Structure:**

| Role | Primary Responsibility | Rotation |
|------|-----------------------|----------|
| **On-Call Engineer (L1)** | First responder, triage, initial investigation | 1 week rotation |
| **Service Owner (L2)** | Deep dive, resolution coordination, escalation | 1 week rotation |
| **Engineering Manager (L3)** | Executive communication, resource allocation | As needed |
| **PR/Messaging (L4)** | External communication, customer notification | As needed |

**Incident Lifecycle:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Detect    │───▶│    Triage   │───▶│  Mitigate   │───▶│   Resolve   │
│ (Automated  │    │   (15 min)  │    │  (Action)   │    │ (Fix Root)  │
│  Alerts)    │    └─────────────┘    └─────────────┘    └─────────────┘
└─────────────┘          │                                       │
                         ▼                                       ▼
                 ┌─────────────┐                        ┌─────────────┐
                 │  Communicate│                        │   Post-Mortem│
                 │ (Stakeholders│                        │ (Analysis)  │
                 │   + Users)  │                        │ + Learning  │
                 └─────────────┘                        └─────────────┘
```

**Incident Communication Flow:**

```
Detection (Auto)
    ↓
PagerDuty Alert → On-Call Engineer
    ↓
#incident-response Slack Channel
    ↓
[15 min] Initial Assessment → SEV Level
    ↓
If SEV-1/2: Escalate to Service Owner, Engineering Manager
    ↓
[30 min] Status Update → #incidents (public)
    ↓
[Every hour] Status Update → #incidents
    ↓
Resolution → Post-Incident Review
```

**Incident Timeline Template:**

```markdown
# Incident: [Title]

**Incident ID:** INC-2026-01-28-001
**Severity:** SEV-1
**Service:** Game Engine
**Start Time:** 2026-01-28T15:30:00Z
**End Time:** 2026-01-28T18:45:00Z
**Duration:** 3h 15m

## Timeline

| Time (UTC) | Event | Owner |
|------------|-------|-------|
| 15:30 | Alert: Error rate > 5% | PagerDuty |
| 15:35 | On-call engineer investigating | @jane |
| 15:45 | Root cause identified: Deadlock in game state mutex | @jane |
| 16:00 | Mitigation: Rolling restart of game engine instances | @jane |
| 16:30 | Error rate dropping, service stabilizing | @jane |
| 17:00 | Postmortem draft started | @jane |
| 18:45 | Incident resolved, postmortem approved | @john |

## Impact
- **Users Affected:** ~2,500 players
- **Duration:** 45 minutes of elevated errors
- **Tables Disrupted:** ~150 tables
- **Financial Impact:** Minimal (no incorrect payouts)

## Root Cause
Deadlock in game state mutex when multiple players acted simultaneously during high concurrency.

## Resolution
Rolling restart of game engine instances, plus code fix for mutex ordering.

## Follow-Up Actions
1. [ ] Fix mutex ordering bug (PR #1234) - @jane
2. [ ] Add integration test for concurrent actions - @jane
3. [ ] Update monitoring for mutex contention - @mike
```

### 12.3.3 Rollback and Mitigation

**Rollback Triggers:**

| Condition | Trigger |
|-----------|---------|
| **Error Rate** | > 1% for > 5 minutes |
| **Latency** | P99 > 1s for > 5 minutes |
| **Data Corruption** | Any detected |
| **Payment Failures** | > 0.5% error rate |
| **Game Logic Bug** | Incorrect payouts detected |

**Rollback Procedure:**

```bash
# Automated rollback script
#!/bin/bash
SERVICE="game-engine"
PREVIOUS_VERSION="v1.2.3"

# Check rollback triggers
ERROR_RATE=$(curl -s "http://prometheus:9090/api/v1/query?query=poker_error_rate" | jq '.data.result[0].value[1]')

if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
    echo "Error rate exceeded threshold, initiating rollback..."

    # Roll back to previous version
    kubectl rollout undo deployment/$SERVICE

    # Wait for rollback to complete
    kubectl rollout status deployment/$SERVICE --timeout=300s

    # Verify health
    HEALTH=$(kubectl get pods -l app=$SERVICE -o jsonpath='{.items[*].status.phase}')
    if [[ "$HEALTH" == "Running" ]]; then
        echo "Rollback successful"
    else
        echo "Rollback failed, manual intervention required"
        exit 1
    fi
fi
```

**Mitigation Playbooks:**

| Incident Type | Immediate Action | Long-Term Fix |
|---------------|-----------------|---------------|
| **Database Deadlock** | Kill long-running transactions, restart connection pool | Add connection pooling, optimize queries |
| **Redis Failure** | Failover to replica, cache miss handling | Implement Redis cluster, backup |
| **Game Engine Crash** | Restart instances, check core dumps | Add circuit breakers, improve error handling |
| **WebSocket Storm** | Rate limit connections, reject new connections | Implement connection pooling, backpressure |
| **Payment Service Down** | Queue transactions, manual review | Implement payment retries, fallback providers |

### 12.3.4 Post-Incident Review (PIR)

**PIR Template:**

```markdown
# Post-Incident Review: [Incident Title]

**Incident ID:** INC-2026-01-28-001
**Date:** 2026-01-28
**Participants:** @jane, @john, @mike
**Reviewer:** @sarah

## Executive Summary
[Brief 2-3 sentence summary]

## Impact Analysis
| Metric | Value |
|--------|-------|
| Users Affected | 2,500 |
| Downtime | 45 minutes |
| Financial Impact | $0 (no incorrect payouts) |
| Error Budget Consumed | 15.3 minutes |

## Timeline
[Same timeline format as incident report]

## Root Cause Analysis
[Five Whys analysis]

## What Went Well
- Automated alerting worked correctly
- On-call engineer responded within SLA
- Rollback restored service quickly

## What Could Be Improved
- Detection time could be reduced (15 min to 5 min)
- Manual rollback process (should be automated)
- Missing integration test for this scenario

## Action Items
| Priority | Action | Owner | Due Date |
|----------|--------|-------|----------|
| P0 | Fix root cause bug | @jane | 2026-01-29 |
| P1 | Add integration test | @jane | 2026-01-30 |
| P1 | Automate rollback | @mike | 2026-02-05 |
| P2 | Improve alert thresholds | @mike | 2026-02-10 |

## Attachments
- Log excerpts: `logs_incident_001.log`
- Core dumps: `core_dumps_incident_001.tar.gz`
- Metrics: `grafana_incident_001.json`
```

---

## 12.4 Backup and Recovery

### 12.4.1 Backup Strategy

**Backup Targets and Schedules:**

| Data Type | Frequency | Retention | Storage Location |
|-----------|-----------|-----------|------------------|
| **PostgreSQL** (Full) | Daily | 30 days | Cross-region S3 |
| **PostgreSQL** (WAL) | Continuous | 7 days | Local + Cross-region |
| **Redis** (RDB) | Hourly | 7 days | S3 |
| **Redis** (AOF) | Continuous | 1 day | Local disk |
| **Audit Logs** (Kafka) | Continuous | 7 years | S3 Glacier |
| **Application Logs** | Daily | 30 days | S3 |
| **Configuration** | On change | 90 days | Git + S3 |

**RPO/RTO Targets:**

| Service | RPO (Data Loss) | RTO (Recovery Time) | Method |
|---------|-----------------|---------------------|--------|
| **PostgreSQL** | 1 minute | 15 minutes | PITR (Point-In-Time Recovery) |
| **Redis** | 5 minutes | 10 minutes | RDB restore + AOF replay |
| **Audit Logs** | 0 seconds | Immediate | Replicated Kafka cluster |
| **Game State** | 30 seconds | 2 minutes | Redis backup + PostgreSQL audit |

**Backup Automation Script:**

```bash
#!/bin/bash
# PostgreSQL daily backup
BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DATABASE="poker_platform"

# Create backup
pg_dump -h postgres-primary -U backup_user -F c -b -v \
  -f "$BACKUP_DIR/poker_$TIMESTAMP.dump" $DATABASE

# Upload to S3 (cross-region)
aws s3 cp "$BACKUP_DIR/poker_$TIMESTAMP.dump" \
  s3://poker-backups-dr-region/postgres/poker_$TIMESTAMP.dump

# Verify backup
if pg_restore --list "$BACKUP_DIR/poker_$TIMESTAMP.dump" > /dev/null 2>&1; then
    echo "Backup verified successfully"
    # Cleanup old backups (keep 30 days)
    find $BACKUP_DIR -name "poker_*.dump" -mtime +30 -delete
else
    echo "Backup verification failed" >&2
    exit 1
fi
```

### 12.4.2 Restore Procedures

**PostgreSQL Restore (PITR):**

```bash
#!/bin/bash
# Restore to specific point-in-time
TIMESTAMP="2026-01-28 14:30:00"
RESTORE_DIR="/restore/postgres"

# Stop application
kubectl scale deployment game-engine --replicas=0

# Create empty instance
initdb -D $RESTORE_DIR

# Restore base backup
pg_restore -h localhost -U postgres -d poker_platform \
  -j 4 /backups/postgres/poker_20260128_010000.dump

# Replay WAL logs to target timestamp
# (configure recovery.conf or postgresql.conf recovery_target_time)
echo "recovery_target_time = '$TIMESTAMP'" >> $RESTORE_DIR/recovery.conf

# Start PostgreSQL in recovery mode
pg_ctl -D $RESTORE_DIR start

# Verify recovery
psql -h localhost -U postgres -d poker_platform -c "SELECT NOW();"

# Start application
kubectl scale deployment game-engine --replicas=3
```

**Redis Restore:**

```bash
#!/bin/bash
# Restore Redis from RDB snapshot
BACKUP="/backups/redis/dump_20260128_150000.rdb"

# Stop Redis
redis-cli shutdown

# Replace current dump file
cp $BACKUP /var/lib/redis/dump.rdb

# Start Redis
redis-server /etc/redis/redis.conf

# Verify restore
redis-cli info replication
redis-cli DBSIZE
```

### 12.4.3 Backup Verification

**Weekly Backup Validation:**

| Task | Frequency | Owner |
|------|-----------|-------|
| **PostgreSQL restore test** | Weekly | @dba_team |
| **Redis restore test** | Weekly | @dba_team |
| **Backup integrity check** | Daily | @monitoring |
| **Cross-region replication check** | Daily | @ops_team |

**Restore Test Checklist:**

- [ ] Verify backup exists in all locations (local + DR region)
- [ ] Verify backup file integrity (checksum)
- [ ] Test restore to staging environment
- [ ] Verify data consistency after restore
- [ ] Document any issues encountered
- [ ] Update restore procedures if needed

---

## 12.5 Disaster Recovery

### 12.5.1 DR Architecture

**Multi-Region Design:**

```
┌─────────────────────────────────────────────────────────┐
│                    Primary Region (US-East)              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │Game Engine  │  │ PostgreSQL  │  │   Redis     │      │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘      │
└─────────┼─────────────────┼─────────────────┼────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────┐
│                  Cross-Region Replication                 │
│         (Async, eventual consistency)                   │
└─────────────────────────────────────────────────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────┐
│                   DR Region (US-West)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │Game Engine  │  │ PostgreSQL  │  │   Redis     │      │
│  │  (Standby)  │  │ (Replica)   │  │ (Replica)   │      │
│  └─────────────┘  └─────────────┘  └─────────────┘      │
└─────────────────────────────────────────────────────────┘
```

**Failover Scenarios:**

| Scenario | Trigger | Failover Time | Data Loss |
|----------|---------|----------------|------------|
| **Region outage** | Complete primary region unavailable | 15-30 minutes | < 1 minute |
| **Database corruption** | PostgreSQL corruption detected | 30 minutes | < 5 minutes |
| **Major attack** | DDoS, security breach | Immediate switch | < 1 minute |
| **Planned maintenance** | Scheduled maintenance window | 0 minutes (graceful) | None |

### 12.5.2 Game State Recovery

**State Recovery Strategy:**

When a table fails mid-hand, the system must reconstruct the game state from the audit log to resume play or declare a winner.

**Recovery Process:**

```go
func RecoverGameState(tableID string, handID string) (*GameState, error) {
    // 1. Check if state exists in hot cache (Redis)
    cachedState, err := redis.Get("game_state:" + tableID)
    if err == nil && cachedState != nil {
        return cachedState, nil
    }

    // 2. Check if state exists in warm cache (PostgreSQL)
    warmState, err := pg.QueryRow(
        "SELECT game_state FROM game_states WHERE table_id = $1 AND hand_id = $2",
        tableID, handID,
    )
    if err == nil {
        return warmState, nil
    }

    // 3. Reconstruct from audit log (replay)
    events, err := LoadAuditEvents(handID)
    if err != nil {
        return nil, fmt.Errorf("cannot recover: audit log not found")
    }

    state := NewGameState(tableID, handID)
    for _, event := range events {
        err := ApplyAction(state, event)
        if err != nil {
            return nil, fmt.Errorf("replay failed: %w", err)
        }
    }

    // 4. Cache recovered state
    redis.Set("game_state:"+tableID, state, 5*time.Minute)

    return state, nil
}
```

**Hand Completion on Failure:**

When a hand cannot be recovered or resumed:

1. **Analyze progress**: Determine which streets were completed
2. **Calculate equity**: Use Monte Carlo simulation to estimate each player's equity
3. **Award chips**: Distribute pot proportionally based on equity
4. **Log decision**: Record the award decision in audit log
5. **Notify players**: Send notification explaining the resolution

### 12.5.3 DR Failover Procedure

**Failover Checklist:**

| Step | Action | Owner | ETA |
|------|--------|-------|-----|
| 1 | Confirm primary region outage | On-Call | 5 min |
| 2 | Verify DR region health | On-Call | 5 min |
| 3 | Promote PostgreSQL replica to primary | DBA | 5 min |
| 4 | Promote Redis replica to primary | DBA | 2 min |
| 5 | Scale game engine instances | Ops | 5 min |
| 6 | Update DNS to point to DR region | Ops | 5 min |
| 7 | Verify player connections | Ops | 5 min |
| 8 | Notify stakeholders | PR | 10 min |
| **Total** | | | **~37 min** |

**Failover Script:**

```bash
#!/bin/bash
REGION="us-west-2"
PRIMARY_REGION="us-east-1"

# Step 1: Verify DR region health
echo "Verifying DR region health..."
aws elasticache describe-replication-groups --region $REGION \
  --replication-group-id poker-redis-primary

aws rds describe-db-instances --region $REGION \
  --db-instance-identifier poker-postgres-primary

# Step 2: Promote PostgreSQL replica
echo "Promoting PostgreSQL replica..."
aws rds promote-read-replica \
  --db-instance-identifier poker-postgres-replica \
  --region $REGION

# Step 3: Promote Redis replica
echo "Promoting Redis replica..."
aws elasticache increase-replica-count \
  --replication-group-id poker-redis-primary \
  --apply-immediately \
  --region $REGION

# Step 4: Scale game engine
echo "Scaling game engine instances..."
kubectl scale deployment game-engine --replicas=10 --context dr-cluster

# Step 5: Update DNS (Route53)
echo "Updating DNS to point to DR region..."
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --change-batch file://route53-failover.json

# Step 6: Verify failover
echo "Verifying failover..."
curl -I https://<your-domain>/health

echo "Failover complete"
```

### 12.5.4 Failback Procedure

**Failback Checklist:**

| Step | Action | Owner | ETA |
|------|--------|-------|-----|
| 1 | Confirm primary region recovery | Ops | 10 min |
| 2 | Verify data synchronization | DBA | 15 min |
| 3 | Set up primary as replica | DBA | 10 min |
| 4 | Wait for replication sync | DBA | 30-60 min |
| 5 | Promote primary back to primary | DBA | 5 min |
| 6 | Scale down DR instances | Ops | 5 min |
| 7 | Update DNS back to primary | Ops | 5 min |
| 8 | Verify service health | Ops | 10 min |
| **Total** | | | **~90-120 min** |

**Data Synchronization:**

```bash
#!/bin/bash
# Set up primary as replica after recovery
PRIMARY_HOST="poker-postgres-primary.us-east-1.rds.amazonaws.com"
DR_HOST="poker-postgres-primary.us-west-2.rds.amazonaws.com"

# Create replication slot on DR
psql -h $DR_HOST -U postgres -c "
  SELECT * FROM pg_create_physical_replication_slot('failback_slot');
"

# Stop primary, enable replication
aws rds stop-db-instance --db-instance-identifier poker-postgres-primary \
  --region us-east-1 --skip-final-snapshot

# Configure primary as replica (via RDS parameter group)
aws rds modify-db-instance \
  --db-instance-identifier poker-postgres-primary \
  --region us-east-1 \
  --apply-immediately \
  --db-parameter-group-name postgres-replica

# Wait for replication to sync
echo "Waiting for replication to sync..."
# Monitor replication lag
while true; do
  LAG=$(psql -h $DR_HOST -U postgres -t -c "SELECT pg_lsn_diff(pg_current_wal_lsn(), replay_lsn) FROM pg_stat_replication;")
  if [[ "$LAG" == "0" ]]; then
    break
  fi
  sleep 10
done

# Promote back to primary
aws rds promote-read-replica \
  --db-instance-identifier poker-postgres-primary \
  --region us-east-1
```

---

## 12.6 Chaos Engineering

### 12.6.1 Chaos Testing Strategy

**Chaos Experiments:**

| Experiment | Frequency | Success Criteria |
|-----------|-----------|-------------------|
| **Pod Kill** | Weekly | Auto-restart < 1 min, no data loss |
| **Network Partition** | Monthly | Degradation handled gracefully |
| **Database Failover** | Monthly | Failover < 5 min, RPO < 1 min |
| **Redis Failover** | Weekly | Failover < 2 min, RPO < 5 min |
| **CPU Starvation** | Monthly | Service degrades, doesn't crash |
| **Memory Pressure** | Monthly | OOM killer triggers, restart < 2 min |

**Chaos Tooling:**

| Tool | Use Case |
|------|----------|
| **Chaos Mesh** | Kubernetes pod/network chaos |
| **Litmus Chaos** | Pod kill, disk fill, latency injection |
| **Gremlin** | Network partition, resource exhaustion |

**Chaos Experiment Example (Pod Kill):**

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: game-engine-pod-kill
spec:
  action: pod-kill
  mode: random-max-percent
  value: "20"
  selector:
    namespaces:
      - production
    labelSelectors:
      app: game-engine
  scheduler:
    cron: "@weekly"
  duration: "5m"
```

### 12.6.2 Game Over Testing

**Purpose:** Verify that the platform fails gracefully under extreme load.

**Game Over Scenarios:**

| Scenario | Description | Expected Behavior |
|----------|-------------|-------------------|
| **Connection Storm** | 50K connections in 1 minute | Rate limiting, queue, graceful degradation |
| **Spike in Actions** | 100K actions/second | Queue, throttle, no crashes |
| **Database Lock Contention** | 1K concurrent writes | Connection pooling, retries, exponential backoff |
| **Redis Exhaustion** | Memory > 90% | Eviction policy, alerts, failover |
| **Kafka Backlog** | 1M messages in queue | Consumer scaling, no data loss |

**Game Over Test Script:**

```bash
#!/bin/bash
# Simulate connection storm
CONNECTIONS=50000
RATE_LIMIT=1000  # connections per second

echo "Starting connection storm with $CONNECTIONS connections..."

for ((i=1; i<=$CONNECTIONS; i++)); do
    (
        # Connect to WebSocket
        wscat -c wss://<your-domain>/game &
    ) &

    # Rate limit
    if (( $i % $RATE_LIMIT == 0 )); then
        sleep 1
    fi
done

echo "Connection storm complete, monitoring..."
# Monitor metrics
curl -s "http://prometheus:9090/api/v1/query?query=poker_active_connections"
```

---

## 12.7 Runbooks

### 12.7.1 Common Operational Procedures

**Deploy New Version:**

```bash
#!/bin/bash
SERVICE="game-engine"
NEW_VERSION="v1.3.0"
CANARY_PERCENT=10

# Step 1: Update deployment
kubectl set image deployment/$SERVICE \
  game-engine=poker/$SERVICE:$NEW_VERSION

# Step 2: Wait for rollout
kubectl rollout status deployment/$SERVICE --timeout=300s

# Step 3: Monitor for 10 minutes
echo "Monitoring for 10 minutes..."
sleep 600

# Step 4: Check error rate
ERROR_RATE=$(curl -s "http://prometheus:9090/api/v1/query?query=poker_error_rate" | jq '.data.result[0].value[1]')

if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
    echo "Error rate too high, rolling back..."
    kubectl rollout undo deployment/$SERVICE
    exit 1
fi

echo "Deployment successful"
```

**Clear Cache (Redis):**

```bash
#!/bin/bash
# Clear specific cache patterns safely
REDIS_HOST="redis-primary.poker.internal"

# Clear game state cache (safe, reconstructed from DB)
redis-cli -h $REDIS_HOST --scan --pattern "game_state:*" | xargs redis-cli -h $REDIS_HOST DEL

# Clear leaderboard cache (safe, recalculated)
redis-cli -h $REDIS_HOST --scan --pattern "leaderboard:*" | xargs redis-cli -h $REDIS_HOST DEL

echo "Cache cleared"
```

**Restart Service:**

```bash
#!/bin/bash
SERVICE="game-engine"

# Rolling restart (zero downtime)
kubectl rollout restart deployment/$SERVICE

# Wait for restart
kubectl rollout status deployment/$SERVICE --timeout=300s

# Verify health
kubectl get pods -l app=$SERVICE -o jsonpath='{.items[*].status.phase}'
```

### 12.7.2 Runbook Index

| Runbook | Scenario | Owner |
|---------|----------|-------|
| [Database Failover](#database-failover) | PostgreSQL primary fails | DBA |
| [Redis Failover](#redis-failover) | Redis primary fails | Ops |
| [Region Failover](#region-failover) | Complete region outage | Ops |
| [Game State Recovery](#game-state-recovery) | Table state lost | Game Team |
| [Rollback Deployment](#rollback-deployment) | Deployment causing errors | DevOps |
| [Clear Cache](#clear-cache) | Cache corruption / stale data | Ops |
| [Restart Service](#restart-service) | Service unresponsive | Ops |
| [Handle Connection Storm](#handle-connection-storm) | Excessive connections | Ops |
| [Payment Failure Investigation](#payment-failure) | Payment processing errors | Payments |
| [Anti-Cheat Escalation](#anti-cheat-escalation) | High-risk bot detected | Security |

---

## 12.8 Operational Excellence

### 12.8.1 On-Call Rotation

**On-Call Schedule:**

- Primary on-call: 1 week rotation, 24/7 coverage
- Secondary on-call: Backup, escalation path
- Manager on-call: For SEV-1 incidents, executive communication

**On-Call Responsibilities:**

| Time | Responsibility |
|------|----------------|
| **0-15 min** | Acknowledge PagerDuty alert, triage severity |
| **15-60 min** | Initial investigation, mitigation |
| **1-4 hours** | Deep dive, resolution (SEV-1/2) |
| **Next business day** | Post-incident review draft |

**On-Call Handoff:**

```markdown
# On-Call Handoff Template

**Date:** 2026-01-28
**On-Call Engineer:** @jane
**Next On-Call:** @john

## Outstanding Issues
- Investigating sporadic latency spikes (ticket #123)
- Monitoring Redis memory usage at 85%

## Active Incidents
- None

## Scheduled Maintenance
- Database index rebuild: 2026-01-29 02:00 UTC (4-hour window)

## Recent Changes
- Deployed v1.3.0 (game engine) - monitoring
- Updated Redis memory limit to 8GB

## Notes
- Payment service upgrade scheduled for next week
- New on-call engineer needs training on rollback procedure
```

### 12.8.2 Capacity Planning

**Capacity Model:**

| Resource | Metric | Scale Factor | Current | Forecast (6 months) |
|----------|--------|-------------|---------|---------------------|
| **Game Engine** | Concurrent players | 1K players = 2 pods | 5K players (10 pods) | 10K players (20 pods) |
| **PostgreSQL** | Active tables | 1K tables = 100 IOPS | 1.5K tables (150 IOPS) | 3K tables (300 IOPS) |
| **Redis** | Game state cache | 1K tables = 2GB | 1.5K tables (3GB) | 3K tables (6GB) |
| **Kafka** | Audit events | 100 events/sec = 1 partition | 500 events/sec (5 partitions) | 1K events/sec (10 partitions) |

**Scaling Triggers:**

| Metric | Threshold | Action |
|--------|-----------|--------|
| **CPU Usage** | > 70% for 5 min | Scale up (+2 pods) |
| **Memory Usage** | > 80% | Scale up (+2 pods) |
| **Queue Depth** | > 1000 | Scale up consumers |
| **Connection Count** | > 80% of max | Add load balancer, scale services |
| **Disk Usage** | > 70% | Expand storage |

**Autoscaling Configuration:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: game-engine-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: game-engine
  minReplicas: 3
  maxReplicas: 30
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
      - type: Pods
        value: 4
        periodSeconds: 60
      selectPolicy: Max
```

### 12.8.3 Security Operations

**Security Monitoring:**

| Event | Detection Method | Response |
|-------|------------------|----------|
| **Failed login attempts** | > 10 failures / IP / minute | IP ban rate limiting |
| **API abuse** | > 1000 requests / minute / token | Token revoke, alert |
| **SQL injection attempts** | WAF detection | Block IP, alert security team |
| **Payment fraud** | ML model alert | Freeze transaction, manual review |
| **Bot detection** | Anti-cheat model | Flag account, review queue |

**Security Incident Response:**

1. **Detection**: Automated alert from security tools
2. **Triage**: Assess severity, determine scope
3. **Containment**: Block malicious traffic, freeze accounts
4. **Eradication**: Patch vulnerability, remove malware
5. **Recovery**: Restore from backups if needed
6. **Lessons Learned**: Post-incident review, update procedures

---

## 12.9 Documentation and Knowledge Management

### 12.9.1 Documentation Requirements

**Required Documentation:**

| Document | Owner | Review Frequency |
|----------|-------|------------------|
| **Architecture Diagrams** | Architecture Team | Quarterly |
| **API Documentation** | Tech Writers | Every release |
| **Runbooks** | Ops Team | Monthly |
| **On-Call Procedures** | Engineering Manager | Monthly |
| **Backup/Restore Procedures** | DBA | Quarterly |
| **DR Playbook** | Ops Team | Quarterly |
| **Post-Incident Reviews** | All | Every incident |
| **Capacity Planning Report** | Ops Team | Quarterly |

### 12.9.2 Knowledge Sharing

**Weekly Engineering Sync:**

- Incident retrospective (15 min)
- Upcoming changes (15 min)
- Knowledge sharing: One engineer deep-dive on a topic (30 min)

**Monthly Tech Talk:**

- Deep dive into architecture decisions
- Lessons learned from incidents
- New technology evaluations

**Quarterly All-Hands:**

- Product roadmap
- Technical achievements
- Customer feedback
- Q&A

---

## 12.10 Compliance and Auditing

### 12.10.1 Regulatory Compliance

**Compliance Requirements:**

| Regulation | Requirement | Evidence |
|------------|-------------|----------|
| **PCI DSS** | Payment card security (Phase 3+ real-money only) | Quarterly SAQ, annual ROC |
| **GLI** | RNG certification, game fairness | Certification reports |
| **eCOGRA** | Payout verification, responsible gaming | Monthly audits |
| **GDPR** | Data protection, user rights | DPIA, data processing agreements |
| **SOC 2** | Security, availability, processing integrity | Annual SOC 2 Type II report |

### 12.10.2 Audit Readiness

**Audit Checklist:**

- [ ] All logs retained for 7 years
- [ ] Audit trail integrity verified (hash chain)
- [ ] Access control review completed
- [ ] Penetration testing performed (quarterly)
- [ ] Vulnerability scan results documented
- [ ] Backup/restore test results available
- [ ] Incident response procedures documented
- [ ] Employee training records current
- [ ] Third-party assessments completed
- [ ] Regulatory filings up to date

---

*Section 12 provides a comprehensive operations and disaster recovery framework ensuring high availability, rapid incident response, and regulatory compliance for the B2B poker platform.*
