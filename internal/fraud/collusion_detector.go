package fraud

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// CollusionDetectionConfig holds configuration for collusion detection
type CollusionDetectionConfig struct {
	// Thresholds for collusion detection
	CoOccurrenceThreshold     int           // Hands played together to flag
	SeatingAdjacencyThreshold int           // Repeated seating adjacent
	StakeOverlapThreshold     float64       // Stake selection overlap percentage
	ArrivalSyncThreshold      time.Duration // Time window for synchronized arrival
	DepartureSyncThreshold    time.Duration // Time window for synchronized departure

	// Pairwise behavior thresholds
	AggressionDeltaThreshold float64 // Change in aggression vs colluder
	PotDeltaThreshold        float64 // Change in pot sizes vs colluder
	ShowdownDeltaThreshold   float64 // Change in showdown rate vs colluder
	CheckDownThreshold       float64 // Heads-up check-down frequency
	VPIPDeltaThreshold       float64 // VPIP change vs colluder
	PFRDeltaThreshold        float64 // PFR change vs colluder
	ThreeBetDeltaThreshold   float64 // 3-bet frequency change vs colluder

	// Chip flow thresholds
	ChipTransferThreshold      int64   // Net chip transfer threshold
	EVLossThreshold            float64 // Suspicious EV loss threshold
	TransferFrequencyThreshold int     // Transfer frequency per 100 hands

	// Network/device thresholds
	IPMatchThreshold      int // Same IP occurrences
	DeviceMatchThreshold  int // Same device occurrences
	NetworkMatchThreshold int // Same network occurrences

	// Score weights
	CoOccurrenceWeight     float64
	SeatingAdjacencyWeight float64
	StakeOverlapWeight     float64
	ArrivalSyncWeight      float64
	AggressionDeltaWeight  float64
	PotDeltaWeight         float64
	ShowdownDeltaWeight    float64
	ChipTransferWeight     float64
	EVLossWeight           float64
	IPMatchWeight          float64
	DeviceMatchWeight      float64
	NetworkMatchWeight     float64

	// Score thresholds
	CollusionScoreThreshold float64
	SoftPlayScoreThreshold  float64
	ReviewScoreThreshold    float64
	CriticalScoreThreshold  float64
}

// DefaultCollusionDetectionConfig returns comprehensive configuration
func DefaultCollusionDetectionConfig() *CollusionDetectionConfig {
	return &CollusionDetectionConfig{
		// Basic thresholds
		CoOccurrenceThreshold:     50,
		SeatingAdjacencyThreshold: 10,
		StakeOverlapThreshold:     0.7,
		ArrivalSyncThreshold:      5 * time.Minute,
		DepartureSyncThreshold:    5 * time.Minute,

		// Pairwise behavior thresholds
		AggressionDeltaThreshold: 0.3,  // 30% change in aggression
		PotDeltaThreshold:        0.4,  // 40% change in pot sizes
		ShowdownDeltaThreshold:   0.25, // 25% change in showdown rate
		CheckDownThreshold:       0.8,  // 80% check-down in heads-up
		VPIPDeltaThreshold:       0.2,  // 20% VPIP change
		PFRDeltaThreshold:        0.2,  // 20% PFR change
		ThreeBetDeltaThreshold:   0.3,  // 30% 3-bet frequency change

		// Chip flow thresholds
		ChipTransferThreshold:      10000, // 10K chips net transfer
		EVLossThreshold:            0.15,  // 15% EV loss consistently
		TransferFrequencyThreshold: 5,     // 5 transfers per 100 hands

		// Network/device thresholds
		IPMatchThreshold:      5,
		DeviceMatchThreshold:  2,
		NetworkMatchThreshold: 10,

		// Weights
		CoOccurrenceWeight:     0.10,
		SeatingAdjacencyWeight: 0.08,
		StakeOverlapWeight:     0.07,
		ArrivalSyncWeight:      0.05,
		AggressionDeltaWeight:  0.15,
		PotDeltaWeight:         0.10,
		ShowdownDeltaWeight:    0.10,
		ChipTransferWeight:     0.20,
		EVLossWeight:           0.10,
		IPMatchWeight:          0.03,
		DeviceMatchWeight:      0.05,
		NetworkMatchWeight:     0.02,

		// Score thresholds
		CollusionScoreThreshold: 0.65,
		SoftPlayScoreThreshold:  0.50,
		ReviewScoreThreshold:    0.35,
		CriticalScoreThreshold:  0.85,
	}
}

