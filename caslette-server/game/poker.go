package game

import (
	"sort"
)

// HandRank represents the rank of a poker hand
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

// String returns the string representation of a hand rank
func (hr HandRank) String() string {
	switch hr {
	case HighCard:
		return "High Card"
	case OnePair:
		return "One Pair"
	case TwoPair:
		return "Two Pair"
	case ThreeOfAKind:
		return "Three of a Kind"
	case Straight:
		return "Straight"
	case Flush:
		return "Flush"
	case FullHouse:
		return "Full House"
	case FourOfAKind:
		return "Four of a Kind"
	case StraightFlush:
		return "Straight Flush"
	case RoyalFlush:
		return "Royal Flush"
	default:
		return "Unknown"
	}
}

// PokerHand represents a evaluated poker hand
type PokerHand struct {
	Rank      HandRank `json:"rank"`
	Cards     []Card   `json:"cards"`
	Kickers   []Rank   `json:"kickers"`   // For tie-breaking
	HighCards []Rank   `json:"highCards"` // Primary cards for the hand type
}

// String returns a string representation of the poker hand
func (ph *PokerHand) String() string {
	return ph.Rank.String()
}

// Compare compares two poker hands. Returns:
// 1 if this hand wins
// -1 if the other hand wins
// 0 if it's a tie
func (ph *PokerHand) Compare(other *PokerHand) int {
	// First compare hand ranks
	if ph.Rank > other.Rank {
		return 1
	}
	if ph.Rank < other.Rank {
		return -1
	}

	// Same rank, compare high cards
	for i := 0; i < len(ph.HighCards) && i < len(other.HighCards); i++ {
		if ph.HighCards[i] > other.HighCards[i] {
			return 1
		}
		if ph.HighCards[i] < other.HighCards[i] {
			return -1
		}
	}

	// Compare kickers
	for i := 0; i < len(ph.Kickers) && i < len(other.Kickers); i++ {
		if ph.Kickers[i] > other.Kickers[i] {
			return 1
		}
		if ph.Kickers[i] < other.Kickers[i] {
			return -1
		}
	}

	return 0 // Tie
}

// PokerEvaluator evaluates poker hands
type PokerEvaluator struct{}

// NewPokerEvaluator creates a new poker evaluator
func NewPokerEvaluator() *PokerEvaluator {
	return &PokerEvaluator{}
}

