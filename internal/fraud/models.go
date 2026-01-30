package fraud

import (
	"encoding/json"
	"time"
)

// PlayerAction represents a player's action in a poker hand for ML training
type PlayerAction struct {
	ID            string    `json:"id"`
	PlayerID      string    `json:"player_id"`
	TableID       string    `json:"table_id"`
	HandID        string    `json:"hand_id"`
	AgentID       string    `json:"agent_id"`
	ClubID        string    `json:"club_id"`
	ActionType    string    `json:"action_type"` // "bet", "fold", "raise", "check", "call", "all_in"
	Amount        int64     `json:"amount,omitempty"`
	Position      int       `json:"position"` // 0-8 seat position
	Timestamp     time.Time `json:"timestamp"`
	DecisionTime  int       `json:"decision_time_ms"` // Time since last action
	HandPhase     string    `json:"hand_phase"`       // "preflop", "flop", "turn", "river"
	PotSize       int64     `json:"pot_size"`
	StackSize     int64     `json:"stack_size"`
	Cards         []string  `json:"cards,omitempty"` // Only visible cards
	IPAddress     string    `json:"ip_address,omitempty"`
	DeviceID      string    `json:"device_id,omitempty"`
	SessionID     string    `json:"session_id,omitempty"`
}

// PlayerBehavioralFeatures represents extracted features for bot detection
type PlayerBehavioralFeatures struct {
	PlayerID           string    `json:"player_id"`
	TimeRange          string    `json:"time_range"` // "1h", "24h", "7d", "30d"
	ExtractedAt        time.Time `json:"extracted_at"`

	// Timing Features
	AvgActionTime      float64   `json:"avg_action_time"`       // Mean action time in seconds
	ActionTimeStdDev   float64   `json:"action_time_std_dev"`   // Standard deviation of action times
	ActionTimeMin      float64   `json:"action_time_min"`       // Minimum action time
	ActionTimeMax      float64   `json:"action_time_max"`       // Maximum action time

	// Bet Sizing Features
	BetPrecision       float64   `json:"bet_precision"` // Percentage of bets that are round numbers or exact percentages
	AvgBetToPotRatio   float64   `json:"avg_bet_to_pot_ratio"`
	BetSizeVariance    float64   `json:"bet_size_variance"`

	// Volume Features
	HandsPlayed        int       `json:"hands_played"`
	HandsPerHour       float64   `json:"hands_per_hour"`
	TablesConcurrent   int       `json:"tables_concurrent"` // Max concurrent tables

	// Performance Features
	WinRate            float64   `json:"win_rate"`
	WinRateVariance    float64   `json:"win_rate_variance"`
	ShowdownRate       float64   `json:"showdown_rate"` // Percentage of hands reaching showdown
	VPIP               float64   `json:"vpip"`           // Voluntarily Put Money In Pot
	PFR                float64   `json:"pfr"`            // Preflop Raise

	// Error/Strange Behavior
	ErrorRate          float64   `json:"error_rate"` // Mistake rate
	TimeoutRate        float64   `json:"timeout_rate"`

	// Consistency Score (0-1, higher = more consistent/bot-like)
	ConsistencyScore   float64   `json:"consistency_score"`
}

// DeviceFingerprint represents collected device characteristics
type DeviceFingerprint struct {
	PlayerID         string    `json:"player_id"`
	Fingerprint      string    `json:"fingerprint"`
	UserAgent        string    `json:"user_agent,omitempty"`
	ScreenResolution string    `json:"screen_resolution,omitempty"`
	ColorDepth       int       `json:"color_depth,omitempty"`
	Timezone         string    `json:"timezone,omitempty"`
	Language         string    `json:"language,omitempty"`
	Platform         string    `json:"platform,omitempty"`
	HardwareConcurrency int    `json:"hardware_concurrency,omitempty"`
	DeviceMemory     float64   `json:"device_memory,omitempty"`
	TouchSupport     bool      `json:"touch_support"`
	WebGLRenderer    string    `json:"webgl_renderer,omitempty"`
	IPAddress        string    `json:"ip_address,omitempty"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
}

// PlayerSession represents a player's gaming session
type PlayerSession struct {
	SessionID    string    `json:"session_id"`
	PlayerID     string    `json:"player_id"`
	TableID      string    `json:"table_id"`
	AgentID      string    `json:"agent_id"`
	ClubID       string    `json:"club_id"`
	ConnectedAt  time.Time `json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`
	Duration     time.Duration `json:"duration"`
	TotalHands   int       `json:"total_hands"`
	TotalWins    int       `json:"total_wins"`
	TotalLosses  int       `json:"total_losses"`
	TotalChips   int64     `json:"total_chips"`
	StartingChips int64    `json:"starting_chips"`
	EndingChips  int64     `json:"ending_chips"`
}

// HandHistory represents a completed poker hand
type HandHistory struct {
	HandID        string    `json:"hand_id"`
	TableID       string    `json:"table_id"`
	AgentID       string    `json:"agent_id"`
	ClubID        string    `json:"club_id"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at"`
	GameType      string    `json:"game_type"` // "texas_hold'em", "omaha"
	BettingType   string    `json:"betting_type"` // "no_limit", "pot_limit", "fixed_limit"
	SmallBlind    int64     `json:"small_blind"`
	BigBlind      int64     `json:"big_blind"`
	PotAmount     int64     `json:"pot_amount"`
	RakeAmount    int64     `json:"rake_amount"`
	WinnerIDs     []string  `json:"winner_ids"`
	ActionHistory []ActionRecord `json:"action_history"`
	CommunityCards []string `json:"community_cards"`
	ShowdownCards map[string][]string `json:"showdown_cards"` // player_id -> cards
}

