package storage

import (
	"context"
	"time"
)

// AnalyticsEventType represents the type of analytics event
type AnalyticsEventType string

const (
	AnalyticsEventHandStarted   AnalyticsEventType = "hand_started"
	AnalyticsEventHandCompleted AnalyticsEventType = "hand_completed"
	AnalyticsEventPlayerAction  AnalyticsEventType = "player_action"
	AnalyticsEventFraudAlert    AnalyticsEventType = "fraud_alert"
	AnalyticsEventRiskScore     AnalyticsEventType = "risk_score"
	AnalyticsEventSessionStart  AnalyticsEventType = "session_start"
	AnalyticsEventSessionEnd    AnalyticsEventType = "session_end"
	AnalyticsEventTableStats    AnalyticsEventType = "table_stats"
)

// HandAnalyticsEvent represents a hand-related analytics event
type HandAnalyticsEvent struct {
	EventID       string             `json:"event_id" ch:"event_id"`
	EventType     AnalyticsEventType `json:"event_type" ch:"event_type"`
	HandID        string             `json:"hand_id" ch:"hand_id"`
	TableID       string             `json:"table_id" ch:"table_id"`
	GameType      string             `json:"game_type" ch:"game_type"`
	BettingType   string             `json:"betting_type" ch:"betting_type"`
	PlayerID      string             `json:"player_id" ch:"player_id"`
	SeatNumber    int                `json:"seat_number" ch:"seat_number"`
	Position      string             `json:"position" ch:"position"`
	ChipsBefore   int64              `json:"chips_before" ch:"chips_before"`
	ChipsAfter    int64              `json:"chips_after" ch:"chips_after"`
	TotalPot      int64              `json:"total_pot" ch:"total_pot"`
	RakeAmount    int64              `json:"rake_amount" ch:"rake_amount"`
	ActionType    string             `json:"action_type" ch:"action_type"`
	ActionAmount  int64              `json:"action_amount" ch:"action_amount"`
	ActionTime    time.Duration      `json:"action_time" ch:"action_time"`
	Timestamp     time.Time          `json:"timestamp" ch:"timestamp"`
	SessionID     string             `json:"session_id" ch:"session_id"`
	AgentID       string             `json:"agent_id" ch:"agent_id"`
	ClubID        string             `json:"club_id" ch:"club_id"`
	DurationMS    int64              `json:"duration_ms" ch:"duration_ms"`
	NumPlayers    int                `json:"num_players" ch:"num_players"`
	StreetReached string             `json:"street_reached" ch:"street_reached"`
}

// FraudAnalyticsEvent represents a fraud detection analytics event
type FraudAnalyticsEvent struct {
	EventID        string             `json:"event_id" ch:"event_id"`
	EventType      AnalyticsEventType `json:"event_type" ch:"event_type"`
	AlertID        string             `json:"alert_id" ch:"alert_id"`
	AlertType      string             `json:"alert_type" ch:"alert_type"`
	Severity       string             `json:"severity" ch:"severity"`
	PlayerID       string             `json:"player_id" ch:"player_id"`
	TableID        string             `json:"table_id" ch:"table_id"`
	HandID         string             `json:"hand_id" ch:"hand_id"`
	DetectionType  string             `json:"detection_type" ch:"detection_type"`
	RiskScore      float64            `json:"risk_score" ch:"risk_score"`
	SignalStrength float64            `json:"signal_strength" ch:"signal_strength"`
	Details        string             `json:"details" ch:"details"`
	Timestamp      time.Time          `json:"timestamp" ch:"timestamp"`
	AgentID        string             `json:"agent_id" ch:"agent_id"`
	ClubID         string             `json:"club_id" ch:"club_id"`
	Resolved       bool               `json:"resolved" ch:"resolved"`
	ResolutionTime *time.Time         `json:"resolution_time" ch:"resolution_time"`
}

