package poker

import (
	"fmt"
	"sort"
)

// Card represents a playing card
type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

// Rank enumeration
type Rank int8

const (
	Rank2 Rank = iota
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJ
	RankQ
	RankK
	RankA
)

func (r Rank) String() string {
	names := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	if r >= 0 && int(r) < len(names) {
		return names[r]
	}
	return "?"
}

// Suit enumeration
type Suit int8

const (
	SuitClubs Suit = iota
	SuitDiamonds
	SuitHearts
	SuitSpades
)

func (s Suit) String() string {
	names := []string{"♣", "♦", "♥", "♠"}
	if s >= 0 && int(s) < len(names) {
		return names[s]
	}
	return "?"
}

// NewCard creates a card from rank and suit
func NewCard(rank Rank, suit Suit) Card {
	return Card{Rank: rank, Suit: suit}
}

// ToID converts card to 0-51 ID for efficient storage
func (c Card) ToID() int {
	return int(c.Rank)*4 + int(c.Suit)
}

// FromID creates card from 0-51 ID
func FromID(id int) Card {
	return Card{
		Rank: Rank(id / 4),
		Suit: Suit(id % 4),
	}
}

// String returns card representation like "A♠"
func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit)
}

// HandRank represents the strength of a poker hand
type HandRank int

const (
	HighCard HandRank = iota
	Pair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

func (h HandRank) String() string {
	names := []string{
		"High Card", "Pair", "Two Pair", "Three of a Kind",
		"Straight", "Flush", "Full House", "Four of a Kind",
		"Straight Flush", "Royal Flush",
	}
	if h >= 0 && int(h) < len(names) {
		return names[h]
	}
	return "Unknown"
}

// EvaluatedHand represents a hand with its evaluation result
type EvaluatedHand struct {
	Cards       []Card      `json:"cards"`
	Rank        HandRank    `json:"rank"`
	Kickers     []Rank      `json:"kickers"`
	TieBreakers []Rank      `json:"tie_breakers"`
	Score       uint32      `json:"score"` // Higher is better
}

// HandEvaluator evaluates poker hands
type HandEvaluator struct{}

// NewHandEvaluator creates a new evaluator with precomputed lookup tables
func NewHandEvaluator() *HandEvaluator {
	return &HandEvaluator{}
}

// Evaluate7Card evaluates the best 5-card hand from 7 cards
func (e *HandEvaluator) Evaluate7Card(cards []Card) (*EvaluatedHand, error) {
	if len(cards) != 7 {
		return nil, fmt.Errorf("exactly 7 cards required, got %d", len(cards))
	}

	// Use direct evaluation for 7 cards
	// For production, use the Rust FFI evaluator for maximum performance
	return e.evaluateDirect(cards)
}

// evaluateDirect uses a direct algorithm for hand evaluation
// Production code should use Rust FFI for better performance
func (e *HandEvaluator) evaluateDirect(cards []Card) (*EvaluatedHand, error) {
	// Sort by rank for easier evaluation
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank > sorted[j].Rank
	})

	// Check for flush
	flushSuit, flushCards := e.checkFlush(sorted)
	if flushCards != nil {
		// Check for straight flush
		if sf := e.checkStraightFlush(flushCards); sf != nil {
			return sf, nil
		}
		// Return best 5 cards of flush
		return &EvaluatedHand{
			Cards:       flushCards[:5],
			Rank:        Flush,
			Kickers:     e.getKickers(flushCards, 5),
			TieBreakers: e.getTieBreakers(flushCards),
		}, nil
	}

	// Check for pairs, three of a kind, four of a kind
	fourKind := e.checkFourOfAKind(sorted)
	if fourKind != nil {
		return fourKind, nil
	}

	fullHouse := e.checkFullHouse(sorted)
	if fullHouse != nil {
		return fullHouse, nil
	}

	// Check for straight
	if straight := e.checkStraight(sorted); straight != nil {
		return straight, nil
	}

	threeKind := e.checkThreeOfAKind(sorted)
	if threeKind != nil {
		return threeKind, nil
	}

	twoPair := e.checkTwoPair(sorted)
	if twoPair != nil {
		return twoPair, nil
	}

	pair := e.checkPair(sorted)
	if pair != nil {
		return pair, nil
	}

	// High card
	return &EvaluatedHand{
		Cards:       sorted[:5],
		Rank:        HighCard,
		Kickers:     e.getKickers(sorted, 5),
		TieBreakers: e.getTieBreakers(sorted),
	}, nil
}

