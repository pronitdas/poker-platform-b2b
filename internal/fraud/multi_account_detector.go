package fraud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// MultiAccountConfig holds configuration for multi-account detection
type MultiAccountConfig struct {
	// Detection thresholds
	SameDeviceThreshold     int // Accounts on same device = suspicious
	SameIPThreshold         int // Accounts from same IP = suspicious
	SameNetworkThreshold    int // Accounts from same /24 subnet = suspicious
	AccountAgeWindow        time.Duration
	SessionOverlapThreshold float64 // Session overlap percentage

	// Score weights
	DeviceMatchWeight     float64
	IPMatchWeight         float64
	NetworkMatchWeight    float64
	SessionOverlapWeight  float64
	BehavioralMatchWeight float64

	// Score thresholds
	MultiAccountScoreThreshold float64
	ReviewScoreThreshold       float64
}

// DefaultMultiAccountConfig returns default configuration
func DefaultMultiAccountConfig() *MultiAccountConfig {
	return &MultiAccountConfig{
		SameDeviceThreshold:     3,                   // 3+ accounts on same device
		SameIPThreshold:         10,                  // 10+ accounts from same IP
		SameNetworkThreshold:    20,                  // 20+ accounts from same /24
		AccountAgeWindow:        30 * 24 * time.Hour, // 30 days
		SessionOverlapThreshold: 0.8,                 // 80% session overlap

		DeviceMatchWeight:     0.40, // Very strong signal
		IPMatchWeight:         0.30, // Strong signal
		NetworkMatchWeight:    0.15, // Medium signal
		SessionOverlapWeight:  0.10, // Medium signal
		BehavioralMatchWeight: 0.05, // Weak signal

		MultiAccountScoreThreshold: 0.75,
		ReviewScoreThreshold:       0.50,
	}
}

// MultiAccountDetector detects multiple accounts by the same player
type MultiAccountDetector struct {
	config        *MultiAccountConfig
	fingerprintDB FingerprintDatabase
	ipTracker     IPTracker
	sessionStore  SessionStore
	mu            sync.RWMutex
}

// FingerprintDatabase stores and queries device fingerprints
type FingerprintDatabase interface {
	FindAccountsByFingerprint(fingerprint string) ([]string, error)
	FindAccountsByIP(ip string, networkMask int) ([]string, error)
	StoreFingerprint(fp DeviceFingerprint) error
	GetFingerprintHistory(playerID string) ([]DeviceFingerprint, error)
}

// IPTracker tracks IP addresses for multi-account detection
type IPTracker interface {
	RecordIPUsage(playerID, ipAddress string, timestamp time.Time) error
	GetPlayerIPs(playerID string) ([]string, error)
	GetIPPlayers(ip string) ([]string, error)
	GetNetworkPlayers(network string) ([]string, error)
}

// SessionStore stores player sessions for overlap analysis
type SessionStore interface {
	GetPlayerSessions(playerID string, startTime, endTime time.Time) ([]PlayerSession, error)
	GetAllPlayerIDs() ([]string, error)
}

// MultiAccountDetectionResult contains the result of multi-account detection
type MultiAccountDetectionResult struct {
	IsMultiAccount    bool
	Score             float64 // 0-1, higher = more likely multi-accounting
	PlayerID          string
	RelatedAccounts   []RelatedAccount
	Evidence          []string
	RecommendedAction string
}

// RelatedAccount represents a potentially related account
type RelatedAccount struct {
	PlayerID       string
	Similarity     float64
	ConnectionType string // "device", "ip", "network", "behavioral"
	Evidence       []string
	FirstSeen      time.Time
	LastSeen       time.Time
}

// NewMultiAccountDetector creates a new multi-account detector
func NewMultiAccountDetector(
	config *MultiAccountConfig,
	fingerprintDB FingerprintDatabase,
	ipTracker IPTracker,
	sessionStore SessionStore,
) *MultiAccountDetector {
	if config == nil {
		config = DefaultMultiAccountConfig()
	}
	return &MultiAccountDetector{
		config:        config,
		fingerprintDB: fingerprintDB,
		ipTracker:     ipTracker,
		sessionStore:  sessionStore,
	}
}