// SessionAnalyticsEvent represents a session analytics event
type SessionAnalyticsEvent struct {
	EventID        string             `json:"event_id" ch:"event_id"`
	EventType      AnalyticsEventType `json:"event_type" ch:"event_type"`
	SessionID      string             `json:"session_id" ch:"session_id"`
	PlayerID       string             `json:"player_id" ch:"player_id"`
	AgentID        string             `json:"agent_id" ch:"agent_id"`
	ClubID         string             `json:"club_id" ch:"club_id"`
	TableID        string             `json:"table_id" ch:"table_id"`
	DeviceID       string             `json:"device_id" ch:"device_id"`
	IPAddress      string             `json:"ip_address" ch:"ip_address"`
	Country        string             `json:"country" ch:"country"`
	Platform       string             `json:"platform" ch:"platform"`
	ChipsDeposited int64              `json:"chips_deposited" ch:"chips_deposited"`
	ChipsWithdrawn int64              `json:"chips_withdrawn" ch:"chips_withdrawn"`
	NetProfit      int64              `json:"net_profit" ch:"net_profit"`
	HandsPlayed    int                `json:"hands_played" ch:"hands_played"`
	Duration       time.Duration      `json:"duration" ch:"duration"`
	Timestamp      time.Time          `json:"timestamp" ch:"timestamp"`
}

// TableAnalyticsEvent represents table statistics analytics
type TableAnalyticsEvent struct {
	EventID          string             `json:"event_id" ch:"event_id"`
	EventType        AnalyticsEventType `json:"event_type" ch:"event_type"`
	TableID          string             `json:"table_id" ch:"table_id"`
	GameType         string             `json:"game_type" ch:"game_type"`
	BettingType      string             `json:"betting_type" ch:"betting_type"`
	StakeLevel       string             `json:"stake_level" ch:"stake_level"`
	AgentID          string             `json:"agent_id" ch:"agent_id"`
	ClubID           string             `json:"club_id" ch:"club_id"`
	AvgPotSize       int64              `json:"avg_pot_size" ch:"avg_pot_size"`
	AvgHandsPerHour  float64            `json:"avg_hands_per_hour" ch:"avg_hands_per_hour"`
	AvgPlayersActive float64            `json:"avg_players_active" ch:"avg_players_active"`
	TotalRake        int64              `json:"total_rake" ch:"total_rake"`
	Timestamp        time.Time          `json:"timestamp" ch:"timestamp"`
	PeriodStart      time.Time          `json:"period_start" ch:"period_start"`
	PeriodEnd        time.Time          `json:"period_end" ch:"period_end"`
}

// AnalyticsRepository defines the interface for analytics storage
type AnalyticsRepository interface {
	// Hand Analytics
	RecordHandEvent(ctx context.Context, event *HandAnalyticsEvent) error
	RecordHandEvents(ctx context.Context, events []*HandAnalyticsEvent) error
	GetHandAnalytics(ctx context.Context, query HandAnalyticsQuery) ([]HandAnalyticsEvent, error)

	// Fraud Analytics
	RecordFraudEvent(ctx context.Context, event *FraudAnalyticsEvent) error
	RecordFraudEvents(ctx context.Context, events []*FraudAnalyticsEvent) error
	GetFraudAnalytics(ctx context.Context, query FraudAnalyticsQuery) ([]FraudAnalyticsEvent, error)
	GetFraudTrend(ctx context.Context, query FraudTrendQuery) ([]FraudTrendPoint, error)

	// Session Analytics
	RecordSessionEvent(ctx context.Context, event *SessionAnalyticsEvent) error
	GetSessionAnalytics(ctx context.Context, query SessionAnalyticsQuery) ([]SessionAnalyticsEvent, error)
	GetPlayerStats(ctx context.Context, playerID string, period time.Duration) (*PlayerAnalyticsStats, error)

	// Table Analytics
	RecordTableStats(ctx context.Context, event *TableAnalyticsEvent) error
	GetTableAnalytics(ctx context.Context, query TableAnalyticsQuery) ([]TableAnalyticsEvent, error)

	// Aggregation Queries
	GetRevenueStats(ctx context.Context, query RevenueQuery) (*RevenueStats, error)
	GetPlayerActivityStats(ctx context.Context, query ActivityQuery) ([]PlayerActivityStat, error)

	// Connection management
	Close() error
	Ping(ctx context.Context) error
}

