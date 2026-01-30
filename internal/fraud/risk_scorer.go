package fraud

import (
	"context"
	"sync"
	"time"
)

// RiskScorer calculates overall risk scores for players
type RiskScorer struct {
	config               *RiskScoringConfig
	botDetector          *BotDetector
	collusionDetector    *CollusionDetector
	multiAccountDetector *MultiAccountDetector
	ruleEngine           *RuleEngine
	alertStorage         AlertStorage
	playerRiskCache      *RiskScoreCache
	mu                   sync.RWMutex
}

// RiskScoringConfig holds configuration for risk scoring
type RiskScoringConfig struct {
	// Score weights
	BotScoreWeight          float64
	CollusionScoreWeight    float64
	MultiAccountScoreWeight float64
	RuleViolationWeight     float64
	AlertHistoryWeight      float64

	// Time windows
	RecentAlertWindow     time.Duration
	HistoricalAlertWindow time.Duration

	// Score thresholds
	ReviewThreshold   float64
	FlagThreshold     float64
	CriticalThreshold float64

	// Cache settings
	CacheTTL             time.Duration
	CacheRefreshInterval time.Duration
}

// DefaultRiskScoringConfig returns default configuration
func DefaultRiskScoringConfig() *RiskScoringConfig {
	return &RiskScoringConfig{
		BotScoreWeight:          0.30,
		CollusionScoreWeight:    0.25,
		MultiAccountScoreWeight: 0.20,
		RuleViolationWeight:     0.15,
		AlertHistoryWeight:      0.10,

		RecentAlertWindow:     24 * time.Hour,
		HistoricalAlertWindow: 30 * 24 * time.Hour,

		ReviewThreshold:   0.50,
		FlagThreshold:     0.75,
		CriticalThreshold: 0.90,

		CacheTTL:             5 * time.Minute,
		CacheRefreshInterval: 1 * time.Minute,
	}
}

// RiskScoreCache caches player risk scores
type RiskScoreCache struct {
	scores map[string]*CachedRiskScore
	mu     sync.RWMutex
}

// CachedRiskScore represents a cached risk score
type CachedRiskScore struct {
	Score      *RiskScore
	ComputedAt time.Time
	ExpiresAt  time.Time
}

// NewRiskScorer creates a new risk scorer
func NewRiskScorer(
	config *RiskScoringConfig,
	botDetector *BotDetector,
	collusionDetector *CollusionDetector,
	multiAccountDetector *MultiAccountDetector,
	ruleEngine *RuleEngine,
	alertStorage AlertStorage,
) *RiskScorer {
	if config == nil {
		config = DefaultRiskScoringConfig()
	}

	return &RiskScorer{
		config:               config,
		botDetector:          botDetector,
		collusionDetector:    collusionDetector,
		multiAccountDetector: multiAccountDetector,
		ruleEngine:           ruleEngine,
		alertStorage:         alertStorage,
		playerRiskCache: &RiskScoreCache{
			scores: make(map[string]*CachedRiskScore),
		},
	}
}

