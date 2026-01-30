package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"poker-platform/internal/game/rules"
	"poker-platform/pkg/poker"
)

// Table is the main game engine for a poker table using the rules engine
type Table struct {
	config      TableConfig
	state       TableState
	rulesEngine rules.RulesEngine
	actions     chan PlayerActionRequest
	stateChange chan struct{}
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	evaluator   *poker.HandEvaluator
	tickRate    time.Duration
}

// NewTable creates a new poker table with the given configuration
func NewTable(config TableConfig) (*Table, error) {
	// Create rules engine
	engine, err := rules.GetRegistry().CreateEngine(config.GameType)
	if err != nil {
		return nil, fmt.Errorf("failed to create rules engine: %w", err)
	}

	// Validate and apply config
	if err := engine.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid table config: %w", err)
	}

	// Apply defaults
	if config.GameType == "" {
		config.GameType = rules.GameTypeTexasHoldem
	}
	if config.BettingType == "" {
		config.BettingType = rules.BettingTypeNoLimit
	}
	if config.SmallBlind == 0 {
		config.SmallBlind = 5
	}
	if config.BigBlind == 0 {
		config.BigBlind = 10
	}

	return &Table{
		config:      config,
		actions:     make(chan PlayerActionRequest, 10),
		stateChange: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
		rulesEngine: engine,
		evaluator:   poker.NewHandEvaluator(),
		tickRate:    50 * time.Millisecond,
	}, nil
}

// NewTableWithEngine creates a table with a custom rules engine
func NewTableWithEngine(config TableConfig, engine rules.RulesEngine) (*Table, error) {
	if err := engine.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid table config: %w", err)
	}

	return &Table{
		config:      config,
		actions:     make(chan PlayerActionRequest, 10),
		stateChange: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
		rulesEngine: engine,
		evaluator:   poker.NewHandEvaluator(),
		tickRate:    50 * time.Millisecond,
	}, nil
}

// Start begins the table's game loop in a goroutine
func (t *Table) Start(ctx context.Context) {
	t.wg.Add(1)
	go t.gameLoop(ctx)
}

// Stop gracefully shuts down the table
func (t *Table) Stop() {
	close(t.stopChan)
	t.wg.Wait()
}

// GetState returns a copy of the current table state
func (t *Table) GetState() TableState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.copyState()
}

// GetRulesEngine returns the current rules engine
func (t *Table) GetRulesEngine() rules.RulesEngine {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.rulesEngine
}

// copyState creates a deep copy of the table state
func (t *Table) copyState() TableState {
	state := TableState{
		TableID:        t.config.TableID,
		GameType:       t.config.GameType,
		BettingType:    t.config.BettingType,
		Phase:          t.state.Phase,
		DealerButton:   t.state.DealerButton,
		CurrentPlayer:  t.state.CurrentPlayer,
		CommunityCards: make([]poker.Card, len(t.state.CommunityCards)),
		Pots:           make([]rules.Pot, len(t.state.Pots)),
		Players:        make([]*rules.Player, len(t.state.Players)),
		LastBet:        t.state.LastBet,
		MinRaise:       t.state.MinRaise,
		PotTotal:       t.state.PotTotal,
		HandNumber:     t.state.HandNumber,
		PhaseStartTime: t.state.PhaseStartTime,
		PlayersActed:   make(map[int]bool),
		PlayersToAct:   make([]int, len(t.state.PlayersToAct)),
		Deck:           make([]poker.Card, len(t.state.Deck)),
	}

	copy(state.CommunityCards, t.state.CommunityCards)
	copy(state.Deck, t.state.Deck)
	copy(state.PlayersToAct, t.state.PlayersToAct)

	// Convert internal players to rules players for external use
	for i, p := range t.state.Players {
		if p != nil {
			playerCopy := rules.Player{
				ID:            p.ID,
				Name:          p.Name,
				Chips:         p.Chips,
				HoleCards:     make([]poker.Card, len(p.HoleCards)),
				Status:        rules.PlayerStatus(p.Status),
				CurrentBet:    p.CurrentBet,
				TotalInvested: p.TotalInvested,
				IsConnected:   p.IsConnected,
				IsDealer:      p.IsDealer,
				SeatNumber:    p.SeatNumber,
			}
			copy(playerCopy.HoleCards, p.HoleCards)
			state.Players[i] = &playerCopy
		}
	}

	for k, v := range t.state.PlayersActed {
		state.PlayersActed[k] = v
	}

	return state
}