// CollusionDetector detects collusion between players using graph-based analysis
type CollusionDetector struct {
	config            *CollusionDetectionConfig
	graph             *PlayerInteractionGraph
	pairwiseScorer    *PairwiseSoftPlayScorer
	chipFlowAnalyzer  *ChipFlowAnalyzer
	communityDetector *CommunityDetector
	alertGenerator    *CollusionAlertGenerator
	mu                sync.RWMutex
}

// PlayerInteractionGraph represents player relationships with rich metadata
type PlayerInteractionGraph struct {
	nodes        map[string]*PlayerNode
	edges        map[string]*InteractionEdge
	sessionStore SessionStore
	mu           sync.RWMutex
}

// PlayerNode represents a player with interaction history
type PlayerNode struct {
	PlayerID      string
	Sessions      map[string]bool
	StakesPlayed  map[int64]int // stake -> hand count
	FirstSeen     time.Time
	LastSeen      time.Time
	TotalHands    int
	TotalWinnings int64
}

// InteractionEdge represents an interaction between two players
type InteractionEdge struct {
	PlayerA          string
	PlayerB          string
	CoOccurrences    int
	SeatingAdjacency int
	StakeOverlap     float64
	ArrivalSyncs     int
	DepartureSyncs   int
	AggressionDelta  float64
	PotDelta         float64
	ShowdownDelta    float64
	CheckDownRate    float64
	VPIPDelta        float64
	PFRDelta         float64
	ThreeBetDelta    float64
	NetChipTransfer  int64
	EVLossRate       float64
	TransferCount    int
	IPMatches        int
	DeviceMatches    int
	NetworkMatches   int
	FirstInteraction time.Time
	LastInteraction  time.Time
	Weight           float64
}

// PairwiseSoftPlayScorer scores pairwise soft-play behavior
type PairwiseSoftPlayScorer struct {
	config        *CollusionDetectionConfig
	gradientBoost *GradientBoostingModel
	mu            sync.RWMutex
}

// GradientBoostingModel implements gradient boosting for scoring
type GradientBoostingModel struct {
	trees        []DecisionTree
	learningRate float64
	maxDepth     int
	numTrees     int
}

// DecisionTree represents a single decision tree
type DecisionTree struct {
	root     *TreeNode
	maxDepth int
}

// TreeNode represents a node in a decision tree
type TreeNode struct {
	feature    string
	threshold  float64
	left       *TreeNode
	right      *TreeNode
	isLeaf     bool
	prediction float64
}

// ChipFlowAnalyzer analyzes chip transfer patterns
type ChipFlowAnalyzer struct {
	config     *CollusionDetectionConfig
	transferDB TransferDatabase
	mu         sync.RWMutex
}

// TransferRecord records a chip transfer between players
type TransferRecord struct {
	FromPlayer string
	ToPlayer   string
	TableID    string
	HandID     string
	Amount     int64
	EVImpact   float64
	Context    string // "bet", "raise", "call", "showdown"
	Timestamp  time.Time
}

// TransferDatabase interface for chip flow tracking
type TransferDatabase interface {
	RecordTransfer(record *TransferRecord) error
	GetPlayerTransfers(playerID string, startTime, endTime time.Time) ([]*TransferRecord, error)
	GetPairTransfers(playerA, playerB string, startTime, endTime time.Time) ([]*TransferRecord, error)
}

// CommunityDetector implements Louvain/Leiden community detection
type CommunityDetector struct {
	graph *PlayerInteractionGraph
	mu    sync.RWMutex
}

// CollusionAlertGenerator generates comprehensive alerts with evidence
type CollusionAlertGenerator struct {
	config       *CollusionDetectionConfig
	alertStorage AlertStorage
	mu           sync.RWMutex
}

// NewCollusionDetector creates a new comprehensive collusion detector
func NewCollusionDetector(config *CollusionDetectionConfig) *CollusionDetector {
	if config == nil {
		config = DefaultCollusionDetectionConfig()
	}

	return &CollusionDetector{
		config:            config,
		graph:             NewPlayerInteractionGraph(),
		pairwiseScorer:    NewPairwiseSoftPlayScorer(config),
		chipFlowAnalyzer:  NewChipFlowAnalyzer(config),
		communityDetector: NewCommunityDetector(),
		alertGenerator:    NewCollusionAlertGenerator(config),
	}
}

