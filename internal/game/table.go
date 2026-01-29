package game

import (
	"context"
	"poker-platform/pkg/poker"
	"sync"
	"time"
)

// GamePhase represents the current phase of a Texas Hold'em hand
type GamePhase int

const (
	// PhaseWaiting indicates the table is waiting for players
	PhaseWaiting GamePhase = iota
	// PhasePreflop is the first betting round before community cards
	PhasePreflop
	// PhaseFlop is the first three community cards
	PhaseFlop
	// PhaseTurn is the fourth community card
	PhaseTurn
	// PhaseRiver is the fifth community card
	PhaseRiver
	// PhaseShowdown is when hands are evaluated
	PhaseShowdown
	// PhaseHandComplete indicates the hand has finished
	PhaseHandComplete
)

func (p GamePhase) String() string {
	switch p {
	case PhaseWaiting:
		return "waiting"
	case PhasePreflop:
		return "preflop"
	case PhaseFlop:
		return "flop"
	case PhaseTurn:
		return "turn"
	case PhaseRiver:
		return "river"
	case PhaseShowdown:
		return "showdown"
	case PhaseHandComplete:
		return "hand_complete"
	default:
		return "unknown"
	}
}

// PlayerAction represents an action a player can take
type PlayerAction int

const (
	// ActionFold means the player surrenders their hand
	ActionFold PlayerAction = iota
	// ActionCheck means the player passes when no bet is required
	ActionCheck
	// ActionCall means the player matches the current bet
	ActionCall
	// ActionBet means the player places a bet (first action in a round)
	ActionBet
	// ActionRaise means the player increases the current bet
	ActionRaise
	// ActionAllIn means the player bets all their chips
	ActionAllIn
)

func (a PlayerAction) String() string {
	switch a {
	case ActionFold:
		return "fold"
	case ActionCheck:
		return "check"
	case ActionCall:
		return "call"
	case ActionBet:
		return "bet"
	case ActionRaise:
		return "raise"
	case ActionAllIn:
		return "all_in"
	default:
		return "unknown"
	}
}

// PlayerActionRequest represents a player's action request
type PlayerActionRequest struct {
	PlayerID string
	Action   PlayerAction
	Amount   int64 // For bet/raise amounts
}

// SeatPosition represents a player's position at the table
type SeatPosition int

// PlayerStatus represents a player's current state
type PlayerStatus int

const (
	// PlayerActive means the player is still in the hand
	PlayerActive PlayerStatus = iota
	// PlayerFolded means the player has folded
	PlayerFolded
	// PlayerAllIn means the player has gone all-in
	PlayerAllIn
	// PlayerSittingOut means the player is not participating this hand
	PlayerSittingOut
	// PlayerDisconnected means the player has lost connection
	PlayerDisconnected
)

