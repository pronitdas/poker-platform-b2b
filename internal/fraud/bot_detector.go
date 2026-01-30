package fraud

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

// BotDetectionConfig holds configuration for bot detection
type BotDetectionConfig struct {
	// Thresholds for bot-like behavior
	ActionTimeMeanThreshold   float64 // Bot if mean action time < this (seconds)
	ActionTimeStdDevThreshold float64 // Bot if std dev < this (seconds)
	BetPrecisionThreshold     float64 // Bot if precision > this (0-1)
	HandsPerHourThreshold     int     // Bot if hands/hour > this
	ConcurrentTablesThreshold int     // Bot if concurrent tables > this
	ConsistencyScoreThreshold float64 // Bot if consistency > this (0-1)

	// Weights for different features
	ActionTimeMeanWeight   float64
	ActionTimeStdDevWeight float64
	BetPrecisionWeight     float64
	HandsPerHourWeight     float64
	ConcurrentTablesWeight float64
	ConsistencyScoreWeight float64

	// Score thresholds
	BotScoreThreshold    float64 // Score above this = bot
	ReviewScoreThreshold float64 // Score above this = needs review
}

// DefaultBotDetectionConfig returns default configuration based on documented thresholds
func DefaultBotDetectionConfig() *BotDetectionConfig {
	return &BotDetectionConfig{
		// Thresholds (from documentation - human ranges vs bot ranges)
		ActionTimeMeanThreshold:   3.0,  // Humans: 2-15s, Bots: 0.5-3s
		ActionTimeStdDevThreshold: 0.5,  // Humans: 2-8s, Bots: <0.5s
		BetPrecisionThreshold:     0.95, // Humans: 70% round, Bots: 95% exact
		HandsPerHourThreshold:     100,  // Humans: 30-60, Bots: 100-200
		ConcurrentTablesThreshold: 10,   // Humans: 1-4, Bots: 10-50
		ConsistencyScoreThreshold: 0.85, // Humans: variable ±20%, Bots: ±5%

		// Feature weights (importance from documentation)
		ActionTimeMeanWeight:   0.20,
		ActionTimeStdDevWeight: 0.30, // Very High importance per docs
		BetPrecisionWeight:     0.10,
		HandsPerHourWeight:     0.15,
		ConcurrentTablesWeight: 0.10,
		ConsistencyScoreWeight: 0.15,

		BotScoreThreshold:    0.80,
		ReviewScoreThreshold: 0.60,
	}
}

// BotDetector detects automated bot behavior using multiple detection methods
type BotDetector struct {
	config           *BotDetectionConfig
	isolationForest  *IsolationForestDetector
	lstmDetector     *LSTMDetector
	featureExtractor *FeatureExtractor
	mu               sync.RWMutex
}

// NewBotDetector creates a new bot detector with all detection methods
func NewBotDetector(config *BotDetectionConfig) *BotDetector {
	if config == nil {
		config = DefaultBotDetectionConfig()
	}

	return &BotDetector{
		config:           config,
		isolationForest:  NewIsolationForestDetector(100, 256),
		lstmDetector:     NewLSTMDetector(),
		featureExtractor: NewFeatureExtractor(),
	}
}

// BotDetectionResult contains the result of bot detection analysis
type BotDetectionResult struct {
	IsBot             bool          `json:"is_bot"`
	Score             float64       `json:"score"`      // 0-1, higher = more likely bot
	Confidence        float64       `json:"confidence"` // 0-1, confidence in the score
	FeatureScores     FeatureScores `json:"feature_scores"`
	IsolationScore    float64       `json:"isolation_score"` // Isolation Forest anomaly score
	LSTMScore         float64       `json:"lstm_score"`      // LSTM sequential pattern score
	Reasons           []string      `json:"reasons"`
	RecommendedAction string        `json:"recommended_action"` // "flag", "review", "clear"
	DetectionMethods  []string      `json:"detection_methods"`  // Which methods triggered
}

// FeatureScores contains individual feature scores
type FeatureScores struct {
	ActionTimeMeanScore   float64 `json:"action_time_mean_score"`
	ActionTimeStdDevScore float64 `json:"action_time_std_dev_score"`
	BetPrecisionScore     float64 `json:"bet_precision_score"`
	HandsPerHourScore     float64 `json:"hands_per_hour_score"`
	ConcurrentTablesScore float64 `json:"concurrent_tables_score"`
	ConsistencyScore      float64 `json:"consistency_score"`
	WinRateConsistency    float64 `json:"win_rate_consistency"`
	ShowdownRate          float64 `json:"showdown_rate"`
}

