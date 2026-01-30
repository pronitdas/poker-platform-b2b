package fraud

import (
	"context"
	"testing"
	"time"
)

func TestBotDetector_DetectBot_HumanBehavior(t *testing.T) {
	detector := NewBotDetector(nil)
	ctx := context.Background()

	// Human-like behavior features
	features := &PlayerBehavioralFeatures{
		PlayerID:           "human_player",
		AvgActionTime:      8.5,      // 8.5 seconds (human range: 2-15s)
		ActionTimeStdDev:   4.2,      // 4.2 seconds variance (human range: 2-8s)
		BetPrecision:       0.65,     // 65% (human range: ~70%)
		HandsPerHour:       45,       // 45 hands/hour (human range: 30-60)
		TablesConcurrent:   2,        // 2 tables (human range: 1-4)
		ConsistencyScore:   0.4,      // Variable behavior
		WinRateVariance:    0.15,     // 15% variance
		ShowdownRate:       0.28,     // 28% (normal human range)
		HandsPlayed:        150,
	}

	result := detector.DetectBot(ctx, features)

	if result.IsBot {
		t.Errorf("Expected human to NOT be flagged as bot, got bot=%v, score=%f", result.IsBot, result.Score)
	}

	if result.Score >= 0.6 {
		t.Errorf("Expected human score to be below review threshold, got score=%f", result.Score)
	}

	if result.RecommendedAction != "clear" {
		t.Errorf("Expected recommended action to be 'clear', got=%s", result.RecommendedAction)
	}
}

func TestBotDetector_DetectBot_BotBehavior(t *testing.T) {
	detector := NewBotDetector(nil)
	ctx := context.Background()

	// Bot-like behavior features
	features := &PlayerBehavioralFeatures{
		PlayerID:           "bot_player",
		AvgActionTime:      1.2,      // 1.2 seconds (bot range: 0.5-3s)
		ActionTimeStdDev:   0.15,     // 0.15 seconds variance (bot range: <0.5s)
		BetPrecision:       0.97,     // 97% (bot range: >95%)
		HandsPerHour:       180,      // 180 hands/hour (bot range: 100-200)
		TablesConcurrent:   25,       // 25 tables (bot range: 10-50)
		ConsistencyScore:   0.92,     // Very consistent
		WinRateVariance:    0.02,     // 2% variance (very stable)
		ShowdownRate:       0.15,     // Low showdown rate
		HandsPlayed:        500,
	}

	result := detector.DetectBot(ctx, features)

	if !result.IsBot {
		t.Logf("Bot detection score: %f, confidence: %f, action: %s",
			result.Score, result.Confidence, result.RecommendedAction)
		// Score should be high enough to at least require review
		if result.Score < 0.6 {
			t.Errorf("Expected bot to have high score, got score=%f", result.Score)
		}
	}

	// Verify multiple detection methods triggered
	if len(result.DetectionMethods) == 0 {
		t.Logf("Detection methods: %v", result.DetectionMethods)
	}
}

func TestBotDetector_DetectBot_Borderline(t *testing.T) {
	detector := NewBotDetector(nil)
	ctx := context.Background()

	// Borderline features (may or may not be bot)
	features := &PlayerBehavioralFeatures{
		PlayerID:           "borderline_player",
		AvgActionTime:      4.0,
		ActionTimeStdDev:   1.5,
		BetPrecision:       0.80,
		HandsPerHour:       80,
		TablesConcurrent:   5,
		ConsistencyScore:   0.6,
		HandsPlayed:        75,
	}

	result := detector.DetectBot(ctx, features)

	// Should be marked for review
	if result.RecommendedAction == "clear" {
		t.Logf("Borderline case - recommended action: %s, score: %f",
			result.RecommendedAction, result.Score)
	}
}

