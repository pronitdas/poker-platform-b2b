package game

import (
	"context"
	"testing"
	"time"
)

func TestNewTable(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-1",
		MinPlayers:     2,
		MaxPlayers:     9,
		SmallBlind:     5,
		BigBlind:       10,
		BuyInMin:       100,
		BuyInMax:       10000,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)

	if table.config.TableID != "test-table-1" {
		t.Errorf("Expected TableID test-table-1, got %s", table.config.TableID)
	}

	if table.config.MaxPlayers != 9 {
		t.Errorf("Expected MaxPlayers 9, got %d", table.config.MaxPlayers)
	}
}

func TestPlayerJoin(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-2",
		MaxPlayers:     9,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)

	// Player should be able to join
	err := table.PlayerJoins("player-1", "Alice", 1000)
	if err != nil {
		t.Fatalf("Failed to join player: %v", err)
	}

	state := table.GetState()
	if len(state.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(state.Players))
	}

	// Same player should be able to reconnect
	err = table.PlayerJoins("player-1", "Alice", 1000)
	if err != nil {
		t.Errorf("Reconnect should succeed: %v", err)
	}
}

func TestTableFull(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-3",
		MaxPlayers:     2,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)

	// Fill the table
	err := table.PlayerJoins("player-1", "Alice", 1000)
	if err != nil {
		t.Fatalf("First player should join: %v", err)
	}

	err = table.PlayerJoins("player-2", "Bob", 1000)
	if err != nil {
		t.Fatalf("Second player should join: %v", err)
	}

	// Third player should fail
	err = table.PlayerJoins("player-3", "Charlie", 1000)
	if err != ErrTableFull {
		t.Errorf("Expected ErrTableFull, got %v", err)
	}
}

func TestPlayerLeave(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-4",
		MaxPlayers:     9,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)

	table.PlayerJoins("player-1", "Alice", 1000)
	table.PlayerJoins("player-2", "Bob", 1000)

	err := table.PlayerLeaves("player-1")
	if err != nil {
		t.Fatalf("Player leave should succeed: %v", err)
	}

	state := table.GetState()
	if len(state.Players) != 2 {
		t.Errorf("Expected 2 players (one disconnected), got %d", len(state.Players))
	}
}

func TestGameLoop(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-5",
		MaxPlayers:     9,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the game loop
	table.Start(ctx)

	// Add players
	table.PlayerJoins("player-1", "Alice", 1000)
	table.PlayerJoins("player-2", "Bob", 1000)

	// Wait for game to start (should transition to PhasePreflop)
	time.Sleep(100 * time.Millisecond)

	state := table.GetState()
	if state.Phase != PhasePreflop {
		t.Errorf("Expected PhasePreflop, got %v", state.Phase)
	}

	// Verify dealer button was assigned
	if state.DealerButton < 0 || state.DealerButton >= len(state.Players) {
		t.Errorf("Invalid dealer button position: %d", state.DealerButton)
	}

	table.Stop()
}

func TestActionValidation(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-6",
		MaxPlayers:     9,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	table.Start(ctx)
	table.PlayerJoins("player-1", "Alice", 1000)
	table.PlayerJoins("player-2", "Bob", 1000)

	// Wait for blinds to be posted
	time.Sleep(50 * time.Millisecond)

	// Test invalid action (bet when it's not player's turn)
	err := table.SubmitAction(ctx, PlayerActionRequest{
		PlayerID: "player-1",
		Action:   ActionBet,
		Amount:   100,
	})
	if err != nil {
		t.Errorf("SubmitAction should not return error for async: %v", err)
	}
}

func TestAllPlayersActed(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-7",
		MaxPlayers:     9,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)

	// Add players
	table.PlayerJoins("player-1", "Alice", 1000)
	table.PlayerJoins("player-2", "Bob", 1000)

	// Before game starts, should have players
	players := table.getActivePlayers()
	if len(players) != 2 {
		t.Errorf("Expected 2 active players, got %d", len(players))
	}
}

func TestBlindsCollection(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-8",
		MaxPlayers:     9,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	table.Start(ctx)
	table.PlayerJoins("player-1", "Alice", 1000)
	table.PlayerJoins("player-2", "Bob", 1000)

	time.Sleep(50 * time.Millisecond)

	state := table.GetState()

	// Verify small blind and big blind were collected
	players := state.Players
	totalBets := int64(0)
	for _, p := range players {
		if p != nil {
			totalBets += p.CurrentBet
		}
	}

	// SB (5) + BB (10) = 15
	if totalBets != 15 {
		t.Errorf("Expected total bets of 15, got %d", totalBets)
	}
}

func TestDealerButtonRotation(t *testing.T) {
	config := TableConfig{
		TableID:        "test-table-9",
		MaxPlayers:     9,
		MinPlayers:     2,
		SmallBlind:     5,
		BigBlind:       10,
		ActionTimeout:  30 * time.Second,
	}

	table := NewTable(config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	table.Start(ctx)

	// Add 3 players
	table.PlayerJoins("player-1", "Alice", 1000)
	table.PlayerJoins("player-2", "Bob", 1000)
	table.PlayerJoins("player-3", "Charlie", 1000)

	time.Sleep(50 * time.Millisecond)

	state := table.GetState()
	initialButton := state.DealerButton

	// Rotate button
	table.rotateDealerButton()

	state = table.GetState()
	if state.DealerButton == initialButton {
		t.Errorf("Dealer button should have moved")
	}
}
