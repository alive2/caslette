package game

import (
	"testing"
)

func TestPokerHandEvaluation(t *testing.T) {
	evaluator := NewPokerEvaluator()

	t.Run("RoyalFlush", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Hearts, King),
			NewCard(Hearts, Queen),
			NewCard(Hearts, Jack),
			NewCard(Hearts, Ten),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != RoyalFlush {
			t.Errorf("Expected RoyalFlush, got %v", hand.Rank)
		}
	})

	t.Run("StraightFlush", func(t *testing.T) {
		cards := []Card{
			NewCard(Spades, Nine),
			NewCard(Spades, Eight),
			NewCard(Spades, Seven),
			NewCard(Spades, Six),
			NewCard(Spades, Five),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != StraightFlush {
			t.Errorf("Expected StraightFlush, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Nine {
			t.Errorf("Expected high card Nine, got %v", hand.HighCards[0])
		}
	})

	t.Run("FourOfAKind", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, Ace),
			NewCard(Clubs, Ace),
			NewCard(Hearts, King),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != FourOfAKind {
			t.Errorf("Expected FourOfAKind, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ace {
			t.Errorf("Expected four Aces, got %v", hand.HighCards[0])
		}
		if hand.Kickers[0] != King {
			t.Errorf("Expected kicker King, got %v", hand.Kickers[0])
		}
	})

	t.Run("FullHouse", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, Ace),
			NewCard(Clubs, King),
			NewCard(Hearts, King),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != FullHouse {
			t.Errorf("Expected FullHouse, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ace {
			t.Errorf("Expected trip Aces, got %v", hand.HighCards[0])
		}
		if hand.HighCards[1] != King {
			t.Errorf("Expected pair Kings, got %v", hand.HighCards[1])
		}
	})

	t.Run("Flush", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Hearts, Jack),
			NewCard(Hearts, Nine),
			NewCard(Hearts, Seven),
			NewCard(Hearts, Five),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != Flush {
			t.Errorf("Expected Flush, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ace {
			t.Errorf("Expected high card Ace, got %v", hand.HighCards[0])
		}
	})

	t.Run("Straight", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ten),
			NewCard(Spades, Nine),
			NewCard(Diamonds, Eight),
			NewCard(Clubs, Seven),
			NewCard(Hearts, Six),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != Straight {
			t.Errorf("Expected Straight, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ten {
			t.Errorf("Expected high card Ten, got %v", hand.HighCards[0])
		}
	})

	t.Run("StraightWheel", func(t *testing.T) {
		// A-2-3-4-5 straight (wheel)
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Five),
			NewCard(Diamonds, Four),
			NewCard(Clubs, Three),
			NewCard(Hearts, Two),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != Straight {
			t.Errorf("Expected Straight (wheel), got %v", hand.Rank)
		}
	})

	t.Run("ThreeOfAKind", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, Ace),
			NewCard(Clubs, King),
			NewCard(Hearts, Queen),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != ThreeOfAKind {
			t.Errorf("Expected ThreeOfAKind, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ace {
			t.Errorf("Expected trip Aces, got %v", hand.HighCards[0])
		}
		if len(hand.Kickers) != 2 {
			t.Errorf("Expected 2 kickers, got %d", len(hand.Kickers))
		}
		if hand.Kickers[0] != King || hand.Kickers[1] != Queen {
			t.Errorf("Expected kickers King and Queen, got %v and %v", hand.Kickers[0], hand.Kickers[1])
		}
	})

	t.Run("TwoPair", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, King),
			NewCard(Clubs, King),
			NewCard(Hearts, Queen),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != TwoPair {
			t.Errorf("Expected TwoPair, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ace || hand.HighCards[1] != King {
			t.Errorf("Expected pairs Aces and Kings, got %v and %v", hand.HighCards[0], hand.HighCards[1])
		}
		if hand.Kickers[0] != Queen {
			t.Errorf("Expected kicker Queen, got %v", hand.Kickers[0])
		}
	})

	t.Run("OnePair", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, King),
			NewCard(Clubs, Queen),
			NewCard(Hearts, Jack),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != OnePair {
			t.Errorf("Expected OnePair, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ace {
			t.Errorf("Expected pair Aces, got %v", hand.HighCards[0])
		}
		if len(hand.Kickers) != 3 {
			t.Errorf("Expected 3 kickers, got %d", len(hand.Kickers))
		}
	})

	t.Run("HighCard", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Jack),
			NewCard(Diamonds, Nine),
			NewCard(Clubs, Seven),
			NewCard(Hearts, Five),
		}

		hand := evaluator.EvaluateHand(cards)
		if hand.Rank != HighCard {
			t.Errorf("Expected HighCard, got %v", hand.Rank)
		}
		if hand.HighCards[0] != Ace {
			t.Errorf("Expected high card Ace, got %v", hand.HighCards[0])
		}
	})
}