// HandAnalyticsQuery represents a query for hand analytics
type HandAnalyticsQuery struct {
	AgentID   string
	ClubID    string
	TableID   string
	PlayerID  string
	GameType  string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}

// FraudAnalyticsQuery represents a query for fraud analytics
type FraudAnalyticsQuery struct {
	AgentID       string
	ClubID        string
	PlayerID      string
	AlertType     string
	Severity      string
	DetectionType string
	Resolved      *bool
	StartTime     time.Time
	EndTime       time.Time
	Limit         int
	Offset        int
}

// FraudTrendQuery represents a query for fraud trend analysis
type FraudTrendQuery struct {
	AgentID   string
	ClubID    string
	GroupBy   time.Duration // hour, day, week, month
	StartTime time.Time
	EndTime   time.Time
}

// FraudTrendPoint represents a single point in fraud trend data
type FraudTrendPoint struct {
	TimeBucket     time.Time `json:"time_bucket" ch:"time_bucket"`
	TotalAlerts    int       `json:"total_alerts" ch:"total_alerts"`
	HighSeverity   int       `json:"high_severity" ch:"high_severity"`
	MediumSeverity int       `json:"medium_severity" ch:"medium_severity"`
	LowSeverity    int       `json:"low_severity" ch:"low_severity"`
	ResolvedCount  int       `json:"resolved_count" ch:"resolved_count"`
	AvgRiskScore   float64   `json:"avg_risk_score" ch:"avg_risk_score"`
}

// SessionAnalyticsQuery represents a query for session analytics
type SessionAnalyticsQuery struct {
	AgentID   string
	ClubID    string
	PlayerID  string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}

// PlayerAnalyticsStats represents aggregated player statistics for analytics
type PlayerAnalyticsStats struct {
	PlayerID           string        `json:"player_id"`
	TotalHandsPlayed   int           `json:"total_hands_played"`
	TotalProfit        int64         `json:"total_profit"`
	TotalRakePaid      int64         `json:"total_rake_paid"`
	AvgSessionDuration time.Duration `json:"avg_session_duration"`
	WinRate            float64       `json:"win_rate"`
	AvgPotSize         int64         `json:"avg_pot_size"`
	LastActive         time.Time     `json:"last_active"`
	FirstSeen          time.Time     `json:"first_seen"`
}

// TableAnalyticsQuery represents a query for table analytics
type TableAnalyticsQuery struct {
	AgentID   string
	ClubID    string
	TableID   string
	GameType  string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}

// RevenueQuery represents a query for revenue statistics
type RevenueQuery struct {
	AgentID   string
	ClubID    string
	StartTime time.Time
	EndTime   time.Time
	GroupBy   time.Duration
}

// RevenueStats represents revenue statistics
type RevenueStats struct {
	TotalRake        int64         `json:"total_rake"`
	TotalDeposits    int64         `json:"total_deposits"`
	TotalWithdrawals int64         `json:"total_withdrawals"`
	NetRevenue       int64         `json:"net_revenue"`
	PeriodStart      time.Time     `json:"period_start"`
	PeriodEnd        time.Time     `json:"period_end"`
	BreakdownByClub  []ClubRevenue `json:"breakdown_by_club"`
}

// ClubRevenue represents revenue breakdown by club
type ClubRevenue struct {
	ClubID    string `json:"club_id"`
	ClubName  string `json:"club_name"`
	TotalRake int64  `json:"total_rake"`
}

// ActivityQuery represents a query for player activity statistics
type ActivityQuery struct {
	AgentID   string
	ClubID    string
	StartTime time.Time
	EndTime   time.Time
	OrderBy   string
	Limit     int
}

// PlayerActivityStat represents player activity statistics
type PlayerActivityStat struct {
	PlayerID       string        `json:"player_id"`
	HandsPlayed    int           `json:"hands_played"`
	TotalProfit    int64         `json:"total_profit"`
	AvgSessionTime time.Duration `json:"avg_session_time"`
	LastActive     time.Time     `json:"last_active"`
}
