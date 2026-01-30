package rules

import (
	"fmt"
	"time"

	"poker-platform/pkg/poker"
)

// GameType represents the type of poker game
type GameType string

const (
	GameTypeTexasHoldem   GameType = "texas_hold'em"
	GameTypeOmaha         GameType = "omaha"
	GameTypeOmahaHiLo     GameType = "omaha_hi_lo"
	GameTypeSevenCardStud GameType = "seven_card_stud"
	GameTypeFiveCardDraw  GameType = "five_card_draw"
)

// BettingType represents the betting structure
type BettingType string

const (
	BettingTypeNoLimit    BettingType = "no_limit"
	BettingTypePotLimit   BettingType = "pot_limit"
	BettingTypeFixedLimit BettingType = "fixed_limit"
)

// GamePhase represents a phase in the poker game
type GamePhase int

const (
	PhaseWaiting GamePhase = iota
	PhasePreDeal
	PhasePreflop
	PhaseFlop
	PhaseTurn
	PhaseRiver
	PhaseShowdown
	PhaseHandComplete
)

func (p GamePhase) String() string {
	switch p {
	case PhaseWaiting:
		return "waiting"
	case PhasePreDeal:
		return "pre_deal"
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
	ActionFold PlayerAction = iota
	ActionCheck
	ActionCall
	ActionBet
	ActionRaise
	ActionAllIn
	ActionSitOut
	ActionReturn
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
	case ActionSitOut:
		return "sit_out"
	case ActionReturn:
		return "return"
	default:
		return "unknown"
	}
}

// Position represents a position at the table
type Position int

const (
	PositionBTN Position = iota // Button/Dealer
	PositionSB                  // Small Blind
	PositionBB                  // Big Blind
	PositionUTG                 // Under the Gun
	PositionUTG1
	PositionUTG2
	PositionUTG3
	PositionLJ // Lojack
	PositionHJ // Hijack
	PositionCO // Cutoff
)

func (p Position) String() string {
	switch p {
	case PositionBTN:
		return "button"
	case PositionSB:
		return "small_blind"
	case PositionBB:
		return "big_blind"
	case PositionUTG:
		return "under_the_gun"
	case PositionUTG1:
		return "under_the_gun_1"
	case PositionUTG2:
		return "under_the_gun_2"
	case PositionUTG3:
		return "under_the_gun_3"
	case PositionLJ:
		return "lojack"
	case PositionHJ:
		return "hijack"
	case PositionCO:
		return "cutoff"
	default:
		return "unknown"
	}
}

// PlayerStatus represents a player's current state
type PlayerStatus int

const (
	PlayerActive PlayerStatus = iota
	PlayerFolded
	PlayerAllIn
	PlayerSittingOut
	PlayerDisconnected
	PlayerBusted
)

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
	case PlayerBusted:
		return "busted"
	default:
		return "unknown"
	}
}

// Player represents a player at the table
type Player struct {
	ID            string
	Name          string
	Chips         int64
	HoleCards     []poker.Card
	Status        PlayerStatus
	CurrentBet    int64
	TotalInvested int64 // Total chips put into this pot (for all-in calculations)
	IsConnected   bool
	IsDealer      bool
	Position      Position
	SeatNumber    int
}

// Pot represents a pot in the game (main pot or side pot)
type Pot struct {
	ID               string
	Amount           int64
	EligiblePlayers  map[string]bool // PlayerIDs eligible to win this pot
	WinnerIDs        []string
	IsSidePot        bool
	AssociatedPlayer string // For side pots, the player who went all-in
}

// TableConfig holds the configuration for a table
type TableConfig struct {
	TableID           string
	GameType          GameType
	BettingType       BettingType
	MinPlayers        int
	MaxPlayers        int
	SmallBlind        int64
	BigBlind          int64
	BuyInMin          int64
	BuyInMax          int64
	MaxSessionTime    time.Duration
	ActionTimeout     time.Duration
	AutoRebuyEnabled  bool
	StraddleEnabled   bool
	RunItTwiceEnabled bool
}