// NewPlayerInteractionGraph creates a new player interaction graph
func NewPlayerInteractionGraph() *PlayerInteractionGraph {
	return &PlayerInteractionGraph{
		nodes: make(map[string]*PlayerNode),
		edges: make(map[string]*InteractionEdge),
	}
}

// NewPairwiseSoftPlayScorer creates a pairwise soft-play scorer
func NewPairwiseSoftPlayScorer(config *CollusionDetectionConfig) *PairwiseSoftPlayScorer {
	return &PairwiseSoftPlayScorer{
		config:        config,
		gradientBoost: NewGradientBoostingModel(),
	}
}

// NewGradientBoostingModel creates a gradient boosting model
func NewGradientBoostingModel() *GradientBoostingModel {
	return &GradientBoostingModel{
		trees:        make([]DecisionTree, 0),
		learningRate: 0.1,
		maxDepth:     5,
		numTrees:     100,
	}
}

// NewChipFlowAnalyzer creates a chip flow analyzer
func NewChipFlowAnalyzer(config *CollusionDetectionConfig) *ChipFlowAnalyzer {
	return &ChipFlowAnalyzer{
		config:     config,
		transferDB: nil, // Would be injected in production
	}
}

// NewCommunityDetector creates a community detector
func NewCommunityDetector() *CommunityDetector {
	return &CommunityDetector{
		graph: NewPlayerInteractionGraph(),
	}
}

// NewCollusionAlertGenerator creates an alert generator
func NewCollusionAlertGenerator(config *CollusionDetectionConfig) *CollusionAlertGenerator {
	return &CollusionAlertGenerator{
		config: config,
	}
}

// CollusionDetectionResult contains comprehensive collusion detection results
type CollusionDetectionResult struct {
	IsCollusion       bool
	Score             float64
	PlayerA           string
	PlayerB           string
	RingID            string
	RingMembers       []string
	CollusionType     string // "soft_play", "chip_dumping", "information_sharing", "squeeze_ring"
	SoftPlayScore     float64
	ChipFlowScore     float64
	NetworkScore      float64
	PairwiseFeatures  map[string]float64
	TopEvidence       []EvidenceItem
	RecommendedAction string
	Confidence        float64
	HandEvidence      []HandEvidence
	ChipFlowSummary   *ChipFlowSummary
	GraphSnapshot     *GraphSnapshot
}

// EvidenceItem represents a single piece of evidence
type EvidenceItem struct {
	Type        string
	Description string
	Value       float64
	Threshold   float64
	Severity    string
}

// HandEvidence represents evidence from a specific hand
type HandEvidence struct {
	HandID         string
	Timestamp      time.Time
	Actions        []ActionEvidence
	Potsize        int64
	Winner         string
	Loser          string
	SuspiciousType string
	EVImpact       float64
}

// ActionEvidence represents a suspicious action in a hand
type ActionEvidence struct {
	PlayerID    string
	ActionType  string
	Amount      int64
	EVImpact    float64
	Explanation string
}

// ChipFlowSummary summarizes chip transfer patterns
type ChipFlowSummary struct {
	NetTransfer         int64
	TransferCount       int
	EVLossTotal         float64
	SuspiciousTransfers int
	AverageTransferSize float64
	TransferDirection   string // "A_to_B", "B_to_A", "bidirectional"
}

// GraphSnapshot represents a snapshot of the interaction graph
type GraphSnapshot struct {
	Nodes          []NodeSnapshot
	Edges          []EdgeSnapshot
	Communities    []CommunitySnapshot
	ClusterCoeff   float64
	AveragePathLen float64
}

// NodeSnapshot represents a player node in the graph
type NodeSnapshot struct {
	PlayerID       string
	Degree         int
	WeightedDegree float64
	CommunityID    int
}

// EdgeSnapshot represents an interaction edge
type EdgeSnapshot struct {
	PlayerA         string
	PlayerB         string
	Weight          float64
	InteractionType string
}

// CommunitySnapshot represents a detected community
type CommunitySnapshot struct {
	CommunityID   int
	Members       []string
	Density       float64
	TotalHands    int
	SuspectedRing bool
}

