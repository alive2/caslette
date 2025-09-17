package poker

import (
	"sort"
)

// EvaluateBestHand finds the best 5-card hand from available cards
func EvaluateBestHand(cards []Card) *Hand {
	if len(cards) < 5 {
		return nil
	}

	// Sort cards by value (descending)
	sortedCards := make([]Card, len(cards))
	copy(sortedCards, cards)
	sort.Slice(sortedCards, func(i, j int) bool {
		return sortedCards[i].Value > sortedCards[j].Value
	})

	// Check for each hand type from highest to lowest
	if hand := checkRoyalFlush(sortedCards); hand != nil {
		return hand
	}
	if hand := checkStraightFlush(sortedCards); hand != nil {
		return hand
	}
	if hand := checkFourOfAKind(sortedCards); hand != nil {
		return hand
	}
	if hand := checkFullHouse(sortedCards); hand != nil {
		return hand
	}
	if hand := checkFlush(sortedCards); hand != nil {
		return hand
	}
	if hand := checkStraight(sortedCards); hand != nil {
		return hand
	}
	if hand := checkThreeOfAKind(sortedCards); hand != nil {
		return hand
	}
	if hand := checkTwoPair(sortedCards); hand != nil {
		return hand
	}
	if hand := checkOnePair(sortedCards); hand != nil {
		return hand
	}

	return checkHighCard(sortedCards)
}

// checkRoyalFlush checks for royal flush (A, K, Q, J, 10 of same suit)
func checkRoyalFlush(cards []Card) *Hand {
	flushSuit := getFlushSuit(cards)
	if flushSuit == "" {
		return nil
	}

	// Get cards of flush suit
	flushCards := getCardsOfSuit(cards, flushSuit)
	if len(flushCards) < 5 {
		return nil
	}

	// Check for A, K, Q, J, 10
	royalValues := []int{14, 13, 12, 11, 10}
	hasAll := true
	for _, value := range royalValues {
		found := false
		for _, card := range flushCards {
			if card.Value == value {
				found = true
				break
			}
		}
		if !found {
			hasAll = false
			break
		}
	}

	if hasAll {
		royalCards := make([]Card, 5)
		cardIndex := 0
		for _, value := range royalValues {
			for _, card := range flushCards {
				if card.Value == value {
					royalCards[cardIndex] = card
					cardIndex++
					break
				}
			}
		}

		return &Hand{
			Cards:   royalCards,
			Rank:    RoyalFlush,
			Name:    "Royal Flush",
			Kickers: []int{}, // No kickers needed for royal flush
		}
	}

	return nil
}

// checkStraightFlush checks for straight flush
func checkStraightFlush(cards []Card) *Hand {
	flushSuit := getFlushSuit(cards)
	if flushSuit == "" {
		return nil
	}

	flushCards := getCardsOfSuit(cards, flushSuit)
	straightCards := findStraight(flushCards)

	if len(straightCards) == 5 {
		return &Hand{
			Cards:   straightCards,
			Rank:    StraightFlush,
			Name:    "Straight Flush",
			Kickers: []int{straightCards[0].Value}, // High card of straight
		}
	}

	return nil
}

// checkFourOfAKind checks for four of a kind
func checkFourOfAKind(cards []Card) *Hand {
	valueCounts := getValueCounts(cards)

	var fourValue int
	var kicker int

	for value, count := range valueCounts {
		if count >= 4 {
			fourValue = value
		} else if count >= 1 && value != fourValue {
			if kicker == 0 || value > kicker {
				kicker = value
			}
		}
	}

	if fourValue > 0 {
		handCards := make([]Card, 0, 5)

		// Add four of a kind
		for _, card := range cards {
			if card.Value == fourValue && len(handCards) < 4 {
				handCards = append(handCards, card)
			}
		}

		// Add highest kicker
		for _, card := range cards {
			if card.Value == kicker && len(handCards) < 5 {
				handCards = append(handCards, card)
				break
			}
		}

		return &Hand{
			Cards:   handCards,
			Rank:    FourOfAKind,
			Name:    "Four of a Kind",
			Kickers: []int{fourValue, kicker},
		}
	}

	return nil
}

// checkFullHouse checks for full house
func checkFullHouse(cards []Card) *Hand {
	valueCounts := getValueCounts(cards)

	var threeValue int
	var pairValue int

	// Find three of a kind and pair
	for value, count := range valueCounts {
		if count >= 3 {
			if threeValue == 0 || value > threeValue {
				if threeValue > 0 && threeValue < value {
					pairValue = threeValue // Previous three becomes pair
				}
				threeValue = value
			} else if pairValue == 0 || value > pairValue {
				pairValue = value
			}
		} else if count >= 2 {
			if pairValue == 0 || value > pairValue {
				pairValue = value
			}
		}
	}

	if threeValue > 0 && pairValue > 0 {
		handCards := make([]Card, 0, 5)

		// Add three of a kind
		for _, card := range cards {
			if card.Value == threeValue && len(handCards) < 3 {
				handCards = append(handCards, card)
			}
		}

		// Add pair
		for _, card := range cards {
			if card.Value == pairValue && len(handCards) < 5 {
				handCards = append(handCards, card)
			}
		}

		return &Hand{
			Cards:   handCards,
			Rank:    FullHouse,
			Name:    "Full House",
			Kickers: []int{threeValue, pairValue},
		}
	}

	return nil
}

