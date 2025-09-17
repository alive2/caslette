package poker

import (
	"caslette-server/models"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Card represents a playing card
type Card struct {
	Suit  string `json:"suit"`  // hearts, diamonds, clubs, spades
	Rank  string `json:"rank"`  // 2-9, T, J, Q, K, A
	Value int    `json:"value"` // 2-14 (Ace high)
}

// Deck represents a deck of cards
type Deck struct {
	Cards []Card `json:"cards"`
}

// Hand represents a 5-card poker hand
type Hand struct {
	Cards   []Card `json:"cards"`
	Rank    int    `json:"rank"`    // 1-10 (1=high card, 10=royal flush)
	Name    string `json:"name"`    // Human readable name
	Kickers []int  `json:"kickers"` // Tiebreaker values
}

// GameEngine manages poker game logic
type GameEngine struct {
	Table                 *models.PokerTable
	Players               []*GamePlayer
	Deck                  *Deck
	CommunityCards        []Card
	CurrentBet            int64
	Pot                   int64
	DealerPosition        int
	SmallBlindPosition    int
	BigBlindPosition      int
	CurrentPlayerPosition int
	BettingRound          string // preflop, flop, turn, river
	HandNumber            int
}

// GamePlayer represents a player in a game
type GamePlayer struct {
	*models.TablePlayer
	HoleCards  []Card
	BestHand   *Hand
	CurrentBet int64
	TotalBet   int64
	HasActed   bool
	IsInHand   bool
	IsAllIn    bool
}

// Card values for comparison
var cardValues = map[string]int{
	"2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9,
	"T": 10, "J": 11, "Q": 12, "K": 13, "A": 14,
}

// Hand rankings
const (
	HighCard      = 1
	OnePair       = 2
	TwoPair       = 3
	ThreeOfAKind  = 4
	Straight      = 5
	Flush         = 6
	FullHouse     = 7
	FourOfAKind   = 8
	StraightFlush = 9
	RoyalFlush    = 10
)

// NewGameEngine creates a new poker game engine
func NewGameEngine(table *models.PokerTable, players []*models.TablePlayer) *GameEngine {
	gamePlayers := make([]*GamePlayer, len(players))
	for i, player := range players {
		gamePlayers[i] = &GamePlayer{
			TablePlayer: player,
			IsInHand:    true,
		}
	}

	return &GameEngine{
		Table:      table,
		Players:    gamePlayers,
		Deck:       NewDeck(),
		HandNumber: 1,
	}
}

// NewDeck creates a standard 52-card deck
func NewDeck() *Deck {
	suits := []string{"hearts", "diamonds", "clubs", "spades"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}

	var cards []Card
	for _, suit := range suits {
		for _, rank := range ranks {
			cards = append(cards, Card{
				Suit:  suit,
				Rank:  rank,
				Value: cardValues[rank],
			})
		}
	}

	deck := &Deck{Cards: cards}
	deck.Shuffle()
	return deck
}

// Shuffle shuffles the deck
func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// DrawCard draws a card from the top of the deck
func (d *Deck) DrawCard() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, errors.New("deck is empty")
	}

	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card, nil
}

// StartNewHand starts a new poker hand
func (g *GameEngine) StartNewHand() error {
	// Reset for new hand
	g.Deck = NewDeck()
	g.CommunityCards = []Card{}
	g.CurrentBet = 0
	g.Pot = 0
	g.BettingRound = "preflop"

	// Reset all players
	for _, player := range g.Players {
		player.HoleCards = []Card{}
		player.BestHand = nil
		player.CurrentBet = 0
		player.TotalBet = 0
		player.HasActed = false
		player.IsInHand = true
		player.IsAllIn = false
	}

	// Set positions
	g.setPositions()

	// Post blinds
	if err := g.postBlinds(); err != nil {
		return err
	}

	// Deal hole cards
	if err := g.dealHoleCards(); err != nil {
		return err
	}

	return nil
}

// setPositions sets dealer, small blind, and big blind positions
func (g *GameEngine) setPositions() {
	activePlayerCount := 0
	for _, player := range g.Players {
		if player.Status == "playing" {
			activePlayerCount++
		}
	}

	if activePlayerCount < 2 {
		return
	}

	// Move dealer position to next active player
	g.DealerPosition = g.getNextActivePlayer(g.DealerPosition)

	if activePlayerCount == 2 {
		// Heads up: dealer is small blind
		g.SmallBlindPosition = g.DealerPosition
		g.BigBlindPosition = g.getNextActivePlayer(g.SmallBlindPosition)
	} else {
		// Regular game: small blind is next after dealer
		g.SmallBlindPosition = g.getNextActivePlayer(g.DealerPosition)
		g.BigBlindPosition = g.getNextActivePlayer(g.SmallBlindPosition)
	}

	// First to act preflop is after big blind
	g.CurrentPlayerPosition = g.getNextActivePlayer(g.BigBlindPosition)
}