// DetectCollusion performs comprehensive collusion detection
func (cd *CollusionDetector) DetectCollusion(ctx context.Context, playerA, playerB string) *CollusionDetectionResult {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	result := &CollusionDetectionResult{
		IsCollusion:       false,
		Score:             0.0,
		PlayerA:           playerA,
		PlayerB:           playerB,
		PairwiseFeatures:  make(map[string]float64),
		TopEvidence:       make([]EvidenceItem, 0),
		RecommendedAction: "clear",
	}

	// Get interaction edge
	edge := cd.getEdge(playerA, playerB)
	if edge == nil {
		return result
	}

	// Calculate pairwise soft-play score
	softPlayScore := cd.pairwiseScorer.ScoreSoftPlay(edge)
	result.SoftPlayScore = softPlayScore
	result.PairwiseFeatures["aggression_delta"] = edge.AggressionDelta
	result.PairwiseFeatures["pot_delta"] = edge.PotDelta
	result.PairwiseFeatures["showdown_delta"] = edge.ShowdownDelta
	result.PairwiseFeatures["check_down_rate"] = edge.CheckDownRate
	result.PairwiseFeatures["vpip_delta"] = edge.VPIPDelta
	result.PairwiseFeatures["pfr_delta"] = edge.PFRDelta
	result.PairwiseFeatures["three_bet_delta"] = edge.ThreeBetDelta

	// Calculate chip flow score
	chipFlowScore := cd.chipFlowAnalyzer.AnalyzeTransferPattern(playerA, playerB)
	result.ChipFlowScore = chipFlowScore
	result.PairwiseFeatures["net_chip_transfer"] = float64(edge.NetChipTransfer)
	result.PairwiseFeatures["ev_loss_rate"] = edge.EVLossRate
	result.PairwiseFeatures["transfer_count"] = float64(edge.TransferCount)

	// Calculate network score
	networkScore := cd.calculateNetworkScore(edge)
	result.NetworkScore = networkScore
	result.PairwiseFeatures["ip_matches"] = float64(edge.IPMatches)
	result.PairwiseFeatures["device_matches"] = float64(edge.DeviceMatches)
	result.PairwiseFeatures["network_matches"] = float64(edge.NetworkMatches)

	// Calculate co-occurrence score
	coOccurrenceScore := cd.scoreCoOccurrence(edge.CoOccurrences)
	result.PairwiseFeatures["co_occurrence"] = coOccurrenceScore

	// Calculate seating adjacency score
	seatingScore := cd.scoreSeatingAdjacency(edge.SeatingAdjacency)
	result.PairwiseFeatures["seating_adjacency"] = seatingScore

	// Calculate stake overlap score
	stakeOverlapScore := cd.scoreStakeOverlap(edge.StakeOverlap)
	result.PairwiseFeatures["stake_overlap"] = stakeOverlapScore

	// Calculate arrival sync score
	arrivalSyncScore := cd.scoreArrivalSync(edge.ArrivalSyncs)
	result.PairwiseFeatures["arrival_sync"] = arrivalSyncScore

	// Calculate weighted overall score
	result.Score = cd.calculateCollusionScore(coOccurrenceScore, seatingScore, stakeOverlapScore,
		arrivalSyncScore, softPlayScore, chipFlowScore, networkScore)

	// Determine collusion type
	result.CollusionType = cd.determineCollusionType(softPlayScore, chipFlowScore, networkScore)

	// Generate evidence
	result.TopEvidence = cd.generateEvidence(edge, softPlayScore, chipFlowScore, networkScore)

	// Determine action based on score and thresholds
	result.RecommendedAction = cd.determineAction(result.Score)
	result.Confidence = cd.calculateConfidence(edge, result)

	return result
}

// getEdge returns the edge between two players
func (cd *CollusionDetector) getEdge(playerA, playerB string) *InteractionEdge {
	cd.graph.mu.RLock()
	defer cd.graph.mu.RUnlock()

	key := cd.graph.edgeKey(playerA, playerB)
	return cd.graph.edges[key]
}

// edgeKey creates a consistent edge key
func (pg *PlayerInteractionGraph) edgeKey(playerA, playerB string) string {
	if playerA < playerB {
		return playerA + ":" + playerB
	}
	return playerB + ":" + playerA
}

// scoreCoOccurrence scores co-occurrence count
func (cd *CollusionDetector) scoreCoOccurrence(count int) float64 {
	if count >= cd.config.CoOccurrenceThreshold {
		return 1.0
	}
	return float64(count) / float64(cd.config.CoOccurrenceThreshold)
}

