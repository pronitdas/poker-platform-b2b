package storage

import (
	"context"
	"time"

	"poker-platform/internal/fraud"
)

// AlertStorage defines the interface for storing and retrieving anti-cheat alerts
type AlertStorage interface {
	// Create a new alert
	CreateAlert(ctx context.Context, alert *fraud.AntiCheatAlert) error
	
	// Get alert by ID
	GetAlert(ctx context.Context, alertID string) (*fraud.AntiCheatAlert, error)
	
	// Get all alerts for a player
	GetPlayerAlerts(ctx context.Context, playerID string, limit int) ([]*fraud.AntiCheatAlert, error)
	
	// Get alerts by time range
	GetAlertsByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*fraud.AntiCheatAlert, error)
	
	// Get alerts by type
	GetAlertsByType(ctx context.Context, alertType string, limit int) ([]*fraud.AntiCheatAlert, error)
	
	// Get alerts by severity
	GetAlertsBySeverity(ctx context.Context, severity string, limit int) ([]*fraud.AntiCheatAlert, error)
	
	// Update alert status (for review workflow)
	UpdateAlertStatus(ctx context.Context, alertID, status, reviewerID, notes string) error
	
	// Get pending alerts
	GetPendingAlerts(ctx context.Context, limit int) ([]*fraud.AntiCheatAlert, error)
	
	// Get alert statistics
	GetAlertStats(ctx context.Context, start, end time.Time) (*AlertStats, error)
	
	// Delete old alerts (for retention policy)
	DeleteOldAlerts(ctx context.Context, before time.Time) (int64, error)
}

// SessionStore defines the interface for storing player sessions
type SessionStore interface {
	// Create a new session
	CreateSession(ctx context.Context, session *fraud.PlayerSession) error
	
	// Get session by ID
	GetSession(ctx context.Context, sessionID string) (*fraud.PlayerSession, error)
	
	// Get all sessions for a player
	GetPlayerSessions(ctx context.Context, playerID string, startTime, endTime time.Time) ([]fraud.PlayerSession, error)
	
	// Get active sessions for a player
	GetActiveSessions(ctx context.Context, playerID string) ([]fraud.PlayerSession, error)
	
	// End a session
	EndSession(ctx context.Context, sessionID string, endingChips int64) error
	
	// Update session statistics
	UpdateSessionStats(ctx context.Context, sessionID string, handsPlayed, totalWins, totalLosses int) error
	
	// Get all player IDs (for batch processing)
	GetAllPlayerIDs(ctx context.Context) ([]string, error)
	
	// Cleanup old sessions
	DeleteOldSessions(ctx context.Context, before time.Time) (int64, error)
}

// FingerprintDatabase defines the interface for device fingerprint storage
type FingerprintDatabase interface {
	// Store a device fingerprint
	StoreFingerprint(ctx context.Context, fp fraud.DeviceFingerprint) error
	
	// Get fingerprint history for a player
	GetFingerprintHistory(ctx context.Context, playerID string) ([]fraud.DeviceFingerprint, error)
	
	// Find accounts sharing a device fingerprint
	FindAccountsByFingerprint(ctx context.Context, fingerprint string) ([]string, error)
	
	// Find accounts sharing an IP address
	FindAccountsByIP(ctx context.Context, ip string) ([]string, error)
	
	// Find accounts in a network range
	FindAccountsByNetwork(ctx context.Context, networkPrefix string) ([]string, error)
	
	// Get most recent fingerprint for a player
	GetLatestFingerprint(ctx context.Context, playerID string) (*fraud.DeviceFingerprint, error)
	
	// Check if fingerprint exists
	FingerprintExists(ctx context.Context, fingerprint string) (bool, error)
}