func (e *HandEvaluator) checkFlush(cards []Card) (Suit, []Card) {
	suitCounts := make(map[Suit][]Card)
	for _, c := range cards {
		suitCounts[c.Suit] = append(suitCounts[c.Suit], c)
	}

	for suit, suitCards := range suitCounts {
		if len(suitCards) >= 5 {
			// Sort by rank
			sort.Slice(suitCards, func(i, j int) bool {
				return suitCards[i].Rank > suitCards[j].Rank
			})
			return suit, suitCards
		}
	}
	return 0, nil
}

func (e *HandEvaluator) checkStraightFlush(flushCards []Card) *EvaluatedHand {
	if len(flushCards) < 5 {
		return nil
	}

	// Check for straight in flush cards
	for i := 0; i <= len(flushCards)-5; i++ {
		hand := flushCards[i : i+5]
		if e.isConsecutiveRanks(hand) {
			rank := StraightFlush
			if hand[0].Rank == RankA && hand[4].Rank == Rank10 {
				rank = RoyalFlush
			}
			return &EvaluatedHand{
				Cards:       hand,
				Rank:        rank,
				TieBreakers: e.getTieBreakers(hand),
			}
		}
	}
	return nil
}

func (e *HandEvaluator) checkFourOfAKind(cards []Card) *EvaluatedHand {
	rankCounts := e.countRanks(cards)

	for rank, count := range rankCounts {
		if count >= 4 {
			// Find the four of a kind cards
			var fourOfKind []Card
			var kickers []Card
			for _, c := range cards {
				if c.Rank == rank {
					fourOfKind = append(fourOfKind, c)
				} else {
					kickers = append(kickers, c)
				}
			}
			// Sort kickers by rank
			sort.Slice(kickers, func(i, j int) bool {
				return kickers[i].Rank > kickers[j].Rank
			})
			return &EvaluatedHand{
				Cards:       fourOfKind,
				Rank:        FourOfAKind,
				Kickers:     kickers[:1],
				TieBreakers: append([]Rank{rank}, kickers[0].Rank),
			}
		}
	}
	return nil
}

func (e *HandEvaluator) checkFullHouse(cards []Card) *EvaluatedHand {
	rankCounts := e.countRanks(cards)

	var threeOfKindRank Rank
	var pairRank Rank

	for rank, count := range rankCounts {
		if count >= 3 {
			if threeOfKindRank == 0 || rank > threeOfKindRank {
				threeOfKindRank = rank
			}
		}
	}

	if threeOfKindRank == 0 {
		return nil
	}

	for rank, count := range rankCounts {
		if count >= 2 && rank != threeOfKindRank {
			if pairRank == 0 || rank > pairRank {
				pairRank = rank
			}
		}
	}

	if pairRank == 0 {
		return nil
	}

	var threeOfKind []Card
	var pair []Card
	for _, c := range cards {
		if c.Rank == threeOfKindRank {
			threeOfKind = append(threeOfKind, c)
		} else if c.Rank == pairRank && len(pair) < 2 {
			pair = append(pair, c)
		}
	}

	return &EvaluatedHand{
		Cards:       append(threeOfKind[:3], pair...),
		Rank:        FullHouse,
		TieBreakers: []Rank{threeOfKindRank, pairRank},
	}
}

func (e *HandEvaluator) checkStraight(cards []Card) *EvaluatedHand {
	for i := 0; i <= len(cards)-5; i++ {
		hand := cards[i : i+5]
		if e.isConsecutiveRanks(hand) {
			return &EvaluatedHand{
				Cards:       hand,
				Rank:        Straight,
				TieBreakers: e.getTieBreakers(hand),
			}
		}
	}
	return nil
}

func (e *HandEvaluator) checkThreeOfAKind(cards []Card) *EvaluatedHand {
	rankCounts := e.countRanks(cards)

	var threeOfKindRank Rank
	for rank, count := range rankCounts {
		if count >= 3 {
			if threeOfKindRank == 0 || rank > threeOfKindRank {
				threeOfKindRank = rank
			}
		}
	}

	if threeOfKindRank == 0 {
		return nil
	}

	var threeOfKind []Card
	var kickers []Card
	for _, c := range cards {
		if c.Rank == threeOfKindRank {
			threeOfKind = append(threeOfKind, c)
		} else {
			kickers = append(kickers, c)
		}
	}
	sort.Slice(kickers, func(i, j int) bool {
		return kickers[i].Rank > kickers[j].Rank
	})

	return &EvaluatedHand{
		Cards:       append(threeOfKind[:3], kickers[:2]...),
		Rank:        ThreeOfAKind,
		Kickers:     kickers[:2],
		TieBreakers: append([]Rank{threeOfKindRank}, kickers[0].Rank, kickers[1].Rank),
	}
}