// getNextActivePlayer finds the next active player from given position
func (g *GameEngine) getNextActivePlayer(from int) int {
	for i := 1; i < len(g.Players); i++ {
		pos := (from + i) % len(g.Players)
		if g.Players[pos].Status == "playing" {
			return pos
		}
	}
	return from
}

// postBlinds posts small and big blinds
func (g *GameEngine) postBlinds() error {
	// Post small blind
	sbPlayer := g.Players[g.SmallBlindPosition]
	sbAmount := g.Table.SmallBlind
	if sbPlayer.ChipCount < sbAmount {
		sbAmount = sbPlayer.ChipCount
		sbPlayer.IsAllIn = true
	}
	sbPlayer.ChipCount -= sbAmount
	sbPlayer.CurrentBet = sbAmount
	sbPlayer.TotalBet = sbAmount
	g.Pot += sbAmount
	g.CurrentBet = sbAmount

	// Post big blind
	bbPlayer := g.Players[g.BigBlindPosition]
	bbAmount := g.Table.BigBlind
	if bbPlayer.ChipCount < bbAmount {
		bbAmount = bbPlayer.ChipCount
		bbPlayer.IsAllIn = true
	}
	bbPlayer.ChipCount -= bbAmount
	bbPlayer.CurrentBet = bbAmount
	bbPlayer.TotalBet = bbAmount
	g.Pot += bbAmount
	g.CurrentBet = bbAmount

	return nil
}

// dealHoleCards deals 2 cards to each active player
func (g *GameEngine) dealHoleCards() error {
	for i := 0; i < 2; i++ {
		for _, player := range g.Players {
			if player.Status == "playing" {
				card, err := g.Deck.DrawCard()
				if err != nil {
					return err
				}
				player.HoleCards = append(player.HoleCards, card)
			}
		}
	}
	return nil
}

// DealFlop deals the flop (3 community cards)
func (g *GameEngine) DealFlop() error {
	if g.BettingRound != "preflop" {
		return errors.New("cannot deal flop: not in preflop")
	}

	// Burn one card
	_, err := g.Deck.DrawCard()
	if err != nil {
		return err
	}

	// Deal 3 cards
	for i := 0; i < 3; i++ {
		card, err := g.Deck.DrawCard()
		if err != nil {
			return err
		}
		g.CommunityCards = append(g.CommunityCards, card)
	}

	g.BettingRound = "flop"
	g.resetBettingRound()
	return nil
}

// DealTurn deals the turn (4th community card)
func (g *GameEngine) DealTurn() error {
	if g.BettingRound != "flop" {
		return errors.New("cannot deal turn: not after flop")
	}

	// Burn one card
	_, err := g.Deck.DrawCard()
	if err != nil {
		return err
	}

	// Deal 1 card
	card, err := g.Deck.DrawCard()
	if err != nil {
		return err
	}
	g.CommunityCards = append(g.CommunityCards, card)

	g.BettingRound = "turn"
	g.resetBettingRound()
	return nil
}

// DealRiver deals the river (5th community card)
func (g *GameEngine) DealRiver() error {
	if g.BettingRound != "turn" {
		return errors.New("cannot deal river: not after turn")
	}

	// Burn one card
	_, err := g.Deck.DrawCard()
	if err != nil {
		return err
	}

	// Deal 1 card
	card, err := g.Deck.DrawCard()
	if err != nil {
		return err
	}
	g.CommunityCards = append(g.CommunityCards, card)

	g.BettingRound = "river"
	g.resetBettingRound()
	return nil
}

// resetBettingRound resets betting for new round
func (g *GameEngine) resetBettingRound() {
	g.CurrentBet = 0
	for _, player := range g.Players {
		player.CurrentBet = 0
		player.HasActed = false
	}
	// First to act post-flop is first active player after dealer
	g.CurrentPlayerPosition = g.getNextActivePlayer(g.DealerPosition)
}