// ActionRecord represents a single action in a hand
type ActionRecord struct {
	PlayerID   string    `json:"player_id"`
	Round      string    `json:"round"` // "preflop", "flop", "turn", "river"
	Action     string    `json:"action"`
	Amount     int64     `json:"amount,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	ElapsedMs  int       `json:"elapsed_ms"` // Time from hand start
}

// PlayerRelationship represents a relationship between two players
type PlayerRelationship struct {
	PlayerA          string    `json:"player_a"`
	PlayerB          string    `json:"player_b"`
	AgentID          string    `json:"agent_id"`
	CoOccurrenceCount int      `json:"co_occurrence_count"` // Number of hands played together
	TotalHandsA      int       `json:"total_hands_a"`
	TotalHandsB      int       `json:"total_hands_b"`
	WinRateA         float64   `json:"win_rate_a"` // Win rate when both playing
	WinRateB         float64   `json:"win_rate_b"`
	MutualWins       int       `json:"mutual_wins"` // Hands where one won and other was in hand
	AvgPotSize       float64   `json:"avg_pot_size"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
	IPMatchCount     int       `json:"ip_match_count"`
	DeviceMatchCount int       `json:"device_match_count"`
}

// AntiCheatAlert represents a generated fraud alert
type AntiCheatAlert struct {
	ID            string          `json:"id"`
	PlayerID      string          `json:"player_id"`
	AlertType     string          `json:"alert_type"` // "bot", "collusion", "multi_account", "chip_dumping"
	Severity      string          `json:"severity"` // "low", "medium", "high", "critical"
	Score         float64         `json:"score"` // Confidence score 0-1
	TableID       string          `json:"table_id,omitempty"`
	HandID        string          `json:"hand_id,omitempty"`
	AgentID       string          `json:"agent_id"`
	ClubID        string          `json:"club_id"`
	Evidence      []string        `json:"evidence"` // List of suspicious behaviors
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	ReviewedAt    *time.Time      `json:"reviewed_at,omitempty"`
	ReviewedBy    string          `json:"reviewed_by,omitempty"`
	Status        string          `json:"status"` // "pending", "reviewed", "dismissed", "confirmed"
	Notes         string          `json:"notes,omitempty"`
}

// RiskScore represents a player's overall risk assessment
type RiskScore struct {
	PlayerID         string    `json:"player_id"`
	AgentID          string    `json:"agent_id"`
	OverallScore     float64   `json:"overall_score"` // 0-1, higher = more risky
	BotScore         float64   `json:"bot_score"`
	CollusionScore   float64   `json:"collusion_score"`
	MultiAccountScore float64  `json:"multi_account_score"`
	ChipDumpScore    float64   `json:"chip_dump_score"`
	LastCalculated   time.Time `json:"last_calculated"`
	CalculatedFrom   time.Time `json:"calculated_from"` // Start of time window
	CalculatedTo     time.Time `json:"calculated_to"`   // End of time window
	FlagCount24h     int       `json:"flag_count_24h"`
	FlagCount7d      int       `json:"flag_count_7d"`
	FlagCount30d     int       `json:"flag_count_30d"`
	ReviewRecommended bool     `json:"review_recommended"`
}

// ChatMessage represents a chat message for collusion detection
type ChatMessage struct {
	ID          string    `json:"id"`
	TableID     string    `json:"table_id"`
	SenderID    string    `json:"sender_id"`
	RecipientID string    `json:"recipient_id,omitempty"` // For private messages
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
	Redacted    bool      `json:"redacted"` // For privacy
}

// ConnectionEvent tracks player connections for analysis
type ConnectionEvent struct {
	ID          string    `json:"id"`
	PlayerID    string    `json:"player_id"`
	TableID     string    `json:"table_id"`
	AgentID     string    `json:"agent_id"`
	EventType   string    `json:"event_type"` // "connect", "disconnect", "reconnect"
	IPAddress   string    `json:"ip_address"`
	DeviceID    string    `json:"device_id"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    time.Duration `json:"duration,omitempty"` // For disconnect events
}
