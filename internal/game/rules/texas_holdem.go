package rules

import (
	"fmt"

	"poker-platform/pkg/poker"
)

// TexasHoldem implements the RulesEngine interface for Texas Hold'em
type TexasHoldem struct {
	*BaseRulesEngine
}

// NewTexasHoldem creates a new Texas Hold'em rules engine
func NewTexasHoldem() *TexasHoldem {
	return &TexasHoldem{
		BaseRulesEngine: &BaseRulesEngine{},
	}
}

func (t *TexasHoldem) GameType() GameType {
	return GameTypeTexasHoldem
}

func (t *TexasHoldem) BettingType() BettingType {
	return BettingTypeNoLimit
}

func (t *TexasHoldem) Name() string {
	return "Texas Hold'em"
}

func (t *TexasHoldem) Version() string {
	return "1.0.0"
}

func (t *TexasHoldem) DefaultConfig() TableConfig {
	return TableConfig{
		GameType:          GameTypeTexasHoldem,
		BettingType:       BettingTypeNoLimit,
		MinPlayers:        2,
		MaxPlayers:        9,
		SmallBlind:        5,
		BigBlind:          10,
		BuyInMin:          100,
		BuyInMax:          10000,
		ActionTimeout:     30,
		AutoRebuyEnabled:  false,
		StraddleEnabled:   false,
		RunItTwiceEnabled: false,
	}
}

func (t *TexasHoldem) CalculateBlinds(handNumber int, players []*Player, config TableConfig) (sbAmount, bbAmount int64) {
	// Standard blind structure
	sbAmount = config.SmallBlind
	bbAmount = config.BigBlind

	// Optional: Tournament blind escalation
	// if handNumber > 100 {
	//     sbAmount *= 2
	//     bbAmount *= 2
	// }

	return sbAmount, bbAmount
}

func (t *TexasHoldem) CalculateMinBet(config TableConfig) int64 {
	return config.BigBlind
}

func (t *TexasHoldem) CalculateMinRaise(currentBet, minBet int64, config TableConfig) int64 {
	// No-limit: minimum raise is the size of the previous bet/raise
	raiseAmount := minBet
	if currentBet > 0 {
		raiseAmount = currentBet
	}
	return raiseAmount
}

func (t *TexasHoldem) CalculateBetAmount(player *Player, targetAmount, config TableConfig) (amount int64, isAllIn bool) {
	if targetAmount >= player.Chips {
		return player.Chips, true
	}
	return targetAmount, false
}

func (t *TexasHoldem) ValidateBetSizing(amount int64, player *Player, currentBet, minBet, potSize int64, config TableConfig) error {
	if amount <= 0 {
		return ErrInvalidAction
	}

	if currentBet == 0 {
		// Opening bet must be at least min bet (big blind)
		if amount < config.BigBlind {
			return ErrBetTooSmall
		}
	} else {
		// Raise must be at least min raise
		if amount < minBet {
			return ErrBetTooSmall
		}
	}

	if amount > player.Chips {
		return ErrNotEnoughChips
	}

	return nil
}