// scoreSeatingAdjacency scores seating adjacency frequency
func (cd *CollusionDetector) scoreSeatingAdjacency(count int) float64 {
	if count >= cd.config.SeatingAdjacencyThreshold {
		return 1.0
	}
	return float64(count) / float64(cd.config.SeatingAdjacencyThreshold)
}

// scoreStakeOverlap scores stake selection overlap
func (cd *CollusionDetector) scoreStakeOverlap(overlap float64) float64 {
	if overlap >= cd.config.StakeOverlapThreshold {
		return 1.0
	}
	return overlap / cd.config.StakeOverlapThreshold
}

// scoreArrivalSync scores arrival synchronization
func (cd *CollusionDetector) scoreArrivalSync(count int) float64 {
	// Normalize based on total co-occurrences
	// This is simplified - in production, calculate based on sessions
	return math.Min(1.0, float64(count)*0.1)
}

// calculateNetworkScore calculates network/device correlation score
func (cd *CollusionDetector) calculateNetworkScore(edge *InteractionEdge) float64 {
	score := 0.0

	if edge.IPMatches >= cd.config.IPMatchThreshold {
		score += cd.config.IPMatchWeight
	} else {
		score += float64(edge.IPMatches) / float64(cd.config.IPMatchThreshold) * cd.config.IPMatchWeight
	}

	if edge.DeviceMatches >= cd.config.DeviceMatchThreshold {
		score += cd.config.DeviceMatchWeight
	} else {
		score += float64(edge.DeviceMatches) / float64(cd.config.DeviceMatchThreshold) * cd.config.DeviceMatchWeight
	}

	if edge.NetworkMatches >= cd.config.NetworkMatchThreshold {
		score += cd.config.NetworkMatchWeight
	} else {
		score += float64(edge.NetworkMatches) / float64(cd.config.NetworkMatchThreshold) * cd.config.NetworkMatchWeight
	}

	return math.Min(1.0, score)
}

// calculateCollusionScore calculates weighted collusion score
func (cd *CollusionDetector) calculateCollusionScore(coOccurrence, seating, stake, arrival, softPlay, chipFlow, network float64) float64 {
	return coOccurrence*cd.config.CoOccurrenceWeight +
		seating*cd.config.SeatingAdjacencyWeight +
		stake*cd.config.StakeOverlapWeight +
		arrival*cd.config.ArrivalSyncWeight +
		softPlay*cd.config.AggressionDeltaWeight +
		chipFlow*cd.config.ChipTransferWeight +
		network*cd.config.IPMatchWeight
}

// determineCollusionType determines the type of collusion
func (cd *CollusionDetector) determineCollusionType(softPlay, chipFlow, network float64) string {
	maxScore := softPlay
	collusionType := "soft_play"

	if chipFlow > maxScore {
		maxScore = chipFlow
		collusionType = "chip_dumping"
	}

	if network > maxScore && network > 0.5 {
		collusionType = "information_sharing"
	}

	// Check for squeeze ring (high combination of soft play + chip flow)
	if softPlay > 0.5 && chipFlow > 0.4 {
		collusionType = "squeeze_ring"
	}

	return collusionType
}