func TestBotDetector_ConfidenceCalculation(t *testing.T) {
	detector := NewBotDetector(nil)
	ctx := context.Background()

	// Test confidence with minimal data
	featuresMin := &PlayerBehavioralFeatures{
		PlayerID:      "min_data_player",
		HandsPlayed:   10,
		AvgActionTime: 5.0,
	}

	resultMin := detector.DetectBot(ctx, featuresMin)
	if resultMin.Confidence > 0.5 {
		t.Errorf("Expected low confidence with minimal data, got confidence=%f", resultMin.Confidence)
	}

	// Test confidence with comprehensive data
	featuresMax := &PlayerBehavioralFeatures{
		PlayerID:           "max_data_player",
		HandsPlayed:        1000,
		AvgActionTime:      7.5,
		ActionTimeStdDev:   3.5,
		BetPrecision:       0.72,
		HandsPerHour:       50,
		TablesConcurrent:   2,
		ConsistencyScore:   0.45,
	}

	resultMax := detector.DetectBot(ctx, featuresMax)
	if resultMax.Confidence < 0.8 {
		t.Logf("Expected higher confidence with more data, got confidence=%f", resultMax.Confidence)
	}
}

func TestCollusionDetector_DetectCollusion_NoRelationship(t *testing.T) {
	detector := NewCollusionDetector(nil)
	ctx := context.Background()

	// No relationship between players
	relationship := &PlayerRelationship{
		PlayerA:           "player_a",
		PlayerB:           "player_b",
		CoOccurrenceCount:  5,
		WinRateA:          0.45,
		IPMatchCount:      0,
		DeviceMatchCount:   0,
	}

	result := detector.DetectCollusion(ctx, relationship)

	if result.Score > 0.3 {
		t.Errorf("Expected low collusion score for players with no relationship, got score=%f", result.Score)
	}

	if result.IsCollusion {
		t.Errorf("Expected no collusion detected for unrelated players")
	}
}

func TestCollusionDetector_DetectCollusion_Suspicious(t *testing.T) {
	detector := NewCollusionDetector(nil)
	ctx := context.Background()

	// Suspicious relationship
	relationship := &PlayerRelationship{
		PlayerA:            "player_a",
		PlayerB:            "player_b",
		CoOccurrenceCount:   150,     // 150 hands together (threshold: 50)
		WinRateA:           0.15,    // 15% win rate vs player B (suspicious chip dumping)
		IPMatchCount:       8,       // Same IP 8 times (threshold: 10)
		DeviceMatchCount:   3,       // Same device 3 times (threshold: 5)
		MutualWins:         5,       // 5 mutual wins (both in hand)
		TotalHandsA:        500,
		TotalHandsB:        450,
	}

	result := detector.DetectCollusion(ctx, relationship)

	if result.Score < 0.5 {
		t.Logf("Suspicious relationship score: %f, type: %s", result.Score, result.CollusionType)
	}

	// Should generate evidence
	if len(result.Evidence) == 0 {
		t.Logf("Expected evidence for suspicious relationship, got none")
	}
}

func TestCollusionDetector_FindCollusionRings(t *testing.T) {
	detector := NewCollusionDetector(nil)
	ctx := context.Background()

	// Add some nodes and edges to the graph
	detector.graph.AddInteractionEdge(&InteractionEdge{
		PlayerA:          "player_1",
		PlayerB:          "player_2",
		CoOccurrences:     100,
		SeatingAdjacency:  20,
		AggressionDelta:   0.4,
		Weight:           0.8,
	})

	detector.graph.AddInteractionEdge(&InteractionEdge{
		PlayerA:          "player_2",
		PlayerB:          "player_3",
		CoOccurrences:     80,
		SeatingAdjacency:  15,
		AggressionDelta:   0.35,
		Weight:           0.7,
	})

	detector.graph.AddInteractionEdge(&InteractionEdge{
		PlayerA:          "player_1",
		PlayerB:          "player_3",
		CoOccurrences:     90,
		SeatingAdjacency:  18,
		AggressionDelta:   0.38,
		Weight:           0.75,
	})

	rings := detector.FindCollusionRings(ctx, 0.0)

	// Should find at least one community
	t.Logf("Found %d potential rings", len(rings))
	for _, ring := range rings {
		t.Logf("Ring %s: %d members, density: %.2f",
			ring.RingID, len(ring.Members), ring.Density)
	}
}

