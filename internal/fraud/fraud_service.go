package fraud

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// FraudService provides real-time fraud detection integration with the game server
type FraudService struct {
	config              *FraudServiceConfig
	botDetector         *BotDetector
	collusionDetector   *CollusionDetector
	multiAccountDetector *MultiAccountDetector
	ruleEngine         *RuleEngine
	riskScorer         *RiskScorer
	alertService       *AlertService
	eventProcessor     *EventProcessor
	metrics            *FraudMetrics
	mu                 sync.RWMutex
}

// FraudServiceConfig holds configuration for the fraud service
type FraudServiceConfig struct {
	// Event processing settings
	ActionBufferSize    int
	ProcessingInterval time.Duration
	BatchSize          int

	// Detection settings
	EnableRealTimeBotDetection    bool
	EnableRealTimeCollusionDetection bool
	EnableRealTimeMultiAccountDetection bool
	EnableRuleEngine              bool

	// Alert thresholds
	HighRiskThreshold      float64
	CriticalRiskThreshold  float64
	AlertCooldown         time.Duration

	// Metrics settings
	MetricsEnabled        bool
	MetricsInterval       time.Duration
}

// DefaultFraudServiceConfig returns default configuration
func DefaultFraudServiceConfig() *FraudServiceConfig {
	return &FraudServiceConfig{
		ActionBufferSize:      1000,
		ProcessingInterval:    100 * time.Millisecond,
		BatchSize:            50,

		EnableRealTimeBotDetection:     true,
		EnableRealTimeCollusionDetection: true,
		EnableRealTimeMultiAccountDetection: true,
		EnableRuleEngine:               true,

		HighRiskThreshold:     0.75,
		CriticalRiskThreshold: 0.90,
		AlertCooldown:        5 * time.Minute,

		MetricsEnabled:        true,
		MetricsInterval:       1 * time.Minute,
	}
}

// FraudMetrics tracks fraud detection metrics
type FraudMetrics struct {
	TotalEventsProcessed   int64
	BotAlertsGenerated    int64
	CollusionAlertsGenerated int64
	MultiAccountAlertsGenerated int64
	RuleAlertsGenerated   int64
	HighRiskPlayers       int64
	CriticalRiskPlayers   int64
	LastProcessedAt       time.Time
	mu                   sync.RWMutex
}

// NewFraudService creates a new fraud service
func NewFraudService(
	config *FraudServiceConfig,
	botDetector *BotDetector,
	collusionDetector *CollusionDetector,
	multiAccountDetector *MultiAccountDetector,
	ruleEngine *RuleEngine,
	riskScorer *RiskScorer,
	alertService *AlertService,
) *FraudService {
	if config == nil {
		config = DefaultFraudServiceConfig()
	}

	return &FraudService{
		config:               config,
		botDetector:          botDetector,
		collusionDetector:    collusionDetector,
		multiAccountDetector: multiAccountDetector,
		ruleEngine:           ruleEngine,
		riskScorer:           riskScorer,
		alertService:         alertService,
		eventProcessor:       NewEventProcessor(config),
		metrics:             &FraudMetrics{},
	}
}