// generateEvidence creates evidence items for the detection
func (cd *CollusionDetector) generateEvidence(edge *InteractionEdge, softPlay, chipFlow, network float64) []EvidenceItem {
	evidence := make([]EvidenceItem, 0)

	// Co-occurrence evidence
	if edge.CoOccurrences >= cd.config.CoOccurrenceThreshold {
		evidence = append(evidence, EvidenceItem{
			Type:        "co_occurrence",
			Description: fmt.Sprintf("Played together in %d hands", edge.CoOccurrences),
			Value:       float64(edge.CoOccurrences),
			Threshold:   float64(cd.config.CoOccurrenceThreshold),
			Severity:    "medium",
		})
	}

	// Seating adjacency evidence
	if edge.SeatingAdjacency >= cd.config.SeatingAdjacencyThreshold {
		evidence = append(evidence, EvidenceItem{
			Type:        "seating_adjacency",
			Description: fmt.Sprintf("Seated adjacent %d times", edge.SeatingAdjacency),
			Value:       float64(edge.SeatingAdjacency),
			Threshold:   float64(cd.config.SeatingAdjacencyThreshold),
			Severity:    "medium",
		})
	}

	// Soft play evidence
	if edge.AggressionDelta > cd.config.AggressionDeltaThreshold {
		evidence = append(evidence, EvidenceItem{
			Type:        "aggression_delta",
			Description: fmt.Sprintf("Aggression drops %.0f%% when facing each other", edge.AggressionDelta*100),
			Value:       edge.AggressionDelta,
			Threshold:   cd.config.AggressionDeltaThreshold,
			Severity:    "high",
		})
	}

	if edge.CheckDownRate > cd.config.CheckDownThreshold {
		evidence = append(evidence, EvidenceItem{
			Type:        "check_down",
			Description: fmt.Sprintf("%.0f%% check-down rate in heads-up pots", edge.CheckDownRate*100),
			Value:       edge.CheckDownRate,
			Threshold:   cd.config.CheckDownThreshold,
			Severity:    "high",
		})
	}

	// Chip flow evidence
	if edge.NetChipTransfer != 0 {
		direction := "A_to_B"
		if edge.NetChipTransfer < 0 {
			direction = "B_to_A"
		}
		evidence = append(evidence, EvidenceItem{
			Type:        "chip_transfer",
			Description: fmt.Sprintf("Net chip transfer: %d from %s", int64(math.Abs(float64(edge.NetChipTransfer))), direction),
			Value:       float64(edge.NetChipTransfer),
			Threshold:   float64(cd.config.ChipTransferThreshold),
			Severity:    "high",
		})
	}

	if edge.EVLossRate > cd.config.EVLossThreshold {
		evidence = append(evidence, EvidenceItem{
			Type:        "ev_loss",
			Description: fmt.Sprintf("Consistent EV loss rate: %.1f%%", edge.EVLossRate*100),
			Value:       edge.EVLossRate,
			Threshold:   cd.config.EVLossThreshold,
			Severity:    "high",
		})
	}

	// Network evidence
	if edge.DeviceMatches >= cd.config.DeviceMatchThreshold {
		evidence = append(evidence, EvidenceItem{
			Type:        "device_match",
			Description: fmt.Sprintf("Same device used %d times", edge.DeviceMatches),
			Value:       float64(edge.DeviceMatches),
			Threshold:   float64(cd.config.DeviceMatchThreshold),
			Severity:    "critical",
		})
	}

	// Sort by severity
	sort.Slice(evidence, func(i, j int) bool {
		severityOrder := map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3}
		return severityOrder[evidence[i].Severity] < severityOrder[evidence[j].Severity]
	})

	return evidence
}

// determineAction determines recommended action based on score
func (cd *CollusionDetector) determineAction(score float64) string {
	if score >= cd.config.CriticalScoreThreshold {
		return "immediate_action"
	}
	if score >= cd.config.CollusionScoreThreshold {
		return "flag_review"
	}
	if score >= cd.config.ReviewScoreThreshold {
		return "monitor"
	}
	return "clear"
}

// calculateConfidence calculates confidence in the detection
func (cd *CollusionDetector) calculateConfidence(edge *InteractionEdge, result *CollusionDetectionResult) float64 {
	confidence := 0.0

	// More data = higher confidence
	if edge.CoOccurrences >= 100 {
		confidence += 0.3
	} else if edge.CoOccurrences >= 50 {
		confidence += 0.2
	} else if edge.CoOccurrences >= 20 {
		confidence += 0.1
	}

	// Multiple evidence types = higher confidence
	evidenceTypes := len(result.TopEvidence)
	if evidenceTypes >= 5 {
		confidence += 0.3
	} else if evidenceTypes >= 3 {
		confidence += 0.2
	} else if evidenceTypes >= 1 {
		confidence += 0.1
	}

	// Network evidence adds confidence
	if result.NetworkScore > 0.3 {
		confidence += 0.2
	}

	return math.Min(1.0, confidence)
}

// ScoreSoftPlay scores soft-play behavior between two players
func (pss *PairwiseSoftPlayScorer) ScoreSoftPlay(edge *InteractionEdge) float64 {
	features := map[string]float64{
		"aggression_delta": edge.AggressionDelta,
		"pot_delta":        edge.PotDelta,
		"showdown_delta":   edge.ShowdownDelta,
		"check_down_rate":  edge.CheckDownRate,
		"vpip_delta":       edge.VPIPDelta,
		"pfr_delta":        edge.PFRDelta,
		"three_bet_delta":  edge.ThreeBetDelta,
	}

	// Use gradient boosting to score
	return pss.gradientBoost.Predict(features)
}

// Predict makes a prediction using the gradient boosting model
func (gb *GradientBoostingModel) Predict(features map[string]float64) float64 {
	prediction := 0.0
	for _, tree := range gb.trees {
		prediction += gb.learningRate * tree.root.traverse(features)
	}
	return math.Min(1.0, math.Max(0.0, prediction))
}

