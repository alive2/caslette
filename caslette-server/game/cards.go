package game

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// Suit represents a playing card suit
type Suit string

const (
	Hearts   Suit = "hearts"
	Diamonds Suit = "diamonds"
	Clubs    Suit = "clubs"
	Spades   Suit = "spades"
)

// Rank represents a playing card rank
type Rank int

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

// String returns the string representation of a rank
func (r Rank) String() string {
	switch r {
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ace:
		return "A"
	default:
		return fmt.Sprintf("%d", int(r))
	}
}

// Card represents a playing card
type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

// String returns the string representation of a card
func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank.String(), string(c.Suit)[0:1])
}

// Value returns the numeric value of a card for comparison
func (c Card) Value() int {
	return int(c.Rank)
}

// NewCard creates a new card
func NewCard(suit Suit, rank Rank) Card {
	return Card{Suit: suit, Rank: rank}
}

// Deck represents a deck of playing cards
type Deck struct {
	cards []Card
	rng   *rand.Rand
}

// NewDeck creates a new standard 52-card deck
func NewDeck() *Deck {
	deck := &Deck{
		cards: make([]Card, 0, 52),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Create all 52 cards
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	for _, suit := range suits {
		for rank := Two; rank <= Ace; rank++ {
			deck.cards = append(deck.cards, Card{Suit: suit, Rank: rank})
		}
	}

	return deck
}

// Shuffle shuffles the deck
func (d *Deck) Shuffle() {
	d.rng.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

// Deal deals a card from the top of the deck
func (d *Deck) Deal() (Card, error) {
	if len(d.cards) == 0 {
		return Card{}, fmt.Errorf("cannot deal from empty deck")
	}

	card := d.cards[0]
	d.cards = d.cards[1:]
	return card, nil
}

// DealHand deals multiple cards
func (d *Deck) DealHand(count int) ([]Card, error) {
	if count > len(d.cards) {
		return nil, fmt.Errorf("cannot deal %d cards, only %d cards remaining", count, len(d.cards))
	}

	hand := make([]Card, count)
	for i := 0; i < count; i++ {
		card, err := d.Deal()
		if err != nil {
			return nil, err
		}
		hand[i] = card
	}

	return hand, nil
}

// Remaining returns the number of cards remaining in the deck
func (d *Deck) Remaining() int {
	return len(d.cards)
}

// Reset resets the deck to a full 52-card deck and shuffles it
func (d *Deck) Reset() {
	d.cards = make([]Card, 0, 52)
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	for _, suit := range suits {
		for rank := Two; rank <= Ace; rank++ {
			d.cards = append(d.cards, Card{Suit: suit, Rank: rank})
		}
	}
	d.Shuffle()
}

// Hand represents a collection of cards
type Hand struct {
	Cards []Card `json:"cards"`
}

// NewHand creates a new empty hand
func NewHand() *Hand {
	return &Hand{Cards: make([]Card, 0)}
}

// AddCard adds a card to the hand
func (h *Hand) AddCard(card Card) {
	h.Cards = append(h.Cards, card)
}

// AddCards adds multiple cards to the hand
func (h *Hand) AddCards(cards []Card) {
	h.Cards = append(h.Cards, cards...)
}

// Clear removes all cards from the hand
func (h *Hand) Clear() {
	h.Cards = h.Cards[:0]
}

// Size returns the number of cards in the hand
func (h *Hand) Size() int {
	return len(h.Cards)
}

// String returns a string representation of the hand
func (h *Hand) String() string {
	if len(h.Cards) == 0 {
		return "Empty hand"
	}

	result := ""
	for i, card := range h.Cards {
		if i > 0 {
			result += " "
		}
		result += card.String()
	}
	return result
}

// Sort sorts the cards in the hand by rank (descending)
func (h *Hand) Sort() {
	sort.Slice(h.Cards, func(i, j int) bool {
		return h.Cards[i].Rank > h.Cards[j].Rank
	})
}

// Contains checks if the hand contains a specific card
func (h *Hand) Contains(card Card) bool {
	for _, c := range h.Cards {
		if c.Suit == card.Suit && c.Rank == card.Rank {
			return true
		}
	}
	return false
}

// Remove removes a specific card from the hand
func (h *Hand) Remove(card Card) bool {
	for i, c := range h.Cards {
		if c.Suit == card.Suit && c.Rank == card.Rank {
			h.Cards = append(h.Cards[:i], h.Cards[i+1:]...)
			return true
		}
	}
	return false
}

// Copy creates a copy of the hand
func (h *Hand) Copy() *Hand {
	newHand := NewHand()
	newHand.Cards = make([]Card, len(h.Cards))
	copy(newHand.Cards, h.Cards)
	return newHand
}
