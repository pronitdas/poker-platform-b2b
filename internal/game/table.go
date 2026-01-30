package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"poker-platform/internal/game/rules"
	"poker-platform/pkg/poker"
)

// Table errors
var (
	ErrTableFull        = fmt.Errorf("table is full")
	ErrNoSeatsAvailable = fmt.Errorf("no seats available")
	ErrPlayerNotFound   = fmt.Errorf("player not found")
	ErrNotYourTurn      = fmt.Errorf("not your turn")
	ErrPlayerNotActive  = fmt.Errorf("player is not active")
)

// Table is the main game engine for a poker table
type Table struct {
	config      rules.TableConfig
	state       rules.TableState
	rulesEngine rules.RulesEngine
	actions     chan rules.PlayerActionRequest
	stateChange chan struct{}
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	evaluator   *poker.HandEvaluator
	tickRate    time.Duration
}

// NewTable creates a new poker table with the given configuration
func NewTable(config rules.TableConfig) (*Table, error) {
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
		state:       rules.TableState{Players: make([]*rules.Player, config.MaxPlayers)},
		actions:     make(chan rules.PlayerActionRequest, 10),
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
func (t *Table) GetState() rules.TableState {
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
func (t *Table) copyState() rules.TableState {
	state := t.state
	state.CommunityCards = make([]poker.Card, len(t.state.CommunityCards))
	copy(state.CommunityCards, t.state.CommunityCards)

	state.Pots = make([]rules.Pot, len(t.state.Pots))
	copy(state.Pots, t.state.Pots)

	state.Players = make([]*rules.Player, len(t.state.Players))
	copy(state.Players, t.state.Players)

	state.Deck = make([]poker.Card, len(t.state.Deck))
	copy(state.Deck, t.state.Deck)

	state.PlayersToAct = make([]int, len(t.state.PlayersToAct))
	copy(state.PlayersToAct, t.state.PlayersToAct)

	state.PlayersActed = make(map[int]bool)
	for k, v := range t.state.PlayersActed {
		state.PlayersActed[k] = v
	}

	return state
}

// SubmitAction allows a player to submit an action
func (t *Table) SubmitAction(ctx context.Context, action rules.PlayerActionRequest) error {
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
			t.state.Players[i] = &rules.Player{
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
func (t *Table) ChangeGameType(gameType rules.GameType, config rules.TableConfig) error {
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
	if t.allPlayersActed() || t.allActivePlayersAllIn() {
		t.completeBettingRound()
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
	t.collectBlinds()

	// Deal hole cards
	t.rulesEngine.DealHoleCards(&t.state, t.state.Players)

	// Determine first player to act
	positions := t.rulesEngine.CalculatePositions(t.state.Players, t.state.DealerButton, t.config)
	t.state.CurrentPlayer = t.rulesEngine.DetermineFirstActor(rules.PhasePreflop, &t.state, positions)

	// Build players to act list
	t.buildPlayersToActList()

	t.state.PhaseStartTime = time.Now()
}

// collectBlinds collects the small and big blinds
func (t *Table) collectBlinds() {
	// Find positions relative to dealer button
	players := t.getActivePlayers()
	if len(players) < 2 {
		return
	}

	// Small blind is first player after dealer button
	// Big blind is second player after dealer button
	sbPos := (t.state.DealerButton + 1) % len(t.state.Players)
	bbPos := (t.state.DealerButton + 2) % len(t.state.Players)

	for sbPos == bbPos {
		bbPos = (bbPos + 1) % len(t.state.Players)
	}

	// Collect small blind
	if t.state.Players[sbPos] != nil && t.state.Players[sbPos].IsConnected {
		amount := t.config.SmallBlind
		if t.state.Players[sbPos].Chips < amount {
			amount = t.state.Players[sbPos].Chips
		}
		t.state.Players[sbPos].Chips -= amount
		t.state.Players[sbPos].CurrentBet = amount
		t.state.Players[sbPos].TotalInvested += amount
		t.state.LastBet = amount
	}

	// Collect big blind
	if t.state.Players[bbPos] != nil && t.state.Players[bbPos].IsConnected {
		amount := t.config.BigBlind
		if t.state.Players[bbPos].Chips < amount {
			amount = t.state.Players[bbPos].Chips
			t.state.Players[bbPos].Status = rules.PlayerAllIn
		}
		t.state.Players[bbPos].Chips -= amount
		t.state.Players[bbPos].CurrentBet = amount
		t.state.Players[bbPos].TotalInvested += amount
		t.state.LastBet = amount
	}

	t.state.MinRaise = t.config.BigBlind
	t.rulesEngine.UpdatePot(&t.state)
}

// handleAction processes a player action
func (t *Table) handleAction(action rules.PlayerActionRequest) {
	// Validate action
	player := t.getPlayerByID(action.PlayerID)
	if player == nil {
		return
	}

	if err := t.rulesEngine.ValidateAction(action, &t.state, player, t.config); err != nil {
		return
	}

	// Process action through rules engine
	t.rulesEngine.ProcessAction(action, &t.state, player, t.config)

	// Move to next player if needed
	if t.rulesEngine.IsBettingPhase(t.state.Phase) {
		t.advanceToNextPlayer()
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

// completeBettingRound advances to the next betting phase
func (t *Table) completeBettingRound() {
	switch t.state.Phase {
	case rules.PhasePreflop:
		t.state.Phase = rules.PhaseFlop
		t.rulesEngine.DealCommunityCards(&t.state, rules.PhaseFlop)
	case rules.PhaseFlop:
		t.state.Phase = rules.PhaseTurn
		t.rulesEngine.DealCommunityCards(&t.state, rules.PhaseTurn)
	case rules.PhaseTurn:
		t.state.Phase = rules.PhaseRiver
		t.rulesEngine.DealCommunityCards(&t.state, rules.PhaseRiver)
	case rules.PhaseRiver:
		t.state.Phase = rules.PhaseShowdown
	default:
		return
	}

	// Reset betting state
	for _, p := range t.state.Players {
		if p != nil {
			p.CurrentBet = 0
		}
	}
	t.state.LastBet = 0
	t.state.MinRaise = t.config.BigBlind

	// Set first player to act in new round
	t.buildPlayersToActList()
}

// getPlayerByID returns a player by ID
func (t *Table) getPlayerByID(playerID string) *rules.Player {
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

// getLastPlayerNotFolded returns the last player who didn't fold
func (t *Table) getLastPlayerNotFolded() *rules.Player {
	var last *rules.Player
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

// getActivePlayers returns indices of active (connected, not sitting out) players
func (t *Table) getActivePlayers() []int {
	var indices []int
	for i, p := range t.state.Players {
		if p != nil && p.IsConnected && p.Status != rules.PlayerSittingOut {
			indices = append(indices, i)
		}
	}
	return indices
}

// allPlayersActed returns true if all active players have acted this round
func (t *Table) allPlayersActed() bool {
	players := t.getActivePlayers()
	if len(players) == 0 {
		return true
	}

	for _, idx := range players {
		if t.state.Players[idx] != nil &&
			t.state.Players[idx].Status == rules.PlayerActive &&
			!t.state.PlayersActed[idx] {
			return false
		}
	}

	return true
}

// allActivePlayersAllIn returns true if all active players are all-in
func (t *Table) allActivePlayersAllIn() bool {
	players := t.getActivePlayers()
	if len(players) == 0 {
		return true
	}

	for _, idx := range players {
		if t.state.Players[idx] != nil && t.state.Players[idx].Status == rules.PlayerActive {
			return false
		}
	}

	return true
}

// advanceToNextPlayer moves to the next player who needs to act
func (t *Table) advanceToNextPlayer() {
	players := t.getActivePlayers()

	currentIdx := -1
	for i, idx := range players {
		if idx == t.state.CurrentPlayer {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 {
		return
	}

	for i := 1; i <= len(players); i++ {
		nextIdx := players[(currentIdx+i)%len(players)]
		if t.state.Players[nextIdx] != nil && t.state.Players[nextIdx].Status == rules.PlayerActive {
			t.state.CurrentPlayer = nextIdx
			return
		}
	}
}

// buildPlayersToActList builds the list of players who need to act this round
func (t *Table) buildPlayersToActList() {
	t.state.PlayersToAct = nil
	t.state.PlayersActed = make(map[int]bool)

	players := t.getActivePlayers()
	if len(players) == 0 {
		return
	}

	// Start from current player and go around
	idx := 0
	for i, p := range players {
		if p == t.state.CurrentPlayer {
			idx = i
			break
		}
	}

	// Add all players starting from current player
	for i := 0; i < len(players); i++ {
		playerIdx := players[(idx+i)%len(players)]
		if t.state.Players[playerIdx] != nil && t.state.Players[playerIdx].Status == rules.PlayerActive {
			t.state.PlayersToAct = append(t.state.PlayersToAct, playerIdx)
		}
	}
}