// traverse traverses the decision tree
func (node *TreeNode) traverse(features map[string]float64) float64 {
	if node.isLeaf {
		return node.prediction
	}

	value := features[node.feature]
	if value < node.threshold {
		return node.left.traverse(features)
	}
	return node.right.traverse(features)
}

// AnalyzeTransferPattern analyzes chip transfer patterns between players
func (cfa *ChipFlowAnalyzer) AnalyzeTransferPattern(playerA, playerB string) float64 {
	// Simplified implementation - in production, query transfer database
	score := 0.0

	// Check net transfer
	netTransfer := cfa.getNetTransfer(playerA, playerB)
	if math.Abs(float64(netTransfer)) > float64(cfa.config.ChipTransferThreshold) {
		score += 0.4
	}

	// Check transfer frequency
	transferCount := cfa.getTransferCount(playerA, playerB)
	if transferCount > cfa.config.TransferFrequencyThreshold {
		score += 0.3
	}

	// Check EV loss pattern
	evLossRate := cfa.getEVLossRate(playerA, playerB)
	if evLossRate > cfa.config.EVLossThreshold {
		score += 0.3
	}

	return math.Min(1.0, score)
}

// getNetTransfer returns net chip transfer between players
func (cfa *ChipFlowAnalyzer) getNetTransfer(playerA, playerB string) int64 {
	// In production, query database
	return 0
}

// getTransferCount returns number of transfers between players
func (cfa *ChipFlowAnalyzer) getTransferCount(playerA, playerB string) int {
	// In production, query database
	return 0
}

// getEVLossRate returns EV loss rate between players
func (cfa *ChipFlowAnalyzer) getEVLossRate(playerA, playerB string) float64 {
	// In production, query database
	return 0.0
}

// FindCollusionRings finds groups of colluding players using community detection
func (cd *CollusionDetector) FindCollusionRings(ctx context.Context, minConfidence float64) []CollusionRing {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	// Use Louvain community detection
	communities := cd.communityDetector.DetectCommunities(cd.graph)

	rings := make([]CollusionRing, 0)
	for _, community := range communities {
		if community.Density > 0.5 && community.TotalHands >= 100 {
			ring := CollusionRing{
				RingID:        fmt.Sprintf("ring_%d", community.CommunityID),
				Members:       community.Members,
				Density:       community.Density,
				TotalHands:    community.TotalHands,
				Confidence:    cd.calculateRingConfidence(community),
				CollusionType: "multi_player_ring",
			}

			if ring.Confidence >= minConfidence {
				rings = append(rings, ring)
			}
		}
	}

	return rings
}

// CollusionRing represents a detected collusion ring
type CollusionRing struct {
	RingID        string
	Members       []string
	Density       float64
	TotalHands    int
	Confidence    float64
	CollusionType string
	SqueezePlays  int
	ChipDumps     int
	SoftPlays     int
}

// DetectCommunities performs community detection on the graph
func (cd *CommunityDetector) DetectCommunities(graph *PlayerInteractionGraph) []DetectedCommunity {
	// Simplified Louvain implementation
	// In production, use proper Louvain/Leiden algorithm

	communities := make([]DetectedCommunity, 0)
	visited := make(map[string]bool)

	for nodeID := range graph.nodes {
		if visited[nodeID] {
			continue
		}

		community := DetectedCommunity{
			CommunityID: len(communities),
			Members:     make([]string, 0),
			Density:     0.0,
			TotalHands:  0,
		}

		// BFS to find connected component
		queue := []string{nodeID}
		visited[nodeID] = true

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			community.Members = append(community.Members, current)

			// Find connected nodes
			for edgeKey := range graph.edges {
				var neighbor string
				if strings.HasPrefix(edgeKey, current+":") {
					neighbor = strings.Split(edgeKey, ":")[1]
				} else if strings.HasSuffix(edgeKey, ":"+current) {
					parts := strings.Split(edgeKey, ":")
					neighbor = parts[0]
				}

				if neighbor != "" && !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}

		// Calculate density
		if len(community.Members) >= 2 {
			community.Density = cd.calculateDensity(graph, community.Members)
		}

		communities = append(communities, community)
	}

	return communities
}

// DetectedCommunity represents a detected community
type DetectedCommunity struct {
	CommunityID int
	Members     []string
	Density     float64
	TotalHands  int
}