func (e *HandEvaluator) checkTwoPair(cards []Card) *EvaluatedHand {
	rankCounts := e.countRanks(cards)

	var pairs []Rank
	for rank, count := range rankCounts {
		if count >= 2 {
			pairs = append(pairs, rank)
		}
	}

	if len(pairs) < 2 {
		return nil
	}

	// Sort pairs by rank (higher first)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i] > pairs[j]
	})
	pairs = pairs[:2]

	var pairCards []Card
	var kicker Card
	for _, c := range cards {
		if c.Rank == pairs[0] || c.Rank == pairs[1] {
			if len(pairCards) < 4 {
				pairCards = append(pairCards, c)
			}
		} else if kicker.Rank == 0 || c.Rank > kicker.Rank {
			kicker = c
		}
	}

	return &EvaluatedHand{
		Cards:       append(pairCards, kicker),
		Rank:        TwoPair,
		Kickers:     []Rank{kicker.Rank},
		TieBreakers: []Rank{pairs[0], pairs[1], kicker.Rank},
	}
}

func (e *HandEvaluator) checkPair(cards []Card) *EvaluatedHand {
	rankCounts := e.countRanks(cards)

	var pairRank Rank
	for rank, count := range rankCards {
		if count >= 2 {
			if pairRank == 0 || rank > pairRank {
				pairRank = rank
			}
		}
	}

	if pairRank == 0 {
		return nil
	}

	var pair []Card
	var kickers []Card
	for _, c := range cards {
		if c.Rank == pairRank {
			pair = append(pair, c)
		} else {
			kickers = append(kickers, c)
		}
	}
	sort.Slice(kickers, func(i, j int) bool {
		return kickers[i].Rank > kickers[j].Rank
	})

	return &EvaluatedHand{
		Cards:       append(pair[:2], kickers[:3]...),
		Rank:        Pair,
		Kickers:     kickers[:3],
		TieBreakers: append([]Rank{pairRank}, kickers[0].Rank, kickers[1].Rank, kickers[2].Rank),
	}
}

func (e *HandEvaluator) countRanks(cards []Card) map[Rank]int {
	counts := make(map[Rank]int)
	for _, c := range cards {
		counts[c.Rank]++
	}
	return counts
}

func (e *HandEvaluator) isConsecutiveRanks(cards []Card) bool {
	if len(cards) != 5 {
		return false
	}
	for i := 0; i < 4; i++ {
		if cards[i].Rank-cards[i+1].Rank != 1 {
			// Check for wheel (A-2-3-4-5)
			if i == 0 && cards[0].Rank == RankA && cards[1].Rank == Rank5 &&
				cards[2].Rank == Rank4 && cards[3].Rank == Rank3 && cards[4].Rank == Rank2 {
				return true
			}
			return false
		}
	}
	return true
}

func (e *HandEvaluator) getKickers(cards []Card, count int) []Rank {
	if len(cards) < count {
		count = len(cards)
	}
	kickers := make([]Rank, count)
	for i := 0; i < count; i++ {
		kickers[i] = cards[i].Rank
	}
	return kickers
}

func (e *HandEvaluator) getTieBreakers(cards []Card) []Rank {
	tieBreakers := make([]Rank, len(cards))
	for i, c := range cards {
		tieBreakers[i] = c.Rank
	}
	return tieBreakers
}

// CompareHands returns 1 if hand1 > hand2, -1 if hand1 < hand2, 0 if equal
func (e *HandEvaluator) CompareHands(h1, h2 *EvaluatedHand) int {
	if h1.Rank != h2.Rank {
		if h1.Rank > h2.Rank {
			return 1
		}
		return -1
	}

	// Same rank, compare tie breakers
	for i := 0; i < len(h1.TieBreakers) && i < len(h2.TieBreakers); i++ {
		if h1.TieBreakers[i] != h2.TieBreakers[i] {
			if h1.TieBreakers[i] > h2.TieBreakers[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}
