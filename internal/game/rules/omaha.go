package rules

import (
	"fmt"

	"poker-platform/pkg/poker"
)

// Omaha implements the RulesEngine interface for Pot-Limit Omaha
type Omaha struct {
	*BaseRulesEngine
}

// NewOmaha creates a new Pot-Limit Omaha rules engine
func NewOmaha() *Omaha {
	return &Omaha{
		BaseRulesEngine: &BaseRulesEngine{},
	}
}

func (o *Omaha) GameType() GameType {
	return GameTypeOmaha
}

func (o *Omaha) BettingType() BettingType {
	return BettingTypePotLimit
}

func (o *Omaha) Name() string {
	return "Omaha"
}

func (o *Omaha) Version() string {
	return "1.0.0"
}

func (o *Omaha) DefaultConfig() TableConfig {
	return TableConfig{
		GameType:          GameTypeOmaha,
		BettingType:       BettingTypePotLimit,
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

func (o *Omaha) CalculateMinBet(config TableConfig) int64 {
	return config.BigBlind
}

func (o *Omaha) CalculateMinRaise(currentBet, minBet int64, config TableConfig) int64 {
	// Pot-limit: minimum raise is the size of the pot
	return minBet
}

func (o *Omaha) CalculateBetAmount(player *Player, targetAmount, config TableConfig) (amount int64, isAllIn bool) {
	if targetAmount >= player.Chips {
		return player.Chips, true
	}
	return targetAmount, false
}

func (o *Omaha) ValidateBetSizing(amount int64, player *Player, currentBet, minBet, potSize int64, config TableConfig) error {
	if amount <= 0 {
		return ErrInvalidAction
	}

	// Pot-limit: bet cannot exceed pot size
	maxBet := potSize + currentBet
	if amount > maxBet {
		return fmt.Errorf("bet exceeds pot limit: max %d", maxBet)
	}

	if amount > player.Chips {
		return ErrNotEnoughChips
	}

	return nil
}

func (o *Omaha) Phases() []GamePhase {
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

func (o *Omaha) NextPhase(currentPhase GamePhase, state *TableState) GamePhase {
	switch currentPhase {
	case PhaseWaiting:
		if o.ShouldStartHand(state.Players, TableConfig{}) {
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

func (o *Omaha) DetermineFirstActor(phase GamePhase, state *TableState, positions map[int]Position) int {
	players := state.Players
	if len(players) == 0 {
		return -1
	}

	switch phase {
	case PhasePreflop:
		bbPos := state.BigBlindPos
		idx := o.findPlayerIndex(players, bbPos)
		if idx >= 0 {
			return (idx + 1) % len(players)
		}
	case PhaseFlop, PhaseTurn, PhaseRiver:
		sbPos := state.SmallBlindPos
		idx := o.findPlayerIndex(players, sbPos)
		if idx >= 0 {
			return (idx + 1) % len(players)
		}
	}

	for i := 0; i < len(players); i++ {
		if players[i] != nil && players[i].Status == PlayerActive {
			return i
		}
	}

	return -1
}

func (o *Omaha) findPlayerIndex(players []*Player, targetPos int) int {
	for i := 0; i < len(players); i++ {
		if players[i] != nil && players[i].SeatNumber == targetPos {
			return i
		}
	}
	return -1
}

func (o *Omaha) DealHoleCards(state *TableState, players []*Player) error {
	// Deal 4 hole cards to each active player
	for _, player := range players {
		if player != nil && player.Status == PlayerActive && player.IsConnected {
			if len(state.Deck) >= 4 {
				player.HoleCards = []poker.Card{
					state.Deck[0], state.Deck[1],
					state.Deck[2], state.Deck[3],
				}
				state.Deck = state.Deck[4:]
			} else {
				return fmt.Errorf("deck exhausted")
			}
		}
	}
	return nil
}

func (o *Omaha) DealCommunityCards(state *TableState, phase GamePhase) error {
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

func (o *Omaha) GetHoleCardCount() int {
	return 4
}

func (o *Omaha) GetCommunityCardCount(phase GamePhase) int {
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

func (o *Omaha) ValidateAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error {
	if err := o.BaseRulesEngine.ValidateAction(action, state, player, config); err != nil {
		return err
	}

	// Pot-limit specific validation
	if action.Action == ActionBet || action.Action == ActionRaise {
		if err := o.ValidateBetSizing(action.Amount, player, state.LastBet, state.MinRaise, state.PotTotal, config); err != nil {
			return err
		}
	}

	return nil
}

func (o *Omaha) GetValidActions(state *TableState, player *Player, config TableConfig) []PlayerAction {
	return o.BaseRulesEngine.GetValidActions(state, player, config)
}

func (o *Omaha) ProcessAction(action PlayerActionRequest, state *TableState, player *Player, config TableConfig) error {
	return o.BaseRulesEngine.ProcessAction(action, state, player, config)
}

func (o *Omaha) ProcessBet(player *Player, amount int64, state *TableState, config TableConfig) {
	o.BaseRulesEngine.ProcessBet(player, amount, state, config)
}

func (o *Omaha) ProcessCall(player *Player, state *TableState, config TableConfig) {
	o.BaseRulesEngine.ProcessCall(player, state, config)
}

func (o *Omaha) ProcessRaise(player *Player, amount int64, state *TableState, config TableConfig) {
	o.BaseRulesEngine.ProcessRaise(player, amount, state, config)
}

func (o *Omaha) ProcessFold(player *Player, state *TableState) {
	o.BaseRulesEngine.ProcessFold(player, state)
}

func (o *Omaha) ProcessCheck(player *Player, state *TableState) {
	o.BaseRulesEngine.ProcessCheck(player, state)
}

func (o *Omaha) ProcessAllIn(player *Player, state *TableState, config TableConfig) {
	o.BaseRulesEngine.ProcessAllIn(player, state, config)
}

func (o *Omaha) UpdatePot(state *TableState) {
	o.BaseRulesEngine.UpdatePot(state)
}

func (o *Omaha) CalculateSidePots(state *TableState, players []*Player) []Pot {
	return o.BaseRulesEngine.CalculateSidePots(state, players)
}

func (o *Omaha) DistributePot(state *TableState, winners []int, evaluator poker.HandEvaluator) error {
	if len(winners) == 0 {
		return nil
	}

	potPerWinner := state.PotTotal / int64(len(winners))
	for _, winnerIdx := range winners {
		if winnerIdx >= 0 && winnerIdx < len(state.Players) && state.Players[winnerIdx] != nil {
			state.Players[winnerIdx].Chips += potPerWinner
		}
	}

	remainder := state.PotTotal % int64(len(winners))
	if remainder > 0 && len(winners) > 0 {
		if winners[0] >= 0 && winners[0] < len(state.Players) && state.Players[winners[0]] != nil {
			state.Players[winners[0]].Chips += remainder
		}
	}

	return nil
}

func (o *Omaha) EvaluateHand(holeCards []poker.Card, communityCards []poker.Card) (poker.HandRank, error) {
	// Omaha: Must use exactly 2 hole cards and 3 community cards
	if len(holeCards) != 4 || len(communityCards) < 3 {
		return nil, fmt.Errorf("invalid omaha hand: need 4 hole cards and at least 3 community cards")
	}

	eval := poker.NewHandEvaluator()
	bestHand := poker.HandRank{}

	// Evaluate all combinations of 2 hole cards + 3 community cards
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			// Use exactly 2 hole cards
			handCards := []poker.Card{holeCards[i], holeCards[j]}

			// Add all 5 community cards
			handCards = append(handCards, communityCards...)

			if len(handCards) != 7 {
				continue
			}

			hand, err := eval.Evaluate7Card(handCards)
			if err != nil {
				continue
			}

			if bestHand == nil || eval.CompareHands(hand, bestHand) > 0 {
				bestHand = hand
			}
		}
	}

	return bestHand, nil
}

func (o *Omaha) CompareHands(hand1, hand2 poker.HandRank) int {
	eval := poker.NewHandEvaluator()
	return eval.CompareHands(hand1, hand2)
}

func (o *Omaha) DetermineWinners(players []*Player, communityCards []poker.Card, evaluator poker.HandEvaluator) []int {
	var bestHand poker.HandRank
	var winners []int

	for i, player := range players {
		if player == nil || player.Status == PlayerFolded || len(player.HoleCards) != 4 {
			continue
		}

		hand, err := o.EvaluateHand(player.HoleCards, communityCards)
		if err != nil {
			continue
		}

		if bestHand == nil {
			bestHand = hand
			winners = []int{i}
		} else {
			cmp := o.CompareHands(hand, bestHand)
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

func (o *Omaha) ShouldStartHand(players []*Player, config TableConfig) bool {
	activePlayers := o.countActivePlayers(players)
	return activePlayers >= config.MinPlayers
}

func (o *Omaha) countActivePlayers(players []*Player) int {
	return o.BaseRulesEngine.countActivePlayers(players)
}

func (o *Omaha) ShouldEndHand(state *TableState) bool {
	activePlayers := o.getPlayersNotFolded(state.Players)
	if len(activePlayers) <= 1 {
		return true
	}
	return false
}

func (o *Omaha) getPlayersNotFolded(players []*Player) []int {
	return o.BaseRulesEngine.getPlayersNotFolded(players)
}

func (o *Omaha) RotateDealerButton(state *TableState, players []*Player) {
	o.BaseRulesEngine.RotateDealerButton(state, players)
}

func (o *Omaha) ResetHandState(state *TableState, config TableConfig) {
	o.BaseRulesEngine.ResetHandState(state, config)
}

func (o *Omaha) StringAction(action PlayerAction) string {
	return action.String()
}

func (o *Omaha) StringPhase(phase GamePhase) string {
	return phase.String()
}

func (o *Omaha) StringPosition(pos Position) string {
	return pos.String()
}

// OmahaHiLo implements Omaha High-Low Split (8 or Better)
type OmahaHiLo struct {
	*Omaha
}

// NewOmahaHiLo creates a new Omaha Hi-Lo rules engine
func NewOmahaHiLo() *OmahaHiLo {
	return &OmahaHiLo{
		Omaha: NewOmaha(),
	}
}

func (o *OmahaHiLo) GameType() GameType {
	return GameTypeOmahaHiLo
}

func (o *OmahaHiLo) Name() string {
	return "Omaha Hi-Lo"
}

func (o *OmahaHiLo) EvaluateHand(holeCards []poker.Card, communityCards []poker.Card) (poker.HandRank, error) {
	// For Hi-Lo, we need to evaluate both high and low hands
	// This is a simplified version - full implementation would track both
	return o.Omaha.EvaluateHand(holeCards, communityCards)
}

// FiveCardDraw implements Five Card Draw poker
type FiveCardDraw struct {
	*BaseRulesEngine
}

// NewFiveCardDraw creates a new Five Card Draw rules engine
func NewFiveCardDraw() *FiveCardDraw {
	return &FiveCardDraw{
		BaseRulesEngine: &BaseRulesEngine{},
	}
}

func (f *FiveCardDraw) GameType() GameType {
	return GameTypeFiveCardDraw
}

func (f *FiveCardDraw) BettingType() BettingType {
	return BettingTypeNoLimit
}

func (f *FiveCardDraw) Name() string {
	return "Five Card Draw"
}

func (f *FiveCardDraw) Version() string {
	return "1.0.0"
}

func (f *FiveCardDraw) DefaultConfig() TableConfig {
	return TableConfig{
		GameType:     GameTypeFiveCardDraw,
		BettingType:  BettingTypeNoLimit,
		MinPlayers:   2,
		MaxPlayers:   6,
		SmallBlind:   5,
		BigBlind:     10,
		BuyInMin:     100,
		BuyInMax:     5000,
		ActionTimeout: 30,
	}
}

func (f *FiveCardDraw) Phases() []GamePhase {
	return []GamePhase{
		PhaseWaiting,
		PhasePreflop, // First betting round (antes)
		PhasePreDeal, // Draw phase
		PhaseFlop,    // Second betting round after draw
		PhaseShowdown,
		PhaseHandComplete,
	}
}

func (f *FiveCardDraw) DealHoleCards(state *TableState, players []*Player) error {
	// Deal 5 hole cards to each active player
	for _, player := range players {
		if player != nil && player.Status == PlayerActive && player.IsConnected {
			if len(state.Deck) >= 5 {
				player.HoleCards = make([]poker.Card, 5)
				copy(player.HoleCards, state.Deck[:5])
				state.Deck = state.Deck[5:]
			} else {
				return fmt.Errorf("deck exhausted")
			}
		}
	}
	return nil
}

func (f *FiveCardDraw) GetHoleCardCount() int {
	return 5
}

func (f *FiveCardDraw) GetCommunityCardCount(phase GamePhase) int {
	return 0 // No community cards in 5 Card Draw
}

func (f *FiveCardDraw) EvaluateHand(holeCards []poker.Card, communityCards []poker.Card) (poker.HandRank, error) {
	if len(holeCards) != 5 {
		return nil, fmt.Errorf("invalid five card draw hand: need 5 cards")
	}

	eval := poker.NewHandEvaluator()
	return eval.Evaluate5Card(holeCards)
}

// SevenCardStud implements Seven Card Stud poker
type SevenCardStud struct {
	*BaseRulesEngine
}

// NewSevenCardStud creates a new Seven Card Stud rules engine
func NewSevenCardStud() *SevenCardStud {
	return &SevenCardStud{
		BaseRulesEngine: &BaseRulesEngine{},
	}
}

func (s *SevenCardStud) GameType() GameType {
	return GameTypeSevenCardStud
}

func (s *SevenCardStud) BettingType() BettingType {
	return BettingTypeFixedLimit
}

func (s *SevenCardStud) Name() string {
	return "Seven Card Stud"
}

func (s *SevenCardStud) Version() string {
	return "1.0.0"
}

func (s *SevenCardStud) DefaultConfig() TableConfig {
	return TableConfig{
		GameType:     GameTypeSevenCardStud,
		BettingType:  BettingTypeFixedLimit,
		MinPlayers:   2,
		MaxPlayers:   8,
		SmallBlind:   5,
		BigBlind:     10,
		BuyInMin:     100,
		BuyInMax:     5000,
		ActionTimeout: 30,
	}
}

func (s *SevenCardStud) Phases() []GamePhase {
	return []GamePhase{
		PhaseWaiting,
		PhasePreDeal,  // Third street
		PhasePreflop,  // Fourth street
		PhaseFlop,     // Fifth street
		PhaseTurn,     // Sixth street
		PhaseRiver,    // Seventh street (river)
		PhaseShowdown,
		PhaseHandComplete,
	}
}

func (s *SevenCardStud) DealHoleCards(state *TableState, players []*Player) error {
	// Deal 3 hole cards (2 face down, 1 face up) in 7 Card Stud
	for _, player := range players {
		if player != nil && player.Status == PlayerActive && player.IsConnected {
			if len(state.Deck) >= 3 {
				player.HoleCards = []poker.Card{state.Deck[0], state.Deck[1], state.Deck[2]}
				state.Deck = state.Deck[3:]
			} else {
				return fmt.Errorf("deck exhausted")
			}
		}
	}
	return nil
}

func (s *SevenCardStud) GetHoleCardCount() int {
	return 7
}

func (s *SevenCardStud) GetCommunityCardCount(phase GamePhase) int {
	return 0 // No community cards in 7 Card Stud
}

func (s *SevenCardStud) EvaluateHand(holeCards []poker.Card, communityCards []poker.Card) (poker.HandRank, error) {
	// 7 Card Stud: use best 5 cards from 7
	if len(holeCards) != 7 {
		return nil, fmt.Errorf("invalid seven card stud hand: need 7 cards")
	}

	eval := poker.NewHandEvaluator()
	return eval.Evaluate7Card(holeCards)
}
