package game

import (
	"context"
	"testing"
	"time"

	"poker-platform/internal/game/rules"
	"poker-platform/pkg/poker"
)

func TestNewTable(t *testing.T) {
	config := rules.TableConfig{
		TableID:       "test-table",
		GameType:      rules.GameTypeTexasHoldem,
		BettingType:   rules.BettingTypeNoLimit,
		MinPlayers:    2,
		MaxPlayers:    9,
		SmallBlind:    1,
		BigBlind:      2,
		BuyInMin:      100,
		BuyInMax:      1000,
		ActionTimeout: 30 * time.Second,
	}

	table, err := NewTable(config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if table == nil {
		t.Fatal("expected non-nil table")
	}

	if table.config.TableID != "test-table" {
		t.Errorf("expected TableID 'test-table', got '%s'", table.config.TableID)
	}

	if table.config.MinPlayers != 2 {
		t.Errorf("expected MinPlayers 2, got %d", table.config.MinPlayers)
	}

	if table.config.MaxPlayers != 9 {
		t.Errorf("expected MaxPlayers 9, got %d", table.config.MaxPlayers)
	}
}

func TestPlayerJoins(t *testing.T) {
	config := rules.TableConfig{
		TableID:       "test-table",
		GameType:      rules.GameTypeTexasHoldem,
		BettingType:   rules.BettingTypeNoLimit,
		MinPlayers:    2,
		MaxPlayers:    9,
		SmallBlind:    1,
		BigBlind:      2,
		BuyInMin:      100,
		BuyInMax:      1000,
		ActionTimeout: 30 * time.Second,
	}

	table, err := NewTable(config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test joining a player
	err = table.PlayerJoins("player1", "Alice", 1000)
	if err != nil {
		t.Fatalf("expected no error joining, got %v", err)
	}

	// Test that player can be retrieved
	state := table.GetState()
	if len(state.Players) == 0 {
		t.Error("expected at least one player")
	}
}

func TestPlayerLeaves(t *testing.T) {
	config := rules.TableConfig{
		TableID:       "test-table",
		GameType:      rules.GameTypeTexasHoldem,
		BettingType:   rules.BettingTypeNoLimit,
		MinPlayers:    2,
		MaxPlayers:    9,
		SmallBlind:    1,
		BigBlind:      2,
		BuyInMin:      100,
		BuyInMax:      1000,
		ActionTimeout: 30 * time.Second,
	}

	table, err := NewTable(config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Add a player
	err = table.PlayerJoins("player1", "Alice", 1000)
	if err != nil {
		t.Fatalf("expected no error joining, got %v", err)
	}

	// Remove the player
	err = table.PlayerLeaves("player1")
	if err != nil {
		t.Fatalf("expected no error leaving, got %v", err)
	}

	// Check player is disconnected
	state := table.GetState()
	found := false
	for _, p := range state.Players {
		if p != nil && p.ID == "player1" {
			found = true
			if p.IsConnected {
				t.Error("expected player to be disconnected")
			}
		}
	}

	if !found {
		t.Error("expected to find player")
	}
}

func TestPlayerSitsOut(t *testing.T) {
	config := rules.TableConfig{
		TableID:       "test-table",
		GameType:      rules.GameTypeTexasHoldem,
		BettingType:   rules.BettingTypeNoLimit,
		MinPlayers:    2,
		MaxPlayers:    9,
		SmallBlind:    1,
		BigBlind:      2,
		BuyInMin:      100,
		BuyInMax:      1000,
		ActionTimeout: 30 * time.Second,
	}

	table, err := NewTable(config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Add a player
	err = table.PlayerJoins("player1", "Alice", 1000)
	if err != nil {
		t.Fatalf("expected no error joining, got %v", err)
	}

	// Sit out
	err = table.PlayerSitsOut("player1")
	if err != nil {
		t.Fatalf("expected no error sitting out, got %v", err)
	}

	// Check player is sitting out
	state := table.GetState()
	found := false
	for _, p := range state.Players {
		if p != nil && p.ID == "player1" {
			found = true
			if p.Status != rules.PlayerSittingOut {
				t.Errorf("expected PlayerSittingOut, got %v", p.Status)
			}
		}
	}

	if !found {
		t.Error("expected to find player")
	}
}

func TestTableStart(t *testing.T) {
	config := rules.TableConfig{
		TableID:       "test-table",
		GameType:      rules.GameTypeTexasHoldem,
		BettingType:   rules.BettingTypeNoLimit,
		MinPlayers:    2,
		MaxPlayers:    9,
		SmallBlind:    1,
		BigBlind:      2,
		BuyInMin:      100,
		BuyInMax:      1000,
		ActionTimeout: 30 * time.Second,
	}

	table, err := NewTable(config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Start the table
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	table.Start(ctx)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Check initial state
	state := table.GetState()
	if state.Phase != rules.PhaseWaiting {
		t.Errorf("expected PhaseWaiting, got %v", state.Phase)
	}
}

// Helper function to create a test deck
func createTestDeck() []poker.Card {
	deck := make([]poker.Card, 52)
	for rank := poker.Rank2; rank <= poker.RankA; rank++ {
		for suit := poker.SuitClubs; suit <= poker.SuitSpades; suit++ {
			deck[int(rank)*4+int(suit)] = poker.NewCard(rank, suit)
		}
	}
	return deck
}