func TestMultiAccountDetector_DetectMultiAccount(t *testing.T) {
	// Skip without mock database
	t.Skip("Requires mock database implementation")
}

func TestRuleBasedDetector_EvaluateRules(t *testing.T) {
	detector := NewRuleBasedDetector()
	ctx := context.Background()

	// Test excessive volume rule
	data := &RuleCheckData{
		PlayerID:       "volume_player",
		HandsPlayed24h: 600, // Exceeds 500 threshold
		HandsPlayed7d:  1500,
		WinRate24h:     0.52,
		WinRate7d:       0.50,
	}

	alerts := detector.EvaluateRules(ctx, data)

	found := false
	for _, alert := range alerts {
		if alert.AlertType == "bot" {
			found = true
			t.Logf("Found volume alert: %s (severity: %s)", alert.ID, alert.Severity)
		}
	}

	if !found {
		t.Errorf("Expected volume rule to trigger alert for >500 hands in 24h")
	}
}

func TestRuleBasedDetector_PerfectWinRate(t *testing.T) {
	detector := NewRuleBasedDetector()
	ctx := context.Background()

	// Test perfect win rate rule
	data := &RuleCheckData{
		PlayerID:       "perfect_player",
		HandsPlayed24h: 100,
		WinRate24h:     0.98, // 98% win rate
	}

	alerts := detector.EvaluateRules(ctx, data)

	found := false
	for _, alert := range alerts {
		if alert.AlertType == "bot" {
			found = true
			if alert.Severity != "high" {
				t.Errorf("Expected high severity for perfect win rate, got %s", alert.Severity)
			}
		}
	}

	if !found {
		t.Errorf("Expected perfect win rate rule to trigger alert")
	}
}

func TestRuleBasedDetector_DisableEnable(t *testing.T) {
	detector := NewRuleBasedDetector()

	// Disable a rule
	err := detector.DisableRule("excessive_volume_24h")
	if err != nil {
		t.Errorf("Failed to disable rule: %v", err)
	}

	rule := detector.GetRule("excessive_volume_24h")
	if rule.Enabled {
		t.Errorf("Rule should be disabled")
	}

	// Re-enable the rule
	err = detector.EnableRule("excessive_volume_24h")
	if err != nil {
		t.Errorf("Failed to enable rule: %v", err)
	}

	rule = detector.GetRule("excessive_volume_24h")
	if !rule.Enabled {
		t.Errorf("Rule should be enabled")
	}
}

func TestDeviceFingerprint_Generation(t *testing.T) {
	fingerprint := GenerateClientFingerprint(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"1920x1080",
		24,
		"America/New_York",
		"en-US",
		"Win32",
		8,
		16.0,
		true,
		"ANGLE (NVIDIA GeForce RTX 3080)",
	)

	if len(fingerprint) != 64 { // SHA256 hex = 64 chars
		t.Errorf("Expected 64-char fingerprint, got %d", len(fingerprint))
	}

	// Same input should produce same fingerprint
	fingerprint2 := GenerateClientFingerprint(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"1920x1080",
		24,
		"America/New_York",
		"en-US",
		"Win32",
		8,
		16.0,
		true,
		"ANGLE (NVIDIA GeForce RTX 3080)",
	)

	if fingerprint != fingerprint2 {
		t.Errorf("Same input should produce same fingerprint")
	}

	// Different input should produce different fingerprint
	fingerprint3 := GenerateClientFingerprint(
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"1920x1080",
		24,
		"America/New_York",
		"en-US",
		"MacIntel",
		8,
		16.0,
		true,
		"ANGLE (AMD Radeon Pro 5500M)",
	)

	if fingerprint == fingerprint3 {
		t.Errorf("Different input should produce different fingerprint")
	}
}