// EvaluateHand evaluates a 5-card poker hand
func (pe *PokerEvaluator) EvaluateHand(cards []Card) *PokerHand {
	if len(cards) != 5 {
		// For hands with more than 5 cards, find the best 5-card combination
		return pe.FindBestHand(cards)
	}

	// Sort cards by rank (descending)
	sortedCards := make([]Card, len(cards))
	copy(sortedCards, cards)
	sort.Slice(sortedCards, func(i, j int) bool {
		return sortedCards[i].Rank > sortedCards[j].Rank
	})

	// Check for each hand type from highest to lowest
	if hand := pe.checkRoyalFlush(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkStraightFlush(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkFourOfAKind(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkFullHouse(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkFlush(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkStraight(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkThreeOfAKind(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkTwoPair(sortedCards); hand != nil {
		return hand
	}
	if hand := pe.checkOnePair(sortedCards); hand != nil {
		return hand
	}

	return pe.checkHighCard(sortedCards)
}

// FindBestHand finds the best 5-card hand from 7 cards
func (pe *PokerEvaluator) FindBestHand(cards []Card) *PokerHand {
	if len(cards) < 5 {
		return &PokerHand{Rank: HighCard, Cards: cards}
	}
	if len(cards) == 5 {
		return pe.EvaluateHand(cards)
	}

	// Generate all possible 5-card combinations
	var bestHand *PokerHand
	pe.generateCombinations(cards, 5, 0, []Card{}, func(combination []Card) {
		hand := pe.EvaluateHand(combination)
		if bestHand == nil || hand.Compare(bestHand) > 0 {
			bestHand = hand
		}
	})

	return bestHand
}

// generateCombinations generates all combinations of k cards from the given cards
func (pe *PokerEvaluator) generateCombinations(cards []Card, k, start int, current []Card, callback func([]Card)) {
	if len(current) == k {
		combination := make([]Card, len(current))
		copy(combination, current)
		callback(combination)
		return
	}

	for i := start; i < len(cards); i++ {
		current = append(current, cards[i])
		pe.generateCombinations(cards, k, i+1, current, callback)
		current = current[:len(current)-1]
	}
}

// Helper functions for checking specific hand types

func (pe *PokerEvaluator) checkRoyalFlush(cards []Card) *PokerHand {
	if !pe.isFlush(cards) {
		return nil
	}

	// Check for A, K, Q, J, 10 of same suit
	expectedRanks := []Rank{Ace, King, Queen, Jack, Ten}
	for i, expectedRank := range expectedRanks {
		if cards[i].Rank != expectedRank {
			return nil
		}
	}

	return &PokerHand{
		Rank:      RoyalFlush,
		Cards:     cards,
		HighCards: []Rank{Ace},
		Kickers:   []Rank{},
	}
}

func (pe *PokerEvaluator) checkStraightFlush(cards []Card) *PokerHand {
	if !pe.isFlush(cards) || !pe.isStraight(cards) {
		return nil
	}

	return &PokerHand{
		Rank:      StraightFlush,
		Cards:     cards,
		HighCards: []Rank{cards[0].Rank},
		Kickers:   []Rank{},
	}
}

func (pe *PokerEvaluator) checkFourOfAKind(cards []Card) *PokerHand {
	rankCounts := pe.getRankCounts(cards)

	var fourRank, kicker Rank
	found := false

	for rank, count := range rankCounts {
		if count == 4 {
			fourRank = rank
			found = true
		} else if count == 1 {
			kicker = rank
		}
	}

	if !found {
		return nil
	}

	return &PokerHand{
		Rank:      FourOfAKind,
		Cards:     cards,
		HighCards: []Rank{fourRank},
		Kickers:   []Rank{kicker},
	}
}

func (pe *PokerEvaluator) checkFullHouse(cards []Card) *PokerHand {
	rankCounts := pe.getRankCounts(cards)

	var threeRank, pairRank Rank
	hasThree, hasPair := false, false

	for rank, count := range rankCounts {
		if count == 3 {
			threeRank = rank
			hasThree = true
		} else if count == 2 {
			pairRank = rank
			hasPair = true
		}
	}

	if !hasThree || !hasPair {
		return nil
	}

	return &PokerHand{
		Rank:      FullHouse,
		Cards:     cards,
		HighCards: []Rank{threeRank, pairRank},
		Kickers:   []Rank{},
	}
}

func (pe *PokerEvaluator) checkFlush(cards []Card) *PokerHand {
	if !pe.isFlush(cards) {
		return nil
	}

	highCards := make([]Rank, len(cards))
	for i, card := range cards {
		highCards[i] = card.Rank
	}

	return &PokerHand{
		Rank:      Flush,
		Cards:     cards,
		HighCards: highCards,
		Kickers:   []Rank{},
	}
}

func (pe *PokerEvaluator) checkStraight(cards []Card) *PokerHand {
	if !pe.isStraight(cards) {
		return nil
	}

	return &PokerHand{
		Rank:      Straight,
		Cards:     cards,
		HighCards: []Rank{cards[0].Rank},
		Kickers:   []Rank{},
	}
}

func (pe *PokerEvaluator) checkThreeOfAKind(cards []Card) *PokerHand {
	rankCounts := pe.getRankCounts(cards)

	var threeRank Rank
	kickers := make([]Rank, 0)
	found := false

	for rank, count := range rankCounts {
		if count == 3 {
			threeRank = rank
			found = true
		} else if count == 1 {
			kickers = append(kickers, rank)
		}
	}

	if !found {
		return nil
	}

	// Sort kickers descending
	sort.Slice(kickers, func(i, j int) bool {
		return kickers[i] > kickers[j]
	})

	return &PokerHand{
		Rank:      ThreeOfAKind,
		Cards:     cards,
		HighCards: []Rank{threeRank},
		Kickers:   kickers,
	}
}

func (pe *PokerEvaluator) checkTwoPair(cards []Card) *PokerHand {
	rankCounts := pe.getRankCounts(cards)

	pairs := make([]Rank, 0)
	var kicker Rank

	for rank, count := range rankCounts {
		if count == 2 {
			pairs = append(pairs, rank)
		} else if count == 1 {
			kicker = rank
		}
	}

	if len(pairs) != 2 {
		return nil
	}

	// Sort pairs descending
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i] > pairs[j]
	})

	return &PokerHand{
		Rank:      TwoPair,
		Cards:     cards,
		HighCards: pairs,
		Kickers:   []Rank{kicker},
	}
}

func (pe *PokerEvaluator) checkOnePair(cards []Card) *PokerHand {
	rankCounts := pe.getRankCounts(cards)

	var pairRank Rank
	kickers := make([]Rank, 0)
	found := false

	for rank, count := range rankCounts {
		if count == 2 {
			pairRank = rank
			found = true
		} else if count == 1 {
			kickers = append(kickers, rank)
		}
	}

	if !found {
		return nil
	}

	// Sort kickers descending
	sort.Slice(kickers, func(i, j int) bool {
		return kickers[i] > kickers[j]
	})

	return &PokerHand{
		Rank:      OnePair,
		Cards:     cards,
		HighCards: []Rank{pairRank},
		Kickers:   kickers,
	}
}

func (pe *PokerEvaluator) checkHighCard(cards []Card) *PokerHand {
	highCards := make([]Rank, len(cards))
	for i, card := range cards {
		highCards[i] = card.Rank
	}

	return &PokerHand{
		Rank:      HighCard,
		Cards:     cards,
		HighCards: highCards,
		Kickers:   []Rank{},
	}
}

// Helper functions

func (pe *PokerEvaluator) isFlush(cards []Card) bool {
	if len(cards) == 0 {
		return false
	}

	suit := cards[0].Suit
	for _, card := range cards[1:] {
		if card.Suit != suit {
			return false
		}
	}
	return true
}

func (pe *PokerEvaluator) isStraight(cards []Card) bool {
	if len(cards) != 5 {
		return false
	}

	// Check for A-2-3-4-5 straight (wheel)
	if cards[0].Rank == Ace && cards[1].Rank == Five &&
		cards[2].Rank == Four && cards[3].Rank == Three && cards[4].Rank == Two {
		return true
	}

	// Check for normal straight
	for i := 1; i < len(cards); i++ {
		if int(cards[i-1].Rank)-int(cards[i].Rank) != 1 {
			return false
		}
	}
	return true
}

func (pe *PokerEvaluator) getRankCounts(cards []Card) map[Rank]int {
	counts := make(map[Rank]int)
	for _, card := range cards {
		counts[card.Rank]++
	}
	return counts
}