// TableState represents the complete state of a poker table
type TableState struct {
	TableID          string
	GameType         GameType
	BettingType      BettingType
	Phase            GamePhase
	DealerButton     int // Index of player with dealer button
	CurrentPlayer    int // Index of player whose turn it is
	CommunityCards   []poker.Card
	Pots             []Pot
	Players          []*Player
	SidePots         []Pot
	LastBet          int64
	MinRaise         int64
	PotTotal         int64
	Deck             []poker.Card
	HandNumber       int
	PhaseStartTime   time.Time
	PlayersActed     map[int]bool
	PlayersToAct     []int
	SmallBlindPos    int
	BigBlindPos      int
	CurrentBetToCall int64
}

// PlayerActionRequest represents a player's action request
type PlayerActionRequest struct {
	PlayerID string
	Action   PlayerAction
	Amount   int64 // For bet/raise amounts
}

// RulesEngine defines the interface for poker game rules
type RulesEngine interface {
	// Game identification
	GameType() GameType
	BettingType() BettingType
	Name() string
	Version() string

	// Table configuration validation
	ValidateConfig(config TableConfig) error
	DefaultConfig() TableConfig

	// Deck management
	CreateDeck() []poker.Card
	ShuffleDeck(deck []poker.Card) []poker.Card

	// Betting structure
	CalculateBlinds(handNumber int, players []*Player, config TableConfig) (sbAmount, bbAmount int64)
	CalculateMinBet(config TableConfig) int64
	CalculateMinRaise(currentBet, minBet int64, config TableConfig) int64
	CalculateBetAmount(player *Player, targetAmount int64, config TableConfig) (amount int64, isAllIn bool)
	ValidateBetSizing(amount int64, player *Player, currentBet, minBet, potSize int64, config TableConfig) error

	// Hand phases
	Phases() []GamePhase
	NextPhase(currentPhase GamePhase, state *TableState) GamePhase
	IsBettingPhase(phase GamePhase) bool
	IsCompletePhase(phase GamePhase) bool

	// Position management
	CalculatePositions(players []*Player, dealerButton int, config TableConfig) map[int]Position
	DetermineFirstActor(phase GamePhase, state *TableState, positions map[int]Position) int

	// Card dealing
	DealHoleCards(state *TableState, players []*Player) error
	DealCommunityCards(state *TableState, phase GamePhase) error
	GetHoleCardCount() int
	GetCommunityCardCount(phase GamePhase) int

	// Action validation
	ValidateAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error
	GetValidActions(state *TableState, player *Player, config TableConfig) []PlayerAction

	// Action processing
	ProcessAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error
	ProcessBet(player *Player, amount int64, state *TableState, config TableConfig)
	ProcessCall(player *Player, state *TableState, config TableConfig)
	ProcessRaise(player *Player, amount int64, state *TableState, config TableConfig)
	ProcessFold(player *Player, state *TableState)
	ProcessCheck(player *Player, state *TableState)
	ProcessAllIn(player *Player, state *TableState, config TableConfig)

	// Pot management
	UpdatePot(state *TableState)
	CalculateSidePots(state *TableState, players []*Player) []Pot
	DistributePot(state *TableState, winners []int, evaluator poker.HandEvaluator) error

	// Hand evaluation
	EvaluateHand(holeCards []poker.Card, communityCards []poker.Card) (*poker.EvaluatedHand, error)
	CompareHands(hand1, hand2 *poker.EvaluatedHand) int
	DetermineWinners(players []*Player, communityCards []poker.Card, evaluator poker.HandEvaluator) []int

	// Game flow
	ShouldStartHand(players []*Player, config TableConfig) bool
	ShouldEndHand(state *TableState) bool
	RotateDealerButton(state *TableState, players []*Player)
	ResetHandState(state *TableState, config TableConfig)

	// String representations
	StringAction(action PlayerAction) string
	StringPhase(phase GamePhase) string
	StringPosition(pos Position) string
}

// BaseRulesEngine provides common functionality for all poker variants
type BaseRulesEngine struct{}

