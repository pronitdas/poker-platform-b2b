package fraud

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAlertStorage implements AlertStorage for testing
type MockAlertStorage struct {
	mu      sync.RWMutex
	alerts  []*AntiCheatAlert
	players map[string][]*AntiCheatAlert
}

func NewMockAlertStorage() *MockAlertStorage {
	return &MockAlertStorage{
		alerts:  make([]*AntiCheatAlert, 0),
		players: make(map[string][]*AntiCheatAlert),
	}
}

func (m *MockAlertStorage) CreateAlert(ctx context.Context, alert *AntiCheatAlert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = append(m.alerts, alert)
	m.players[alert.PlayerID] = append(m.players[alert.PlayerID], alert)
	return nil
}

func (m *MockAlertStorage) GetAlert(ctx context.Context, alertID string) (*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.alerts {
		if a.ID == alertID {
			return a, nil
		}
	}
	return nil, nil
}

func (m *MockAlertStorage) GetPlayerAlerts(ctx context.Context, playerID string, limit int) ([]*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alerts := m.players[playerID]
	if len(alerts) > limit {
		return alerts[:limit], nil
	}
	return alerts, nil
}

func (m *MockAlertStorage) GetAlertsByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*AntiCheatAlert
	for _, a := range m.alerts {
		if (a.CreatedAt.After(start) || a.CreatedAt.Equal(start)) &&
			(a.CreatedAt.Before(end) || a.CreatedAt.Equal(end)) {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAlertStorage) GetAlertsByType(ctx context.Context, alertType string, limit int) ([]*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*AntiCheatAlert
	for _, a := range m.alerts {
		if a.AlertType == alertType {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAlertStorage) GetAlertsBySeverity(ctx context.Context, severity string, limit int) ([]*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*AntiCheatAlert
	for _, a := range m.alerts {
		if a.Severity == severity {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAlertStorage) UpdateAlertStatus(ctx context.Context, alertID, status, reviewerID, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.alerts {
		if a.ID == alertID {
			a.Status = status
			a.ReviewedBy = reviewerID
			a.Notes = notes
			now := time.Now()
			a.ReviewedAt = &now
			break
		}
	}
	return nil
}

func (m *MockAlertStorage) GetPendingAlerts(ctx context.Context, limit int) ([]*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*AntiCheatAlert
	for _, a := range m.alerts {
		if a.Status == "pending" {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAlertStorage) GetAlertStats(ctx context.Context, start, end time.Time) (*AlertStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &AlertStats{}
	for _, a := range m.alerts {
		if (a.CreatedAt.After(start) || a.CreatedAt.Equal(start)) &&
			(a.CreatedAt.Before(end) || a.CreatedAt.Equal(end)) {
			switch a.Severity {
			case "critical":
				stats.Critical++
			case "high":
				stats.High++
			case "medium":
				stats.Medium++
			case "low":
				stats.Low++
			}
			switch a.Status {
			case "pending":
				stats.PendingReview++
			case "reviewed":
				stats.ConfirmedFraud++
			case "dismissed":
				stats.Dismissed++
			}
		}
	}
	return stats, nil
}

func (m *MockAlertStorage) DeleteOldAlerts(ctx context.Context, before time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	var remaining []*AntiCheatAlert
	for _, a := range m.alerts {
		if a.CreatedAt.Before(before) {
			count++
		} else {
			remaining = append(remaining, a)
		}
	}
	m.alerts = remaining
	return count, nil
}

// MockSessionStore implements SessionStore for testing
type MockSessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]PlayerSession
}

func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: make(map[string][]PlayerSession),
	}
}

func (m *MockSessionStore) CreateSession(ctx context.Context, session *PlayerSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.PlayerID] = append(m.sessions[session.PlayerID], *session)
	return nil
}

func (m *MockSessionStore) GetSession(ctx context.Context, sessionID string) (*PlayerSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, sessions := range m.sessions {
		for i := range sessions {
			if sessions[i].SessionID == sessionID {
				return &sessions[i], nil
			}
		}
	}
	return nil, nil
}

func (m *MockSessionStore) GetPlayerSessions(ctx context.Context, playerID string, startTime, endTime time.Time) ([]PlayerSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []PlayerSession
	for _, s := range m.sessions[playerID] {
		if (s.ConnectedAt.After(startTime) || s.ConnectedAt.Equal(startTime)) &&
			(s.ConnectedAt.Before(endTime) || s.ConnectedAt.Equal(endTime)) {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockSessionStore) GetActiveSessions(ctx context.Context, playerID string) ([]PlayerSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var active []PlayerSession
	for _, s := range m.sessions[playerID] {
		if s.DisconnectedAt == nil {
			active = append(active, s)
		}
	}
	return active, nil
}

func (m *MockSessionStore) EndSession(ctx context.Context, sessionID string, endingChips int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, sessions := range m.sessions {
		for i := range sessions {
			if sessions[i].SessionID == sessionID {
				now := time.Now()
				sessions[i].DisconnectedAt = &now
				sessions[i].EndingChips = endingChips
				break
			}
		}
	}
	return nil
}

func (m *MockSessionStore) UpdateSessionStats(ctx context.Context, sessionID string, handsPlayed, totalWins, totalLosses int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, sessions := range m.sessions {
		for i := range sessions {
			if sessions[i].SessionID == sessionID {
				sessions[i].HandsPlayed = handsPlayed
				sessions[i].TotalWins = totalWins
				sessions[i].TotalLosses = totalLosses
				break
			}
		}
	}
	return nil
}

func (m *MockSessionStore) GetAllPlayerIDs(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	playerIDs := make([]string, 0, len(m.sessions))
	for playerID := range m.sessions {
		playerIDs = append(playerIDs, playerID)
	}
	return playerIDs, nil
}

func (m *MockSessionStore) DeleteOldSessions(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

// MockFingerprintDatabase implements FingerprintDatabase for testing
type MockFingerprintDatabase struct {
	mu           sync.RWMutex
	fingerprints map[string][]DeviceFingerprint
}

func NewMockFingerprintDatabase() *MockFingerprintDatabase {
	return &MockFingerprintDatabase{
		fingerprints: make(map[string][]DeviceFingerprint),
	}
}

func (m *MockFingerprintDatabase) StoreFingerprint(ctx context.Context, fp DeviceFingerprint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fingerprints[fp.PlayerID] = append(m.fingerprints[fp.PlayerID], fp)
	return nil
}

func (m *MockFingerprintDatabase) GetFingerprintHistory(ctx context.Context, playerID string) ([]DeviceFingerprint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fingerprints[playerID], nil
}

func (m *MockFingerprintDatabase) FindAccountsByFingerprint(ctx context.Context, fingerprint string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var players []string
	for playerID, fps := range m.fingerprints {
		for _, fp := range fps {
			if fp.Fingerprint == fingerprint {
				players = append(players, playerID)
				break
			}
		}
	}
	return players, nil
}

func (m *MockFingerprintDatabase) FindAccountsByIP(ctx context.Context, ip string) ([]string, error) {
	return nil, nil
}

func (m *MockFingerprintDatabase) FindAccountsByNetwork(ctx context.Context, networkPrefix string) ([]string, error) {
	return nil, nil
}

func (m *MockFingerprintDatabase) GetLatestFingerprint(ctx context.Context, playerID string) (*DeviceFingerprint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fps := m.fingerprints[playerID]
	if len(fps) > 0 {
		return &fps[len(fps)-1], nil
	}
	return nil, nil
}

func (m *MockFingerprintDatabase) FingerprintExists(ctx context.Context, fingerprint string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, fps := range m.fingerprints {
		for _, fp := range fps {
			if fp.Fingerprint == fingerprint {
				return true, nil
			}
		}
	}
	return false, nil
}

// MockTransferDatabase implements TransferDatabase for testing
type MockTransferDatabase struct {
	mu        sync.RWMutex
	transfers []TransferRecord
	chipFlows map[string]map[string][]TransferRecord
}

func NewMockTransferDatabase() *MockTransferDatabase {
	return &MockTransferDatabase{
		transfers: make([]TransferRecord, 0),
		chipFlows: make(map[string]map[string][]TransferRecord),
	}
}

func (m *MockTransferDatabase) RecordTransfer(ctx context.Context, record *TransferRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transfers = append(m.transfers, *record)

	p1, p2 := record.FromPlayer, record.ToPlayer
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	if m.chipFlows[p1] == nil {
		m.chipFlows[p1] = make(map[string][]TransferRecord)
	}
	m.chipFlows[p1][p2] = append(m.chipFlows[p1][p2], *record)
	return nil
}

func (m *MockTransferDatabase) GetPlayerTransfers(ctx context.Context, playerID string, startTime, endTime time.Time) ([]TransferRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []TransferRecord
	for _, t := range m.transfers {
		if (t.Timestamp.After(startTime) || t.Timestamp.Equal(startTime)) &&
			(t.Timestamp.Before(endTime) || t.Timestamp.Equal(endTime)) {
			if t.FromPlayer == playerID || t.ToPlayer == playerID {
				result = append(result, t)
			}
		}
	}
	return result, nil
}

func (m *MockTransferDatabase) GetPlayerTransferStats(ctx context.Context, playerID string) (*TransferStats, error) {
	return nil, nil
}

func NewMockAlertStorage() *MockAlertStorage {
	return &MockAlertStorage{
		alerts:  make([]*AntiCheatAlert, 0),
		players: make(map[string][]*AntiCheatAlert),
	}
}

func (m *MockAlertStorage) CreateAlertTable(ctx context.Context) error {
	return nil
}

func (m *MockAlertStorage) StoreAlert(ctx context.Context, alert *AntiCheatAlert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = append(m.alerts, alert)
	m.players[alert.PlayerID] = append(m.players[alert.PlayerID], alert)
	return nil
}

func (m *MockAlertStorage) GetAlert(ctx context.Context, alertID string) (*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.alerts {
		if a.ID == alertID {
			return a, nil
		}
	}
	return nil, nil
}

func (m *MockAlertStorage) GetPlayerAlerts(ctx context.Context, playerID string, limit int) ([]*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alerts := m.players[playerID]
	if len(alerts) > limit {
		return alerts[:limit], nil
	}
	return alerts, nil
}

func (m *MockAlertStorage) GetAlertsByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*AntiCheatAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*AntiCheatAlert
	for _, a := range m.alerts {
		if (a.CreatedAt.After(start) || a.CreatedAt.Equal(start)) &&
			(a.CreatedAt.Before(end) || a.CreatedAt.Equal(end)) {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAlertStorage) GetAlertStats(ctx context.Context, agentID string) (*AlertStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &AlertStats{AgentID: agentID}
	for _, a := range m.alerts {
		if agentID == "" || a.AgentID == agentID {
			switch a.Severity {
			case "critical":
				stats.Critical++
			case "high":
				stats.High++
			case "medium":
				stats.Medium++
			case "low":
				stats.Low++
			}
			switch a.Status {
			case "pending":
				stats.PendingReview++
			case "reviewed":
				stats.ConfirmedFraud++
			case "dismissed":
				stats.Dismissed++
			}
		}
	}
	return stats, nil
}

func (m *MockAlertStorage) UpdateAlertStatus(ctx context.Context, alertID, status, reviewerID, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.alerts {
		if a.ID == alertID {
			a.Status = status
			a.ReviewedBy = reviewerID
			a.Notes = notes
			now := time.Now()
			a.ReviewedAt = &now
			break
		}
	}
	return nil
}

func (m *MockAlertStorage) BulkUpdateAlerts(ctx context.Context, alertIDs []string, status string) error {
	return nil
}

func (m *MockAlertStorage) ResolveAlert(ctx context.Context, alertID, resolution string) error {
	return m.UpdateAlertStatus(ctx, alertID, "reviewed", "system", resolution)
}

// MockSessionStore implements SessionStore for testing
type MockSessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]PlayerSession
}

func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: make(map[string][]PlayerSession),
	}
}

func (m *MockSessionStore) CreateSessionTable(ctx context.Context) error {
	return nil
}

func (m *MockSessionStore) RecordSession(ctx context.Context, session *PlayerSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.PlayerID] = append(m.sessions[session.PlayerID], *session)
	return nil
}

func (m *MockSessionStore) GetPlayerSessions(ctx context.Context, playerID string, limit int) ([]PlayerSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessions := m.sessions[playerID]
	if len(sessions) > limit {
		return sessions[:limit], nil
	}
	return sessions, nil
}

func (m *MockSessionStore) UpdateSessionEnd(ctx context.Context, sessionID string, endTime time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, sessions := range m.sessions {
		for i := range sessions {
			if sessions[i].SessionID == sessionID {
				now := endTime
				sessions[i].DisconnectedAt = &now
				break
			}
		}
	}
	return nil
}

func (m *MockSessionStore) GetActiveSessions(ctx context.Context, playerID string) ([]PlayerSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var active []PlayerSession
	for _, s := range m.sessions[playerID] {
		if s.DisconnectedAt == nil {
			active = append(active, s)
		}
	}
	return active, nil
}

// MockFingerprintDatabase implements FingerprintDatabase for testing
type MockFingerprintDatabase struct {
	mu           sync.RWMutex
	fingerprints map[string][]DeviceFingerprint
}

func NewMockFingerprintDatabase() *MockFingerprintDatabase {
	return &MockFingerprintDatabase{
		fingerprints: make(map[string][]DeviceFingerprint),
	}
}

func (m *MockFingerprintDatabase) CreateFingerprintTable(ctx context.Context) error {
	return nil
}

func (m *MockFingerprintDatabase) StoreFingerprint(ctx context.Context, fp DeviceFingerprint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fingerprints[fp.PlayerID] = append(m.fingerprints[fp.PlayerID], fp)
	return nil
}

func (m *MockFingerprintDatabase) GetFingerprintHistory(ctx context.Context, playerID string) ([]DeviceFingerprint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fingerprints[playerID], nil
}

func (m *MockFingerprintDatabase) FindPlayersByFingerprint(ctx context.Context, fingerprint string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var players []string
	for playerID, fps := range m.fingerprints {
		for _, fp := range fps {
			if fp.Fingerprint == fingerprint {
				players = append(players, playerID)
				break
			}
		}
	}
	return players, nil
}

// MockTransferDatabase implements TransferDatabase for testing
type MockTransferDatabase struct {
	mu        sync.RWMutex
	transfers []TransferRecord
	chipFlows map[string]map[string][]TransferRecord
}

func NewMockTransferDatabase() *MockTransferDatabase {
	return &MockTransferDatabase{
		transfers: make([]TransferRecord, 0),
		chipFlows: make(map[string]map[string][]TransferRecord),
	}
}

func (m *MockTransferDatabase) CreateTransferTable(ctx context.Context) error {
	return nil
}

func (m *MockTransferDatabase) RecordTransfer(ctx context.Context, transfer TransferRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transfers = append(m.transfers, transfer)

	p1, p2 := transfer.FromPlayer, transfer.ToPlayer
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	if m.chipFlows[p1] == nil {
		m.chipFlows[p1] = make(map[string][]TransferRecord)
	}
	m.chipFlows[p1][p2] = append(m.chipFlows[p1][p2], transfer)
	return nil
}

func (m *MockTransferDatabase) GetPlayerTransfers(ctx context.Context, playerID string, limit int) ([]TransferRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []TransferRecord
	for _, t := range m.transfers {
		if t.FromPlayer == playerID || t.ToPlayer == playerID {
			result = append(result, t)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// Integration test for the full fraud detection pipeline
func TestFraudDetectionPipeline(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" && os.Getenv("CI") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	ctx := context.Background()

	// Setup mock storage
	alertStorage := NewMockAlertStorage()
	sessionStore := NewMockSessionStore()
	fingerprintDB := NewMockFingerprintDatabase()
	transferDB := NewMockTransferDatabase()

	// Setup detectors
	botConfig := DefaultBotDetectionConfig()
	botDetector := NewBotDetector(botConfig)

	collusionConfig := DefaultCollusionDetectionConfig()
	collusionDetector := NewCollusionDetector(collusionConfig, transferDB)

	multiAccountConfig := DefaultMultiAccountConfig()
	multiAccountDetector := NewMultiAccountDetector(
		multiAccountConfig,
		fingerprintDB,
		nil,
		sessionStore,
	)

	ruleDetector := NewRuleBasedDetector()

	riskConfig := DefaultRiskScoringConfig()
	riskScorer := NewRiskScorer(riskConfig,
		botDetector,
		collusionDetector,
		multiAccountDetector,
		ruleDetector,
		alertStorage,
	)

	// Test data: Create a bot-like player
	t.Run("Bot Detection Pipeline", func(t *testing.T) {
		playerID := "player_bot_001"
		tableID := "table_001"

		// Create bot-like behavioral features
		features := PlayerBehavioralFeatures{
			PlayerID:         playerID,
			TimeRange:        "1h",
			ExtractedAt:      time.Now(),
			AvgActionTime:    1.5, // Very fast, bot-like
			ActionTimeStdDev: 0.3, // Too consistent
			ActionTimeMin:    1.0,
			ActionTimeMax:    2.0,
			BetPrecision:     0.98, // Extremely precise
			AvgBetToPotRatio: 0.75,
			BetSizeVariance:  0.05,
			HandsPlayed:      150,
			HandsPerHour:     120,  // Too fast
			TablesConcurrent: 8,    // Too many
			ConsistencyScore: 0.95, // Too consistent
			WinRate:          0.72,
			TotalProfit:      50000,
			SessionDuration:  2 * time.Hour,
			AggressionFactor: 0.25,
			VPIP:             0.35,
			PFR:              0.30,
			ThreeBetRate:     0.08,
			ShowdownRate:     0.45,
			CbetSuccess:      0.70,
		}

		// Run bot detection
		result := botDetector.DetectBot(ctx, &features)

		assert.NotNil(t, result)
		assert.True(t, result.Score > 0.6, "Bot score should be high for bot-like behavior")
		assert.Equal(t, "bot", result.RecommendedAction, "Should recommend bot action")

		t.Logf("Bot detection result: Score=%.3f, Action=%s, Reasons=%v",
			result.Score, result.RecommendedAction, result.Reasons)
	})

	// Test data: Create colluding players
	t.Run("Collusion Detection Pipeline", func(t *testing.T) {
		playerA := "player_colluder_a"
		playerB := "player_colluder_b"

		// Record chip transfers between them
		transferDB.RecordTransfer(ctx, TransferRecord{
			FromPlayer: playerA,
			ToPlayer:   playerB,
			TableID:    "table_001",
			HandID:     "hand_001",
			Amount:     15000,
			EVImpact:   -0.12,
			Context:    "showdown",
			Timestamp:  time.Now().Add(-1 * time.Hour),
		})

		transferDB.RecordTransfer(ctx, TransferRecord{
			FromPlayer: playerB,
			ToPlayer:   playerA,
			TableID:    "table_002",
			HandID:     "hand_002",
			Amount:     8000,
			EVImpact:   -0.08,
			Context:    "bet",
			Timestamp:  time.Now().Add(-30 * time.Minute),
		})

		// Record sessions showing they play together frequently
		now := time.Now()
		sessionStore.RecordSession(ctx, &PlayerSession{
			SessionID:      "session_a_001",
			PlayerID:       playerA,
			ConnectedAt:    now.Add(-3 * time.Hour),
			DisconnectedAt: &now,
		})

		sessionStore.RecordSession(ctx, &PlayerSession{
			SessionID:      "session_b_001",
			PlayerID:       playerB,
			ConnectedAt:    now.Add(-3 * time.Hour),
			DisconnectedAt: &now,
		})

		// Run collusion detection
		result := collusionDetector.DetectCollusion(ctx, playerA, playerB)

		assert.NotNil(t, result)
		assert.True(t, result.Score > 0.5, "Collusion score should be elevated")

		t.Logf("Collusion detection result: Score=%.3f, Type=%s, Confidence=%.3f",
			result.Score, result.CollusionType, result.Confidence)
	})

	// Test data: Create multi-account player
	t.Run("Multi-Account Detection Pipeline", func(t *testing.T) {
		mainPlayer := "player_main_account"
		altPlayer1 := "player_alt_001"
		altPlayer2 := "player_alt_002"
		fingerprint := "fp_abc123xyz"

		// Record same device fingerprint for multiple accounts
		fingerprintDB.StoreFingerprint(ctx, DeviceFingerprint{
			PlayerID:    mainPlayer,
			Fingerprint: fingerprint,
			DeviceType:  "desktop",
			OS:          "Windows 10",
			Browser:     "Chrome 120",
			CreatedAt:   time.Now().Add(-30 * time.Day),
		})

		fingerprintDB.StoreFingerprint(ctx, DeviceFingerprint{
			PlayerID:    altPlayer1,
			Fingerprint: fingerprint,
			DeviceType:  "desktop",
			OS:          "Windows 10",
			Browser:     "Chrome 120",
			CreatedAt:   time.Now().Add(-20 * time.Day),
		})

		fingerprintDB.StoreFingerprint(ctx, DeviceFingerprint{
			PlayerID:    altPlayer2,
			Fingerprint: fingerprint,
			DeviceType:  "desktop",
			OS:          "Windows 10",
			Browser:     "Chrome 120",
			CreatedAt:   time.Now().Add(-10 * time.Day),
		})

		// Run multi-account detection
		result := multiAccountDetector.DetectMultiAccount(ctx, mainPlayer)

		assert.NotNil(t, result)
		assert.True(t, len(result.RelatedAccounts) >= 2, "Should find related accounts")
		assert.Equal(t, "device", result.RelatedAccounts[0].ConnectionType)

		t.Logf("Multi-account detection result: Score=%.3f, Related=%d",
			result.Score, len(result.RelatedAccounts))
	})

	// Test data: Rule violation detection
	t.Run("Rule Detection Pipeline", func(t *testing.T) {
		playerID := "player_rule_violator"

		ruleData := &RuleCheckData{
			PlayerID:            playerID,
			HandsPlayed24h:      200,
			WinRate24h:          0.85, // Suspiciously high
			HandsPlayed7d:       500,
			WinRate7d:           0.72,
			AvgActionTime:       1.2,
			AccountsFromIP:      map[string]int{"192.168.1.1": 5},
			AccountsFromDevice:  map[string]int{"fp_xyz": 3},
			CurrentSessionHands: 50,
			TotalChipsWon:       100000,
			AlertCount24h:       3,
		}

		// Run rule detection
		results := ruleDetector.CheckAllRules(ctx, playerID, ruleData)

		assert.NotEmpty(t, results, "Should trigger at least one rule")

		highSeverityRules := 0
		for _, r := range results {
			if r.Rule.Severity == "high" || r.Rule.Severity == "critical" {
				highSeverityRules++
			}
		}
		assert.True(t, highSeverityRules >= 2, "Should trigger high severity rules")

		t.Logf("Rule detection results: Total=%d, HighSeverity=%d", len(results), highSeverityRules)
		for _, r := range results {
			t.Logf("  - %s (%s): %v", r.Rule.Name, r.Rule.Severity, r.Triggered)
		}
	})

	// Test end-to-end risk scoring
	t.Run("End-to-End Risk Scoring", func(t *testing.T) {
		highRiskPlayer := "player_high_risk"

		// Store some alerts for this player
		alertStorage.StoreAlert(ctx, &AntiCheatAlert{
			ID:        "alert_001",
			PlayerID:  highRiskPlayer,
			AlertType: "bot",
			Severity:  "high",
			Score:     0.85,
			AgentID:   "agent_001",
			ClubID:    "club_001",
			Evidence:  []string{"Bot-like timing patterns"},
			Status:    "pending",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		})

		// Calculate overall risk score
		riskScore := riskScorer.CalculateOverallRisk(ctx, highRiskPlayer)

		assert.NotNil(t, riskScore)
		assert.True(t, riskScore.OverallScore > 0.5, "Should have elevated risk score")
		assert.True(t, riskScore.ReviewRecommended, "Should recommend review")

		t.Logf("Overall risk score: Overall=%.3f, Bot=%.3f, Collusion=%.3f, MultiAccount=%.3f, Rules=%.3f",
			riskScore.OverallScore,
			riskScore.BotScore,
			riskScore.CollusionScore,
			riskScore.MultiAccountScore,
			riskScore.RulesScore)
	})

	// Test alert generation
	t.Run("Alert Generation Pipeline", func(t *testing.T) {
		alertPlayer := "player_alert_test"

		// Create a risk score that should trigger an alert
		riskScore := &RiskScore{
			PlayerID:          alertPlayer,
			AgentID:           "agent_001",
			OverallScore:      0.85,
			BotScore:          0.90,
			CollusionScore:    0.30,
			MultiAccountScore: 0.20,
			ChipDumpScore:     0.15,
			LastCalculated:    time.Now(),
			CalculatedFrom:    time.Now().Add(-1 * time.Hour),
			CalculatedTo:      time.Now(),
			FlagCount24h:      2,
			FlagCount7d:       5,
			ReviewRecommended: true,
		}

		// Generate alert
		alert := riskScorer.GenerateAlert(ctx, riskScore)

		assert.NotNil(t, alert)
		assert.Equal(t, "bot", alert.AlertType)
		assert.Equal(t, "high", alert.Severity)
		assert.Equal(t, alertPlayer, alert.PlayerID)

		// Store the alert
		err := alertStorage.StoreAlert(ctx, alert)
		require.NoError(t, err)

		// Verify alert was stored
		storedAlerts, err := alertStorage.GetPlayerAlerts(ctx, alertPlayer, 10)
		require.NoError(t, err)
		assert.Len(t, storedAlerts, 1)

		t.Logf("Generated alert: ID=%s, Type=%s, Severity=%s, Score=%.3f",
			alert.ID, alert.AlertType, alert.Severity, alert.Score)
	})
}

// TestMetricsRecording tests that metrics are properly recorded
func TestMetricsRecording(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" && os.Getenv("CI") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	t.Run("Record Bot Detection Metrics", func(t *testing.T) {
		RecordBotDetection("heuristic", 0.015, 0.85, true)
		RecordBotDetection("heuristic", 0.012, 0.72, false)
		RecordBotDetection("isolation_forest", 0.025, 0.91, true)
		RecordBotDetection("lstm", 0.045, 0.78, true)

		t.Log("Bot detection metrics recorded successfully")
	})

	t.Run("Record Collusion Metrics", func(t *testing.T) {
		RecordCollusionDetection("graph", 0.120, 0.72, "high")
		RecordCollusionDetection("pairwise", 0.085, 0.65, "medium")
		RecordCollusionRing(3)
		RecordCollusionRing(5)

		t.Log("Collusion detection metrics recorded successfully")
	})

	t.Run("Record Multi-Account Metrics", func(t *testing.T) {
		RecordMultiAccountDetection("device", 0.008, 0.88, "critical", 2)
		RecordMultiAccountDetection("ip", 0.005, 0.72, "high", 1)
		RecordMultiAccountDetection("session", 0.012, 0.45, "low", 0)

		t.Log("Multi-account detection metrics recorded successfully")
	})

	t.Run("Record Risk Score Metrics", func(t *testing.T) {
		breakdown := map[string]float64{
			"bot":           0.85,
			"collusion":     0.30,
			"multi_account": 0.20,
			"rules":         0.45,
		}
		RecordRiskScore(0.72, breakdown, "high_activity")

		t.Log("Risk score metrics recorded successfully")
	})

	t.Run("Record Alert Metrics", func(t *testing.T) {
		RecordAlert("bot", "high", 3600, "reviewed")
		RecordAlert("bot", "critical", 7200, "confirmed")
		RecordAlert("collusion", "high", 1800, "dismissed")
		UpdateAlertStatus("pending", 15)
		UpdateAlertStatus("reviewed", 8)
		UpdateAlertStatus("dismissed", 3)

		t.Log("Alert metrics recorded successfully")
	})

	t.Run("Record Feature Extraction Metrics", func(t *testing.T) {
		RecordFeatureExtraction("timing", 0.005, true)
		RecordFeatureExtraction("betting", 0.003, true)
		RecordFeatureExtraction("volume", 0.002, true)
		RecordFeatureExtraction("pattern", 0.008, false) // This one fails

		t.Log("Feature extraction metrics recorded successfully")
	})

	t.Run("Record Overall Detection Metrics", func(t *testing.T) {
		RecordFraudDetection(0.125)
		RecordFraudDetection(0.098)
		RecordFraudDetection(0.156)
		RecordError("bot_detector", "timeout")
		RecordError("collusion_detector", "database")

		t.Log("Overall detection metrics recorded successfully")
	})
}

// TestKafkaProducerIntegration tests the Kafka producer
func TestKafkaProducerIntegration(t *testing.T) {
	if os.Getenv("KAFKA_TEST") == "" {
		t.Skip("Skipping Kafka test. Set KAFKA_TEST=1 to run.")
	}

	ctx := context.Background()

	config := KafkaAlertProducerConfig{
		Brokers:      []string{"localhost:9092"},
		Topic:        "fraud-alerts-test",
		MaxRetries:   3,
		RetryBackoff: 100 * time.Millisecond,
		RequiredAcks: sarama.WaitForAll,
		Compression:  sarama.LZ4,
		AsyncMode:    false,
	}

	// This will fail if Kafka is not running
	producer, err := NewKafkaAlertProducer(config)
	if err != nil {
		t.Skipf("Kafka not available: %v", err)
	}
	defer producer.Close()

	t.Run("Publish Single Alert", func(t *testing.T) {
		alert := &AntiCheatAlert{
			ID:        "test_alert_001",
			PlayerID:  "test_player_001",
			AlertType: "bot",
			Severity:  "high",
			Score:     0.85,
			AgentID:   "test_agent",
			ClubID:    "test_club",
			Evidence:  []string{"Test evidence"},
			Status:    "pending",
			CreatedAt: time.Now(),
		}

		breakdown := map[string]float64{
			"bot":           0.85,
			"collusion":     0.15,
			"multi_account": 0.10,
		}

		err := producer.PublishAlert(ctx, alert, breakdown)
		require.NoError(t, err)

		stats := producer.GetStats()
		assert.Equal(t, int64(1), stats.MessagesSent)

		t.Logf("Alert published successfully, partition info available in stats")
	})

	t.Run("Publish Batch Alerts", func(t *testing.T) {
		alerts := make([]*AntiCheatAlert, 5)
		breakdowns := make([]map[string]float64, 5)

		for i := 0; i < 5; i++ {
			alerts[i] = &AntiCheatAlert{
				ID:        string("test_alert_batch_", i),
				PlayerID:  string("test_player_batch_", i),
				AlertType: "collusion",
				Severity:  "medium",
				Score:     0.65,
				AgentID:   "test_agent",
				ClubID:    "test_club",
				Evidence:  []string{"Batch test evidence"},
				Status:    "pending",
				CreatedAt: time.Now(),
			}
			breakdowns[i] = map[string]float64{"collusion": 0.65}
		}

		err := producer.PublishBatch(ctx, alerts, breakdowns)
		require.NoError(t, err)

		stats := producer.GetStats()
		assert.Equal(t, int64(6), stats.MessagesSent) // 1 + 5

		t.Logf("Batch alerts published successfully")
	})
}

// TestConfigurationLoading tests configuration loading
func TestConfigurationLoading(t *testing.T) {
	t.Run("Default Bot Detection Config", func(t *testing.T) {
		config := DefaultBotDetectionConfig()

		assert.NotNil(t, config)
		assert.Equal(t, 3.0, config.ActionTimeMeanThreshold)
		assert.Equal(t, 0.5, config.ActionTimeStdDevThreshold)
		assert.Equal(t, 0.95, config.BetPrecisionThreshold)
		assert.Equal(t, 100, config.HandsPerHourThreshold)
		assert.Equal(t, 10, config.ConcurrentTablesThreshold)
		assert.Equal(t, 0.85, config.ConsistencyScoreThreshold)

		// Check weights sum to 1
		weightSum := config.ActionTimeMeanWeight +
			config.ActionTimeStdDevWeight +
			config.BetPrecisionWeight +
			config.HandsPerHourWeight +
			config.ConcurrentTablesWeight +
			config.ConsistencyScoreWeight
		assert.InDelta(t, 1.0, weightSum, 0.01)
	})

	t.Run("Default Collusion Detection Config", func(t *testing.T) {
		config := DefaultCollusionDetectionConfig()

		assert.NotNil(t, config)
		assert.Equal(t, 50, config.CoOccurrenceThreshold)
		assert.Equal(t, 10, config.SeatingAdjacencyThreshold)
		assert.Equal(t, 0.7, config.StakeOverlapThreshold)
		assert.Equal(t, 5*time.Minute, config.ArrivalSyncThreshold)
		assert.Equal(t, 5*time.Minute, config.DepartureSyncThreshold)

		// Check weights
		assert.Equal(t, 0.15, config.CoOccurrenceWeight)
		assert.Equal(t, 0.10, config.SeatingAdjacencyWeight)
		assert.Equal(t, 0.10, config.StakeOverlapWeight)
		assert.Equal(t, 0.10, config.ArrivalSyncWeight)
		assert.Equal(t, 0.20, config.AggressionDeltaWeight)
		assert.Equal(t, 0.15, config.ChipTransferWeight)
	})

	t.Run("Default Multi-Account Config", func(t *testing.T) {
		config := DefaultMultiAccountConfig()

		assert.NotNil(t, config)
		assert.Equal(t, 3, config.DeviceMatchThreshold)
		assert.Equal(t, 5, config.IPMatchThreshold)
		assert.Equal(t, 3, config.SessionOverlapThreshold)
		assert.Equal(t, 0.8, config.BehavioralMatchThreshold)
		assert.Equal(t, 10, config.Network24hThreshold)
		assert.Equal(t, 20, config.Network7dThreshold)
	})

	t.Run("Default Risk Scoring Config", func(t *testing.T) {
		config := DefaultRiskScoringConfig()

		assert.NotNil(t, config)

		// Check weights sum to 1
		weightSum := config.BotWeight + config.CollusionWeight +
			config.MultiAccountWeight + config.RulesWeight + config.HistoricalWeight
		assert.InDelta(t, 1.0, weightSum, 0.01)

		assert.Equal(t, 0.80, config.HighRiskThreshold)
		assert.Equal(t, 0.60, config.MediumRiskThreshold)
		assert.Equal(t, 0.50, config.ReviewThreshold)
	})
}

// TestAlertMessageSerialization tests Kafka message serialization
func TestAlertMessageSerialization(t *testing.T) {
	t.Run("Serialize AlertMessage", func(t *testing.T) {
		msg := AlertMessage{
			ID:         "alert_001",
			PlayerID:   "player_001",
			AlertType:  "bot",
			Severity:   "high",
			Score:      0.85,
			TableID:    "table_001",
			HandID:     "hand_001",
			AgentID:    "agent_001",
			ClubID:     "club_001",
			Evidence:   []string{"Fast timing", "Consistent bet sizing"},
			Timestamp:  time.Now(),
			DetectedAt: time.Now(),
			RiskBreakdown: map[string]float64{
				"bot":           0.85,
				"collusion":     0.15,
				"multi_account": 0.10,
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		var decoded AlertMessage
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, msg.ID, decoded.ID)
		assert.Equal(t, msg.PlayerID, decoded.PlayerID)
		assert.Equal(t, msg.AlertType, decoded.AlertType)
		assert.Equal(t, msg.Severity, decoded.Severity)
		assert.Equal(t, msg.Score, decoded.Score)
		assert.Equal(t, len(msg.Evidence), len(decoded.Evidence))
		assert.Equal(t, len(msg.RiskBreakdown), len(decoded.RiskBreakdown))

		t.Logf("AlertMessage serialized and deserialized successfully")
	})
}
