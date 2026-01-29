# Section 11: Testing and Validation

## 11.1 Testing Strategy Overview

The B2B poker platform requires a comprehensive testing framework covering functional correctness, game integrity, performance at scale, and security validation. This section outlines the testing methodology, tools, quality gates, and validation processes.

### Testing Philosophy

| Principle | Implementation | Rationale |
|-----------|----------------|-----------|
| **Test Pyramid** | 70% unit tests, 20% integration tests, 10% E2E tests | Fast feedback loops, maintainable test suite |
| **Shift Left** | Unit tests written during implementation, TDD for critical paths | Early defect detection, lower remediation cost |
| **State Determinism** | All game logic tests use seeded RNG, replayable state | Reproducible bugs, verifiable behavior |
| **Contract Testing** | API contracts validated between services | Prevent integration breaks during deployment |
| **Game Integrity First** | Dedicated test suites for fairness, RNG, anti-cheat | Regulatory compliance, player trust |

---

## 11.2 Test Pyramid and Coverage

### 11.2.1 Unit Testing (70% - Target: 85%+ Code Coverage)

**Scope:**
- Game logic: hand evaluation, bet validation, pot calculation
- Business rules: club management, rake calculation, player limits
- Utility functions: RNG algorithms, cryptographic primitives
- Data models: validation, serialization, transformation
- Core algorithms: anti-cheat scoring, rate limiting logic

**Tools:**
- Go: `testing` package, `testify/assert`, `testify/mock`
- TypeScript: `Jest`, `ts-mockito`
- Coverage reporting: `go test -cover`, `jest --coverage`

**Unit Test Checklist:**

| Category | Test Cases Required | Acceptance Criteria |
|----------|-------------------|---------------------|
| **Hand Evaluation** | Comprehensive hand evaluation test suite (edge cases, boundary values, large randomized corpus) | 100% correct ranking against reference implementation |
| **Bet Validation** | Valid/invalid bets across all game states | Zero false positives/negatives |
| **Pot Calculation** | Side pots, all-in scenarios, rake deduction | Exact match with reference implementation |
| **RNG Output** | Statistical tests (NIST SP 800-22) on large sample corpus (e.g., 10M samples) | Pass all statistical randomness tests |
| **Anti-Cheat Scoring** | Known bot behavior vs. legitimate play patterns | ROC AUC > 0.95 on labeled dataset (example target; define based on stakeholder requirements) |
| **Rate Limiting** | Edge cases: burst, steady, threshold crossing | Exact limit enforcement, no overage |

**Example Unit Test (Go - Hand Evaluation):**

```go
// Pseudocode illustrating test structure
func TestHandEvaluation_RoyalFlush(t *testing.T) {
    // Seeded test for reproducibility
    deck := NewDeck(WithSeed(12345))  // Adjusted for valid Go syntax
    hand := []Card{
        {Rank: Ace, Suit: Hearts},
        {Rank: King, Suit: Hearts},
        {Rank: Queen, Suit: Hearts},
        {Rank: Jack, Suit: Hearts},
        {Rank: Ten, Suit: Hearts},
    }
    board := []Card{
        {Rank: Two, Suit: Spades},
        {Rank: Three, Suit: Clubs},
        {Rank: Four, Suit: Diamonds},
    }

    result := EvaluateHand(hand, board)
    assert.Equal(t, RoyalFlush, result.Rank)
    assert.Equal(t, []Card{Ace, King, Queen, Jack, Ten}, result.Kickers)
}
```

### 11.2.2 Integration Testing (20% - Target: 80%+ Path Coverage)

**Scope:**
- Service-to-service communication: Game Engine ↔ Real-Time ↔ Auth
- Database integration: PostgreSQL queries, Redis caching
- External API mocks: Payment providers, SMS providers
- WebSocket message flow: Connect → Join → Action → Disconnect
- Event streaming: Kafka producer/consumer pipelines