func (b *BaseRulesEngine) ValidateConfig(config TableConfig) error {
	if config.MinPlayers < 2 {
		return ErrMinPlayers
	}
	if config.MaxPlayers > 10 {
		return ErrMaxPlayers
	}
	if config.SmallBlind <= 0 {
		return ErrInvalidBlind
	}
	if config.BigBlind <= config.SmallBlind {
		return ErrInvalidBlindStructure
	}
	if config.BuyInMin > config.BuyInMax {
		return ErrInvalidBuyIn
	}
	return nil
}

func (b *BaseRulesEngine) CreateDeck() []poker.Card {
	deck := make([]poker.Card, 52)
	for rank := poker.Rank2; rank <= poker.RankA; rank++ {
		for suit := poker.SuitClubs; suit <= poker.SuitSpades; suit++ {
			deck[int(rank)*4+int(suit)] = poker.NewCard(rank, suit)
		}
	}
	return deck
}

func (b *BaseRulesEngine) ShuffleDeck(deck []poker.Card) []poker.Card {
	// Fisher-Yates shuffle - in production, use crypto/rand
	n := len(deck)
	for i := n - 1; i > 0; i-- {
		j := i // In production: crypto rand
		deck[i], deck[j] = deck[j], deck[i]
	}
	return deck
}

func (b *BaseRulesEngine) CalculateBlinds(handNumber int, players []*Player, config TableConfig) (sbAmount, bbAmount int64) {
	// Standard blind structure
	return config.SmallBlind, config.BigBlind
}

func (b *BaseRulesEngine) CalculateMinBet(config TableConfig) int64 {
	return config.BigBlind
}

func (b *BaseRulesEngine) CalculateMinRaise(currentBet, minBet int64, config TableConfig) int64 {
	raiseAmount := currentBet + minBet
	return raiseAmount
}

func (b *BaseRulesEngine) CalculateBetAmount(player *Player, targetAmount int64, config TableConfig) (amount int64, isAllIn bool) {
	if targetAmount >= player.Chips {
		return player.Chips, true
	}
	return targetAmount, false
}

func (b *BaseRulesEngine) ValidateBetSizing(amount int64, player *Player, currentBet, minBet, potSize int64, config TableConfig) error {
	if amount <= 0 {
		return ErrInvalidBetAmount
	}

	// Check minimum bet/raise
	if currentBet > 0 && amount < minBet {
		return ErrRaiseTooSmall
	}

	// Check player has enough chips
	if amount > player.Chips {
		return ErrInsufficientChips
	}

	return nil
}

func (b *BaseRulesEngine) Phases() []GamePhase {
	return []GamePhase{
		PhaseWaiting,
		PhasePreflop,
		PhaseFlop,
		PhaseTurn,
		PhaseRiver,
		PhaseShowdown,
		PhaseHandComplete,
	}
}

func (b *BaseRulesEngine) IsBettingPhase(phase GamePhase) bool {
	return phase == PhasePreflop || phase == PhaseFlop || phase == PhaseTurn || phase == PhaseRiver
}

func (b *BaseRulesEngine) IsCompletePhase(phase GamePhase) bool {
	return phase == PhaseHandComplete || phase == PhaseShowdown
}

func (b *BaseRulesEngine) CalculatePositions(players []*Player, dealerButton int, config TableConfig) map[int]Position {
	positions := make(map[int]Position)
	activePlayers := b.getActivePlayerIndices(players)

	if len(activePlayers) == 0 {
		return positions
	}

	// Find dealer's index in active players
	dealerIdx := -1
	for i, idx := range activePlayers {
		if idx == dealerButton {
			dealerIdx = i
			break
		}
	}
	if dealerIdx == -1 {
		dealerIdx = 0
	}

	// Assign positions based on table size
	positionOrder := b.getPositionOrder(len(activePlayers))

	for i, playerIdx := range activePlayers {
		posIdx := (dealerIdx + i) % len(activePlayers)
		if posIdx < len(positionOrder) {
			positions[playerIdx] = positionOrder[posIdx]
		}
	}

	return positions
}