// checkFlush checks for flush
func checkFlush(cards []Card) *Hand {
	flushSuit := getFlushSuit(cards)
	if flushSuit == "" {
		return nil
	}

	flushCards := getCardsOfSuit(cards, flushSuit)

	// Sort flush cards by value (descending)
	sort.Slice(flushCards, func(i, j int) bool {
		return flushCards[i].Value > flushCards[j].Value
	})

	// Take best 5 cards
	handCards := flushCards[:5]
	kickers := make([]int, 5)
	for i, card := range handCards {
		kickers[i] = card.Value
	}

	return &Hand{
		Cards:   handCards,
		Rank:    Flush,
		Name:    "Flush",
		Kickers: kickers,
	}
}

// checkStraight checks for straight
func checkStraight(cards []Card) *Hand {
	straightCards := findStraight(cards)

	if len(straightCards) == 5 {
		return &Hand{
			Cards:   straightCards,
			Rank:    Straight,
			Name:    "Straight",
			Kickers: []int{straightCards[0].Value}, // High card of straight
		}
	}

	return nil
}

// checkThreeOfAKind checks for three of a kind
func checkThreeOfAKind(cards []Card) *Hand {
	valueCounts := getValueCounts(cards)

	var threeValue int
	kickers := make([]int, 0, 2)

	for value, count := range valueCounts {
		if count >= 3 {
			threeValue = value
		}
	}

	if threeValue > 0 {
		// Find two highest kickers
		for _, card := range cards {
			if card.Value != threeValue && len(kickers) < 2 {
				found := false
				for _, k := range kickers {
					if k == card.Value {
						found = true
						break
					}
				}
				if !found {
					kickers = append(kickers, card.Value)
				}
			}
		}

		// Sort kickers descending
		sort.Slice(kickers, func(i, j int) bool {
			return kickers[i] > kickers[j]
		})

		handCards := make([]Card, 0, 5)

		// Add three of a kind
		for _, card := range cards {
			if card.Value == threeValue && len(handCards) < 3 {
				handCards = append(handCards, card)
			}
		}

		// Add kickers
		for _, kickerValue := range kickers {
			for _, card := range cards {
				if card.Value == kickerValue && len(handCards) < 5 {
					handCards = append(handCards, card)
					break
				}
			}
		}

		allKickers := append([]int{threeValue}, kickers...)
		return &Hand{
			Cards:   handCards,
			Rank:    ThreeOfAKind,
			Name:    "Three of a Kind",
			Kickers: allKickers,
		}
	}

	return nil
}

// checkTwoPair checks for two pair
func checkTwoPair(cards []Card) *Hand {
	valueCounts := getValueCounts(cards)

	pairs := make([]int, 0, 2)
	var kicker int

	// Find pairs
	for value, count := range valueCounts {
		if count >= 2 {
			pairs = append(pairs, value)
		}
	}

	if len(pairs) >= 2 {
		// Sort pairs descending
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i] > pairs[j]
		})

		// Take two highest pairs
		pairs = pairs[:2]

		// Find kicker
		for value, count := range valueCounts {
			if count >= 1 && value != pairs[0] && value != pairs[1] {
				if kicker == 0 || value > kicker {
					kicker = value
				}
			}
		}

		handCards := make([]Card, 0, 5)

		// Add pairs
		for _, pairValue := range pairs {
			pairCount := 0
			for _, card := range cards {
				if card.Value == pairValue && pairCount < 2 {
					handCards = append(handCards, card)
					pairCount++
				}
			}
		}

		// Add kicker
		for _, card := range cards {
			if card.Value == kicker && len(handCards) < 5 {
				handCards = append(handCards, card)
				break
			}
		}

		kickers := []int{pairs[0], pairs[1], kicker}
		return &Hand{
			Cards:   handCards,
			Rank:    TwoPair,
			Name:    "Two Pair",
			Kickers: kickers,
		}
	}

	return nil
}