**Tools:**
- `testcontainers` for database/service fixtures
- `WireMock` for HTTP API mocking
- Custom WebSocket test harness for socket.IO
- Local Kafka instance (`testcontainers`)

**Integration Test Checklist:**

| Flow | Test Scenarios | Acceptance Criteria |
|------|----------------|---------------------|
| **Player Join Table** | Valid auth, invalid auth, table full, club mismatch | Correct error codes, state transitions |
| **Game Flow** | Full hand from deal to showdown across 9 players | State consistency across all services |
| **Payment Deposit** | Success, failure, retry, webhook timeout | Idempotency, correct ledger updates |
| **Anti-Cheat Pipeline** | Event ingestion → ML scoring → Alert generation | End-to-end latency < 5s for alert |
| **State Recovery** | Redis crash/reconnect, PostgreSQL failover | Zero data loss, automatic recovery |

### 11.2.3 End-to-End Testing (10% - Critical User Journeys)

**Scope:**
- Critical path: Registration → Deposit → Play → Cashout
- Multi-device: Desktop + Mobile same account
- Cross-platform: iOS ↔ Android gameplay consistency
- Real-time: 6-player table with concurrent actions

**Tools:**
- `Playwright` for web admin panel E2E
- `Cocos Creator` test framework for mobile game E2E
- Custom test orchestrator for multi-user scenarios

**E2E Test Scenarios:**

| Scenario | Precondition | Steps | Validation |
|----------|--------------|-------|------------|
| **New Player Onboarding** | Clean environment | Register → Verify email → Add funds → Join low-stakes table | Account active, balance correct, table joined |
| **Tournament Flow** | 50 registered players | Tournament starts → Blinds increase → Player eliminated → Winner declared | Chip counts correct, payouts calculated, leaderboard accurate |
| **Dispute Replay** | Hand ID from production | Replay from audit log → Verify all actions → Cross-check hand result | 100% match with original outcome |
| **Cross-Device Handoff** | Player on mobile | Login on desktop → Continue same hand → Complete | Session synced, state consistent |

---

## 11.3 Game Integrity Testing

### 11.3.1 RNG Verification and Certification

**RNG Architecture:**
1. Hardware RNG: Intel DRNG / TPM entropy source
2. PRNG: AES-CTR DRBG (NIST SP 800-90A)
3. Per-table seeding: Unique seeds per hand, logged for audits

**Testing Requirements:**

| Test Type | Tool/Method | Sample Size (Example) | Pass Criteria (Example) |
|-----------|-------------|----------------------|-----------------------|
| **Frequency (Monobit)** | NIST SP 800-22 | 10M bits | p-value > 0.01 |
| **Runs Test** | NIST SP 800-22 | 10M bits | p-value > 0.01 |
| **Spectral Test** | Dieharder | 10M bits | Pass all tests |
| **Entropy** | ENT | 10M bits | Entropy > 7.999 bits/byte |
| **Chi-Square** | Custom | 100K decks | χ² < critical value |
| **Card Distribution** | Custom | 10M hands | Uniform distribution, p > 0.01 |

**Audit Data Format:**

```json
{
  "hand_id": "hnd_1234567890",
  "table_id": "tbl_abc123",
  "rng_seed": "0x8f3d2e1a4b5c6d7e",
  "prng_algorithm": "AES-CTR-DRBG",
  "shuffle_sequence": ["7h", "2d", "As", "Kc", ...],
  "timestamp": "2026-01-28T15:30:45Z",
  "audit_hash": "sha256(hexencode(seed + sequence))",
  "regulatory_export_format": "GLI-STD-001"
}
```

**Certification Preparation Checklist:**

- [ ] All NIST SP 800-22 tests passing with documented p-values (example target)
- [ ] Dieharder test suite results report
- [ ] Large hand sample dataset (e.g., 10M) with full audit trail
- [ ] Source code review checklist completed
- [ ] Seed management documentation
- [ ] Hardware entropy source verification report
- [ ] GLI/eCOGRA/iTech Labs audit format exports