func TestPokerHandComparison(t *testing.T) {
	evaluator := NewPokerEvaluator()

	t.Run("DifferentRanks", func(t *testing.T) {
		// Royal flush vs straight flush
		royalFlush := evaluator.EvaluateHand([]Card{
			NewCard(Hearts, Ace),
			NewCard(Hearts, King),
			NewCard(Hearts, Queen),
			NewCard(Hearts, Jack),
			NewCard(Hearts, Ten),
		})

		straightFlush := evaluator.EvaluateHand([]Card{
			NewCard(Spades, Nine),
			NewCard(Spades, Eight),
			NewCard(Spades, Seven),
			NewCard(Spades, Six),
			NewCard(Spades, Five),
		})

		if royalFlush.Compare(straightFlush) <= 0 {
			t.Error("Royal flush should beat straight flush")
		}
		if straightFlush.Compare(royalFlush) >= 0 {
			t.Error("Straight flush should lose to royal flush")
		}
	})

	t.Run("SameRankDifferentHighCards", func(t *testing.T) {
		// Two pair: Aces and Kings vs Aces and Queens
		acesKings := evaluator.EvaluateHand([]Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, King),
			NewCard(Clubs, King),
			NewCard(Hearts, Five),
		})

		acesQueens := evaluator.EvaluateHand([]Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, Queen),
			NewCard(Clubs, Queen),
			NewCard(Hearts, Five),
		})

		if acesKings.Compare(acesQueens) <= 0 {
			t.Error("Aces and Kings should beat Aces and Queens")
		}
	})

	t.Run("SameRankSameHighCardsDifferentKickers", func(t *testing.T) {
		// Pair of Aces with different kickers
		acesKingKicker := evaluator.EvaluateHand([]Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, King),
			NewCard(Clubs, Seven),
			NewCard(Hearts, Five),
		})

		acesQueenKicker := evaluator.EvaluateHand([]Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, Queen),
			NewCard(Clubs, Seven),
			NewCard(Hearts, Five),
		})

		if acesKingKicker.Compare(acesQueenKicker) <= 0 {
			t.Error("Aces with King kicker should beat Aces with Queen kicker")
		}
	})

	t.Run("Tie", func(t *testing.T) {
		// Identical hands
		hand1 := evaluator.EvaluateHand([]Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, King),
			NewCard(Diamonds, Queen),
			NewCard(Clubs, Jack),
			NewCard(Hearts, Ten),
		})

		hand2 := evaluator.EvaluateHand([]Card{
			NewCard(Diamonds, Ace),
			NewCard(Clubs, King),
			NewCard(Hearts, Queen),
			NewCard(Spades, Jack),
			NewCard(Diamonds, Ten),
		})

		if hand1.Compare(hand2) != 0 {
			t.Error("Identical hands should tie")
		}
	})
}

func TestFindBestHand(t *testing.T) {
	evaluator := NewPokerEvaluator()

	t.Run("SevenCardBestHand", func(t *testing.T) {
		// 7 cards that should make a full house (AAA KK)
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, Ace),
			NewCard(Diamonds, Ace),
			NewCard(Clubs, King),
			NewCard(Hearts, King),
			NewCard(Spades, Seven),
			NewCard(Diamonds, Two),
		}

		bestHand := evaluator.FindBestHand(cards)
		if bestHand.Rank != FullHouse {
			t.Errorf("Expected FullHouse from 7 cards, got %v", bestHand.Rank)
		}
		if bestHand.HighCards[0] != Ace {
			t.Errorf("Expected trip Aces, got %v", bestHand.HighCards[0])
		}
		if bestHand.HighCards[1] != King {
			t.Errorf("Expected pair Kings, got %v", bestHand.HighCards[1])
		}
	})

	t.Run("LessThanFiveCards", func(t *testing.T) {
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, King),
			NewCard(Diamonds, Queen),
		}

		hand := evaluator.FindBestHand(cards)
		if hand.Rank != HighCard {
			t.Errorf("Expected HighCard for less than 5 cards, got %v", hand.Rank)
		}
	})
}

func TestHandRankString(t *testing.T) {
	tests := []struct {
		rank     HandRank
		expected string
	}{
		{HighCard, "High Card"},
		{OnePair, "One Pair"},
		{TwoPair, "Two Pair"},
		{ThreeOfAKind, "Three of a Kind"},
		{Straight, "Straight"},
		{Flush, "Flush"},
		{FullHouse, "Full House"},
		{FourOfAKind, "Four of a Kind"},
		{StraightFlush, "Straight Flush"},
		{RoyalFlush, "Royal Flush"},
	}

	for _, test := range tests {
		if result := test.rank.String(); result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}