// String returns the player status as a string
func (s PlayerStatus) String() string {
	switch s {
	case PlayerActive:
		return "active"
	case PlayerFolded:
		return "folded"
	case PlayerAllIn:
		return "all_in"
	case PlayerSittingOut:
		return "sitting_out"
	case PlayerDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

// Player represents a player at the table
type Player struct {
	ID          string
	Name        string
	Chips       int64
	HoleCards   [2]poker.Card
	Status      PlayerStatus
	CurrentBet  int64
	TotalInvested int64 // Total chips put into this pot (for all-in calculations)
	IsConnected bool
	IsDealer    bool
}

// Pot represents a pot in the game (main pot or side pot)
type Pot struct {
	Amount     int64
	EligiblePlayers map[string]bool // PlayerIDs eligible to win this pot
	WinnerIDs  []string
}

// TableConfig holds the configuration for a table
type TableConfig struct {
	TableID          string
	MinPlayers       int
	MaxPlayers       int
	SmallBlind       int64
	BigBlind         int64
	BuyInMin         int64
	BuyInMax         int64
	MaxSessionTime   time.Duration
	ActionTimeout    time.Duration
	AutoRebuyEnabled bool
}

// TableState represents the complete state of a poker table
type TableState struct {
	TableID        string
	Phase          GamePhase
	DealerButton   int // Index of player with dealer button
	CurrentPlayer  int // Index of player whose turn it is
	CommunityCards []poker.Card
	Pots           []Pot
	Players        []*Player
	SidePots       []Pot
	LastBet        int64
	MinRaise       int64
	PotTotal       int64
	Deck           []poker.Card
	HandNumber     int
	PhaseStartTime time.Time
	PlayersActed   map[int]bool // Players who have acted in current round
	PlayersToAct   []int        // Order of players to act this round
}

// Table is the main game engine for a poker table
type Table struct {
	config      TableConfig
	state       TableState
	actions     chan PlayerActionRequest
	stateChange chan struct{}
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	evaluator   *poker.HandEvaluator
	tickRate    time.Duration
}

// NewTable creates a new poker table with the given configuration
func NewTable(config TableConfig) *Table {
	return &Table{
		config:      config,
		actions:     make(chan PlayerActionRequest, 10),
		stateChange: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
		evaluator:   poker.NewHandEvaluator(),
		tickRate:    50 * time.Millisecond,
	}
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

// copyState creates a deep copy of the table state
func (t *Table) copyState() TableState {
	state := TableState{
		TableID:        t.state.TableID,
		Phase:          t.state.Phase,
		DealerButton:   t.state.DealerButton,
		CurrentPlayer:  t.state.CurrentPlayer,
		CommunityCards: make([]poker.Card, len(t.state.CommunityCards)),
		Pots:           make([]Pot, len(t.state.Pots)),
		SidePots:       make([]Pot, len(t.state.SidePots)),
		Players:        make([]*Player, len(t.state.Players)),
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

	for i, p := range t.state.Players {
		if p != nil {
			playerCopy := *p
			state.Players[i] = &playerCopy
		}
	}

	for i, pot := range t.state.Pots {
		potCopy := pot
		potCopy.EligiblePlayers = make(map[string]bool)
		for k, v := range pot.EligiblePlayers {
			potCopy.EligiblePlayers[k] = v
		}
		state.Pots[i] = potCopy
	}
		state.Pots[i] = potCopy
	}

	for i, pot := range t.state.SidePots {
		potCopy := pot
		potCopy.EligiblePlayers = make(map[string]bool)
		for k, v := range pot.EligiblePlayers {
			potCopy.EligiblePlayers[k] = v
		}
		state.SidePots[i] = potCopy
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
	activePlayers := 0
	for _, p := range t.state.Players {
		if p != nil && p.IsConnected {
			activePlayers++
		}
	}
	if activePlayers >= t.config.MaxPlayers {
		return ErrTableFull
	}

	// Check if player is already at the table
	for _, p := range t.state.Players {
		if p != nil && p.ID == playerID {
			p.IsConnected = true
			p.Status = PlayerActive
			return nil
		}
	}

	// Find empty seat
	for i, p := range t.state.Players {
		if p == nil {
			t.state.Players[i] = &Player{
				ID:          playerID,
				Name:        name,
				Chips:       chips,
				Status:      PlayerActive,
				IsConnected: true,
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
				t.state.Players[i].Status = PlayerDisconnected
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
			p.Status = PlayerSittingOut
			return nil
		}
	}

	return ErrPlayerNotFound
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

	switch t.state.Phase {
	case PhaseWaiting:
		t.handleWaitingPhase()
	case PhasePreflop, PhaseFlop, PhaseTurn, PhaseRiver:
		t.handleBettingPhase()
	case PhaseShowdown:
		t.handleShowdownPhase()
	case PhaseHandComplete:
		t.handleHandCompletePhase()
	}
}

// handleWaitingPhase checks if enough players are ready to start a hand
func (t *Table) handleWaitingPhase() {
	activePlayers := t.countActivePlayers()
	if activePlayers >= t.config.MinPlayers {
		t.startNewHand()
	}
}

// handleBettingPhase handles actions during betting phases
func (t *Table) handleBettingPhase() {
	// Check if all players have acted or all-in
	if t.allPlayersActed() || t.allActivePlayersAllIn() {
		t.completeBettingRound()
	}
}

// handleShowdownPhase handles the showdown
func (t *Table) handleShowdownPhase() {
	t.distributePots()
	t.state.Phase = PhaseHandComplete
}

// handleHandCompletePhase handles the end of a hand
func (t *Table) handleHandCompletePhase() {
	// Reset for next hand after a short delay
	t.state.HandNumber++
	t.rotateDealerButton()
	t.resetHandState()

	if t.countActivePlayers() >= t.config.MinPlayers {
		t.startNewHand()
	} else {
		t.state.Phase = PhaseWaiting
	}
}

// startNewHand initializes a new hand
func (t *Table) startNewHand() {
	t.resetHandState()
	t.state.HandNumber++
	t.state.Phase = PhasePreflop

	// Collect antes if configured (simplified - no antes for now)
	// Collect blinds
	t.collectBlinds()

	// Deal hole cards
	t.dealHoleCards()

	// Determine first player to act (under the gun)
	t.determineFirstActor()

	t.state.PhaseStartTime = time.Now()
}

// resetHandState resets the state for a new hand
func (t *Table) resetHandState() {
	t.state.CommunityCards = nil
	t.state.Pots = []Pot{{
		Amount:           0,
		EligiblePlayers:  make(map[string]bool),
	}}
	t.state.SidePots = nil
	t.state.LastBet = 0
	t.state.MinRaise = t.config.BigBlind
	t.state.PotTotal = 0
	t.state.Deck = t.createDeck()
	t.state.PlayersActed = make(map[int]bool)
	t.state.PlayersToAct = nil

	for i, p := range t.state.Players {
		if p != nil {
			p.HoleCards = [2]poker.Card{}
			p.CurrentBet = 0
			p.TotalInvested = 0
			if p.Status == PlayerDisconnected {
				p.Status = PlayerSittingOut
			} else if p.IsConnected {
				p.Status = PlayerActive
			}
		}
	}
}

// createDeck creates a new shuffled deck of cards
func (t *Table) createDeck() []poker.Card {
	deck := make([]poker.Card, 52)
	for rank := poker.Rank2; rank <= poker.RankA; rank++ {
		for suit := poker.SuitClubs; suit <= poker.SuitSpades; suit++ {
			deck[int(rank)*4+int(suit)] = poker.NewCard(rank, suit)
		}
	// Shuffle would go here - using simple shuffle for now
	// In production, use pkg/rng/shuffle.go
	return deck
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
			t.state.Players[bbPos].Status = PlayerAllIn
		}
		t.state.Players[bbPos].Chips -= amount
		t.state.Players[bbPos].CurrentBet = amount
		t.state.Players[bbPos].TotalInvested += amount
		t.state.LastBet = amount
	}

	t.state.MinRaise = t.config.BigBlind * 2
}

// dealHoleCards deals two cards to each active player
func (t *Table) dealHoleCards() {
	players := t.getActivePlayers()
	for _, idx := range players {
		if t.state.Players[idx] != nil && t.state.Players[idx].IsConnected {
			t.state.Players[idx].HoleCards[0] = t.state.Deck[0]
			t.state.Players[idx].HoleCards[1] = t.state.Deck[1]
			t.state.Deck = t.state.Deck[2:]
		}
	}
}

// determineFirstActor determines who acts first in the current round
func (t *Table) determineFirstActor() {
	players := t.getActivePlayers()
	if len(players) == 0 {
		return
	}

	// Preflop: first to act is under the gun (first player after big blind)
	// Flop/Turn/River: first to act is small blind or first active player
	if t.state.Phase == PhasePreflop {
		// Find big blind position
		bbPos := (t.state.DealerButton + 2) % len(t.state.Players)
		for bbPos == t.state.DealerButton {
			bbPos = (bbPos + 1) % len(t.state.Players)
		}

		// First to act is the player after big blind
		t.state.CurrentPlayer = (bbPos + 1) % len(t.state.Players)
	} else {
		// For later streets, first to act is the player after dealer button
		t.state.CurrentPlayer = (t.state.DealerButton + 1) % len(t.state.Players)
	}

	t.buildPlayersToActList()
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
		if t.state.Players[playerIdx] != nil &&
			t.state.Players[playerIdx].Status == PlayerActive {
			t.state.PlayersToAct = append(t.state.PlayersToAct, playerIdx)
		}
	}
}

// handleAction processes a player action
func (t *Table) handleAction(action PlayerActionRequest) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Validate action
	if !t.isValidAction(&action) {
		return
	}

	player := t.state.Players[t.state.CurrentPlayer]
	if player == nil || player.ID != action.PlayerID {
		return
	}

	switch action.Action {
	case ActionFold:
		t.processFold(player)
	case ActionCheck:
		t.processCheck(player)
	case ActionCall:
		t.processCall(player)
	case ActionBet:
		t.processBet(player, action.Amount)
	case ActionRaise:
		t.processRaise(player, action.Amount)
	case ActionAllIn:
		t.processAllIn(player)
	}

	// Move to next player if needed
	if t.state.Phase != PhaseShowdown && t.state.Phase != PhaseHandComplete {
		t.advanceToNextPlayer()
	}
}

// isValidAction validates if an action is legal
func (t *Table) isValidAction(action *PlayerActionRequest) bool {
	// Player must be the current player to act
	if t.state.CurrentPlayer < 0 || t.state.CurrentPlayer >= len(t.state.Players) {
		return false
	}

	player := t.state.Players[t.state.CurrentPlayer]
	if player == nil || player.ID != action.PlayerID {
		return false
	}

	// Player must be active
	if player.Status != PlayerActive {
		return false
	}

	// Validate action type based on current bet
	currentHighBet := t.getCurrentHighBet()
	if currentHighBet == 0 {
		// No bets yet - can check or bet
		switch action.Action {
		case ActionCheck, ActionBet, ActionAllIn:
			return true
		default:
			return false
		}
	} else {
		// There's a bet - can call, raise, fold, or go all-in
		switch action.Action {
		case ActionCall, ActionRaise, ActionFold, ActionAllIn:
			return true
		case ActionCheck:
			// Can only check if player's bet equals current high bet
			return player.CurrentBet == currentHighBet
		default:
			return false
		}
	}
}

// getCurrentHighBet returns the highest current bet among all players
func (t *Table) getCurrentHighBet() int64 {
	var highest int64
	for _, p := range t.state.Players {
		if p != nil && p.CurrentBet > highest {
			highest = p.CurrentBet
		}
	}
	return highest
}

// processFold handles a fold action
func (t *Table) processFold(player *Player) {
	player.Status = PlayerFolded
	t.state.PlayersActed[t.state.CurrentPlayer] = true
}

// processCheck handles a check action
func (t *Table) processCheck(player *Player) {
	player.Status = PlayerActive
	t.state.PlayersActed[t.state.CurrentPlayer] = true
}

// processCall handles a call action
func (t *Table) processCall(player *Player) {
	currentHighBet := t.getCurrentHighBet()
	callAmount := currentHighBet - player.CurrentBet

	if player.Chips <= callAmount {
		// All-in call
		player.Chips = 0
		player.CurrentBet += player.Chips
		player.TotalInvested += player.Chips
		player.Status = PlayerAllIn
	} else {
		player.Chips -= callAmount
		player.CurrentBet += callAmount
		player.TotalInvested += callAmount
	}

	t.updatePot()
	t.state.PlayersActed[t.state.CurrentPlayer] = true
}

// processBet handles a bet action
func (t *Table) processBet(player *Player, amount int64) {
	if amount < t.config.BigBlind {
		amount = t.config.BigBlind
	}

	if amount > player.Chips {
		amount = player.Chips
	}

	player.Chips -= amount
	player.CurrentBet = amount
	player.TotalInvested += amount

	if player.Chips == 0 {
		player.Status = PlayerAllIn
	}

	t.state.LastBet = amount
	t.state.MinRaise = amount * 2

	// Reset other players' "acted" status since a new bet requires action
	for i := range t.state.PlayersActed {
		t.state.PlayersActed[i] = false
	}

	t.updatePot()
	t.state.PlayersActed[t.state.CurrentPlayer] = true
}

// processRaise handles a raise action
func (t *Table) processRaise(player *Player, amount int64) {
	currentHighBet := t.getCurrentHighBet()
	raiseAmount := amount

	if raiseAmount < t.state.MinRaise {
		raiseAmount = t.state.MinRaise
	}

	totalBet := currentHighBet + raiseAmount
	if totalBet > player.Chips+player.CurrentBet {
		totalBet = player.Chips + player.CurrentBet
		raiseAmount = totalBet - player.CurrentBet
	}

	player.Chips -= raiseAmount
	player.CurrentBet = totalBet
	player.TotalInvested += raiseAmount

	if player.Chips == 0 {
		player.Status = PlayerAllIn
	}

	t.state.LastBet = raiseAmount
	t.state.MinRaise = raiseAmount * 2

	// Reset other players' "acted" status
	for i := range t.state.PlayersActed {
		t.state.PlayersActed[i] = false
	}

	t.updatePot()
	t.state.PlayersActed[t.state.CurrentPlayer] = true
}

// processAllIn handles an all-in action
func (t *Table) processAllIn(player *Player) {
	allInAmount := player.Chips
	player.CurrentBet += allInAmount
	player.TotalInvested += allInAmount
	player.Chips = 0
	player.Status = PlayerAllIn

	if allInAmount > 0 {
		if allInAmount > t.state.LastBet {
			t.state.LastBet = allInAmount
			// Reset other players' "acted" status
			for i := range t.state.PlayersActed {
				t.state.PlayersActed[i] = false
			}
		}
	}

	t.updatePot()
	t.state.PlayersActed[t.state.CurrentPlayer] = true
}

// advanceToNextPlayer moves to the next player who needs to act
func (t *Table) advanceToNextPlayer() {
	players := t.getActivePlayers()

	// Find current player's index in the list
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

	// Find next active player
	for i := 1; i <= len(players); i++ {
		nextIdx := players[(currentIdx+i)%len(players)]
		if t.state.Players[nextIdx] != nil &&
			t.state.Players[nextIdx].Status == PlayerActive {
			t.state.CurrentPlayer = nextIdx
			return
		}
	}

	// No more active players to act
}

// allPlayersActed returns true if all active players have acted this round
func (t *Table) allPlayersActed() bool {
	players := t.getActivePlayers()
	if len(players) == 0 {
		return true
	}

	// All active (non-folded) players must have acted
	for _, idx := range players {
		if t.state.Players[idx] != nil &&
			t.state.Players[idx].Status == PlayerActive &&
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
		if t.state.Players[idx] != nil &&
			t.state.Players[idx].Status == PlayerActive {
			return false
		}
	}

	return true
}

// completeBettingRound advances to the next phase
func (t *Table) completeBettingRound() {
	switch t.state.Phase {
	case PhasePreflop:
		t.state.Phase = PhaseFlop
		t.dealCommunityCard(3)
	case PhaseFlop:
		t.state.Phase = PhaseTurn
		t.dealCommunityCard(1)
	case PhaseTurn:
		t.state.Phase = PhaseRiver
		t.dealCommunityCard(1)
	case PhaseRiver:
		t.state.Phase = PhaseShowdown
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
	t.determineFirstActor()
}

// dealCommunityCard deals community cards
func (t *Table) dealCommunityCard(count int) {
	for i := 0; i < count && len(t.state.Deck) > 0; i++ {
		t.state.CommunityCards = append(t.state.CommunityCards, t.state.Deck[0])
		t.state.Deck = t.state.Deck[1:]
	}
}

// updatePot updates the main pot with current bets
func (t *Table) updatePot() {
	var potAmount int64
	for _, p := range t.state.Players {
		if p != nil {
			potAmount += p.CurrentBet
			p.CurrentBet = 0 // Reset current bet after adding to pot
		}
	}
	t.state.PotTotal += potAmount
	if len(t.state.Pots) > 0 {
		t.state.Pots[0].Amount += potAmount
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
		// Only one player left - they win
		winner := t.state.Players[activePlayers[0]]
		if winner != nil {
			winner.Chips += t.state.PotTotal
		}
		return
	}

	// Multiple players - evaluate hands
	winners := t.determineWinners(activePlayers)

	// Split pot among winners
	potPerWinner := t.state.PotTotal / int64(len(winners))
	for _, winnerIdx := range winners {
		if t.state.Players[winnerIdx] != nil {
			t.state.Players[winnerIdx].Chips += potPerWinner
		}
	}

	// Handle remainder
	remainder := t.state.PotTotal % int64(len(winners))
	if remainder > 0 && len(winners) > 0 {
		if t.state.Players[winners[0]] != nil {
			t.state.Players[winners[0]].Chips += remainder
		}
	}
}

// determineWinners determines which players have the best hands
func (t *Table) determineWinners(playerIndices []int) []int {
	var bestHand *poker.EvaluatedHand
	var winners []int

	for _, idx := range playerIndices {
		player := t.state.Players[idx]
		if player == nil || len(player.HoleCards) != 2 {
			continue
		}

		// Combine hole cards with community cards
		allCards := append([]poker.Card{player.HoleCards[0], player.HoleCards[1]}, t.state.CommunityCards...)
		if len(allCards) != 7 {
			continue
		}

		hand, err := t.evaluator.Evaluate7Card(allCards)
		if err != nil {
			continue
		}

		if bestHand == nil {
			bestHand = hand
			winners = []int{idx}
		} else {
			cmp := t.evaluator.CompareHands(hand, bestHand)
			if cmp > 0 {
				bestHand = hand
				winners = []int{idx}
			} else if cmp == 0 {
				winners = append(winners, idx)
			}
		}
	}

	return winners
}

// getActivePlayers returns indices of all active players
func (t *Table) getActivePlayers() []int {
	var players []int
	for i, p := range t.state.Players {
		if p != nil && p.Status == PlayerActive && p.IsConnected {
			players = append(players, i)
		}
	}
	return players
}

// getPlayersNotFolded returns indices of players who haven't folded
func (t *Table) getPlayersNotFolded() []int {
	var players []int
	for i, p := range t.state.Players {
		if p != nil && p.Status != PlayerFolded && p.IsConnected {
			players = append(players, i)
		}
	}
	return players
}

// getLastPlayerNotFolded returns the index of the last player who didn't fold
func (t *Table) getLastPlayerNotFolded() *Player {
	var last *Player
	for _, p := range t.state.Players {
		if p != nil && p.Status != PlayerFolded && p.IsConnected {
			last = p
		}
	}
	return last
}

// countActivePlayers returns the number of active players
func (t *Table) countActivePlayers() int {
	count := 0
	for _, p := range t.state.Players {
		if p != nil && p.IsConnected && p.Status != PlayerSittingOut {
			count++
		}
	}
	return count
}

// rotateDealerButton moves the dealer button to the next player
func (t *Table) rotateDealerButton() {
	players := t.getActivePlayers()
	if len(players) == 0 {
		return
	}

	// Find current button position
	currentButtonIdx := -1
	for i, idx := range players {
		if idx == t.state.DealerButton {
			currentButtonIdx = i
			break
		}
	}

	// Move button to next player
	nextButtonIdx := players[(currentButtonIdx+1)%len(players)]

	// Update dealer status
	for i, p := range t.state.Players {
		if p != nil {
			p.IsDealer = (i == nextButtonIdx)
		}
	}

	t.state.DealerButton = nextButtonIdx
}

// Errors
var (
	ErrTableFull         = &TableError{Message: "table is full"}
	ErrNoSeatsAvailable  = &TableError{Message: "no seats available"}
	ErrPlayerNotFound    = &TableError{Message: "player not found"}
	ErrInvalidAction     = &TableError{Message: "invalid action"}
	ErrNotEnoughPlayers  = &TableError{Message: "not enough players"}
	ErrInvalidBetAmount  = &TableError{Message: "invalid bet amount"}
)

// TableError represents an error from the table
type TableError struct {
	Message string
}

func (e *TableError) Error() string {
	return e.Message
}