### 11.3.2 Deterministic Replay System

**Purpose:** Enable game state replay for dispute resolution, debugging, and regulatory audits.

**Replay Architecture:**
- Every game event logged to immutable audit log (Kafka → PostgreSQL)
- Event includes: timestamp, player_id, action_type, action_data, state_hash
- Replayer reconstructs state by replaying events in order
- Final state hash compared to production for verification

**Replay Implementation:**

```go
type GameEvent struct {
    HandID      string    `json:"hand_id"`
    Sequence    int       `json:"sequence"`
    Timestamp   time.Time `json:"timestamp"`
    PlayerID    string    `json:"player_id"`
    ActionType  string    `json:"action_type"` // "deal", "bet", "fold", etc.
    ActionData  json.RawMessage `json:"action_data"`
    StateHash   string    `json:"state_hash"` // SHA-256 of state after event
}

func ReplayHand(handID string) (*GameState, error) {
    events := LoadAuditEvents(handID)
    state := NewGameState()

    for _, event := range events {
        prevStateHash := HashState(state)

        err := ApplyAction(state, event)
        if err != nil {
            return nil, fmt.Errorf("replay failed at seq %d: %w", event.Sequence, err)
        }

        if event.StateHash != HashState(state) {
            return nil, fmt.Errorf("state mismatch at seq %d: expected %s, got %s",
                event.Sequence, event.StateHash, HashState(state))
        }
    }

    return state, nil
}
```

**Dispute Resolution Workflow:**

1. Agent submits dispute with Hand ID
2. Support team runs `ReplayHand(handID)` automatically
3. System generates full replay report with timeline
4. Replayer flags any state inconsistencies
5. Reviewer validates outcome against audit log
6. Report exported in regulator-compatible format

### 11.3.3 Anti-Cheat System Validation

**Testing Anti-Cheat Models:**

| Test Type | Dataset | Metrics | Threshold (Example) |
|-----------|---------|---------|---------------------|
| **Known Bot Detection** | 10K labeled hands (bot vs. human) | Precision, Recall, F1, ROC AUC | F1 > 0.90, AUC > 0.95 (define based on stakeholder requirements) |
| **Collusion Detection** | 1K hands with labeled collusion | True Positive Rate, False Positive Rate | TPR > 0.85, FPR < 0.01 |
| **Cross-Device Tracking** | 1K accounts with device history | Linkage accuracy | > 95% correct linkages |
| **Model Drift** | Weekly production samples | Score distribution stability | KS statistic < 0.1 |

**False Positive Mitigation:**

- **Review Queue:** Scores above threshold but below auto-ban flag for manual review
- **Contextual Filtering:** Adjust thresholds for new accounts, VIP players, unusual stakes
- **Appeals Process**: 24-hour review window, escalation matrix

**Anti-Cheat Test Scenarios:**

| Scenario | Description | Expected Alert |
|----------|-------------|----------------|
| **Bot with perfect play** | NLHE bot with optimal pre-flop strategy | ML score > 0.8, manual review |
| **Two players colluding** | Always raising together, folding to each other | Collusion score > 0.7, flag |
| **Chip dumping** | Intentional loses to specific player | Chip flow anomaly > 3σ, flag |
| **Legitimate aggressive play** | Human player with VPIP 80%, AF 4 | No alert (human-like variance) |

**Model Monitoring:**

```go
// Monitor model drift weekly
func MonitorModelDrift(model *AntiCheatModel) {
    recentSamples := GetLastWeekSamples()
    scores := model.ScoreBatch(recentSamples)

    baseline := model.LoadBaselineDistribution()
    ks := KolmogorovSmirnov(scores, baseline)

    if ks > 0.1 {
        Alert(fmt.Sprintf("Model drift detected (KS=%.3f), retraining required", ks))
        ScheduleRetraining()
    }
}
```