// calculateDensity calculates edge density of a community
func (cd *CommunityDetector) calculateDensity(graph *PlayerInteractionGraph, members []string) float64 {
	if len(members) < 2 {
		return 0.0
	}

	possibleEdges := float64(len(members) * (len(members) - 1) / 2)
	actualEdges := 0.0

	for i, memberA := range members {
		for _, memberB := range members[i+1:] {
			key := graph.edgeKey(memberA, memberB)
			if edge, exists := graph.edges[key]; exists {
				if edge.Weight > 0.3 { // Threshold for "strong" connection
					actualEdges++
				}
			}
		}
	}

	if possibleEdges == 0 {
		return 0.0
	}

	return actualEdges / possibleEdges
}

// calculateRingConfidence calculates confidence for a detected ring
func (cd *CollusionDetector) calculateRingConfidence(community DetectedCommunity) float64 {
	// Confidence based on density and size
	baseConfidence := community.Density

	// Boost for larger rings (3+ players)
	if len(community.Members) >= 4 {
		baseConfidence += 0.1
	}

	return math.Min(1.0, baseConfidence)
}

// GenerateEvidencePacket generates comprehensive evidence for a detected ring
func (cag *CollusionAlertGenerator) GenerateEvidencePacket(ring CollusionRing) *EvidencePacket {
	packet := &EvidencePacket{
		RingID:        ring.RingID,
		GeneratedAt:   time.Now(),
		Members:       ring.Members,
		Confidence:    ring.Confidence,
		CollusionType: ring.CollusionType,
	}

	// Generate graph snapshot
	packet.GraphSnapshot = &GraphSnapshot{
		Nodes: make([]NodeSnapshot, len(ring.Members)),
		Edges: make([]EdgeSnapshot, 0),
	}

	for i, member := range ring.Members {
		packet.GraphSnapshot.Nodes[i] = NodeSnapshot{
			PlayerID: member,
			Degree:   i + 1, // Simplified
		}
	}

	// Generate chip flow summary
	packet.ChipFlowSummary = &ChipFlowSummary{
		NetTransfer:         0,
		TransferCount:       0,
		SuspiciousTransfers: 0,
	}

	return packet
}

// EvidencePacket represents a complete evidence packet for enforcement
type EvidencePacket struct {
	RingID          string
	GeneratedAt     time.Time
	Members         []string
	Confidence      float64
	CollusionType   string
	GraphSnapshot   *GraphSnapshot
	ChipFlowSummary *ChipFlowSummary
	HandEvidence    []HandEvidence
	ModelFactors    map[string]float64
	Recommendations []string
}

// AddInteractionEdge adds or updates an interaction edge
func (pg *PlayerInteractionGraph) AddInteractionEdge(edge *InteractionEdge) {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	key := pg.edgeKey(edge.PlayerA, edge.PlayerB)
	if existing, exists := pg.edges[key]; exists {
		// Merge updates
		existing.CoOccurrences += edge.CoOccurrences
		existing.SeatingAdjacency += edge.SeatingAdjacency
		existing.NetChipTransfer += edge.NetChipTransfer
		existing.IPMatches += edge.IPMatches
		existing.DeviceMatches += edge.DeviceMatches
		existing.LastInteraction = time.Now()
	} else {
		pg.edges[key] = edge
	}

	// Ensure nodes exist
	if _, exists := pg.nodes[edge.PlayerA]; !exists {
		pg.nodes[edge.PlayerA] = &PlayerNode{
			PlayerID:     edge.PlayerA,
			Sessions:     make(map[string]bool),
			StakesPlayed: make(map[int64]int),
		}
	}
	if _, exists := pg.nodes[edge.PlayerB]; !exists {
		pg.nodes[edge.PlayerB] = &PlayerNode{
			PlayerID:     edge.PlayerB,
			Sessions:     make(map[string]bool),
			StakesPlayed: make(map[int64]int),
		}
	}
}

// SerializeToJSON serializes the graph to JSON
func (pg *PlayerInteractionGraph) SerializeToJSON() ([]byte, error) {
	type GraphExport struct {
		Nodes map[string]*PlayerNode      `json:"nodes"`
		Edges map[string]*InteractionEdge `json:"edges"`
	}

	export := GraphExport{
		Nodes: pg.nodes,
		Edges: pg.edges,
	}

	return json.Marshal(export)
}