func (b *BaseRulesEngine) getPositionOrder(tableSize int) []Position {
	switch tableSize {
	case 2:
		return []Position{PositionBTN, PositionSB}
	case 3:
		return []Position{PositionBTN, PositionSB, PositionBB}
	case 4, 5, 6:
		return []Position{PositionBTN, PositionSB, PositionBB, PositionUTG}
	default:
		return []Position{PositionBTN, PositionSB, PositionBB, PositionUTG, PositionUTG1, PositionUTG2,
			PositionLJ, PositionHJ, PositionCO}
	}
}

func (b *BaseRulesEngine) getActivePlayerIndices(players []*Player) []int {
	var indices []int
	for i, p := range players {
		if p != nil && p.Status == PlayerActive && p.IsConnected {
			indices = append(indices, i)
		}
	}
	return indices
}

func (b *BaseRulesEngine) ValidateAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error {
	if player == nil || player.ID != action.PlayerID {
		return ErrPlayerNotFound
	}
	if player.Status != PlayerActive {
		return ErrPlayerNotActive
	}
	if state.CurrentPlayer < 0 || state.CurrentPlayer >= len(state.Players) {
		return ErrNotYourTurn
	}

	currentHighBet := state.LastBet
	playerBet := player.CurrentBet

	switch action.Action {
	case ActionFold:
		return nil
	case ActionCheck:
		if currentHighBet > 0 && playerBet < currentHighBet {
			return ErrCannotCheck
		}
		return nil
	case ActionCall:
		if currentHighBet == 0 {
			return ErrCannotCall
		}
		return nil
	case ActionBet:
		if currentHighBet > 0 {
			return ErrCannotBet
		}
		return b.ValidateBetSizing(action.Amount, player, currentHighBet, state.MinRaise, state.PotTotal, config)
	case ActionRaise:
		if currentHighBet == 0 {
			return ErrCannotRaise
		}
		return b.ValidateBetSizing(action.Amount, player, currentHighBet, state.MinRaise, state.PotTotal, config)
	case ActionAllIn:
		return nil
	default:
		return ErrInvalidAction
	}
}

func (b *BaseRulesEngine) GetValidActions(state *TableState, player *Player, config TableConfig) []PlayerAction {
	var actions []PlayerAction

	currentHighBet := state.LastBet
	playerBet := player.CurrentBet

	actions = append(actions, ActionFold)

	if currentHighBet == 0 || playerBet == currentHighBet {
		actions = append(actions, ActionCheck)
	}

	actions = append(actions, ActionCall, ActionAllIn)

	if currentHighBet == 0 {
		actions = append(actions, ActionBet)
	} else {
		actions = append(actions, ActionRaise)
	}

	return actions
}

func (b *BaseRulesEngine) ProcessAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error {
	switch action.Action {
	case ActionFold:
		b.ProcessFold(player, state)
	case ActionCheck:
		b.ProcessCheck(player, state)
	case ActionCall:
		b.ProcessCall(player, state, config)
	case ActionBet:
		b.ProcessBet(player, action.Amount, state, config)
	case ActionRaise:
		b.ProcessRaise(player, action.Amount, state, config)
	case ActionAllIn:
		b.ProcessAllIn(player, state, config)
	}
	return nil
}

func (b *BaseRulesEngine) ProcessFold(player *Player, state *TableState) {
	player.Status = PlayerFolded
	state.PlayersActed[state.CurrentPlayer] = true
}

func (b *BaseRulesEngine) ProcessCheck(player *Player, state *TableState) {
	player.Status = PlayerActive
	state.PlayersActed[state.CurrentPlayer] = true
}

func (b *BaseRulesEngine) ProcessCall(player *Player, state *TableState, config TableConfig) {
	callAmount := state.LastBet - player.CurrentBet
	b.ProcessBet(player, callAmount, state, config)
}

func (b *BaseRulesEngine) ProcessBet(player *Player, amount int64, state *TableState, config TableConfig) {
	if amount > player.Chips {
		amount = player.Chips
	}

	player.Chips -= amount
	player.CurrentBet += amount
	player.TotalInvested += amount

	if player.Chips == 0 {
		player.Status = PlayerAllIn
	}

	state.LastBet = player.CurrentBet
	state.MinRaise = b.CalculateMinRaise(state.LastBet, config.BigBlind, config)

	// Reset other players' "acted" status
	for i := range state.PlayersActed {
		state.PlayersActed[i] = false
	}

	state.PlayersActed[state.CurrentPlayer] = true
	b.UpdatePot(state)
}

