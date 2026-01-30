package fraud

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RuleBasedDetector implements deterministic anti-cheat rules
type RuleBasedDetector struct {
	rules     []AntiCheatRule
	ruleIndex map[string]int
	mu        sync.RWMutex
}

// AntiCheatRule represents a deterministic anti-cheat rule
type AntiCheatRule struct {
	Name          string
	Description   string
	Category      string // "volume", "timing", "pattern", "identity"
	Severity      string // "low", "medium", "high", "critical"
	Enabled       bool
	Check         func(ctx context.Context, playerID string, data *RuleCheckData) bool
	Action        func(ctx context.Context, alert *AntiCheatAlert)
	Cooldown      time.Duration
	LastTriggered map[string]time.Time
}

// RuleCheckData contains data needed for rule evaluation
type RuleCheckData struct {
	PlayerID            string
	AgentID             string
	ClubID              string
	TableID             string
	HandID              string
	HandsPlayed24h      int
	HandsPlayed7d       int
	WinRate24h          float64
	WinRate7d           float64
	WinRate30d          float64
	AvgActionTime       float64
	AccountsFromIP      map[string]int
	AccountsFromDevice  map[string]int
	CurrentSessionHands int
	SessionStartTime    time.Time
	TotalChipsWon       int64
	TotalChipsLost      int64
	IPAddress           string
	DeviceFingerprint   string
	PlayDuration24h     time.Duration
	IsNewAccount        bool
	AlertCount24h       int
	AlertCount7d        int
}

// NewRuleBasedDetector creates a new rule-based detector with default rules
func NewRuleBasedDetector() *RuleBasedDetector {
	detector := &RuleBasedDetector{
		rules:     make([]AntiCheatRule, 0),
		ruleIndex: make(map[string]int),
		mu:        sync.RWMutex{},
	}

	detector.registerDefaultRules()
	return detector
}

// registerDefaultRules registers the default anti-cheat rules
func (rbd *RuleBasedDetector) registerDefaultRules() {
	rules := []AntiCheatRule{
		// Volume-based rules
		{
			Name:          "excessive_volume_24h",
			Description:   "Player played unrealistic number of hands in 24 hours",
			Category:      "volume",
			Severity:      "high",
			Enabled:       true,
			Cooldown:      1 * time.Hour,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// 500+ hands in 24 hours is unrealistic for a human
				return data.HandsPlayed24h > 500
			},
		},
		{
			Name:          "excessive_volume_7d",
			Description:   "Player played unrealistic number of hands in 7 days",
			Category:      "volume",
			Severity:      "medium",
			Enabled:       true,
			Cooldown:      6 * time.Hour,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// 2000+ hands in 7 days (avg ~285/day) is high
				return data.HandsPlayed7d > 2000
			},
		},
		{
			Name:          "marathon_session",
			Description:   "Player session exceeds maximum human endurance",
			Category:      "volume",
			Severity:      "medium",
			Enabled:       true,
			Cooldown:      4 * time.Hour,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// 12+ hour continuous session is suspicious
				return data.PlayDuration24h > 12*time.Hour
			},
		},

		// Win rate rules
		{
			Name:          "perfect_win_rate",
			Description:   "Player has suspiciously perfect win rate",
			Category:      "pattern",
			Severity:      "high",
			Enabled:       true,
			Cooldown:      24 * time.Hour,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// >95% win rate over 50+ hands is extremely suspicious
				if data.HandsPlayed24h >= 50 {
					return data.WinRate24h > 0.95
				}
				return false
			},
		},
		{
			Name:          "sustained_win_rate",
			Description:   "Player maintains suspiciously high win rate over extended period",
			Category:      "pattern",
			Severity:      "medium",
			Enabled:       true,
			Cooldown:      48 * time.Hour,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// >80% win rate over 500+ hands in 7 days
				if data.HandsPlayed7d >= 500 {
					return data.WinRate7d > 0.80
				}
				return false
			},
		},
		{
			Name:          "no_losses_7d",
			Description:   "Player has not lost any chips in 7 days",
			Category:      "pattern",
			Severity:      "high",
			Enabled:       true,
			Cooldown:      24 * time.Hour,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// No losses over 100+ hands
				if data.HandsPlayed7d >= 100 {
					return data.TotalChipsLost == 0 && data.TotalChipsWon > 0
				}
				return false
			},
		},

		// Identity rules
		{
			Name:          "same_ip_multi_account",
			Description:   "Multiple accounts from same IP address",
			Category:      "identity",
			Severity:      "medium",
			Enabled:       true,
			Cooldown:      0,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				count := data.AccountsFromIP[data.IPAddress]
				// 3+ accounts from same IP (allow for family/household)
				return count >= 3
			},
		},
		{
			Name:          "same_device_multi_account",
			Description:   "Multiple accounts from same device",
			Category:      "identity",
			Severity:      "high",
			Enabled:       true,
			Cooldown:      0,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				count := data.AccountsFromDevice[data.DeviceFingerprint]
				// 2+ accounts from same device is very suspicious
				return count >= 2
			},
		},
		{
			Name:          "new_account_suspicious",
			Description:   "New account with suspicious activity patterns",
			Category:      "pattern",
			Severity:      "medium",
			Enabled:       true,
			Cooldown:      2 * time.Hour,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				if !data.IsNewAccount {
					return false
				}
				// New account with high volume and high win rate
				return data.HandsPlayed24h > 100 && data.WinRate24h > 0.70
			},
		},

		// Timing rules
		{
			Name:          "instant_actions",
			Description:   "Player actions consistently too fast for human reaction time",
			Category:      "timing",
			Severity:      "medium",
			Enabled:       true,
			Cooldown:      30 * time.Minute,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// Average action time under 0.5 seconds consistently
				return data.AvgActionTime > 0 && data.AvgActionTime < 0.5
			},
		},

		// Alert fatigue rules
		{
			Name:          "alert_fatigue",
			Description:   "Player has triggered too many alerts in short period",
			Category:      "pattern",
			Severity:      "low",
			Enabled:       true,
			Cooldown:      0,
			LastTriggered: make(map[string]time.Time),
			Check: func(ctx context.Context, playerID string, data *RuleCheckData) bool {
				// 10+ alerts in 24 hours
				return data.AlertCount24h >= 10
			},
		},
	}

	for i := range rules {
		rbd.rules = append(rbd.rules, rules[i])
		rbd.ruleIndex[rules[i].Name] = i
	}
}