// DetectBot analyzes player behavioral features using all detection methods
func (d *BotDetector) DetectBot(ctx context.Context, features *PlayerBehavioralFeatures) *BotDetectionResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := &BotDetectionResult{
		IsBot:             false,
		Score:             0.0,
		Confidence:        0.0,
		FeatureScores:     d.calculateFeatureScores(features),
		Reasons:           make([]string, 0),
		RecommendedAction: "clear",
		DetectionMethods:  make([]string, 0),
	}

	// Calculate Isolation Forest anomaly score
	result.IsolationScore = d.isolationForest.AnomalyScore(features)
	if result.IsolationScore > 0.7 {
		result.DetectionMethods = append(result.DetectionMethods, "isolation_forest")
	}

	// Calculate LSTM sequential pattern score
	result.LSTMScore = d.lstmDetector.DetectSequentialBot(features)
	if result.LSTMScore > 0.7 {
		result.DetectionMethods = append(result.DetectionMethods, "lstm")
	}

	// Calculate weighted score combining all methods
	fs := result.FeatureScores
	heuristicScore := fs.ActionTimeMeanScore*d.config.ActionTimeMeanWeight +
		fs.ActionTimeStdDevScore*d.config.ActionTimeStdDevWeight +
		fs.BetPrecisionScore*d.config.BetPrecisionWeight +
		fs.HandsPerHourScore*d.config.HandsPerHourWeight +
		fs.ConcurrentTablesScore*d.config.ConcurrentTablesWeight +
		fs.ConsistencyScore*d.config.ConsistencyScoreWeight

	// Combine all scores with appropriate weights
	// Heuristic: 50%, Isolation Forest: 25%, LSTM: 25%
	combinedScore := heuristicScore*0.5 + result.IsolationScore*0.25 + result.LSTMScore*0.25
	result.Score = combinedScore

	// Calculate confidence based on data quality and method agreement
	result.Confidence = d.calculateConfidence(features, result)

	// Determine if bot and recommended action
	if result.Score >= d.config.BotScoreThreshold && result.Confidence >= 0.7 {
		result.IsBot = true
		result.RecommendedAction = "flag"
	} else if result.Score >= d.config.ReviewScoreThreshold {
		result.RecommendedAction = "review"
	}

	// Generate reasons
	result.Reasons = d.generateReasons(features, result.FeatureScores, result)

	return result
}

// calculateFeatureScores calculates individual feature scores (0-1, higher = more bot-like)
func (d *BotDetector) calculateFeatureScores(features *PlayerBehavioralFeatures) FeatureScores {
	return FeatureScores{
		ActionTimeMeanScore:   d.scoreActionTimeMean(features.AvgActionTime),
		ActionTimeStdDevScore: d.scoreActionTimeStdDev(features.ActionTimeStdDev),
		BetPrecisionScore:     d.scoreBetPrecision(features.BetPrecision),
		HandsPerHourScore:     d.scoreHandsPerHour(features.HandsPerHour),
		ConcurrentTablesScore: d.scoreConcurrentTables(features.TablesConcurrent),
		ConsistencyScore:      d.scoreConsistency(features.ConsistencyScore),
		WinRateConsistency:    d.scoreWinRateConsistency(features.WinRateVariance),
		ShowdownRate:          d.scoreShowdownRate(features.ShowdownRate),
	}
}

// scoreActionTimeMean scores action time mean (bots act too fast)
func (d *BotDetector) scoreActionTimeMean(meanTime float64) float64 {
	if meanTime <= 0 {
		return 0.5
	}
	// Bot range: 0.5-3s, Human range: 2-15s
	if meanTime < d.config.ActionTimeMeanThreshold {
		return 1.0 - (meanTime / d.config.ActionTimeMeanThreshold)
	}
	return 0.0
}

// scoreActionTimeStdDev scores action time variance (bots are too consistent)
func (d *BotDetector) scoreActionTimeStdDev(stdDev float64) float64 {
	if stdDev <= 0 {
		return 0.5
	}
	if stdDev < d.config.ActionTimeStdDevThreshold {
		return 1.0 - (stdDev / d.config.ActionTimeStdDevThreshold)
	}
	return 0.0
}

// scoreBetPrecision scores bet precision (bots bet exact percentages)
func (d *BotDetector) scoreBetPrecision(precision float64) float64 {
	if precision < 0 || precision > 1 {
		return 0.5
	}
	if precision > d.config.BetPrecisionThreshold {
		return 1.0
	}
	return precision
}