---

## 11.4 Performance and Load Testing

### 11.4.1 Load Testing Strategy

**Load Testing Tools:**

| Tool | Use Case | Target Metric |
|------|----------|---------------|
| **k6** | API load testing (HTTP/WebSocket) | Throughput, latency, error rates |
| **Artillery** | WebSocket connection storm simulation | Connection establishment rate |
| **Locust** | Custom game flow simulation (Python) | End-to-end game latency |
| **JMeter** | Admin panel load testing | Page response times, concurrent users |
| **go-wrk** | Go service HTTP benchmarking | Requests/sec under load |

**Load Test Scenarios:**

| Scenario | Target Load | Duration | Success Criteria |
|----------|-------------|----------|------------------|
| **Baseline** | 1K concurrent players | 10 min | P99 latency < 100ms, 0% errors |
| **Stress Test** | 10K concurrent players | 30 min | P99 latency < 200ms, < 0.1% errors |
| **Connection Storm** | 5K connections in 10s | 5 min | All connections accepted, < 2s latency |
| **Spike Test** | 5K → 15K → 5K concurrent players | 20 min | Graceful degradation, no crashes |
| **Endurance Test** | 5K concurrent players, 24h | 24h | No memory leaks, stable latency |

**k6 WebSocket Load Test Example:**

```javascript
import websocket from 'k6/x/websocket';
import { check } from 'k6';

export let options = {
    stages: [
        { duration: '2m', target: 1000 },  // Ramp up
        { duration: '5m', target: 1000 },  // Sustained
        { duration: '2m', target: 5000 },  // Ramp up to peak
        { duration: '10m', target: 5000 }, // Peak load
        { duration: '2m', target: 0 },      // Ramp down
    ],
    thresholds: {
        'ws_connecting': ['rate<1'],     // < 1% connection failures
        'http_req_duration': ['p(99)<200'], // P99 < 200ms
    },
};

export default function () {
    const ws = websocket.connect('wss://<your-domain>/game', null, () => {
        ws.on('open', () => {
            ws.send(JSON.stringify({
                action: 'join_table',
                table_id: 'tbl_test_001',
                player_id: `player_${__VU}`,
            }));
        });

        ws.on('message', (msg) => {
            const data = JSON.parse(msg);
            check(data, {
                'valid message': (d) => d.event_type !== undefined,
                'latency < 100ms': (d) => Date.now() - d.timestamp < 100,
            });

            // Simulate player actions
            if (d.game_state === 'waiting_for_action') {
                ws.send(JSON.stringify({
                    action: 'bet',
                    amount: d.min_bet,
                }));
            }
        });
    });

    ws.close();
}
```

### 11.4.2 Performance Validation

**Latency Targets by Layer:**

| Layer | Operation | P50 | P95 | P99 |
|-------|-----------|-----|-----|-----|
| **Game Engine** | Card deal | 5ms | 10ms | 20ms |
| **Game Engine** | Bet validation | 3ms | 8ms | 15ms |
| **Game Engine** | Hand evaluation | 2ms | 5ms | 10ms |
| **WebSocket** | Message broadcast | 10ms | 30ms | 50ms |
| **End-to-End** | Player action → State update | 50ms | 100ms | 200ms |
| **Database** | Game state write | 5ms | 15ms | 30ms |
| **Database** | Player balance update | 10ms | 25ms | 50ms |

**Database Performance Validation:**

```sql
-- Measure query latency under load
EXPLAIN ANALYZE
SELECT gs.*, p.username, p.balance
FROM game_states gs
JOIN players p ON gs.player_id = p.player_id
WHERE gs.table_id = 'tbl_test_001'
  AND gs.hand_id = 'hnd_test_001'
ORDER BY gs.sequence DESC
LIMIT 10;

-- Validate partition pruning efficiency
EXPLAIN ANALYZE
SELECT COUNT(*)
FROM game_events
WHERE event_timestamp >= '2026-01-01'::timestamptz
  AND event_timestamp < '2026-02-01'::timestamptz;
```

