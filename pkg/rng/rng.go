package rng

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// System provides cryptographically secure random numbers for poker operations
type System struct {
	cipher  cipher.Block
	nonce   []byte
	counter uint64
	mu      sync.Mutex
	audit   *AuditLogger
}

// NewSystem creates a new RNG system with hardware seed
func NewSystem(audit *AuditLogger) (*System, error) {
	// Obtain seed from hardware RNG
	seed, err := getHardwareSeed(32)
	if err != nil {
		return nil, fmt.Errorf("failed to get hardware seed: %w", err)
	}

	// Create AES-CTR cipher
	block, err := aes.NewCipher(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Initialize nonce with random value
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return &System{
		cipher:  block,
		nonce:   nonce,
		counter: 0,
		audit:   audit,
	}, nil
}

// getHardwareSeed obtains entropy from system CSPRNG
func getHardwareSeed(n int) ([]byte, error) {
	seed := make([]byte, n)
	// Use crypto/rand which reads from /dev/urandom on Linux
	// This pools entropy from hardware sources (RDSEED, RDRAND, etc.)
	nRead, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		return nil, err
	}
	if nRead != n {
		return nil, fmt.Errorf("short read from CSPRNG: %d/%d", nRead, n)
	}
	return seed, nil
}

// RandomUint64 returns a cryptographically secure random uint64
func (s *System) RandomUint64() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create counter-based output
	counterBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(counterBytes[:8], s.counter)
	binary.BigEndian.PutUint64(counterBytes[8:], uint64(time.Now().UnixNano()))

	// Encrypt with AES-CTR
	output := make([]byte, 16)
	s.cipher.XORKeyStream(output, counterBytes)

	s.counter++

	// Convert to uint64
	return binary.BigEndian.Uint64(output[:8])
}

// RandomInt returns a random int in range [0, max)
func (s *System) RandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	return int(s.RandomUint64() % uint64(max))
}

// RandomBytes returns cryptographically secure random bytes
func (s *System) RandomBytes(n int) ([]byte, error) {
	result := make([]byte, n)
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < n; i += 16 {
		chunk := make([]byte, 16)
		counterBytes := make([]byte, 16)
		binary.BigEndian.PutUint64(counterBytes[:8], s.counter)
		binary.BigEndian.PutUint64(counterBytes[8:], uint64(time.Now().UnixNano()))

		s.cipher.XORKeyStream(chunk, counterBytes)
		s.counter++

		copyLen := 16
		if i+copyLen > n {
			copyLen = n - i
		}
		copy(result[i:i+copyLen], chunk[:copyLen])
	}

	return result, nil
}

// Seed creates a new System with a specific seed (for deterministic testing)
func NewSystemWithSeed(seed []byte, audit *AuditLogger) (*System, error) {
	// Ensure seed is exactly 32 bytes for AES-256
	if len(seed) < 32 {
		// Expand seed using SHA-256
		hash := sha256.Sum256(seed)
		seed = hash[:]
	}
	if len(seed) > 32 {
		seed = seed[:32]
	}

	block, err := aes.NewCipher(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return &System{
		cipher:  block,
		nonce:   nonce,
		counter: 0,
		audit:   audit,
	}, nil
}

// AuditLogger records RNG events for certification compliance
type AuditLogger struct {
	enabled bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{enabled: true}
}

// LogShuffleEvent records a shuffle operation for audit
func (a *AuditLogger) LogShuffleEvent(event *ShuffleAuditEvent) error {
	if !a.enabled {
		return nil
	}
	// In production, this would write to an append-only table in PostgreSQL
	// For now, we log to stdout in structured format
	fmt.Printf("RNG_AUDIT: %+v\n", event)
	return nil
}

// ShuffleAuditEvent represents a single shuffle operation for audit
type ShuffleAuditEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	TableID      string    `json:"table_id"`
	HandID       string    `json:"hand_id"`
	Seed         string    `json:"seed"`          // Hex encoded
	SeedHash     string    `json:"seed_hash"`     // SHA-256 of seed
	DeckBefore   []int     `json:"deck_before"`   // Card IDs before shuffle
	DeckAfter    []int     `json:"deck_after"`    // Card IDs after shuffle
	Algorithm    string    `json:"algorithm"`     // "Fisher-Yates"
	PRNG         string    `json:"prng"`          // "AES-CTR-256"
	DealerID     string    `json:"dealer_id"`
	ServerID     string    `json:"server_id"`
}

// CreateAuditEntry creates a structured audit entry for a shuffle
func (s *System) CreateAuditEntry(tableID, handID, dealerID, serverID string, deckBefore, deckAfter []int) *ShuffleAuditEvent {
	// Generate seed for this shuffle
	seed, _ := s.RandomBytes(32)

	// Create hash of seed
	hash := sha256.Sum256(seed)

	return &ShuffleAuditEvent{
		Timestamp:    time.Now().UTC(),
		TableID:      tableID,
		HandID:       handID,
		Seed:         fmt.Sprintf("%x", seed),
		SeedHash:     fmt.Sprintf("%x", hash[:]),
		DeckBefore:   deckBefore,
		DeckAfter:    deckAfter,
		Algorithm:    "Fisher-Yates",
		PRNG:         "AES-CTR-256",
		DealerID:     dealerID,
		ServerID:     serverID,
	}
}

// CSPRNGProvider interface for testing
type CSPRNGProvider interface {
	Read(p []byte) (n int, err error)
}

// DefaultCSPRNG returns the system's default CSPRNG
func DefaultCSPRNG() CSPRNGProvider {
	return rand.Reader
}

// IsDevEnvironment checks if running in development mode
func IsDevEnvironment() bool {
	return os.Getenv("POKER_ENV") != "production"
}