// ProcessPlayerAction processes a player action for fraud detection
func (fs *FraudService) ProcessPlayerAction(ctx context.Context, action *PlayerAction) (*FraudDetectionResult, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	fs.incrementCounter("TotalEventsProcessed")

	result := &FraudDetectionResult{
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
	}

	// Process through each detection system
	var wg sync.WaitGroup
	var botResult *BotDetectionResult
	var collusionResults []*CollusionDetectionResult
	var multiAccountResult *MultiAccountDetectionResult
	var ruleAlerts []*AntiCheatAlert
	var riskScore *RiskScore

	// Bot detection
	if fs.config.EnableRealTimeBotDetection {
		wg.Add(1)
		go func() {
			defer wg.Done()
			features := fs.extractFeatures(action)
			botResult = fs.botDetector.DetectBot(ctx, features)
		}()
	}

	// Collusion detection
	if fs.config.EnableRealTimeCollusionDetection {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collusionResults = fs.detectCollusionForAction(ctx, action)
		}()
	}

	// Multi-account detection
	if fs.config.EnableRealTimeMultiAccountDetection {
		wg.Add(1)
		go func() {
			defer wg.Done()
			multiAccountResult = fs.multiAccountDetector.DetectMultiAccount(ctx, action.PlayerID)
		}()
	}

	// Rule engine
	if fs.config.EnableRuleEngine {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ruleAlerts = fs.ruleEngine.ProcessPlayerAction(ctx, action, fs.getPlayerStats(action.PlayerID))
		}()
	}

	// Risk scoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		riskScore, _ = fs.riskScorer.CalculateRiskScore(ctx, action.PlayerID, action.AgentID)
	}()

	wg.Wait()

	// Aggregate results
	result.BotDetection = botResult
	result.CollusionDetection = collusionResults
	result.MultiAccountDetection = multiAccountResult
	result.RuleAlerts = ruleAlerts
	result.RiskScore = riskScore

	// Determine overall verdict
	result.RequiresAction = fs.determineActionRequired(result)
	result.RecommendedActions = fs.generateRecommendedActions(result)

	// Generate alerts if needed
	if result.RequiresAction {
		fs.generateAlerts(ctx, result)
	}

	return result, nil
}

// FraudDetectionResult contains the result of fraud detection
type FraudDetectionResult struct {
	PlayerID              string
	Timestamp             time.Time
	RequiresAction        bool
	RecommendedActions    []string
	BotDetection          *BotDetectionResult
	CollusionDetection    []*CollusionDetectionResult
	MultiAccountDetection *MultiAccountDetectionResult
	RuleAlerts            []*AntiCheatAlert
	RiskScore             *RiskScore
}

// extractFeatures extracts behavioral features from an action
func (fs *FraudService) extractFeatures(action *PlayerAction) *PlayerBehavioralFeatures {
	return &PlayerBehavioralFeatures{
		PlayerID:        action.PlayerID,
		TimeRange:       "realtime",
		ExtractedAt:     time.Now(),
		AvgActionTime:   float64(action.DecisionTime) / 1000.0,
		HandsPlayed:     1,
		ConsistencyScore: 0.5, // Would be calculated from history
	}
}

// detectCollusionForAction detects collusion related to an action
func (fs *FraudService) detectCollusionForAction(ctx context.Context, action *PlayerAction) []*CollusionDetectionResult {
	// In production, this would check table relationships
	return []*CollusionDetectionResult{}
}

// getPlayerStats returns player statistics
func (fs *FraudService) getPlayerStats(playerID string) *PlayerStats {
	return &PlayerStats{
		HandsPlayed24h:  0,
		HandsPlayed7d:   0,
		WinRate24h:      0.5,
		WinRate7d:       0.5,
		AvgActionTime:   5.0,
		IsNewAccount:    false,
		AlertCount24h:   0,
	}
}

// determineActionRequired determines if action is required based on results
func (fs *FraudService) determineActionRequired(result *FraudDetectionResult) bool {
	// Check bot detection
	if result.BotDetection != nil && result.BotDetection.Score >= fs.config.HighRiskThreshold {
		return true
	}

	// Check collusion detection
	for _, cr := range result.CollusionDetection {
		if cr.Score >= fs.config.HighRiskThreshold {
			return true
		}
	}

	// Check multi-account detection
	if result.MultiAccountDetection != nil && result.MultiAccountDetection.Score >= fs.config.HighRiskThreshold {
		return true
	}

	// Check rule alerts
	if len(result.RuleAlerts) > 0 {
		for _, alert := range result.RuleAlerts {
			if alert.Severity == "high" || alert.Severity == "critical" {
				return true
			}
		}
	}

	// Check risk score
	if result.RiskScore != nil && result.RiskScore.OverallScore >= fs.config.HighRiskThreshold {
		return true
	}

	return false
}