**Concurrency Validation:**

| Test | Method | Target | Validation |
|------|--------|--------|------------|
| **Goroutine Scalability** | pprof under load | 10K concurrent goroutines | No goroutine leaks, < 2KB/goroutine |
| **Database Connection Pool** | pgbench | 1000 connections | No connection exhaustion, steady latency |
| **Lock Contention** | race detector, mutex profiling | < 1% blocked goroutines | Minimal lock contention |
| **Memory Usage** | heap profiler under 24h load | Steady state, < 2GB | No memory leaks, GC < 10ms pauses |

---

## 11.5 Cross-Platform Consistency Testing

### 11.5.1 Client Consistency Matrix

| Test Case | iOS | Android | Web Admin | Validation |
|-----------|-----|---------|-----------|------------|
| **Hand rendering** | ✓ | ✓ | N/A | Identical card visuals, suits, ranks |
| **Pot calculation** | ✓ | ✓ | ✓ | Same pot values displayed |
| **Chip stacks** | ✓ | ✓ | ✓ | Consistent stack display |
| **Game timing** | ✓ | ✓ | N/A | Same timeout timers |
| **Animation duration** | ✓ | ✓ | N/A | Consistent timing within ±100ms |
| **Localization** | ✓ | ✓ | ✓ | Same translations for all languages |

### 11.5.2 Cross-Platform Test Automation

**Test Framework:**
- iOS: XCTest with XCUITest for UI automation
- Android: Espresso + UI Automator for UI automation
- Web: Playwright for admin panel E2E
- API: Postman/Newman for API contract tests

**Consistency Validation Script:**

```python
# Compare game state across platforms
def validate_cross_platform_state(hand_id):
    ios_state = fetch_ios_client_state(hand_id)
    android_state = fetch_android_client_state(hand_id)
    web_state = fetch_web_admin_state(hand_id)

    # Compare critical fields
    assert ios_state['pot'] == android_state['pot'] == web_state['pot']
    assert ios_state['winner'] == android_state['winner'] == web_state['winner']
    assert ios_state['rake'] == android_state['rake'] == web_state['rake']

    # Compare hand results
    for player_id in ios_state['players']:
        ios_hand = ios_state['players'][player_id]['hand']
        android_hand = android_state['players'][player_id]['hand']
        web_hand = web_state['players'][player_id]['hand']
        assert ios_hand == android_hand == web_hand

    return True
```

---

## 11.6 Release Quality Gates

### 11.6.1 Pre-Release Checklist

**Code Quality Gates:**

| Gate | Tool | Criteria | Blocker? |
|------|------|----------|----------|
| **Unit Test Coverage** | `go test -cover`, `jest --coverage` | > 85% for critical modules | Yes |
| **Static Analysis** | `golangci-lint`, `ESLint`, `SonarQube` | Zero critical issues | Yes |
| **Security Scan** | `gosec`, `npm audit`, `Snyk` | Zero high-severity vulnerabilities | Yes |
| **Integration Tests** | CI pipeline | 100% pass rate | Yes |
| **API Contract Tests** | Postman/Newman | 100% contract compliance | Yes |
| **Performance Baseline** | k6 benchmarks | Within 10% of baseline | Yes |
| **Load Test** | k6 stress test | Pass stress scenario (10K concurrent) | Yes |
| **RNG Verification** | NIST tests | All statistical tests pass | Yes |
| **Anti-Cheat Model** | ML validation suite | AUC > 0.95 on test set (example target; define based on stakeholder requirements) | Yes |

### 11.6.2 Deployment Pipeline Quality Gates

**CI/CD Pipeline Stages:**