func (b *BaseRulesEngine) ProcessRaise(player *Player, amount int64, state *TableState, config TableConfig) {
	totalBet := state.LastBet + amount
	if totalBet > player.Chips+player.CurrentBet {
		totalBet = player.Chips + player.CurrentBet
		amount = totalBet - player.CurrentBet
	}

	player.Chips -= amount
	player.CurrentBet = totalBet
	player.TotalInvested += amount

	if player.Chips == 0 {
		player.Status = PlayerAllIn
	}

	state.LastBet = player.CurrentBet
	state.MinRaise = b.CalculateMinRaise(state.LastBet, config.BigBlind, config)

	for i := range state.PlayersActed {
		state.PlayersActed[i] = false
	}

	state.PlayersActed[state.CurrentPlayer] = true
	b.UpdatePot(state)
}

func (b *BaseRulesEngine) ProcessAllIn(player *Player, state *TableState, config TableConfig) {
	allInAmount := player.Chips
	player.CurrentBet += allInAmount
	player.TotalInvested += allInAmount
	player.Chips = 0
	player.Status = PlayerAllIn

	if allInAmount > 0 {
		if allInAmount > state.LastBet {
			state.LastBet = allInAmount
			state.MinRaise = b.CalculateMinRaise(state.LastBet, config.BigBlind, config)
			for i := range state.PlayersActed {
				state.PlayersActed[i] = false
			}
		}
	}

	state.PlayersActed[state.CurrentPlayer] = true
	b.UpdatePot(state)
}

func (b *BaseRulesEngine) UpdatePot(state *TableState) {
	var potAmount int64
	for _, p := range state.Players {
		if p != nil {
			potAmount += p.CurrentBet
			p.CurrentBet = 0
		}
	}
	state.PotTotal += potAmount
	if len(state.Pots) > 0 {
		state.Pots[0].Amount += potAmount
	}
}

func (b *BaseRulesEngine) CalculateSidePots(state *TableState, players []*Player) []Pot {
	var sidePots []Pot
	var allInPlayers []*Player

	// Find all-in players
	for _, p := range players {
		if p != nil && p.Status == PlayerAllIn && p.TotalInvested > 0 {
			allInPlayers = append(allInPlayers, p)
		}
	}

	if len(allInPlayers) == 0 {
		return sidePots
	}

	// Sort by total invested (lowest first)
	for i := 0; i < len(allInPlayers)-1; i++ {
		for j := i + 1; j < len(allInPlayers); j++ {
			if allInPlayers[i].TotalInvested > allInPlayers[j].TotalInvested {
				allInPlayers[i], allInPlayers[j] = allInPlayers[j], allInPlayers[i]
			}
		}
	}

	// Calculate side pots
	for i, player := range allInPlayers {
		var pot Pot
		pot.ID = fmt.Sprintf("side_pot_%d", i+1)
		pot.IsSidePot = true
		pot.AssociatedPlayer = player.ID
		pot.EligiblePlayers = make(map[string]bool)

		// Calculate pot amount for this level
		var minBet int64 = player.TotalInvested
		for _, p := range players {
			if p != nil {
				if p.TotalInvested < minBet {
					minBet = p.TotalInvested
				}
			}
		}

		// Add chips to pot
		for _, p := range players {
			if p != nil {
				addAmount := minBet
				if p.TotalInvested < minBet {
					addAmount = p.TotalInvested
				}
				pot.Amount += addAmount
			}
		}

		// Mark eligible players
		for _, p := range players {
			if p != nil && p.Status != PlayerFolded && p.TotalInvested >= player.TotalInvested {
				pot.EligiblePlayers[p.ID] = true
			}
		}

		if pot.Amount > 0 {
			sidePots = append(sidePots, pot)
		}
	}

	return sidePots
}

func (b *BaseRulesEngine) ShouldStartHand(players []*Player, config TableConfig) bool {
	activePlayers := b.countActivePlayers(players)
	return activePlayers >= config.MinPlayers
}