// DetectMultiAccount checks if a player has multiple accounts
func (mad *MultiAccountDetector) DetectMultiAccount(ctx context.Context, playerID string) *MultiAccountDetectionResult {
	mad.mu.RLock()
	defer mad.mu.RUnlock()

	result := &MultiAccountDetectionResult{
		IsMultiAccount:    false,
		Score:             0.0,
		PlayerID:          playerID,
		RelatedAccounts:   make([]RelatedAccount, 0),
		Evidence:          make([]string, 0),
		RecommendedAction: "clear",
	}

	// Get device fingerprint history
	fingerprints, err := mad.fingerprintDB.GetFingerprintHistory(playerID)
	if err == nil && len(fingerprints) > 0 {
		currentFP := fingerprints[len(fingerprints)-1]
		mad.checkDeviceMatches(playerID, currentFP.Fingerprint, result)
	}

	// Get IP addresses used by this player
	ips, _ := mad.ipTracker.GetPlayerIPs(playerID)
	for _, ip := range ips {
		mad.checkIPMatches(playerID, ip, result)
	}

	// Check for session overlaps
	mad.checkSessionOverlaps(playerID, result)

	// Calculate weighted score
	result.Score = mad.calculateMultiAccountScore(result)

	// Determine action
	if result.Score >= mad.config.MultiAccountScoreThreshold {
		result.IsMultiAccount = true
		result.RecommendedAction = "flag"
	} else if result.Score >= mad.config.ReviewScoreThreshold {
		result.RecommendedAction = "review"
	}

	// Sort related accounts by similarity
	sort.Slice(result.RelatedAccounts, func(i, j int) bool {
		return result.RelatedAccounts[i].Similarity > result.RelatedAccounts[j].Similarity
	})

	return result
}

// checkDeviceMatches finds other accounts using the same device
func (mad *MultiAccountDetector) checkDeviceMatches(playerID, fingerprint string, result *MultiAccountDetectionResult) {
	accounts, err := mad.fingerprintDB.FindAccountsByFingerprint(fingerprint)
	if err != nil || len(accounts) <= 1 {
		return
	}

	for _, accountID := range accounts {
		if accountID == playerID {
			continue
		}

		accountFingerprints, _ := mad.fingerprintDB.GetFingerprintHistory(accountID)
		var lastSeen time.Time
		if len(accountFingerprints) > 0 {
			lastSeen = accountFingerprints[len(accountFingerprints)-1].LastSeen
		}

		related := RelatedAccount{
			PlayerID:       accountID,
			Similarity:     1.0,
			ConnectionType: "device",
			Evidence:       []string{fmt.Sprintf("Same device fingerprint: %s...", fingerprint[:16])},
			FirstSeen:      lastSeen.Add(-30 * 24 * time.Hour), // Approximate
			LastSeen:       lastSeen,
		}
		result.RelatedAccounts = append(result.RelatedAccounts, related)
		result.Evidence = append(result.Evidence, fmt.Sprintf("Account %s shares device with %s", accountID, playerID))
	}
}

// checkIPMatches finds other accounts from the same IP
func (mad *MultiAccountDetector) checkIPMatches(playerID, ip string, result *MultiAccountDetectionResult) {
	accounts, err := mad.ipTracker.GetIPPlayers(ip)
	if err != nil || len(accounts) <= 1 {
		return
	}

	for _, accountID := range accounts {
		if accountID == playerID {
			continue
		}

		// Check if already found via device
		alreadyFound := false
		for _, existing := range result.RelatedAccounts {
			if existing.PlayerID == accountID {
				alreadyFound = true
				break
			}
		}

		if !alreadyFound {
			related := RelatedAccount{
				PlayerID:       accountID,
				Similarity:     0.7,
				ConnectionType: "ip",
				Evidence:       []string{fmt.Sprintf("Connected from same IP: %s", ip)},
			}
			result.RelatedAccounts = append(result.RelatedAccounts, related)
			result.Evidence = append(result.Evidence, fmt.Sprintf("Account %s shares IP %s with %s", accountID, ip, playerID))
		}
	}
}

// checkSessionOverlaps checks for suspicious session overlaps
func (mad *MultiAccountDetector) checkSessionOverlaps(playerID string, result *MultiAccountDetectionResult) {
	playerSessions, err := mad.sessionStore.GetPlayerSessions(
		playerID,
		time.Now().Add(-mad.config.AccountAgeWindow),
		time.Now(),
	)
	if err != nil || len(playerSessions) == 0 {
		return
	}

	allPlayers, _ := mad.sessionStore.GetAllPlayerIDs()
	now := time.Now()

	for _, otherID := range allPlayers {
		if otherID == playerID {
			continue
		}

		otherSessions, err := mad.sessionStore.GetPlayerSessions(
			otherID,
			time.Now().Add(-mad.config.AccountAgeWindow),
			now,
		)
		if err != nil || len(otherSessions) == 0 {
			continue
		}

		overlap := mad.calculateSessionOverlap(playerSessions, otherSessions)
		if overlap >= mad.config.SessionOverlapThreshold {
			// Check if already found via device or IP
			alreadyFound := false
			for _, existing := range result.RelatedAccounts {
				if existing.PlayerID == otherID {
					alreadyFound = true
					break
				}
			}

			if !alreadyFound {
				related := RelatedAccount{
					PlayerID:       otherID,
					Similarity:     overlap * 0.5,
					ConnectionType: "behavioral",
					Evidence:       []string{fmt.Sprintf("%.0f%% session overlap", overlap*100)},
				}
				result.RelatedAccounts = append(result.RelatedAccounts, related)
				result.Evidence = append(result.Evidence, fmt.Sprintf("Account %s has %.0f%% session overlap with %s", otherID, overlap*100, playerID))
			}
		}
	}
}