// EvaluateRules runs all enabled rules and returns triggered alerts
func (rbd *RuleBasedDetector) EvaluateRules(ctx context.Context, data *RuleCheckData) []*AntiCheatAlert {
	rbd.mu.RLock()
	defer rbd.mu.RUnlock()

	alerts := make([]*AntiCheatAlert, 0)

	for i, rule := range rbd.rules {
		if !rule.Enabled {
			continue
		}

		// Check cooldown
		if rule.Cooldown > 0 {
			lastTriggered, exists := rule.LastTriggered[data.PlayerID]
			if exists && time.Since(lastTriggered) < rule.Cooldown {
				continue
			}
		}

		// Evaluate rule
		if rule.Check(ctx, data.PlayerID, data) {
			// Create alert
			alert := &AntiCheatAlert{
				ID:        fmt.Sprintf("alert_%s_%d", rule.Name, time.Now().UnixNano()),
				PlayerID:  data.PlayerID,
				AlertType: rbd.categoryToAlertType(rule.Category),
				Severity:  rule.Severity,
				Score:     rbd.severityToScore(rule.Severity),
				TableID:   data.TableID,
				HandID:    data.HandID,
				AgentID:   data.AgentID,
				ClubID:    data.ClubID,
				Evidence:  []string{fmt.Sprintf("Rule '%s' triggered: %s", rule.Name, rule.Description)},
				CreatedAt: time.Now(),
				Status:    "pending",
			}

			alerts = append(alerts, alert)

			// Update last triggered
			rbd.rules[i].LastTriggered[data.PlayerID] = time.Now()

			// Execute rule action if defined
			if rule.Action != nil {
				rule.Action(ctx, alert)
			}
		}
	}

	return alerts
}

// categoryToAlertType maps rule category to alert type
func (rbd *RuleBasedDetector) categoryToAlertType(category string) string {
	switch category {
	case "volume":
		return "bot"
	case "timing":
		return "bot"
	case "pattern":
		return "bot"
	case "identity":
		return "multi_account"
	default:
		return "fraud"
	}
}

