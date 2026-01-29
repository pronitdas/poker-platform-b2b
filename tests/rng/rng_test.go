package rng

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	audit := NewAuditLogger()
	system, err := NewSystem(audit)

	if err != nil {
		t.Fatalf("Failed to create RNG system: %v", err)
	}

	if system == nil {
		t.Fatal("RNG system should not be nil")
	}
}

func TestRandomUint64(t *testing.T) {
	audit := NewAuditLogger()
	system, err := NewSystem(audit)
	if err != nil {
		t.Fatalf("Failed to create RNG system: %v", err)
	}

	// Generate multiple random numbers
	nums := make(map[uint64]bool)
	for i := 0; i < 1000; i++ {
		num := system.RandomUint64()
		if nums[num] {
			t.Errorf("Duplicate random number generated: %d", num)
		}
		nums[num] = true
	}
}

func TestRandomInt(t *testing.T) {
	audit := NewAuditLogger()
	system, err := NewSystem(audit)
	if err != nil {
		t.Fatalf("Failed to create RNG system: %v", err)
	}

	max := 100
	counts := make([]int, max)

	for i := 0; i < 10000; i++ {
		num := system.RandomInt(max)
		if num < 0 || num >= max {
			t.Errorf("RandomInt out of range: %d", num)
		}
		counts[num]++
	}

	// Check for reasonable distribution (chi-square test approximation)
	for i, count := range counts {
		expected := 10000 / max
		if count < expected/2 || count > expected*2 {
			t.Errorf("Unreasonable distribution at index %d: got %d, expected around %d", i, count, expected)
		}
	}
}

func TestRandomBytes(t *testing.T) {
	audit := NewAuditLogger()
	system, err := NewSystem(audit)
	if err != nil {
		t.Fatalf("Failed to create RNG system: %v", err)
	}

	// Test various sizes
	sizes := []int{16, 32, 64, 128}

	for _, size := range sizes {
		bytes, err := system.RandomBytes(size)
		if err != nil {
			t.Errorf("Failed to generate %d random bytes: %v", size, err)
		}

		if len(bytes) != size {
			t.Errorf("Wrong number of bytes: got %d, expected %d", len(bytes), size)
		}

		// All zeros check
		allZero := true
		for _, b := range bytes {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Errorf("Generated all-zero bytes for size %d", size)
		}
	}
}

func TestAuditLogger(t *testing.T) {
	audit := NewAuditLogger()
	if audit.enabled != true {
		t.Error("Audit logger should be enabled by default")
	}

	// Test logging (should not panic)
	event := &ShuffleAuditEvent{
		Timestamp:  time.Now(),
		TableID:    "test-table",
		HandID:     "hand-1",
		Algorithm:  "Fisher-Yates",
		PRNG:       "AES-CTR-256",
	}
	err := audit.LogShuffleEvent(event)
	if err != nil {
		t.Errorf("Failed to log event: %v", err)
	}
}

func TestDeterministicWithSeed(t *testing.T) {
	seed := []byte("test-seed-1234567890123456")
	audit := NewAuditLogger()

	system1, err := NewSystemWithSeed(seed, audit)
	if err != nil {
		t.Fatalf("Failed to create first system: %v", err)
	}

	system2, err := NewSystemWithSeed(seed, audit)
	if err != nil {
		t.Fatalf("Failed to create second system: %v", err)
	}

	// Generate same sequence
	for i := 0; i < 100; i++ {
		if system1.RandomUint64() != system2.RandomUint64() {
			t.Errorf("Systems generated different values at index %d", i)
		}
	}
}

func TestDifferentSeeds(t *testing.T) {
	audit := NewAuditLogger()

	seed1 := []byte("seed-1-1234567890123456")
	seed2 := []byte("seed-2-1234567890123456")

	system1, err := NewSystemWithSeed(seed1, audit)
	if err != nil {
		t.Fatalf("Failed to create first system: %v", err)
	}

	system2, err := NewSystemWithSeed(seed2, audit)
	if err != nil {
		t.Fatalf("Failed to create second system: %v", err)
	}

	// Generate different sequences
	allSame := true
	for i := 0; i < 100; i++ {
		if system1.RandomUint64() != system2.RandomUint64() {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Systems with different seeds should generate different sequences")
	}
}

func TestCreateAuditEntry(t *testing.T) {
	audit := NewAuditLogger()
	system, err := NewSystem(audit)
	if err != nil {
		t.Fatalf("Failed to create RNG system: %v", err)
	}

	entry := system.CreateAuditEntry(
		"table-1",
		"hand-123",
		"dealer-1",
		"server-1",
		[]int{0, 1, 2, 3, 4},
		[]int{51, 50, 49, 48, 47},
	)

	if entry.TableID != "table-1" {
		t.Errorf("Expected TableID table-1, got %s", entry.TableID)
	}

	if entry.HandID != "hand-123" {
		t.Errorf("Expected HandID hand-123, got %s", entry.HandID)
	}

	if entry.Algorithm != "Fisher-Yates" {
		t.Errorf("Expected Algorithm Fisher-Yates, got %s", entry.Algorithm)
	}

	if entry.PRNG != "AES-CTR-256" {
		t.Errorf("Expected PRNG AES-CTR-256, got %s", entry.PRNG)
	}

	if entry.Seed == "" {
		t.Error("Seed should not be empty")
	}

	if entry.SeedHash == "" {
		t.Error("SeedHash should not be empty")
	}
}