// TransferDatabase defines the interface for chip transfer tracking
type TransferDatabase interface {
	// Record a chip transfer
	RecordTransfer(ctx context.Context, record *fraud.TransferRecord) error
	
	// Get all transfers for a player
	GetPlayerTransfers(ctx context.Context, playerID string, startTime, endTime time.Time) ([]fraud.TransferRecord, error)
	
	// Get transfers between two players
	GetPairTransfers(ctx context.Context, playerA, playerB string, startTime, endTime time.Time) ([]fraud.TransferRecord, error)
	
	// Calculate net transfer between two players
	CalculateNetTransfer(ctx context.Context, playerA, playerB string, startTime, endTime time.Time) (int64, error)
	
	// Get transfer count between two players
	GetTransferCount(ctx context.Context, playerA, playerB string, startTime, endTime time.Time) (int, error)
	
	// Calculate EV loss rate for a player pair
	CalculateEVLossRate(ctx context.Context, playerA, playerB string, startTime, endTime time.Time) (float64, error)
}

// PlayerStatsStorage defines the interface for player statistics
type PlayerStatsStorage interface {
	// Update player stats
	UpdatePlayerStats(ctx context.Context, playerID string, stats *PlayerStats) error
	
	// Get player stats
	GetPlayerStats(ctx context.Context, playerID string) (*PlayerStats, error)
	
	// Get hands played in time range
	GetHandsPlayed(ctx context.Context, playerID string, startTime, endTime time.Time) (int, error)
	
	// Get win rate
	GetWinRate(ctx context.Context, playerID string, startTime, endTime time.Time) (float64, error)
	
	// Update chips
	UpdateChips(ctx context.Context, playerID string, delta int64) error
	
	// Get chip balance
	GetChips(ctx context.Context, playerID string) (int64, error)
}

// HandHistoryStorage defines the interface for hand history storage
type HandHistoryStorage interface {
	// Store a completed hand
	StoreHand(ctx context.Context, hand *fraud.HandHistory) error
	
	// Get hand by ID
	GetHand(ctx context.Context, handID string) (*fraud.HandHistory, error)
	
	// Get hands for a player
	GetPlayerHands(ctx context.Context, playerID string, startTime, endTime time.Time, limit int) ([]fraud.HandHistory, error)
	
	// Get hands for a table
	GetTableHands(ctx context.Context, tableID string, startTime, endTime time.Time, limit int) ([]fraud.HandHistory, error)
	
	// Get hands between two players
	GetPlayerPairHands(ctx context.Context, playerA, playerB string, startTime, endTime time.Time) ([]fraud.HandHistory, error)
	
	// Get hand statistics
	GetHandStats(ctx context.Context, tableID string, startTime, endTime time.Time) (*HandStats, error)
	
	// Cleanup old hands
	DeleteOldHands(ctx context.Context, before time.Time) (int64, error)
}

// AlertStats contains aggregated alert statistics
type AlertStats struct {
	TotalAlerts       int
	ByType            map[string]int
	BySeverity        map[string]int
	ByAgent           map[string]int
	PendingReview     int
	ConfirmedFraud    int
	Dismissed         int
	AverageResolution time.Duration
	TopRiskPlayers    []RiskPlayerSummary
}

// PlayerStats contains player statistics for fraud detection
type PlayerStats struct {
	PlayerID           string
	HandsPlayed24h     int
	HandsPlayed7d      int
	HandsPlayed30d     int
	WinRate24h         float64
	WinRate7d          float64
	WinRate30d         float64
	TotalChips         int64
	ChipsWon24h        int64
	ChipsWon7d         int64
	ChipsWon30d        int64
	ChipsLost24h       int64
	ChipsLost7d        int64
	ChipsLost30d       int64
	LastPlayed         time.Time
	AlertCount24h      int
	AlertCount7d       int
	AlertCount30d      int
}

// HandStats contains hand statistics
type HandStats struct {
	TotalHands       int
	TotalPotAmount   int64
	AveragePotSize   int64
	MostCommonWinner string
	WinDistribution  map[string]int
}

// RiskPlayerSummary contains risk information for a player
type RiskPlayerSummary struct {
	PlayerID     string
	AgentID      string
	RiskScore    float64
	AlertCount   int
	LastAlertAt  time.Time
}