// generateRecommendedActions generates recommended actions based on results
func (fs *FraudService) generateRecommendedActions(result *FraudDetectionResult) []string {
	actions := make([]string, 0)

	if result.BotDetection != nil && result.BotDetection.Score >= fs.config.HighRiskThreshold {
		actions = append(actions, fmt.Sprintf("CAPTCHA verification for player %s", result.PlayerID))
		actions = append(actions, "Flag for manual review")
	}

	if result.MultiAccountDetection != nil && result.MultiAccountDetection.Score >= fs.config.HighRiskThreshold {
		actions = append(actions, "Device fingerprint verification")
		actions = append(actions, "Multi-account investigation")
	}

	for _, alert := range result.RuleAlerts {
		if alert.Severity == "critical" {
			actions = append(actions, fmt.Sprintf("Immediate review required: %s", alert.AlertType))
		}
	}

	return actions
}

// generateAlerts generates alerts from detection results
func (fs *FraudService) generateAlerts(ctx context.Context, result *FraudDetectionResult) {
	// Generate bot alert
	if result.BotDetection != nil && result.BotDetection.IsBot {
		alert := &AntiCheatAlert{
			ID:        fmt.Sprintf("bot_%s_%d", result.PlayerID, time.Now().UnixNano()),
			PlayerID:  result.PlayerID,
			AlertType: "bot",
			Severity:  fs.botDetectionToSeverity(result.BotDetection.Score),
			Score:     result.BotDetection.Score,
			Evidence:  result.BotDetection.Reasons,
			CreatedAt: time.Now(),
			Status:    "pending",
		}
		fs.alertService.CreateAlert(ctx, alert)
		fs.incrementCounter("BotAlertsGenerated")
	}

	// Generate collusion alerts
	for _, cr := range result.CollusionDetection {
		if cr.IsCollusion {
			alert := &AntiCheatAlert{
				ID:         fmt.Sprintf("collusion_%s_%s_%d", cr.PlayerA, cr.PlayerB, time.Now().UnixNano()),
				PlayerID:   cr.PlayerA,
				AlertType:  "collusion",
				Severity:   fs.collusionToSeverity(cr.Score),
				Score:      cr.Score,
				Evidence:   fs.collusionEvidenceToStrings(cr.TopEvidence),
				CreatedAt:  time.Now(),
				Status:     "pending",
			}
			fs.alertService.CreateAlert(ctx, alert)
			fs.incrementCounter("CollusionAlertsGenerated")
		}
	}
}

// botDetectionToSeverity converts bot detection score to severity
func (fs *FraudService) botDetectionToSeverity(score float64) string {
	if score >= fs.config.CriticalRiskThreshold {
		return "critical"
	}
	if score >= fs.config.HighRiskThreshold {
		return "high"
	}
	return "medium"
}

// collusionToSeverity converts collusion score to severity
func (fs *FraudService) collusionToSeverity(score float64) string {
	if score >= fs.config.CriticalRiskThreshold {
		return "critical"
	}
	if score >= fs.config.HighRiskThreshold {
		return "high"
	}
	return "medium"
}

// collusionEvidenceToStrings converts evidence items to strings
func (fs *FraudService) collusionEvidenceToStrings(evidence []EvidenceItem) []string {
	result := make([]string, len(evidence))
	for i, e := range evidence {
		result[i] = fmt.Sprintf("[%s] %s", e.Severity, e.Description)
	}
	return result
}

// incrementCounter increments a counter
func (fs *FraudService) incrementCounter(name string) {
	fs.metrics.mu.Lock()
	defer fs.metrics.mu.Unlock()

	switch name {
	case "TotalEventsProcessed":
		fs.metrics.TotalEventsProcessed++
	case "BotAlertsGenerated":
		fs.metrics.BotAlertsGenerated++
	case "CollusionAlertsGenerated":
		fs.metrics.CollusionAlertsGenerated++
	case "MultiAccountAlertsGenerated":
		fs.metrics.MultiAccountAlertsGenerated++
	case "RuleAlertsGenerated":
		fs.metrics.RuleAlertsGenerated++
	}

	fs.metrics.LastProcessedAt = time.Now()
}