// ProcessPlayerAction processes a player's action
func (g *GameEngine) ProcessPlayerAction(playerPos int, action string, amount int64) error {
	player := g.Players[playerPos]

	if !player.IsInHand {
		return errors.New("player is not in hand")
	}

	if player.IsAllIn {
		return errors.New("player is already all-in")
	}

	switch action {
	case "fold":
		player.IsInHand = false

	case "check":
		if g.CurrentBet > player.CurrentBet {
			return errors.New("cannot check: there is a bet to call")
		}
		player.HasActed = true

	case "call":
		callAmount := g.CurrentBet - player.CurrentBet
		if callAmount == 0 {
			return errors.New("nothing to call")
		}

		if player.ChipCount <= callAmount {
			// All-in call
			callAmount = player.ChipCount
			player.IsAllIn = true
		}

		player.ChipCount -= callAmount
		player.CurrentBet += callAmount
		player.TotalBet += callAmount
		g.Pot += callAmount
		player.HasActed = true

	case "bet", "raise":
		betAmount := amount
		totalBet := player.CurrentBet + betAmount

		if totalBet <= g.CurrentBet {
			return errors.New("bet/raise amount must be higher than current bet")
		}

		if player.ChipCount <= betAmount {
			// All-in bet/raise
			betAmount = player.ChipCount
			player.IsAllIn = true
		}

		player.ChipCount -= betAmount
		player.CurrentBet += betAmount
		player.TotalBet += betAmount
		g.Pot += betAmount
		g.CurrentBet = player.CurrentBet
		player.HasActed = true

		// Reset other players' acted status (except all-in players)
		for _, p := range g.Players {
			if p != player && !p.IsAllIn && p.IsInHand {
				p.HasActed = false
			}
		}

	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	return nil
}

// IsBettingRoundComplete checks if current betting round is complete
func (g *GameEngine) IsBettingRoundComplete() bool {
	activePlayersInHand := 0
	playersWhoActed := 0
	playersWhoNeedToAct := 0

	for _, player := range g.Players {
		if player.IsInHand {
			activePlayersInHand++
			if player.HasActed || player.IsAllIn {
				playersWhoActed++
			} else if player.CurrentBet < g.CurrentBet {
				playersWhoNeedToAct++
			}
		}
	}

	// If only one player left, hand is over
	if activePlayersInHand <= 1 {
		return true
	}

	// If all players have acted and no one needs to call
	return playersWhoNeedToAct == 0 && playersWhoActed == activePlayersInHand
}

// GetCurrentPlayer returns the current player to act
func (g *GameEngine) GetCurrentPlayer() *GamePlayer {
	if g.CurrentPlayerPosition >= 0 && g.CurrentPlayerPosition < len(g.Players) {
		return g.Players[g.CurrentPlayerPosition]
	}
	return nil
}

// MoveToNextPlayer moves to the next player to act
func (g *GameEngine) MoveToNextPlayer() {
	for i := 1; i < len(g.Players); i++ {
		pos := (g.CurrentPlayerPosition + i) % len(g.Players)
		player := g.Players[pos]

		if player.IsInHand && !player.IsAllIn && !player.HasActed {
			g.CurrentPlayerPosition = pos
			return
		}
	}
}

// EvaluateHands evaluates all players' best hands
func (g *GameEngine) EvaluateHands() {
	for _, player := range g.Players {
		if player.IsInHand {
			allCards := append(player.HoleCards, g.CommunityCards...)
			player.BestHand = EvaluateBestHand(allCards)
		}
	}
}

// DetermineWinners determines the winner(s) of the hand
func (g *GameEngine) DetermineWinners() []*GamePlayer {
	var winners []*GamePlayer
	var bestRank int
	var bestKickers []int

	for _, player := range g.Players {
		if !player.IsInHand || player.BestHand == nil {
			continue
		}

		if len(winners) == 0 {
			winners = []*GamePlayer{player}
			bestRank = player.BestHand.Rank
			bestKickers = player.BestHand.Kickers
		} else {
			comparison := compareHands(player.BestHand, &Hand{Rank: bestRank, Kickers: bestKickers})
			if comparison > 0 {
				// New best hand
				winners = []*GamePlayer{player}
				bestRank = player.BestHand.Rank
				bestKickers = player.BestHand.Kickers
			} else if comparison == 0 {
				// Tie
				winners = append(winners, player)
			}
		}
	}

	return winners
}

// DistributePot distributes the pot to winners and calculates rake
func (g *GameEngine) DistributePot() (map[uint]int64, int64) {
	winners := g.DetermineWinners()
	if len(winners) == 0 {
		return make(map[uint]int64), 0
	}

	// Calculate rake
	rakeAmount := int64(float64(g.Pot) * g.Table.RakePercent)
	if rakeAmount > g.Table.MaxRake {
		rakeAmount = g.Table.MaxRake
	}

	// Distribute remaining pot among winners
	remainingPot := g.Pot - rakeAmount
	winPerPlayer := remainingPot / int64(len(winners))
	remainder := remainingPot % int64(len(winners))

	winnings := make(map[uint]int64)
	for i, winner := range winners {
		amount := winPerPlayer
		if int64(i) < remainder {
			amount++ // Distribute remainder
		}
		winnings[winner.UserID] = amount
		winner.ChipCount += amount
	}

	return winnings, rakeAmount
}

// SerializeCards converts cards to JSON string
func SerializeCards(cards []Card) string {
	data, _ := json.Marshal(cards)
	return string(data)
}

// DeserializeCards converts JSON string to cards
func DeserializeCards(data string) []Card {
	var cards []Card
	json.Unmarshal([]byte(data), &cards)
	return cards
}