// calculateSessionOverlap calculates the percentage of time two players are online together
func (mad *MultiAccountDetector) calculateSessionOverlap(sessionsA, sessionsB []PlayerSession) float64 {
	if len(sessionsA) == 0 || len(sessionsB) == 0 {
		return 0.0
	}

	// Calculate total online time
	totalOnlineA := time.Duration(0)
	for _, s := range sessionsA {
		if s.DisconnectedAt != nil {
			totalOnlineA += s.DisconnectedAt.Sub(s.ConnectedAt)
		} else {
			totalOnlineA += 1 * time.Hour // Estimate for current session
		}
	}

	totalOverlap := time.Duration(0)
	for _, a := range sessionsA {
		startA := a.ConnectedAt
		endA := a.ConnectedAt.Add(a.Duration)
		if a.DisconnectedAt != nil {
			endA = *a.DisconnectedAt
		}

		for _, b := range sessionsB {
			startB := b.ConnectedAt
			endB := b.ConnectedAt.Add(b.Duration)
			if b.DisconnectedAt != nil {
				endB = *b.DisconnectedAt
			}

			// Calculate overlap
			overlapStart := startA
			if startB.After(overlapStart) {
				overlapStart = startB
			}
			overlapEnd := endA
			if endB.Before(overlapEnd) {
				overlapEnd = endB
			}

			if overlapEnd.After(overlapStart) {
				totalOverlap += overlapEnd.Sub(overlapStart)
			}
		}
	}

	if totalOnlineA == 0 {
		return 0.0
	}

	return float64(totalOverlap) / float64(totalOnlineA)
}

// calculateMultiAccountScore calculates the overall multi-account score
func (mad *MultiAccountDetector) calculateMultiAccountScore(result *MultiAccountDetectionResult) float64 {
	if len(result.RelatedAccounts) == 0 {
		return 0.0
	}

	score := 0.0
	for _, account := range result.RelatedAccounts {
		switch account.ConnectionType {
		case "device":
			score += account.Similarity * mad.config.DeviceMatchWeight
		case "ip":
			score += account.Similarity * mad.config.IPMatchWeight
		case "network":
			score += account.Similarity * mad.config.NetworkMatchWeight
		case "behavioral":
			score += account.Similarity * mad.config.BehavioralMatchWeight
		}
	}

	return math.Min(1.0, score)
}

// DeviceFingerprintHasher generates secure fingerprints from device characteristics
type DeviceFingerprintHasher struct {
	salt string
}

// NewDeviceFingerprintHasher creates a new fingerprint hasher
func NewDeviceFingerprintHasher(salt string) *DeviceFingerprintHasher {
	if salt == "" {
		salt = "poker-platform-fingerprint-v1"
	}
	return &DeviceFingerprintHasher{salt: salt}
}

// GenerateFingerprint generates a fingerprint from device characteristics
func (dfh *DeviceFingerprintHasher) GenerateFingerprint(components map[string]interface{}) string {
	// Normalize and sort components for consistent hashing
	components["salt"] = dfh.salt

	var parts []string
	for key, value := range components {
		parts = append(parts, fmt.Sprintf("%s:%v", key, value))
	}
	sort.Strings(parts)

	data := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateClientFingerprint generates a fingerprint from client-side characteristics
func GenerateClientFingerprint(
	userAgent, screenResolution string,
	colorDepth int,
	timezone, language, platform string,
	hardwareConcurrency int,
	deviceMemory float64,
	touchSupport bool,
	webGLRenderer string,
) string {
	components := map[string]interface{}{
		"userAgent":           userAgent,
		"screenResolution":    screenResolution,
		"colorDepth":          colorDepth,
		"timezone":            timezone,
		"language":            language,
		"platform":            platform,
		"hardwareConcurrency": hardwareConcurrency,
		"deviceMemory":        deviceMemory,
		"touchSupport":        touchSupport,
		"webGLRenderer":       webGLRenderer,
	}

	hasher := NewDeviceFingerprintHasher("")
	return hasher.GenerateFingerprint(components)
}

// NetworkAnalyzer analyzes IP addresses for network-level relationships
type NetworkAnalyzer struct{}

// NewNetworkAnalyzer creates a new network analyzer
func NewNetworkAnalyzer() *NetworkAnalyzer {
	return &NetworkAnalyzer{}
}

// GetNetworkPrefix returns the /24 network prefix for an IP address
func (na *NetworkAnalyzer) GetNetworkPrefix(ipAddress string) string {
	// Simple /24 extraction - in production, use proper CIDR parsing
	parts := strings.Split(ipAddress, ".")
	if len(parts) != 4 {
		return ipAddress
	}
	return fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
}

// IsSameNetwork checks if two IP addresses are on the same network
func (na *NetworkAnalyzer) IsSameNetwork(ip1, ip2 string) bool {
	return na.GetNetworkPrefix(ip1) == na.GetNetworkPrefix(ip2)
}