// GetMetrics returns current metrics
func (fs *FraudService) GetMetrics() FraudMetrics {
	fs.metrics.mu.RLock()
	defer fs.metrics.mu.RUnlock()

	return *fs.metrics
}

// EventProcessor processes fraud events in batches
type EventProcessor struct {
	config   *FraudServiceConfig
	eventCh  chan *FraudEvent
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// FraudEvent represents an event for fraud detection
type FraudEvent struct {
	Type      string
	PlayerID  string
	TableID   string
	HandID    string
	Data      map[string]interface{}
	Timestamp time.Time
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(config *FraudServiceConfig) *EventProcessor {
	return &EventProcessor{
		config:  config,
		eventCh: make(chan *FraudEvent, config.ActionBufferSize),
		stopCh:  make(chan struct{}),
	}
}

// Start starts the event processor
func (ep *EventProcessor) Start(ctx context.Context, handler func(*FraudEvent) error) {
	ep.wg.Add(1)
	go func() {
		defer ep.wg.Done()
		ep.processEvents(ctx, handler)
	}()
}

// Stop stops the event processor
func (ep *EventProcessor) Stop() {
	close(ep.stopCh)
	ep.wg.Wait()
}

// processEvents processes events from the channel
func (ep *EventProcessor) processEvents(ctx context.Context, handler func(*FraudEvent) error) {
	ticker := time.NewTicker(ep.config.ProcessingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ep.stopCh:
			return
		case event := <-ep.eventCh:
			if err := handler(event); err != nil {
				// Log error in production
			}
		case <-ticker.C:
			// Periodic batch processing if needed
		}
	}
}

// PushEvent pushes an event to the processor
func (ep *EventProcessor) PushEvent(event *FraudEvent) error {
	select {
	case ep.eventCh <- event:
		return nil
	default:
		return fmt.Errorf("event buffer full")
	}
}

// FraudEventType constants
const (
	EventTypePlayerAction   = "player_action"
	EventTypePlayerJoin    = "player_join"
	EventTypePlayerLeave   = "player_leave"
	EventTypeHandComplete  = "hand_complete"
	EventTypeChatMessage   = "chat_message"
	EventTypeConnection   = "connection"
)

// CreatePlayerActionEvent creates a player action event
func CreatePlayerActionEvent(action *PlayerAction) *FraudEvent {
	data := map[string]interface{}{
		"action_type":   action.ActionType,
		"amount":        action.Amount,
		"decision_time": action.DecisionTime,
		"hand_phase":    action.HandPhase,
		"position":      action.Position,
		"pot_size":      action.PotSize,
	}

	return &FraudEvent{
		Type:      EventTypePlayerAction,
		PlayerID:  action.PlayerID,
		TableID:   action.TableID,
		HandID:    action.HandID,
		Data:      data,
		Timestamp: action.Timestamp,
	}
}

// CreateChatEvent creates a chat event for collusion detection
func CreateChatEvent(tableID, senderID, content string) *FraudEvent {
	return &FraudEvent{
		Type:      EventTypeChatMessage,
		PlayerID:  senderID,
		TableID:   tableID,
		Data:      map[string]interface{}{"content": content},
		Timestamp: time.Now(),
	}
}

// CreateConnectionEvent creates a connection event for multi-account detection
func CreateConnectionEvent(playerID, ipAddress, deviceID string, connected bool) *FraudEvent {
	eventType := EventTypeConnection
	if !connected {
		eventType = EventTypeConnection + "_disconnect"
	}

	return &FraudEvent{
		Type:      eventType,
		PlayerID:  playerID,
		Data: map[string]interface{}{
			"ip_address":     ipAddress,
			"device_id":      deviceID,
			"connected":      connected,
		},
		Timestamp: time.Now(),
	}
}

// SerializeEvent serializes an event to JSON
func SerializeEvent(event *FraudEvent) ([]byte, error) {
	return json.Marshal(event)
}

// DeserializeEvent deserializes an event from JSON
func DeserializeEvent(data []byte) (*FraudEvent, error) {
	var event FraudEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