// checkOnePair checks for one pair
func checkOnePair(cards []Card) *Hand {
	valueCounts := getValueCounts(cards)

	var pairValue int
	kickers := make([]int, 0, 3)

	for value, count := range valueCounts {
		if count >= 2 {
			pairValue = value
		}
	}

	if pairValue > 0 {
		// Find three highest kickers
		for _, card := range cards {
			if card.Value != pairValue && len(kickers) < 3 {
				found := false
				for _, k := range kickers {
					if k == card.Value {
						found = true
						break
					}
				}
				if !found {
					kickers = append(kickers, card.Value)
				}
			}
		}

		// Sort kickers descending
		sort.Slice(kickers, func(i, j int) bool {
			return kickers[i] > kickers[j]
		})

		handCards := make([]Card, 0, 5)

		// Add pair
		pairCount := 0
		for _, card := range cards {
			if card.Value == pairValue && pairCount < 2 {
				handCards = append(handCards, card)
				pairCount++
			}
		}

		// Add kickers
		for _, kickerValue := range kickers {
			for _, card := range cards {
				if card.Value == kickerValue && len(handCards) < 5 {
					handCards = append(handCards, card)
					break
				}
			}
		}

		allKickers := append([]int{pairValue}, kickers...)
		return &Hand{
			Cards:   handCards,
			Rank:    OnePair,
			Name:    "One Pair",
			Kickers: allKickers,
		}
	}

	return nil
}

// checkHighCard returns the highest card hand
func checkHighCard(cards []Card) *Hand {
	// Sort cards by value (descending)
	sortedCards := make([]Card, len(cards))
	copy(sortedCards, cards)
	sort.Slice(sortedCards, func(i, j int) bool {
		return sortedCards[i].Value > sortedCards[j].Value
	})

	// Take best 5 cards
	handCards := sortedCards[:5]
	kickers := make([]int, 5)
	for i, card := range handCards {
		kickers[i] = card.Value
	}

	return &Hand{
		Cards:   handCards,
		Rank:    HighCard,
		Name:    "High Card",
		Kickers: kickers,
	}
}

// Helper functions

// getFlushSuit returns the suit if there's a flush, empty string otherwise
func getFlushSuit(cards []Card) string {
	suitCounts := make(map[string]int)

	for _, card := range cards {
		suitCounts[card.Suit]++
		if suitCounts[card.Suit] >= 5 {
			return card.Suit
		}
	}

	return ""
}

// getCardsOfSuit returns all cards of specified suit
func getCardsOfSuit(cards []Card, suit string) []Card {
	var result []Card
	for _, card := range cards {
		if card.Suit == suit {
			result = append(result, card)
		}
	}

	// Sort by value descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})

	return result
}

// findStraight finds a straight in the cards
func findStraight(cards []Card) []Card {
	// Remove duplicates and sort
	uniqueValues := make(map[int]Card)
	for _, card := range cards {
		if existing, exists := uniqueValues[card.Value]; !exists || card.Value > existing.Value {
			uniqueValues[card.Value] = card
		}
	}

	var values []int
	for value := range uniqueValues {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i] > values[j]
	})

	// Check for regular straight
	for i := 0; i <= len(values)-5; i++ {
		if values[i]-values[i+4] == 4 { // 5 consecutive values
			straightCards := make([]Card, 5)
			for j := 0; j < 5; j++ {
				straightCards[j] = uniqueValues[values[i+j]]
			}
			return straightCards
		}
	}

	// Check for A-2-3-4-5 straight (wheel)
	if len(values) >= 5 {
		hasAce := false
		hasTwo := false
		hasThree := false
		hasFour := false
		hasFive := false

		for _, value := range values {
			switch value {
			case 14:
				hasAce = true
			case 2:
				hasTwo = true
			case 3:
				hasThree = true
			case 4:
				hasFour = true
			case 5:
				hasFive = true
			}
		}

		if hasAce && hasTwo && hasThree && hasFour && hasFive {
			return []Card{
				uniqueValues[5], // 5 high in wheel straight
				uniqueValues[4],
				uniqueValues[3],
				uniqueValues[2],
				uniqueValues[14], // Ace low
			}
		}
	}

	return nil
}

// getValueCounts returns a map of card values to their counts
func getValueCounts(cards []Card) map[int]int {
	counts := make(map[int]int)
	for _, card := range cards {
		counts[card.Value]++
	}
	return counts
}

// compareHands compares two hands and returns:
// 1 if hand1 > hand2
// 0 if hand1 == hand2
// -1 if hand1 < hand2
func compareHands(hand1, hand2 *Hand) int {
	if hand1.Rank > hand2.Rank {
		return 1
	} else if hand1.Rank < hand2.Rank {
		return -1
	}

	// Same rank, compare kickers
	for i := 0; i < len(hand1.Kickers) && i < len(hand2.Kickers); i++ {
		if hand1.Kickers[i] > hand2.Kickers[i] {
			return 1
		} else if hand1.Kickers[i] < hand2.Kickers[i] {
			return -1
		}
	}

	return 0 // Exactly equal
}