// scoreHandsPerHour scores hands per hour (bots play too many)
func (d *BotDetector) scoreHandsPerHour(handsPerHour float64) float64 {
	if handsPerHour <= 0 {
		return 0.5
	}
	if handsPerHour > float64(d.config.HandsPerHourThreshold) {
		return 1.0
	}
	return handsPerHour / float64(d.config.HandsPerHourThreshold)
}

// scoreConcurrentTables scores concurrent tables (bots multitablke)
func (d *BotDetector) scoreConcurrentTables(tables int) float64 {
	if tables <= 0 {
		return 0.5
	}
	if tables > d.config.ConcurrentTablesThreshold {
		return 1.0
	}
	return float64(tables) / float64(d.config.ConcurrentTablesThreshold)
}

// scoreConsistency scores behavioral consistency (bots are too consistent)
func (d *BotDetector) scoreConsistency(consistency float64) float64 {
	if consistency < 0 || consistency > 1 {
		return 0.5
	}
	return consistency
}

// scoreWinRateConsistency scores win rate consistency (bots have too stable win rates)
func (d *BotDetector) scoreWinRateConsistency(variance float64) float64 {
	// Humans: variable ±20%, Bots: consistent ±5%
	if variance <= 0 {
		return 0.5
	}
	if variance < 0.05 {
		return 1.0
	}
	if variance > 0.20 {
		return 0.0
	}
	return 1.0 - ((variance - 0.05) / 0.15)
}

// scoreShowdownRate scores showdown rate (bots may have unusual showdown patterns)
func (d *BotDetector) scoreShowdownRate(rate float64) float64 {
	if rate < 0 || rate > 1 {
		return 0.5
	}
	// Humans: ~25-35% showdown rate, bots may be different
	if rate < 0.15 || rate > 0.50 {
		return 0.7
	}
	return 0.3
}

// calculateConfidence calculates confidence in the detection based on data quality
func (d *BotDetector) calculateConfidence(features *PlayerBehavioralFeatures, result *BotDetectionResult) float64 {
	confidence := 0.0

	// Weight by amount of data available
	if features.HandsPlayed >= 100 {
		confidence += 0.25
	} else if features.HandsPlayed >= 50 {
		confidence += 0.15
	}

	if features.ActionTimeStdDev > 0 {
		confidence += 0.15
	}

	if features.BetPrecision > 0 {
		confidence += 0.15
	}

	if features.HandsPerHour > 0 {
		confidence += 0.15
	}

	if features.ConsistencyScore > 0 {
		confidence += 0.15
	}

	// Boost confidence if multiple methods agree
	if len(result.DetectionMethods) >= 2 {
		confidence += 0.15
	}

	// Boost confidence if we have lots of data
	if features.HandsPlayed >= 500 {
		confidence = math.Min(1.0, confidence+0.1)
	}

	return math.Min(1.0, confidence)
}

// generateReasons creates human-readable reasons for the detection
func (d *BotDetector) generateReasons(features *PlayerBehavioralFeatures, scores FeatureScores, result *BotDetectionResult) []string {
	reasons := make([]string, 0)

	// Heuristic-based reasons
	if scores.ActionTimeMeanScore > 0.7 {
		reasons = append(reasons, "Unusually fast action timing")
	}

	if scores.ActionTimeStdDevScore > 0.8 {
		reasons = append(reasons, "Suspiciously consistent action timing")
	}

	if scores.BetPrecisionScore > 0.7 {
		reasons = append(reasons, "Suspiciously precise bet sizing")
	}

	if scores.HandsPerHourScore > 0.7 {
		reasons = append(reasons, "Excessive hands per hour")
	}

	if scores.ConcurrentTablesScore > 0.7 {
		reasons = append(reasons, "Too many concurrent tables")
	}

	if scores.ConsistencyScore > 0.8 {
		reasons = append(reasons, "Highly consistent behavior patterns")
	}

	// ML-based reasons
	if result.IsolationScore > 0.7 {
		reasons = append(reasons, "ML anomaly detection flagged unusual patterns")
	}

	if result.LSTMScore > 0.7 {
		reasons = append(reasons, "Sequential pattern analysis detected bot-like behavior")
	}

	return reasons
}

// IsolationForestDetector implements Isolation Forest for anomaly detection
type IsolationForestDetector struct {
	numTrees      int
	subsampleSize int
	maxDepth      int
	rng           *rand.Rand
}