// SubmitAction allows a player to submit an action
func (t *Table) SubmitAction(ctx context.Context, action PlayerActionRequest) error {
	select {
	case t.actions <- action:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-t.stopChan:
		return nil
	}
}

// PlayerJoins adds a new player to the table
func (t *Table) PlayerJoins(playerID, name string, chips int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if table is full
	activePlayers := t.countActivePlayers()
	if activePlayers >= t.config.MaxPlayers {
		return ErrTableFull
	}

	// Check if player is already at the table
	for _, p := range t.state.Players {
		if p != nil && p.ID == playerID {
			p.IsConnected = true
			p.Status = rules.PlayerActive
			return nil
		}
	}

	// Find empty seat
	for i, p := range t.state.Players {
		if p == nil {
			t.state.Players[i] = &player{
				ID:          playerID,
				Name:        name,
				Chips:       chips,
				Status:      rules.PlayerActive,
				IsConnected: true,
				SeatNumber:  i,
			}
			return nil
		}
	}

	return ErrNoSeatsAvailable
}

// PlayerLeaves removes a player from the table
func (t *Table) PlayerLeaves(playerID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, p := range t.state.Players {
		if p != nil && p.ID == playerID {
			if t.state.Players[i] != nil {
				t.state.Players[i].IsConnected = false
				t.state.Players[i].Status = rules.PlayerDisconnected
			}
			return nil
		}
	}

	return ErrPlayerNotFound
}

// PlayerSitsOut marks a player as sitting out
func (t *Table) PlayerSitsOut(playerID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, p := range t.state.Players {
		if p != nil && p.ID == playerID {
			p.Status = rules.PlayerSittingOut
			return nil
		}
	}

	return ErrPlayerNotFound
}

// ChangeGameType changes the poker variant at the table
func (t *Table) ChangeGameType(gameType rules.GameType, config TableConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	engine, err := rules.GetRegistry().CreateEngine(gameType)
	if err != nil {
		return fmt.Errorf("failed to create rules engine: %w", err)
	}

	config.GameType = gameType
	if err := engine.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	t.rulesEngine = engine
	t.config = config
	t.state.GameType = gameType

	return nil
}

// gameLoop is the main game loop that runs in a goroutine
func (t *Table) gameLoop(ctx context.Context) {
	defer t.wg.Done()

	ticker := time.NewTicker(t.tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.stopChan:
			return
		case action := <-t.actions:
			t.handleAction(action)
		case <-ticker.C:
			t.tick()
		}
	}
}

// tick is called on each tick of the game loop
func (t *Table) tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	phase := t.state.Phase

	if t.rulesEngine.IsBettingPhase(phase) {
		t.handleBettingPhase()
	} else {
		t.handleNonBettingPhase(phase)
	}
}

// handleBettingPhase handles actions during betting phases
func (t *Table) handleBettingPhase() {
	if t.rulesEngine.AllPlayersActed(&t.state) || t.rulesEngine.AllActivePlayersAllIn(&t.state) {
		t.rulesEngine.CompleteBettingRound(&t.state, t.config)
	}
}

// handleNonBettingPhase handles non-betting phases
func (t *Table) handleNonBettingPhase(phase rules.GamePhase) {
	switch phase {
	case rules.PhaseWaiting:
		t.handleWaitingPhase()
	case rules.PhaseShowdown:
		t.handleShowdownPhase()
	case rules.PhaseHandComplete:
		t.handleHandCompletePhase()
	}
}

// handleWaitingPhase checks if enough players are ready to start a hand
func (t *Table) handleWaitingPhase() {
	if t.rulesEngine.ShouldStartHand(t.state.Players, t.config) {
		t.startNewHand()
	}
}

// handleShowdownPhase handles the showdown
func (t *Table) handleShowdownPhase() {
	t.distributePots()
	t.state.Phase = rules.PhaseHandComplete
}

// handleHandCompletePhase handles the end of a hand
func (t *Table) handleHandCompletePhase() {
	t.state.HandNumber++
	t.rulesEngine.RotateDealerButton(&t.state, t.state.Players)
	t.rulesEngine.ResetHandState(&t.state, t.config)

	if t.rulesEngine.ShouldStartHand(t.state.Players, t.config) {
		t.startNewHand()
	} else {
		t.state.Phase = rules.PhaseWaiting
	}
}