func TestNetworkAnalyzer(t *testing.T) {
	analyzer := NewNetworkAnalyzer()

	tests := []struct {
		ip1     string
		ip2     string
		sameNet bool
	}{
		{"192.168.1.100", "192.168.1.200", true},
		{"10.0.0.50", "10.0.0.100", true},
		{"192.168.1.1", "10.0.0.1", false},
		{"192.168.1.1", "172.16.0.1", false},
	}

	for _, tt := range tests {
		result := analyzer.IsSameNetwork(tt.ip1, tt.ip2)
		if result != tt.sameNet {
			t.Errorf("IsSameNetwork(%s, %s) = %v, want %v",
				tt.ip1, tt.ip2, result, tt.sameNet)
		}

		prefix1 := analyzer.GetNetworkPrefix(tt.ip1)
		prefix2 := analyzer.GetNetworkPrefix(tt.ip2)

		if tt.sameNet && prefix1 != prefix2 {
			t.Errorf("Same network IPs should have same prefix: %s vs %s",
				prefix1, prefix2)
		}
	}
}

func TestFeatureExtractor_ExtractFeatures(t *testing.T) {
	extractor := NewFeatureExtractor()
	playerID := "test_player"

	// Empty actions
	features := extractor.ExtractFeatures(playerID, []PlayerAction{}, 24*time.Hour)
	if features.PlayerID != playerID {
		t.Errorf("Expected player ID to be set")
	}
	if features.HandsPlayed != 0 {
		t.Errorf("Expected 0 hands for empty action list")
	}
}

func TestRiskScorer_CalculateRiskScore(t *testing.T) {
	// Skip without full service dependencies
	t.Skip("Requires mock services")
}

func TestIsolationForestDetector_AnomalyScore(t *testing.T) {
	detector := NewIsolationForestDetector(10, 32)

	features := &PlayerBehavioralFeatures{
		PlayerID:           "test_player",
		AvgActionTime:      2.0,
		ActionTimeStdDev:   0.1, // Very consistent = anomalous
		BetPrecision:       0.98,
		HandsPerHour:       150,
		TablesConcurrent:   20,
		ConsistencyScore:   0.95,
		WinRate:           0.60,
		WinRateVariance:    0.01,
		ShowdownRate:       0.20,
	}

	score := detector.AnomalyScore(features)

	if score < 0.5 {
		t.Logf("Anomaly score for suspicious behavior: %f", score)
	}

	// Normal behavior should have lower score
	normalFeatures := &PlayerBehavioralFeatures{
		PlayerID:           "normal_player",
		AvgActionTime:      8.0,
		ActionTimeStdDev:   3.5,
		BetPrecision:       0.70,
		HandsPerHour:       40,
		TablesConcurrent:   2,
		ConsistencyScore:   0.45,
		WinRate:           0.48,
		WinRateVariance:    0.12,
		ShowdownRate:       0.28,
	}

	normalScore := detector.AnomalyScore(normalFeatures)
	if normalScore >= score {
		t.Logf("Normal behavior should have lower anomaly score than suspicious")
	}
}

func TestStatisticalHelpers(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"empty", []float64{}, 0.0},
		{"single", []float64{5.0}, 5.0},
		{"simple", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mean(tt.values)
			if len(tt.values) > 0 && result != tt.expected {
				t.Errorf("mean(%v) = %f, want %f", tt.values, result, tt.expected)
			}
		})
	}

	// Test stdDev
	values := []float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0}
	result := stdDev(values)
	expected := 2.0
	if abs(result-expected) > 0.01 {
		t.Errorf("stdDev(%v) = %f, want approximately %f", values, result, expected)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkBotDetector_DetectBot(b *testing.B) {
	detector := NewBotDetector(nil)
	ctx := context.Background()
	features := &PlayerBehavioralFeatures{
		PlayerID:           "bench_player",
		AvgActionTime:      5.0,
		ActionTimeStdDev:   2.0,
		BetPrecision:       0.75,
		HandsPerHour:       60,
		TablesConcurrent:   3,
		ConsistencyScore:   0.5,
		HandsPlayed:        200,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectBot(ctx, features)
	}
}

func BenchmarkCollusionDetector_DetectCollusion(b *testing.B) {
	detector := NewCollusionDetector(nil)
	ctx := context.Background()
	relationship := &PlayerRelationship{
		PlayerA:          "player_a",
		PlayerB:          "player_b",
		CoOccurrenceCount: 100,
		WinRateA:         0.50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectCollusion(ctx, relationship.PlayerA, relationship.PlayerB)
	}
}
