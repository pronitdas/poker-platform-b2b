package poker

import (
	"testing"
)

func TestCardCreation(t *testing.T) {
	// Test card creation and ID conversion
	card := NewCard(RankA, SuitSpades)
	if card.Rank != RankA || card.Suit != SuitSpades {
		t.Errorf("Expected Ace of Spades, got %v", card)
	}

	id := card.ToID()
	if id != 51 { // Ace of Spades should be 51 (12*4 + 3)
		t.Errorf("Expected card ID 51, got %d", id)
	}

	restored := FromID(id)
	if restored.Rank != RankA || restored.Suit != SuitSpades {
		t.Errorf("Card from ID should match original")
	}
}

func TestHandEvaluation(t *testing.T) {
	evaluator := NewHandEvaluator()

	tests := []struct {
		name     string
		cards    []Card
		expected HandRank
	}{
		{
			name:     "High card A",
			cards:    []Card{{RankA, SuitSpades}, {RankK, SuitHearts}, {RankQ, SuitDiamonds}, {RankJ, SuitClubs}, {Rank10, SuitSpades}},
			expected: HighCard,
		},
		{
			name:     "Pair of Aces",
			cards:    []Card{{RankA, SuitSpades}, {RankA, SuitHearts}, {RankK, SuitDiamonds}, {RankQ, SuitClubs}, {RankJ, SuitSpades}},
			expected: Pair,
		},
		{
			name:     "Two Pair",
			cards:    []Card{{RankA, SuitSpades}, {RankA, SuitHearts}, {RankK, SuitDiamonds}, {RankK, SuitClubs}, {RankQ, SuitSpades}},
			expected: TwoPair,
		},
		{
			name:     "Three of a Kind",
			cards:    []Card{{RankA, SuitSpades}, {RankA, SuitHearts}, {RankA, SuitDiamonds}, {RankK, SuitClubs}, {RankQ, SuitSpades}},
			expected: ThreeOfAKind,
		},
		{
			name:     "Straight",
			cards:    []Card{{RankA, SuitSpades}, {RankK, SuitHearts}, {RankQ, SuitDiamonds}, {RankJ, SuitClubs}, {Rank10, SuitSpades}},
			expected: Straight,
		},
		{
			name:     "Flush",
			cards:    []Card{{RankA, SuitSpades}, {RankK, SuitSpades}, {RankQ, SuitSpades}, {RankJ, SuitSpades}, {Rank10, SuitSpades}},
			expected: Flush,
		},
		{
			name:     "Full House",
			cards:    []Card{{RankA, SuitSpades}, {RankA, SuitHearts}, {RankA, SuitDiamonds}, {RankK, SuitClubs}, {RankK, SuitSpades}},
			expected: FullHouse,
		},
		{
			name:     "Four of a Kind",
			cards:    []Card{{RankA, SuitSpades}, {RankA, SuitHearts}, {RankA, SuitDiamonds}, {RankA, SuitClubs}, {RankK, SuitSpades}},
			expected: FourOfAKind,
		},
		{
			name:     "Straight Flush",
			cards:    []Card{{RankA, SuitSpades}, {RankK, SuitSpades}, {RankQ, SuitSpades}, {RankJ, SuitSpades}, {Rank10, SuitSpades}},
			expected: StraightFlush,
		},
		{
			name:     "Royal Flush",
			cards:    []Card{{RankA, SuitSpades}, {RankK, SuitSpades}, {RankQ, SuitSpades}, {RankJ, SuitSpades}, {Rank10, SuitSpades}},
			expected: RoyalFlush,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand, err := evaluator.evaluateDirect(tt.cards)
			if err != nil {
				t.Fatalf("Failed to evaluate hand: %v", err)
			}
			if hand.Rank != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, hand.Rank)
			}
		})
	}
}

func Test7CardEvaluation(t *testing.T) {
	evaluator := NewHandEvaluator()

	// Test 7-card evaluation (2 hole cards + 5 community)
	holeCards := []Card{{RankA, SuitSpades}, {RankK, SuitHearts}}
	communityCards := []Card{{RankQ, SuitDiamonds}, {RankJ, SuitClubs}, {Rank10, SuitSpades}, {Rank9, SuitHearts}, {Rank2, SuitDiamonds}}

	allCards := append(holeCards, communityCards...)
	hand, err := evaluator.Evaluate7Card(allCards)
	if err != nil {
		t.Fatalf("Failed to evaluate 7-card hand: %v", err)
	}

	// Should be a straight (A-K-Q-J-10-9-2, best 5 is A-K-Q-J-10)
	if hand.Rank != Straight {
		t.Errorf("Expected Straight, got %v", hand.Rank)
	}
}

func TestHandComparison(t *testing.T) {
	evaluator := NewHandEvaluator()

	// Straight flush beats flush
	sf := &EvaluatedHand{Rank: StraightFlush}
	flush := &EvaluatedHand{Rank: Flush}

	cmp := evaluator.CompareHands(sf, flush)
	if cmp <= 0 {
		t.Errorf("StraightFlush should beat Flush")
	}

	// Full house beats three of a kind
	fh := &EvaluatedHand{Rank: FullHouse, TieBreakers: []Rank{RankA, RankK}}
	tk := &EvaluatedHand{Rank: ThreeOfAKind, TieBreakers: []Rank{RankA}}

	cmp = evaluator.CompareHands(fh, tk)
	if cmp <= 0 {
		t.Errorf("FullHouse should beat ThreeOfAKind")
	}
}

func TestWheelStraight(t *testing.T) {
	evaluator := NewHandEvaluator()

	// A-2-3-4-5 wheel
	cards := []Card{
		{RankA, SuitSpades},
		{Rank2, SuitHearts},
		{Rank3, SuitDiamonds},
		{Rank4, SuitClubs},
		{Rank5, SuitSpades},
	}

	hand, err := evaluator.evaluateDirect(cards)
	if err != nil {
		t.Fatalf("Failed to evaluate wheel: %v", err)
	}

	if hand.Rank != Straight {
		t.Errorf("Expected Straight (wheel), got %v", hand.Rank)
	}
}