func (b *BaseRulesEngine) countActivePlayers(players []*Player) int {
	count := 0
	for _, p := range players {
		if p != nil && p.IsConnected && p.Status != PlayerSittingOut {
			count++
		}
	}
	return count
}

func (b *BaseRulesEngine) ShouldEndHand(state *TableState) bool {
	activePlayers := b.getPlayersNotFolded(state.Players)
	if len(activePlayers) == 0 {
		return true
	}
	if len(activePlayers) == 1 {
		return true
	}
	return false
}

func (b *BaseRulesEngine) getPlayersNotFolded(players []*Player) []int {
	var indices []int
	for i, p := range players {
		if p != nil && p.Status != PlayerFolded && p.IsConnected {
			indices = append(indices, i)
		}
	}
	return indices
}

func (b *BaseRulesEngine) RotateDealerButton(state *TableState, players []*Player) {
	activePlayers := b.getActivePlayerIndices(players)
	if len(activePlayers) == 0 {
		return
	}

	currentButtonIdx := -1
	for i, idx := range activePlayers {
		if idx == state.DealerButton {
			currentButtonIdx = i
			break
		}
	}

	nextButtonIdx := activePlayers[(currentButtonIdx+1)%len(activePlayers)]

	for i, p := range players {
		if p != nil {
			p.IsDealer = (i == nextButtonIdx)
		}
	}

	state.DealerButton = nextButtonIdx
}

func (b *BaseRulesEngine) ResetHandState(state *TableState, config TableConfig) {
	state.CommunityCards = nil
	state.Pots = []Pot{{
		ID:              "main",
		Amount:          0,
		EligiblePlayers: make(map[string]bool),
	}}
	state.SidePots = nil
	state.LastBet = 0
	state.MinRaise = config.BigBlind
	state.PotTotal = 0
	state.Deck = b.ShuffleDeck(b.CreateDeck())
	state.PlayersActed = make(map[int]bool)
	state.PlayersToAct = nil

	for _, p := range state.Players {
		if p != nil {
			p.HoleCards = nil
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

func (b *BaseRulesEngine) StringAction(action PlayerAction) string {
	return action.String()
}

func (b *BaseRulesEngine) StringPhase(phase GamePhase) string {
	return phase.String()
}

func (b *BaseRulesEngine) StringPosition(pos Position) string {
	return pos.String()
}

// Errors
var (
	ErrMinPlayers            = &RulesError{Message: "minimum 2 players required"}
	ErrMaxPlayers            = &RulesError{Message: "maximum 10 players per table"}
	ErrInvalidBlind          = &RulesError{Message: "blinds must be positive"}
	ErrInvalidBlindStructure = &RulesError{Message: "big blind must be greater than small blind"}
	ErrInvalidBuyIn          = &RulesError{Message: "min buy-in must be less than max buy-in"}
	ErrPlayerNotFound        = &RulesError{Message: "player not found"}
	ErrPlayerNotActive       = &RulesError{Message: "player is not active"}
	ErrNotYourTurn           = &RulesError{Message: "not your turn"}
	ErrCannotCheck           = &RulesError{Message: "cannot check when there's a bet"}
	ErrCannotCall            = &RulesError{Message: "cannot call when there's no bet"}
	ErrCannotBet             = &RulesError{Message: "cannot bet when there's already a bet"}
	ErrCannotRaise           = &RulesError{Message: "cannot raise when there's no bet"}
	ErrInvalidAction         = &RulesError{Message: "invalid action"}
	ErrBetTooSmall           = &RulesError{Message: "bet is too small"}
	ErrBetTooLarge           = &RulesError{Message: "bet exceeds table limits"}
	ErrNotEnoughChips        = &RulesError{Message: "not enough chips"}
	ErrInvalidBetAmount      = &RulesError{Message: "bet amount must be positive"}
	ErrRaiseTooSmall         = &RulesError{Message: "raise must be at least minimum raise"}
	ErrInsufficientChips     = &RulesError{Message: "bet exceeds player's chips"}
)

// RulesError represents an error from the rules engine
type RulesError struct {
	Message string
}

func (e *RulesError) Error() string {
	return e.Message
}