// startNewHand initializes a new hand
func (t *Table) startNewHand() {
	t.rulesEngine.ResetHandState(&t.state, t.config)
	t.state.HandNumber++
	t.state.Phase = rules.PhasePreflop

	// Collect blinds
	t.rulesEngine.CollectBlinds(&t.state, t.state.Players, t.config)

	// Deal hole cards
	t.rulesEngine.DealHoleCards(&t.state, t.state.Players)

	// Determine first player to act
	positions := t.rulesEngine.CalculatePositions(t.state.Players, t.state.DealerButton, t.config)
	t.state.CurrentPlayer = t.rulesEngine.DetermineFirstActor(rules.PhasePreflop, &t.state, positions)

	// Build players to act list
	t.rulesEngine.BuildPlayersToActList(&t.state)

	t.state.PhaseStartTime = time.Now()
}

// handleAction processes a player action
func (t *Table) handleAction(action PlayerActionRequest) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Validate action
	player := t.getPlayerByID(action.PlayerID)
	if player == nil {
		return
	}

	// Convert to rules action
	rulesAction := rules.PlayerActionRequest{
		PlayerID: action.PlayerID,
		Action:   rules.PlayerAction(action.Action),
		Amount:   action.Amount,
	}

	if err := t.rulesEngine.ValidateAction(rulesAction, &t.state, player, t.config); err != nil {
		return
	}

	// Process action through rules engine
	t.rulesEngine.ProcessAction(rulesAction, &t.state, player, t.config)

	// Move to next player if needed
	if t.rulesEngine.IsBettingPhase(t.state.Phase) {
		t.rulesEngine.AdvanceToNextPlayer(&t.state)
	}
}

// distributePots distributes the pot to the winner(s)
func (t *Table) distributePots() {
	activePlayers := t.getPlayersNotFolded()
	if len(activePlayers) == 0 {
		// All players folded - give pot to last player who didn't fold
		lastNonFolded := t.getLastPlayerNotFolded()
		if lastNonFolded != nil {
			lastNonFolded.Chips += t.state.PotTotal
		}
		return
	}

	if len(activePlayers) == 1 {
		winner := t.state.Players[activePlayers[0]]
		if winner != nil {
			winner.Chips += t.state.PotTotal
		}
		return
	}

	// Multiple players - evaluate hands
	winners := t.rulesEngine.DetermineWinners(t.state.Players, t.state.CommunityCards, *t.evaluator)

	// Split pot among winners
	t.rulesEngine.DistributePot(&t.state, winners, *t.evaluator)
}

// getPlayerByID returns a player by ID
func (t *Table) getPlayerByID(playerID string) *player {
	for _, p := range t.state.Players {
		if p != nil && p.ID == playerID {
			return p
		}
	}
	return nil
}

// getPlayersNotFolded returns indices of players who haven't folded
func (t *Table) getPlayersNotFolded() []int {
	var indices []int
	for i, p := range t.state.Players {
		if p != nil && p.Status != rules.PlayerFolded && p.IsConnected {
			indices = append(indices, i)
		}
	}
	return indices
}

// getLastPlayerNotFolded returns the index of the last player who didn't fold
func (t *Table) getLastPlayerNotFolded() *player {
	var last *player
	for _, p := range t.state.Players {
		if p != nil && p.Status != rules.PlayerFolded && p.IsConnected {
			last = p
		}
	}
	return last
}

// countActivePlayers returns the number of active players
func (t *Table) countActivePlayers() int {
	count := 0
	for _, p := range t.state.Players {
		if p != nil && p.IsConnected && p.Status != rules.PlayerSittingOut {
			count++
		}
	}
	return count
}

// player is the internal player representation
type player struct {
	ID            string
	Name          string
	Chips         int64
	HoleCards     []poker.Card
	Status        rules.PlayerStatus
	CurrentBet    int64
	TotalInvested int64
	IsConnected   bool
	IsDealer      bool
	SeatNumber    int
}

// InternalTableState represents the internal table state
type InternalTableState struct {
	TableID         string
	GameType        rules.GameType
	BettingType     rules.BettingType
	Phase           rules.GamePhase
	DealerButton    int
	CurrentPlayer   int
	CommunityCards  []poker.Card
	Pots            []rules.Pot
	Players         []*player
	LastBet         int64
	MinRaise        int64
	PotTotal        int64
	Deck            []poker.Card
	HandNumber      int
	PhaseStartTime  time.Time
	PlayersActed    map[int]bool
	PlayersToAct    []int
	SmallBlindPos   int
	BigBlindPos     int
	CurrentBetToCall int64
}