// CalculateRiskScore calculates the overall risk score for a player
func (rs *RiskScorer) CalculateRiskScore(ctx context.Context, playerID, agentID string) (*RiskScore, error) {
	// Check cache first
	cached := rs.getCachedScore(playerID, agentID)
	if cached != nil {
		return cached, nil
	}

	// Calculate individual risk components
	botScore := rs.calculateBotRisk(ctx, playerID)
	collusionScore := rs.calculateCollusionRisk(ctx, playerID)
	multiAccountScore := rs.calculateMultiAccountRisk(ctx, playerID)
	ruleViolations := rs.calculateRuleViolationRisk(ctx, playerID)
	alertHistory := rs.calculateAlertHistoryRisk(ctx, playerID, agentID)

	// Calculate weighted overall score
	overallScore := botScore*rs.config.BotScoreWeight +
		collusionScore*rs.config.CollusionScoreWeight +
		multiAccountScore*rs.config.MultiAccountScoreWeight +
		ruleViolations*rs.config.RuleViolationWeight +
		alertHistory*rs.config.AlertHistoryWeight

	now := time.Now()
	riskScore := &RiskScore{
		PlayerID:          playerID,
		AgentID:           agentID,
		OverallScore:      overallScore,
		BotScore:          botScore,
		CollusionScore:    collusionScore,
		MultiAccountScore: multiAccountScore,
		ChipDumpScore:     ruleViolations, // Reuse rule violation for chip dumping
		LastCalculated:    now,
		CalculatedFrom:    now.Add(-rs.config.HistoricalAlertWindow),
		CalculatedTo:      now,
		FlagCount24h:      rs.countRecentAlerts(playerID, 24*time.Hour),
		FlagCount7d:       rs.countRecentAlerts(playerID, 7*24*time.Hour),
		FlagCount30d:      rs.countRecentAlerts(playerID, 30*24*time.Hour),
		ReviewRecommended: overallScore >= rs.config.ReviewThreshold,
	}

	// Cache the score
	rs.cacheScore(playerID, agentID, riskScore)

	return riskScore, nil
}

// calculateBotRisk calculates bot-related risk score
func (rs *RiskScorer) calculateBotRisk(ctx context.Context, playerID string) float64 {
	// In production, this would fetch real behavioral features
	features := &PlayerBehavioralFeatures{
		PlayerID: playerID,
	}

	result := rs.botDetector.DetectBot(ctx, features)
	return result.Score
}

// calculateCollusionRisk calculates collusion-related risk score
func (rs *RiskScorer) calculateCollusionRisk(ctx context.Context, playerID string) float64 {
	// Check if player is part of any detected collusion network
	networks := rs.collusionDetector.FindCollusionRings(ctx, 0.5)

	for _, network := range networks {
		for _, member := range network.Members {
			if member == playerID {
				return network.Confidence
			}
		}
	}

	return 0.0
}

// calculateMultiAccountRisk calculates multi-account risk score
func (rs *RiskScorer) calculateMultiAccountRisk(ctx context.Context, playerID string) float64 {
	result := rs.multiAccountDetector.DetectMultiAccount(ctx, playerID)
	return result.Score
}

// calculateRuleViolationRisk calculates risk from rule violations
func (rs *RiskScorer) calculateRuleViolationRisk(ctx context.Context, playerID string) float64 {
	alerts, _ := rs.alertStorage.GetPlayerAlerts(playerID, 100)

	recentViolations := 0
	for _, alert := range alerts {
		if alert.AlertType == "bot" && alert.CreatedAt.After(time.Now().Add(-24*time.Hour)) {
			recentViolations++
		}
	}

	// Normalize to 0-1 scale (10+ violations = max score)
	return float64(recentViolations) / 10.0
}

// calculateAlertHistoryRisk calculates risk from alert history
func (rs *RiskScorer) calculateAlertHistoryRisk(ctx context.Context, playerID, agentID string) float64 {
	recentAlerts := rs.countRecentAlerts(playerID, rs.config.RecentAlertWindow)
	historicalAlerts := rs.countRecentAlerts(playerID, rs.config.HistoricalAlertWindow)

	// Weight recent alerts more heavily
	score := float64(recentAlerts)*0.6 + float64(historicalAlerts)*0.4

	// Normalize (20+ alerts = max score)
	return score / 20.0
}

// countRecentAlerts counts alerts within the given duration
func (rs *RiskScorer) countRecentAlerts(playerID string, duration time.Duration) int {
	cutoff := time.Now().Add(-duration)
	alerts, _ := rs.alertStorage.GetPlayerAlerts(playerID, 1000)

	count := 0
	for _, alert := range alerts {
		if alert.CreatedAt.After(cutoff) {
			count++
		}
	}
	return count
}

// getCachedScore returns cached score if valid
func (rs *RiskScorer) getCachedScore(playerID, agentID string) *RiskScore {
	rs.playerRiskCache.mu.RLock()
	defer rs.playerRiskCache.mu.RUnlock()

	key := playerID + ":" + agentID
	if cached, exists := rs.playerRiskCache.scores[key]; exists {
		if time.Now().Before(cached.ExpiresAt) {
			return cached.Score
		}
	}
	return nil
}

