package game

import (
	"testing"
)

func TestCard(t *testing.T) {
	t.Run("NewCard", func(t *testing.T) {
		card := NewCard(Hearts, Ace)
		if card.Suit != Hearts {
			t.Errorf("Expected suit %v, got %v", Hearts, card.Suit)
		}
		if card.Rank != Ace {
			t.Errorf("Expected rank %v, got %v", Ace, card.Rank)
		}
	})

	t.Run("String", func(t *testing.T) {
		tests := []struct {
			card     Card
			expected string
		}{
			{NewCard(Hearts, Ace), "Ah"},
			{NewCard(Spades, King), "Ks"},
			{NewCard(Diamonds, Queen), "Qd"},
			{NewCard(Clubs, Jack), "Jc"},
			{NewCard(Hearts, Ten), "10h"},
			{NewCard(Spades, Two), "2s"},
		}

		for _, test := range tests {
			if result := test.card.String(); result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		}
	})

	t.Run("Value", func(t *testing.T) {
		aceCard := NewCard(Hearts, Ace)
		kingCard := NewCard(Spades, King)
		twoCard := NewCard(Clubs, Two)

		if aceCard.Value() != 14 {
			t.Errorf("Expected Ace value 14, got %d", aceCard.Value())
		}
		if kingCard.Value() != 13 {
			t.Errorf("Expected King value 13, got %d", kingCard.Value())
		}
		if twoCard.Value() != 2 {
			t.Errorf("Expected Two value 2, got %d", twoCard.Value())
		}
	})
}

func TestDeck(t *testing.T) {
	t.Run("NewDeck", func(t *testing.T) {
		deck := NewDeck()
		if deck.Remaining() != 52 {
			t.Errorf("Expected 52 cards, got %d", deck.Remaining())
		}
	})

	t.Run("Deal", func(t *testing.T) {
		deck := NewDeck()
		card, err := deck.Deal()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if deck.Remaining() != 51 {
			t.Errorf("Expected 51 cards remaining, got %d", deck.Remaining())
		}

		// Verify card is valid
		if card.Suit == "" || card.Rank < Two || card.Rank > Ace {
			t.Errorf("Invalid card dealt: %v", card)
		}
	})

	t.Run("DealHand", func(t *testing.T) {
		deck := NewDeck()
		hand, err := deck.DealHand(5)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(hand) != 5 {
			t.Errorf("Expected 5 cards, got %d", len(hand))
		}
		if deck.Remaining() != 47 {
			t.Errorf("Expected 47 cards remaining, got %d", deck.Remaining())
		}
	})

	t.Run("DealFromEmptyDeck", func(t *testing.T) {
		deck := NewDeck()
		// Deal all cards
		for deck.Remaining() > 0 {
			deck.Deal()
		}

		_, err := deck.Deal()
		if err == nil {
			t.Error("Expected error when dealing from empty deck")
		}
	})

	t.Run("Reset", func(t *testing.T) {
		deck := NewDeck()
		deck.Deal() // Deal one card
		deck.Reset()
		if deck.Remaining() != 52 {
			t.Errorf("Expected 52 cards after reset, got %d", deck.Remaining())
		}
	})

	t.Run("Shuffle", func(t *testing.T) {
		deck1 := NewDeck()
		deck2 := NewDeck()

		// Get initial order
		cards1 := make([]Card, 0, 52)
		for deck1.Remaining() > 0 {
			card, _ := deck1.Deal()
			cards1 = append(cards1, card)
		}

		// Shuffle and get new order
		deck2.Shuffle()
		cards2 := make([]Card, 0, 52)
		for deck2.Remaining() > 0 {
			card, _ := deck2.Deal()
			cards2 = append(cards2, card)
		}

		// Check if orders are different (highly likely with proper shuffle)
		different := false
		for i := 0; i < 52; i++ {
			if cards1[i] != cards2[i] {
				different = true
				break
			}
		}

		if !different {
			t.Error("Shuffle did not change card order (very unlikely)")
		}
	})
}

func TestHand(t *testing.T) {
	t.Run("NewHand", func(t *testing.T) {
		hand := NewHand()
		if hand.Size() != 0 {
			t.Errorf("Expected empty hand, got size %d", hand.Size())
		}
	})

	t.Run("AddCard", func(t *testing.T) {
		hand := NewHand()
		card := NewCard(Hearts, Ace)
		hand.AddCard(card)

		if hand.Size() != 1 {
			t.Errorf("Expected hand size 1, got %d", hand.Size())
		}
		if !hand.Contains(card) {
			t.Error("Hand should contain the added card")
		}
	})

	t.Run("AddCards", func(t *testing.T) {
		hand := NewHand()
		cards := []Card{
			NewCard(Hearts, Ace),
			NewCard(Spades, King),
		}
		hand.AddCards(cards)

		if hand.Size() != 2 {
			t.Errorf("Expected hand size 2, got %d", hand.Size())
		}
	})

	t.Run("Clear", func(t *testing.T) {
		hand := NewHand()
		hand.AddCard(NewCard(Hearts, Ace))
		hand.Clear()

		if hand.Size() != 0 {
			t.Errorf("Expected empty hand after clear, got size %d", hand.Size())
		}
	})

	t.Run("Sort", func(t *testing.T) {
		hand := NewHand()
		hand.AddCard(NewCard(Hearts, Two))
		hand.AddCard(NewCard(Spades, Ace))
		hand.AddCard(NewCard(Clubs, King))

		hand.Sort()

		// Should be sorted Ace, King, Two
		if hand.Cards[0].Rank != Ace {
			t.Errorf("Expected Ace first after sort, got %v", hand.Cards[0].Rank)
		}
		if hand.Cards[1].Rank != King {
			t.Errorf("Expected King second after sort, got %v", hand.Cards[1].Rank)
		}
		if hand.Cards[2].Rank != Two {
			t.Errorf("Expected Two third after sort, got %v", hand.Cards[2].Rank)
		}
	})

	t.Run("Remove", func(t *testing.T) {
		hand := NewHand()
		card := NewCard(Hearts, Ace)
		hand.AddCard(card)
		hand.AddCard(NewCard(Spades, King))

		removed := hand.Remove(card)
		if !removed {
			t.Error("Should have removed the card")
		}
		if hand.Size() != 1 {
			t.Errorf("Expected hand size 1 after removal, got %d", hand.Size())
		}
		if hand.Contains(card) {
			t.Error("Hand should not contain removed card")
		}
	})

	t.Run("Copy", func(t *testing.T) {
		hand := NewHand()
		card := NewCard(Hearts, Ace)
		hand.AddCard(card)

		copy := hand.Copy()
		if copy.Size() != hand.Size() {
			t.Error("Copy should have same size as original")
		}
		if !copy.Contains(card) {
			t.Error("Copy should contain same cards as original")
		}

		// Modify original, copy should be unaffected
		hand.AddCard(NewCard(Spades, King))
		if copy.Size() == hand.Size() {
			t.Error("Copy should be independent of original")
		}
	})

	t.Run("String", func(t *testing.T) {
		hand := NewHand()
		if hand.String() != "Empty hand" {
			t.Errorf("Expected 'Empty hand', got '%s'", hand.String())
		}

		hand.AddCard(NewCard(Hearts, Ace))
		hand.AddCard(NewCard(Spades, King))
		result := hand.String()
		if result != "Ah Ks" {
			t.Errorf("Expected 'Ah Ks', got '%s'", result)
		}
	})
}