1. **Lint & Static Analysis** (Block on critical issues)
2. **Unit Tests** (Block on failures)
3. **Integration Tests** (Block on failures)
4. **Security Scans** (Block on high-severity vulnerabilities)
5. **Performance Benchmarks** (Warn on > 10% regression, Block on > 20%)
6. **Load Test** (Block on stress test failure)
7. **Manual QA Review** (Required for major releases)
8. **Canary Deployment** (Monitor for 1 hour, auto-rollback on error rate > 0.5%)
9. **Full Rollout** (Phase out canary gradually)

**Rollback Criteria:**

- Error rate > 0.5% in canary
- P99 latency > 300ms for > 5 minutes
- Any database corruption or data loss
- Game logic bugs affecting payouts

### 11.6.3 Post-Release Validation

**24-Hour Monitoring Checklist:**

| Metric | Threshold | Action |
|--------|-----------|--------|
| **Error Rate** | < 0.1% | Alert if exceeded |
| **P99 Latency** | < 200ms | Alert if exceeded for > 5 min |
| **Active Players** | ±10% of forecast | Investigate if outside range |
| **Dispute Rate** | < 0.01% of hands | Alert if exceeded |
| **Payment Failure Rate** | < 0.5% | Alert if exceeded |
| **Anti-Cheat Score Distribution** | Stable (KS < 0.1) | Retrain model if drift detected |

---

## 11.7 Test Environment and Tooling

### 11.7.1 Test Infrastructure

**Environments:**

| Environment | Purpose | Data | Scale |
|-------------|---------|------|-------|
| **Dev** | Unit testing, feature development | Synthetic | Single instance |
| **CI** | Automated testing pipeline | Synthetic, anonymized | Small scale (100 users) |
| **QA** | Manual QA, integration testing | Anonymized production snapshot | Medium scale (1K users) |
| **Staging** | Pre-production validation | Production replica | Full scale (5K users) |
| **Performance** | Load and stress testing | Synthetic, anonymized | Stress scale (15K users) |

### 11.7.2 Test Data Management

**Data Anonymization:**

```sql
-- Anonymize production data for QA
UPDATE players
SET email = 'player_' || player_id || '@test.example.com',
    phone = '+1555' || LPAD(player_id::text, 7, '0'),
    username = 'test_user_' || player_id
WHERE environment = 'qa';
```

**Test Data Categories:**

| Category | Source | Refresh Frequency |
|----------|--------|-------------------|
| **Unit Test Data** | Synthetic fixtures | Commits |
| **Integration Data** | Hand-crafted scenarios | Weekly |
| **Performance Data** | Production anonymized snapshot | Monthly |
| **Edge Case Data** | Manually curated bug scenarios | As needed |

---

## 11.8 Testing Metrics and Reporting

### 11.8.1 Test Metrics Dashboard

**Key Metrics:**

| Metric | Target | Alert |
|--------|--------|-------|
| **Unit Test Pass Rate** | 100% | < 100% |
| **Unit Test Coverage** | > 85% | < 80% |
| **Integration Test Pass Rate** | 100% | < 98% |
| **E2E Test Pass Rate** | 100% | < 95% |
| **Test Execution Time** | < 30 min (CI) | > 45 min |
| **Flaky Test Rate** | < 1% | > 2% |
| **RNG Statistical Tests** | 100% pass | Any failure |
| **Anti-Cheat Model Performance** | AUC > 0.95 (example target) | < 0.90 |

### 11.8.2 Defect Tracking and Analysis

**Defect Severity Levels:**

| Severity | Description | SLA |
|----------|-------------|-----|
| **P0 - Critical** | Production outage, data loss, incorrect payouts | 4 hours |
| **P1 - High** | Major feature broken, security vulnerability | 24 hours |
| **P2 - Medium** | Minor feature broken, performance degradation | 1 week |
| **P3 - Low** | Cosmetic issues, documentation errors | 1 release |

**Defect Escape Analysis:**