// severityToScore converts severity to numeric score
func (rbd *RuleBasedDetector) severityToScore(severity string) float64 {
	switch severity {
	case "critical":
		return 1.0
	case "high":
		return 0.8
	case "medium":
		return 0.5
	case "low":
		return 0.25
	default:
		return 0.0
	}
}

// GetRule returns a rule by name
func (rbd *RuleBasedDetector) GetRule(name string) *AntiCheatRule {
	rbd.mu.RLock()
	defer rbd.mu.RUnlock()

	if idx, exists := rbd.ruleIndex[name]; exists {
		return &rbd.rules[idx]
	}
	return nil
}

// EnableRule enables a rule by name
func (rbd *RuleBasedDetector) EnableRule(name string) error {
	rbd.mu.Lock()
	defer rbd.mu.Unlock()

	if idx, exists := rbd.ruleIndex[name]; exists {
		rbd.rules[idx].Enabled = true
		return nil
	}
	return fmt.Errorf("rule not found: %s", name)
}

// DisableRule disables a rule by name
func (rbd *RuleBasedDetector) DisableRule(name string) error {
	rbd.mu.Lock()
	defer rbd.mu.Unlock()

	if idx, exists := rbd.ruleIndex[name]; exists {
		rbd.rules[idx].Enabled = false
		return nil
	}
	return fmt.Errorf("rule not found: %s", name)
}

// GetAllRules returns all registered rules
func (rbd *RuleBasedDetector) GetAllRules() []AntiCheatRule {
	rbd.mu.RLock()
	defer rbd.mu.RUnlock()

	result := make([]AntiCheatRule, len(rbd.rules))
	copy(result, rbd.rules)
	return result
}

// RuleEngine manages and executes anti-cheat rules
type RuleEngine struct {
	detector     *RuleBasedDetector
	alertStorage AlertStorage
	mu           sync.RWMutex
}

// AlertStorage defines the interface for storing and retrieving alerts
type AlertStorage interface {
	SaveAlert(alert *AntiCheatAlert) error
	GetPlayerAlerts(playerID string, limit int) ([]*AntiCheatAlert, error)
	UpdateAlertStatus(alertID, status, reviewerID, notes string) error
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine(detector *RuleBasedDetector, alertStorage AlertStorage) *RuleEngine {
	return &RuleEngine{
		detector:     detector,
		alertStorage: alertStorage,
	}
}

// ProcessPlayerAction processes a player action through all rules
func (re *RuleEngine) ProcessPlayerAction(ctx context.Context, action *PlayerAction, stats *PlayerStats) []*AntiCheatAlert {
	data := &RuleCheckData{
		PlayerID:          action.PlayerID,
		AgentID:           action.AgentID,
		ClubID:            action.ClubID,
		TableID:           action.TableID,
		HandID:            action.HandID,
		HandsPlayed24h:    stats.HandsPlayed24h,
		HandsPlayed7d:     stats.HandsPlayed7d,
		WinRate24h:        stats.WinRate24h,
		WinRate7d:         stats.WinRate7d,
		AvgActionTime:     stats.AvgActionTime,
		IPAddress:         action.IPAddress,
		DeviceFingerprint: action.DeviceID,
		PlayDuration24h:   stats.PlayDuration24h,
		IsNewAccount:      stats.IsNewAccount,
		AlertCount24h:     stats.AlertCount24h,
	}

	// Add account counts from maps
	data.AccountsFromIP = make(map[string]int)
	data.AccountsFromIP[action.IPAddress] = 2 // Example

	data.AccountsFromDevice = make(map[string]int)
	data.AccountsFromDevice[action.DeviceID] = 1 // Example

	// Evaluate rules
	alerts := re.detector.EvaluateRules(ctx, data)

	// Save alerts
	for _, alert := range alerts {
		re.alertStorage.SaveAlert(alert)
	}

	return alerts
}

// PlayerStats contains player statistics for rule evaluation
type PlayerStats struct {
	HandsPlayed24h  int
	HandsPlayed7d   int
	WinRate24h      float64
	WinRate7d       float64
	AvgActionTime   float64
	PlayDuration24h time.Duration
	IsNewAccount    bool
	AlertCount24h   int
	TotalChipsWon   int64
	TotalChipsLost  int64
}