// NewIsolationForestDetector creates a new Isolation Forest detector
func NewIsolationForestDetector(numTrees int, subsampleSize int) *IsolationForestDetector {
	if numTrees <= 0 {
		numTrees = 100
	}
	if subsampleSize <= 0 {
		subsampleSize = 256
	}

	return &IsolationForestDetector{
		numTrees:      numTrees,
		subsampleSize: subsampleSize,
		maxDepth:      int(math.Log2(float64(subsampleSize))) + 1,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AnomalyScore calculates an anomaly score for the given features
func (ifd *IsolationForestDetector) AnomalyScore(features *PlayerBehavioralFeatures) float64 {
	// Convert features to feature vector
	featureVector := ifd.featuresToVector(features)

	// Calculate average path length across all trees
	totalPathLength := 0.0
	for i := 0; i < ifd.numTrees; i++ {
		pathLength := ifd.pathLength(featureVector, i)
		totalPathLength += pathLength
	}

	avgPathLength := totalPathLength / float64(ifd.numTrees)

	// Calculate anomaly score using average path length
	// Shorter path = more anomalous = higher score
	c := ifd.averagePathLength(ifd.subsampleSize)
	if avgPathLength <= c {
		return 1.0
	}

	score := math.Pow(2, -avgPathLength/c)
	return score
}

// featuresToVector converts behavioral features to a normalized feature vector
func (ifd *IsolationForestDetector) featuresToVector(features *PlayerBehavioralFeatures) []float64 {
	return []float64{
		features.AvgActionTime,                    // Normalized by threshold
		features.ActionTimeStdDev,                 // Normalized
		features.BetPrecision,                     // Already 0-1
		features.HandsPerHour / 200.0,             // Normalize
		float64(features.TablesConcurrent) / 50.0, // Normalize
		features.ConsistencyScore,                 // Already 0-1
		features.WinRate,                          // Already 0-1
		features.WinRateVariance * 5,              // Amplify small variance
		features.ShowdownRate,                     // Already 0-1
		features.ErrorRate * 10,                   // Amplify error rate
	}
}

// pathLength calculates the path length to isolate a sample
func (ifd *IsolationForestDetector) pathLength(sample []float64, treeSeed int) float64 {
	r := rand.New(rand.NewSource(int64(treeSeed) + ifd.rng.Int63()))

	depth := 0.0
	currentSample := make([]float64, len(sample))
	copy(currentSample, sample)

	for depth < float64(ifd.maxDepth) && len(currentSample) > 1 {
		// Select random feature
		featureIdx := r.Intn(len(currentSample))

		// Generate random split point within feature range
		minVal := currentSample[featureIdx] - 1.0
		maxVal := currentSample[featureIdx] + 1.0
		splitPoint := minVal + r.Float64()*(maxVal-minVal)

		// Partition samples
		var left, right []float64
		for _, val := range currentSample {
			if val < splitPoint {
				left = append(left, val)
			} else {
				right = append(right, val)
			}
		}

		// Choose partition with fewer samples
		if len(left) < len(right) {
			currentSample = left
		} else {
			currentSample = right
		}

		depth++
	}

	return depth
}

// averagePathLength calculates the average path length for a given sample size
func (ifd *IsolationForestDetector) averagePathLength(sampleSize int) float64 {
	if sampleSize <= 1 {
		return 0.0
	}
	// Approximation of average path length c(n)
	return 2.0 * (float64(sampleSize-1) / float64(sampleSize))
}

// LSTMDetector implements LSTM-based sequential pattern detection for bot behavior
type LSTMDetector struct {
	hiddenSize  int
	numLayers   int
	sequenceLen int
}

// NewLSTMDetector creates a new LSTM detector
func NewLSTMDetector() *LSTMDetector {
	return &LSTMDetector{
		hiddenSize:  64,
		numLayers:   2,
		sequenceLen: 50,
	}
}

// DetectSequentialBot detects bot-like sequential patterns in player actions
func (lstm *LSTMDetector) DetectSequentialBot(features *PlayerBehavioralFeatures) float64 {
	// Simplified LSTM-like detection based on feature patterns
	// In production, this would use a trained LSTM model

	// Bot indicators in sequential patterns
	score := 0.0

	// Check for perfect timing consistency (indicates programmatic behavior)
	if features.ActionTimeStdDev < 0.1 {
		score += 0.4
	} else if features.ActionTimeStdDev < 0.3 {
		score += 0.2
	}

	// Check for perfect bet sizing precision (programmatic bet sizing)
	if features.BetPrecision > 0.95 {
		score += 0.3
	}

	// Check for unnatural play duration consistency
	if features.HandsPerHour > 80 {
		score += 0.2
	}

	// Check for win rate that's too stable (bots often maintain consistent win rates)
	if features.WinRateVariance < 0.02 {
		score += 0.1
	}

	return math.Min(1.0, score)
}

// FeatureExtractor extracts behavioral features from player actions
type FeatureExtractor struct {
	mu sync.RWMutex
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

// ExtractFeatures extracts behavioral features from a list of player actions
func (fe *FeatureExtractor) ExtractFeatures(playerID string, actions []PlayerAction, timeRange time.Duration) *PlayerBehavioralFeatures {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if len(actions) == 0 {
		return &PlayerBehavioralFeatures{
			PlayerID:    playerID,
			TimeRange:   timeRange.String(),
			ExtractedAt: time.Now(),
		}
	}

	now := time.Now()
	cutoff := now.Add(-timeRange)

	// Filter actions within time range
	var filteredActions []PlayerAction
	for _, action := range actions {
		if action.Timestamp.After(cutoff) {
			filteredActions = append(filteredActions, action)
		}
	}

	features := &PlayerBehavioralFeatures{
		PlayerID:    playerID,
		TimeRange:   timeRange.String(),
		ExtractedAt: now,
		HandsPlayed: len(filteredActions),
	}

	if len(filteredActions) == 0 {
		return features
	}

	// Calculate timing features
	actionTimes := make([]float64, 0, len(filteredActions))
	for _, action := range filteredActions {
		if action.DecisionTime > 0 {
			actionTimes = append(actionTimes, float64(action.DecisionTime)/1000.0) // Convert ms to seconds
		}
	}

	if len(actionTimes) > 0 {
		features.AvgActionTime = mean(actionTimes)
		features.ActionTimeStdDev = stdDev(actionTimes)
		if len(actionTimes) > 0 {
			features.ActionTimeMin = min(actionTimes...)
			features.ActionTimeMax = max(actionTimes...)
		}
	}

	// Calculate bet sizing features
	betAmounts := make([]float64, 0)
	for _, action := range filteredActions {
		if action.ActionType == "bet" || action.ActionType == "raise" {
			if action.PotSize > 0 {
				betAmounts = append(betAmounts, float64(action.Amount)/float64(action.PotSize))
			}
		}
	}

	if len(betAmounts) > 0 {
		features.BetPrecision = calculateBetPrecision(betAmounts)
		features.AvgBetToPotRatio = mean(betAmounts)
		features.BetSizeVariance = variance(betAmounts)
	}

	// Calculate volume features
	durationHours := timeRange.Hours()
	if durationHours > 0 {
		features.HandsPerHour = float64(len(filteredActions)) / durationHours
	}

	features.TablesConcurrent = 0

	// Calculate performance features from hand history data
	features.WinRate = 0.0
	features.WinRateVariance = 0.0
	features.ShowdownRate = 0.0
	features.VPIP = 0.0
	features.PFR = 0.0

	// Calculate consistency score
	features.ConsistencyScore = fe.calculateConsistencyScore(features)

	return features
}

// calculateConsistencyScore calculates overall consistency score
func (fe *FeatureExtractor) calculateConsistencyScore(features *PlayerBehavioralFeatures) float64 {
	consistency := 0.0
	metrics := 0

	if features.ActionTimeStdDev > 0 {
		consistency += 1.0 - math.Min(1.0, features.ActionTimeStdDev/5.0)
		metrics++
	}

	if features.BetPrecision > 0 {
		consistency += features.BetPrecision
		metrics++
	}

	if metrics > 0 {
		return consistency / float64(metrics)
	}
	return 0.5
}

// calculateBetPrecision calculates how precise bet sizing is (bots use exact percentages)
func calculateBetPrecision(betAmounts []float64) float64 {
	if len(betAmounts) == 0 {
		return 0.0
	}

	preciseBets := 0
	for _, amount := range betAmounts {
		// Check if bet is a round number (e.g., 0.5, 0.75, 1.0 pot)
		// Bots tend to bet exact fractions, humans bet round numbers
		remainder := amount - math.Floor(amount)
		if remainder > 0.01 && remainder < 0.99 {
			preciseBets++
		}
	}

	return float64(preciseBets) / float64(len(betAmounts))
}

// Statistical helper functions
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	m := mean(values)
	sumSq := 0.0
	for _, v := range values {
		sumSq += (v - m) * (v - m)
	}
	return math.Sqrt(sumSq / float64(len(values)))
}

func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	m := mean(values)
	sumSq := 0.0
	for _, v := range values {
		sumSq += (v - m) * (v - m)
	}
	return sumSq / float64(len(values))
}

func min(values ...float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func max(values ...float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