| Stage | Escape Count | Root Cause Analysis |
|-------|--------------|---------------------|
| **Unit Test** | Track monthly | Inadequate test coverage, edge cases missed |
| **Integration Test** | Track monthly | Service contract changes, API version mismatch |
| **QA** | Track monthly | Test environment discrepancies, insufficient test data |
| **Staging** | Track monthly | Scale differences, production data edge cases |

---

## 11.9 Compliance and Regulatory Testing

### 11.9.1 Regulatory Certification Requirements

**GLI (Gaming Laboratories International) Standards:**

| Standard | Requirement | Testing Method |
|----------|--------------|-----------------|
| **GLI-STD-001** | RNG certification | Statistical tests, source code review |
| **GLI-STD-002** | Game logic verification | Independent verification of all game rules |
| **GLI-STD-003** | Paytable verification | Validate all payouts match rules |
| **GLI-STD-004** | Payout verification | Long-term payout percentage testing |

**eCOGRA Requirements:**

| Requirement | Description | Evidence |
|-------------|-------------|----------|
| **RNG Fairness** | RNG tested and certified by independent lab | Certification report |
| **Game Integrity** | Game rules mathematically verified | Mathematical analysis |
| **Payout Audits** | Monthly payout audits performed | Audit reports |
| **Responsible Gaming** | Self-exclusion, deposit limits | Feature implementation |

### 11.9.2 Audit Trail Validation

**Audit Log Requirements:**

- Every game action logged with timestamp, player_id, action, state
- Immutable log (append-only, signed)
- Retention: 7 years minimum (per regulatory requirements)
- Export format: GLI-compliant JSON/CSV
- Integrity: SHA-256 hash chain verification

**Audit Trail Verification Script:**

```go
func VerifyAuditTrail(handID string) error {
    events := LoadAuditEvents(handID)

    for i := 0; i < len(events); i++ {
        event := events[i]

        // Verify timestamp monotonicity
        if i > 0 && event.Timestamp.Before(events[i-1].Timestamp) {
            return fmt.Errorf("timestamp not monotonic at seq %d", event.Sequence)
        }

        // Verify hash chain
        if i > 0 {
            expectedHash := HashEvent(events[i-1])
            if event.PrevHash != expectedHash {
                return fmt.Errorf("hash chain broken at seq %d", event.Sequence)
            }
        }
    }

    return nil
}
```

---

## 11.10 Continuous Improvement

### 11.10.1 Test Maintenance Strategy

**Flaky Test Mitigation:**

- Automatic flaky test detection (run failed tests 3x, classify if intermittent)
- Flaky test quarantine (separate suite, fix before release)
- Root cause analysis for every flaky test
- Timeouts randomized to avoid fixed-time races

**Test Debt Tracking:**

| Category | Metric | Target | Current | Action |
|----------|--------|--------|---------|--------|
| **Coverage Gap** | % code untested | < 10% | TBD | Add tests |
| **Flaky Tests** | % of test suite | < 1% | TBD | Quarantine/fix |
| **Test Execution Time** | CI pipeline time | < 30 min | TBD | Parallelize |
| **Test Maintenance** | Time spent on test fixes | < 20% of dev time | TBD | Refactor |

### 11.10.2 Testing Best Practices

**Do's:**

- Write tests alongside implementation (TDD for critical paths)
- Use descriptive test names that explain what and why
- Test edge cases and error conditions, not just happy path
- Mock external dependencies (databases, APIs, RNG for deterministic tests)
- Keep tests independent and order-agnostic
- Use seeded RNG in game logic tests for reproducibility

**Don'ts:**

- Don't test implementation details (test behavior, not internals)
- Don't skip tests in CI (all tests must pass)
- Don't ignore flaky tests (fix before merging)
- Don't hardcode test data (use fixtures/data builders)
- Don't use sleep/timeout for synchronization (use proper synchronization primitives)
- Don't test external services directly (mock them)

---

*Section 11 provides a comprehensive testing framework ensuring game integrity, performance at scale, and regulatory compliance for the B2B poker platform.*