func (t *TexasHoldem) Phases() []GamePhase {
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

func (t *TexasHoldem) NextPhase(currentPhase GamePhase, state *TableState) GamePhase {
	switch currentPhase {
	case PhaseWaiting:
		if t.ShouldStartHand(state.Players, TableConfig{}) {
			return PhasePreflop
		}
	case PhasePreflop:
		return PhaseFlop
	case PhaseFlop:
		return PhaseTurn
	case PhaseTurn:
		return PhaseRiver
	case PhaseRiver:
		return PhaseShowdown
	case PhaseShowdown:
		return PhaseHandComplete
	case PhaseHandComplete:
		return PhaseWaiting
	}
	return currentPhase
}

func (t *TexasHoldem) IsBettingPhase(phase GamePhase) bool {
	return phase == PhasePreflop || phase == PhaseFlop || phase == PhaseTurn || phase == PhaseRiver
}

func (t *TexasHoldem) IsCompletePhase(phase GamePhase) bool {
	return phase == PhaseHandComplete || phase == PhaseShowdown
}

func (t *TexasHoldem) DetermineFirstActor(phase GamePhase, state *TableState, positions map[int]Position) int {
	players := state.Players
	if len(players) == 0 {
		return -1
	}

	switch phase {
	case PhasePreflop:
		// First to act is under the gun (first player after big blind)
		bbPos := state.BigBlindPos
		idx := t.findPlayerIndex(players, bbPos)
		if idx >= 0 {
			return (idx + 1) % len(players)
		}
	case PhaseFlop, PhaseTurn, PhaseRiver:
		// First to act is small blind or first active player
		sbPos := state.SmallBlindPos
		idx := t.findPlayerIndex(players, sbPos)
		if idx >= 0 {
			return (idx + 1) % len(players)
		}
	}

	// Fallback: find first active player
	for i := 0; i < len(players); i++ {
		if players[i] != nil && players[i].Status == PlayerActive {
			return i
		}
	}

	return -1
}

func (t *TexasHoldem) findPlayerIndex(players []*Player, targetPos int) int {
	for i := 0; i < len(players); i++ {
		if players[i] != nil && players[i].SeatNumber == targetPos {
			return i
		}
	}
	return -1
}

func (t *TexasHoldem) DealHoleCards(state *TableState, players []*Player) error {
	// Deal 2 hole cards to each active player
	for _, player := range players {
		if player != nil && player.Status == PlayerActive && player.IsConnected {
			if len(state.Deck) >= 2 {
				player.HoleCards = []poker.Card{state.Deck[0], state.Deck[1]}
				state.Deck = state.Deck[2:]
			} else {
				return fmt.Errorf("deck exhausted")
			}
		}
	}
	return nil
}

func (t *TexasHoldem) DealCommunityCards(state *TableState, phase GamePhase) error {
	var count int
	switch phase {
	case PhaseFlop:
		count = 3
	case PhaseTurn, PhaseRiver:
		count = 1
	default:
		return nil
	}

	for i := 0; i < count && len(state.Deck) > 0; i++ {
		state.CommunityCards = append(state.CommunityCards, state.Deck[0])
		state.Deck = state.Deck[1:]
	}

	return nil
}

func (t *TexasHoldem) GetHoleCardCount() int {
	return 2
}

func (t *TexasHoldem) GetCommunityCardCount(phase GamePhase) int {
	switch phase {
	case PhaseFlop:
		return 3
	case PhaseTurn:
		return 4
	case PhaseRiver:
		return 5
	default:
		return 0
	}
}

func (t *TexasHoldem) ValidateAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error {
	// Use base validation first
	if err := t.BaseRulesEngine.ValidateAction(action, state, player, config); err != nil {
		return err
	}

	// Additional Texas Hold'em specific validation
	currentHighBet := state.LastBet
	playerBet := player.CurrentBet

	switch action.Action {
	case ActionBet:
		if currentHighBet > 0 {
			return ErrCannotBet
		}
		if action.Amount < config.BigBlind {
			return ErrBetTooSmall
		}
	case ActionRaise:
		if currentHighBet == 0 {
			return ErrCannotRaise
		}
		minRaise := state.MinRaise
		raiseTotal := action.Amount
		if raiseTotal < minRaise {
			return ErrBetTooSmall
		}
	}

	return nil
}

func (t *TexasHoldem) GetValidActions(state *TableState, player *Player, config TableConfig) []PlayerAction {
	return t.BaseRulesEngine.GetValidActions(state, player, config)
}

func (t *TexasHoldem) ProcessAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error {
	return t.BaseRulesEngine.ProcessAction(action, state, player, config)
}

func (t *TexasHoldem) ProcessBet(player *Player, amount int64, state *TableState, config TableConfig) {
	t.BaseRulesEngine.ProcessBet(player, amount, state, config)
}

func (t *TexasHoldem) ProcessCall(player *Player, state *TableState, config TableConfig) {
	t.BaseRulesEngine.ProcessCall(player, state, config)
}

func (t *TexasHoldem) ProcessRaise(player *Player, amount int64, state *TableState, config TableConfig) {
	t.BaseRulesEngine.ProcessRaise(player, amount, state, config)
}

func (t *TexasHoldem) ProcessFold(player *Player, state *TableState) {
	t.BaseRulesEngine.ProcessFold(player, state)
}

func (t *TexasHoldem) ProcessCheck(player *Player, state *TableState) {
	t.BaseRulesEngine.ProcessCheck(player, state)
}

func (t *TexasHoldem) ProcessAllIn(player *Player, state *TableState, config TableConfig) {
	t.BaseRulesEngine.ProcessAllIn(player, state, config)
}

func (t *TexasHoldem) UpdatePot(state *TableState) {
	t.BaseRulesEngine.UpdatePot(state)
}

func (t *TexasHoldem) CalculateSidePots(state *TableState, players []*Player) []Pot {
	return t.BaseRulesEngine.CalculateSidePots(state, players)
}

func (t *TexasHoldem) DistributePot(state *TableState, winners []int, evaluator poker.HandEvaluator) error {
	if len(winners) == 0 {
		return nil
	}

	// Split main pot
	potPerWinner := state.PotTotal / int64(len(winners))
	for _, winnerIdx := range winners {
		if winnerIdx >= 0 && winnerIdx < len(state.Players) && state.Players[winnerIdx] != nil {
			state.Players[winnerIdx].Chips += potPerWinner
		}
	}

	// Handle remainder
	remainder := state.PotTotal % int64(len(winners))
	if remainder > 0 && len(winners) > 0 {
		if winners[0] >= 0 && winners[0] < len(state.Players) && state.Players[winners[0]] != nil {
			state.Players[winners[0]].Chips += remainder
		}
	}

	return nil
}

func (t *TexasHoldem) EvaluateHand(holeCards []poker.Card, communityCards []poker.Card) (poker.HandRank, error) {
	eval := poker.NewHandEvaluator()
	allCards := append(holeCards, communityCards...)
	if len(allCards) != 7 {
		// Pad with empty cards if needed
		for i := len(allCards); i < 7; i++ {
			allCards = append(allCards, poker.Card{})
		}
	}
	return eval.Evaluate7Card(allCards)
}

func (t *TexasHoldem) CompareHands(hand1, hand2 poker.HandRank) int {
	eval := poker.NewHandEvaluator()
	return eval.CompareHands(hand1, hand2)
}

func (t *TexasHoldem) DetermineWinners(players []*Player, communityCards []poker.Card, evaluator poker.HandEvaluator) []int {
	var bestHand poker.HandRank
	var winners []int

	for i, player := range players {
		if player == nil || player.Status == PlayerFolded || len(player.HoleCards) != 2 {
			continue
		}

		allCards := append(player.HoleCards, communityCards...)
		if len(allCards) != 7 {
			continue
		}

		hand, err := evaluator.Evaluate7Card(allCards)
		if err != nil {
			continue
		}

		if bestHand == nil {
			bestHand = hand
			winners = []int{i}
		} else {
			cmp := evaluator.CompareHands(hand, bestHand)
			if cmp > 0 {
				bestHand = hand
				winners = []int{i}
			} else if cmp == 0 {
				winners = append(winners, i)
			}
		}
	}

	return winners
}

func (t *TexasHoldem) ShouldStartHand(players []*Player, config TableConfig) bool {
	activePlayers := t.countActivePlayers(players)
	return activePlayers >= config.MinPlayers
}

func (t *TexasHoldem) countActivePlayers(players []*Player) int {
	return t.BaseRulesEngine.countActivePlayers(players)
}

func (t *TexasHoldem) ShouldEndHand(state *TableState) bool {
	activePlayers := t.getPlayersNotFolded(state.Players)
	if len(activePlayers) <= 1 {
		return true
	}
	return false
}

func (t *TexasHoldem) getPlayersNotFolded(players []*Player) []int {
	return t.BaseRulesEngine.getPlayersNotFolded(players)
}

func (t *TexasHoldem) RotateDealerButton(state *TableState, players []*Player) {
	t.BaseRulesEngine.RotateDealerButton(state, players)
}

func (t *TexasHoldem) ResetHandState(state *TableState, config TableConfig) {
	t.BaseRulesEngine.ResetHandState(state, config)
}

func (t *TexasHoldem) StringAction(action PlayerAction) string {
	return action.String()
}

func (t *TexasHoldem) StringPhase(phase GamePhase) string {
	return phase.String()
}

func (t *TexasHoldem) StringPosition(pos Position) string {
	return pos.String()
}

// CollectBlinds collects blinds from players
func (t *TexasHoldem) CollectBlinds(state *TableState, players []*Player, config TableConfig) error {
	sbAmount, bbAmount := t.CalculateBlinds(state.HandNumber, players, config)

	// Find SB and BB positions
	var sbIdx, bbIdx int = -1, -1
	activePlayers := t.getActivePlayerIndices(players)

	if len(activePlayers) < 2 {
		return fmt.Errorf("not enough players for blinds")
	}

	// SB is first player after button
	sbIdx = activePlayers[0]
	// BB is second player after button
	if len(activePlayers) >= 2 {
		bbIdx = activePlayers[1]
	} else {
		bbIdx = activePlayers[0]
	}

	state.SmallBlindPos = players[sbIdx].SeatNumber
	state.BigBlindPos = players[bbIdx].SeatNumber

	// Collect SB
	if sbIdx >= 0 && players[sbIdx] != nil && players[sbIdx].IsConnected {
		amount := sbAmount
		if players[sbIdx].Chips < amount {
			amount = players[sbIdx].Chips
		}
		players[sbIdx].Chips -= amount
		players[sbIdx].CurrentBet = amount
		players[sbIdx].TotalInvested += amount
		state.LastBet = amount
	}

	// Collect BB
	if bbIdx >= 0 && players[bbIdx] != nil && players[bbIdx].IsConnected {
		amount := bbAmount
		if players[bbIdx].Chips < amount {
			amount = players[bbIdx].Chips
			players[bbIdx].Status = PlayerAllIn
		}
		players[bbIdx].Chips -= amount
		players[bbIdx].CurrentBet = amount
		players[bbIdx].TotalInvested += amount
		state.LastBet = amount
	}

	state.MinRaise = bbAmount
	t.UpdatePot(state)

	return nil
}

func (t *TexasHoldem) getActivePlayerIndices(players []*Player) []int {
	return t.BaseRulesEngine.getActivePlayerIndices(players)
}

// BuildPlayersToActList builds the list of players who need to act this round
func (t *TexasHoldem) BuildPlayersToActList(state *TableState) {
	state.PlayersToAct = nil
	state.PlayersActed = make(map[int]bool)

	players := t.getActivePlayerIndices(state.Players)
	if len(players) == 0 {
		return
	}

	// Start from current player and go around
	idx := 0
	for i, p := range players {
		if p == state.CurrentPlayer {
			idx = i
			break
		}
	}

	// Add all players starting from current player
	for i := 0; i < len(players); i++ {
		playerIdx := players[(idx+i)%len(players)]
		if state.Players[playerIdx] != nil && state.Players[playerIdx].Status == PlayerActive {
			state.PlayersToAct = append(state.PlayersToAct, playerIdx)
		}
	}
}

// AllPlayersActed returns true if all active players have acted this round
func (t *TexasHoldem) AllPlayersActed(state *TableState) bool {
	players := t.getActivePlayerIndices(state.Players)
	if len(players) == 0 {
		return true
	}

	for _, idx := range players {
		if state.Players[idx] != nil &&
			state.Players[idx].Status == PlayerActive &&
			!state.PlayersActed[idx] {
			return false
		}
	}

	return true
}

// AllActivePlayersAllIn returns true if all active players are all-in
func (t *TexasHoldem) AllActivePlayersAllIn(state *TableState) bool {
	players := t.getActivePlayerIndices(state.Players)
	if len(players) == 0 {
		return true
	}

	for _, idx := range players {
		if state.Players[idx] != nil && state.Players[idx].Status == PlayerActive {
			return false
		}
	}

	return true
}

// AdvanceToNextPlayer moves to the next player who needs to act
func (t *TexasHoldem) AdvanceToNextPlayer(state *TableState) {
	players := t.getActivePlayerIndices(state.Players)

	currentIdx := -1
	for i, idx := range players {
		if idx == state.CurrentPlayer {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 {
		return
	}

	for i := 1; i <= len(players); i++ {
		nextIdx := players[(currentIdx+i)%len(players)]
		if state.Players[nextIdx] != nil && state.Players[nextIdx].Status == PlayerActive {
			state.CurrentPlayer = nextIdx
			return
		}
	}
}

// CompleteBettingRound advances to the next betting phase
func (t *TexasHoldem) CompleteBettingRound(state *TableState, config TableConfig) {
	switch state.Phase {
	case PhasePreflop:
		state.Phase = PhaseFlop
		t.DealCommunityCards(state, PhaseFlop)
	case PhaseFlop:
		state.Phase = PhaseTurn
		t.DealCommunityCards(state, PhaseTurn)
	case PhaseTurn:
		state.Phase = PhaseRiver
		t.DealCommunityCards(state, PhaseRiver)
	case PhaseRiver:
		state.Phase = PhaseShowdown
	default:
		return
	}

	// Reset betting state
	for _, p := range state.Players {
		if p != nil {
			p.CurrentBet = 0
		}
	}
	state.LastBet = 0
	state.MinRaise = config.BigBlind

	// Set first player to act in new round
	t.BuildPlayersToActList(state)
}