// cacheScore stores a score in the cache
func (rs *RiskScorer) cacheScore(playerID, agentID string, score *RiskScore) {
	rs.playerRiskCache.mu.Lock()
	defer rs.playerRiskCache.mu.Unlock()

	key := playerID + ":" + agentID
	rs.playerRiskCache.scores[key] = &CachedRiskScore{
		Score:      score,
		ComputedAt: time.Now(),
		ExpiresAt:  time.Now().Add(rs.config.CacheTTL),
	}
}

// AlertService manages the alert lifecycle
type AlertService struct {
	storage             AlertStorage
	riskScorer          *RiskScorer
	notificationService NotificationService
	mu                  sync.RWMutex
}

// NotificationService handles alert notifications
type NotificationService interface {
	SendHighRiskAlert(alert *AntiCheatAlert) error
	SendReviewRequest(alert *AntiCheatAlert, reviewers []string) error
}

// NewAlertService creates a new alert service
func NewAlertService(
	storage AlertStorage,
	riskScorer *RiskScorer,
	notificationService NotificationService,
) *AlertService {
	return &AlertService{
		storage:             storage,
		riskScorer:          riskScorer,
		notificationService: notificationService,
	}
}

// CreateAlert creates a new alert and triggers appropriate actions
func (as *AlertService) CreateAlert(ctx context.Context, alert *AntiCheatAlert) error {
	// Save alert
	if err := as.storage.SaveAlert(alert); err != nil {
		return err
	}

	// Trigger notifications for high-severity alerts
	if alert.Severity == "high" || alert.Severity == "critical" {
		return as.notificationService.SendHighRiskAlert(alert)
	}

	return nil
}

// ReviewAlert marks an alert as reviewed
func (as *AlertService) ReviewAlert(alertID, reviewerID, notes, status string) error {
	return as.storage.UpdateAlertStatus(alertID, status, reviewerID, notes)
}

// GetPendingAlerts returns all pending alerts
func (as *AlertService) GetPendingAlerts(limit int) ([]*AntiCheatAlert, error) {
	// In production, this would query by status
	return as.storage.GetPlayerAlerts("", limit)
}

// AlertAggregator aggregates and summarizes alerts
type AlertAggregator struct {
	storage AlertStorage
	mu      sync.RWMutex
}

// NewAlertAggregator creates a new alert aggregator
func NewAlertAggregator(storage AlertStorage) *AlertAggregator {
	return &AlertAggregator{storage: storage}
}

// AlertSummary contains aggregated alert statistics
type AlertSummary struct {
	TotalAlerts       int
	BySeverity        map[string]int
	ByType            map[string]int
	ByAgent           map[string]int
	PendingReview     int
	ConfirmedFraud    int
	Dismissed         int
	AverageResolution time.Duration
	TopRiskPlayers    []RiskPlayerSummary
}

// RiskPlayerSummary contains risk information for a player
type RiskPlayerSummary struct {
	PlayerID    string
	AgentID     string
	RiskScore   float64
	AlertCount  int
	LastAlertAt time.Time
}

// AggregateSummary generates an alert summary for a time period
func (aa *AlertAggregator) AggregateSummary(ctx context.Context, startTime, endTime time.Time) (*AlertSummary, error) {
	// In production, this would query and aggregate from storage
	summary := &AlertSummary{
		BySeverity: make(map[string]int),
		ByType:     make(map[string]int),
		ByAgent:    make(map[string]int),
	}

	// Placeholder implementation
	summary.TotalAlerts = 0
	summary.PendingReview = 0

	return summary, nil
}

// GetHighRiskPlayers returns players with elevated risk scores
func (aa *AlertAggregator) GetHighRiskPlayers(ctx context.Context, minScore float64, limit int) ([]RiskPlayerSummary, error) {
	// In production, this would query and rank players by risk
	return []RiskPlayerSummary{}, nil
}
