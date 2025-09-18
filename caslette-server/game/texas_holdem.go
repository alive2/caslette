package game

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// TexasHoldemState represents the current state of a Texas Hold'em game
type TexasHoldemState string

const (
	PreFlop  TexasHoldemState = "preflop"
	Flop     TexasHoldemState = "flop"
	Turn     TexasHoldemState = "turn"
	River    TexasHoldemState = "river"
	Showdown TexasHoldemState = "showdown"
)

// TexasHoldemAction represents actions players can take
type TexasHoldemAction string

const (
	ActionFold  TexasHoldemAction = "fold"
	ActionCall  TexasHoldemAction = "call"
	ActionRaise TexasHoldemAction = "raise"
	ActionCheck TexasHoldemAction = "check"
	ActionBet   TexasHoldemAction = "bet"
	ActionAllIn TexasHoldemAction = "all_in"
)

// TexasHoldemPlayer extends the base Player with poker-specific data
type TexasHoldemPlayer struct {
	*Player
	Hand       *Hand `json:"hand"`
	Chips      int   `json:"chips"`
	CurrentBet int   `json:"currentBet"`
	TotalBet   int   `json:"totalBet"`
	HasFolded  bool  `json:"hasFolded"`
	IsAllIn    bool  `json:"isAllIn"`
	HasActed   bool  `json:"hasActed"`
}

// TexasHoldemEngine implements the Texas Hold'em poker game
type TexasHoldemEngine struct {
	*BaseGameEngine
	deck           *Deck
	communityCards *Hand
	pot            int
	currentBet     int
	dealerPos      int
	smallBlindPos  int
	bigBlindPos    int
	actionPos      int
	roundState     TexasHoldemState
	smallBlind     int
	bigBlind       int
	evaluator      *PokerEvaluator
	winners        []*TexasHoldemPlayer
}

// NewTexasHoldemEngine creates a new Texas Hold'em game engine
func NewTexasHoldemEngine(gameID string) *TexasHoldemEngine {
	base := NewBaseGameEngine(gameID)
	return &TexasHoldemEngine{
		BaseGameEngine: base,
		deck:           NewDeck(),
		communityCards: NewHand(),
		roundState:     PreFlop,
		smallBlind:     5,
		bigBlind:       10,
		evaluator:      NewPokerEvaluator(),
		winners:        make([]*TexasHoldemPlayer, 0),
	}
}

// Initialize sets up the Texas Hold'em game
func (the *TexasHoldemEngine) Initialize(config map[string]interface{}) error {
	if err := the.BaseGameEngine.Initialize(config); err != nil {
		return err
	}

	// Set blinds from config
	if sb, ok := config["smallBlind"].(int); ok {
		the.smallBlind = sb
	}
	if bb, ok := config["bigBlind"].(int); ok {
		the.bigBlind = bb
	}

	return nil
}

// AddPlayer adds a player to the Texas Hold'em game
func (the *TexasHoldemEngine) AddPlayer(player *Player) error {
	if len(the.players) >= 10 {
		return fmt.Errorf("maximum 10 players allowed")
	}

	// Set default chips if not provided
	if player.Data == nil {
		player.Data = make(map[string]interface{})
	}
	if _, hasChips := player.Data["chips"]; !hasChips {
		player.Data["chips"] = 1000
	}

	// Initialize poker-specific data
	player.Data["hand"] = []Card{}
	player.Data["currentBet"] = 0
	player.Data["totalBet"] = 0
	player.Data["hasFolded"] = false
	player.Data["isAllIn"] = false
	player.Data["hasActed"] = false

	return the.BaseGameEngine.AddPlayer(player)
}

// Start begins the Texas Hold'em game
func (the *TexasHoldemEngine) Start() error {
	if len(the.players) < 2 {
		return fmt.Errorf("need at least 2 players to start Texas Hold'em")
	}

	if err := the.BaseGameEngine.Start(); err != nil {
		return err
	}

	// Start new hand
	return the.startNewHand()
}

// startNewHand begins a new hand of poker
func (the *TexasHoldemEngine) startNewHand() error {
	// Reset deck and shuffle
	the.deck.Reset()
	the.communityCards.Clear()
	the.pot = 0
	the.currentBet = 0
	the.roundState = PreFlop
	the.winners = the.winners[:0]

	// Reset all players
	for _, player := range the.players {
		holdemPlayer := the.getHoldemPlayer(player.ID)
		if holdemPlayer != nil {
			holdemPlayer.Hand.Clear()
			holdemPlayer.CurrentBet = 0
			holdemPlayer.TotalBet = 0
			holdemPlayer.HasFolded = false
			holdemPlayer.IsAllIn = false
			holdemPlayer.HasActed = false
		}
	}

	// Set positions
	the.setPositions()

	// Post blinds
	if err := the.postBlinds(); err != nil {
		return err
	}

	// Deal hole cards
	if err := the.dealHoleCards(); err != nil {
		return err
	}

	// Set action to left of big blind for preflop
	the.actionPos = (the.bigBlindPos + 1) % len(the.getActivePlayers())

	the.emitEvent(&GameEvent{
		Type: "hand_started",
		Data: map[string]interface{}{
			"roundState":    the.roundState,
			"dealerPos":     the.dealerPos,
			"smallBlindPos": the.smallBlindPos,
			"bigBlindPos":   the.bigBlindPos,
			"pot":           the.pot,
			"currentBet":    the.currentBet,
		},
	})

	return nil
}

// setPositions sets dealer, small blind, and big blind positions
func (the *TexasHoldemEngine) setPositions() {
	activePlayers := the.getActivePlayers()
	numPlayers := len(activePlayers)

	if numPlayers == 2 {
		// Heads up: dealer is small blind
		the.smallBlindPos = the.dealerPos
		the.bigBlindPos = (the.dealerPos + 1) % numPlayers
	} else {
		// Multi-way: small blind is left of dealer
		the.smallBlindPos = (the.dealerPos + 1) % numPlayers
		the.bigBlindPos = (the.dealerPos + 2) % numPlayers
	}
}

// postBlinds posts the small and big blinds
func (the *TexasHoldemEngine) postBlinds() error {
	activePlayers := the.getActivePlayers()

	// Post small blind
	sbPlayer := the.getHoldemPlayer(activePlayers[the.smallBlindPos].ID)
	if sbPlayer == nil {
		return fmt.Errorf("small blind player not found")
	}

	sbAmount := min(the.smallBlind, sbPlayer.Chips)
	sbPlayer.Chips -= sbAmount
	sbPlayer.CurrentBet = sbAmount
	sbPlayer.TotalBet = sbAmount
	the.pot += sbAmount

	if sbPlayer.Chips == 0 {
		sbPlayer.IsAllIn = true
	}

	the.saveHoldemPlayer(sbPlayer)

	// Post big blind
	bbPlayer := the.getHoldemPlayer(activePlayers[the.bigBlindPos].ID)
	if bbPlayer == nil {
		return fmt.Errorf("big blind player not found")
	}

	bbAmount := min(the.bigBlind, bbPlayer.Chips)
	bbPlayer.Chips -= bbAmount
	bbPlayer.CurrentBet = bbAmount
	bbPlayer.TotalBet = bbAmount
	the.pot += bbAmount
	the.currentBet = bbAmount

	if bbPlayer.Chips == 0 {
		bbPlayer.IsAllIn = true
	}

	the.saveHoldemPlayer(bbPlayer)

	the.emitEvent(&GameEvent{
		Type: "blinds_posted",
		Data: map[string]interface{}{
			"smallBlind": map[string]interface{}{
				"playerID": sbPlayer.ID,
				"amount":   sbAmount,
			},
			"bigBlind": map[string]interface{}{
				"playerID": bbPlayer.ID,
				"amount":   bbAmount,
			},
			"pot": the.pot,
		},
	})

	return nil
}

// dealHoleCards deals 2 cards to each player
func (the *TexasHoldemEngine) dealHoleCards() error {
	activePlayers := the.getActivePlayers()

	// Deal 2 cards to each player
	for i := 0; i < 2; i++ {
		for _, player := range activePlayers {
			holdemPlayer := the.getHoldemPlayer(player.ID)
			if holdemPlayer == nil {
				continue
			}

			card, err := the.deck.Deal()
			if err != nil {
				return fmt.Errorf("error dealing hole cards: %v", err)
			}

			holdemPlayer.Hand.AddCard(card)
			the.saveHoldemPlayer(holdemPlayer)
		}
	}

	the.emitEvent(&GameEvent{
		Type: "hole_cards_dealt",
		Data: map[string]interface{}{
			"playersCount": len(activePlayers),
		},
	})

	return nil
}

// ProcessAction processes a player action
func (the *TexasHoldemEngine) ProcessAction(ctx context.Context, action *GameAction) (*GameEvent, error) {
	if err := the.IsValidAction(action); err != nil {
		return nil, err
	}

	player := the.getHoldemPlayer(action.PlayerID)
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	actionType := action.Data["action"].(string)
	amount := 0
	if val, ok := action.Data["amount"].(float64); ok {
		amount = int(val)
	} else if val, ok := action.Data["amount"].(int); ok {
		amount = val
	}

	var event *GameEvent
	var err error

	switch TexasHoldemAction(actionType) {
	case ActionFold:
		event, err = the.processFold(player)
	case ActionCall:
		event, err = the.processCall(player)
	case ActionRaise:
		event, err = the.processRaise(player, amount)
	case ActionBet:
		event, err = the.processBet(player, amount)
	case ActionCheck:
		event, err = the.processCheck(player)
	case ActionAllIn:
		event, err = the.processAllIn(player)
	default:
		return nil, fmt.Errorf("unknown action: %s", actionType)
	}

	if err != nil {
		return nil, err
	}

	player.HasActed = true
	the.saveHoldemPlayer(player)

	// Check if betting round is complete
	if the.isBettingRoundComplete() {
		if err := the.nextBettingRound(); err != nil {
			return nil, err
		}
	} else {
		// Move to next player
		the.nextPlayer()
	}

	return event, nil
}

// Helper methods for processing specific actions

func (the *TexasHoldemEngine) processFold(player *TexasHoldemPlayer) (*GameEvent, error) {
	player.HasFolded = true
	player.IsActive = false
	the.saveHoldemPlayer(player)

	event := &GameEvent{
		Type:     "player_folded",
		PlayerID: player.ID,
		Data: map[string]interface{}{
			"playerID": player.ID,
		},
	}

	// Check if only one player remains
	activePlayers := the.getActivePlayers()
	if len(activePlayers) == 1 {
		the.winners = []*TexasHoldemPlayer{the.getHoldemPlayer(activePlayers[0].ID)}
		the.SetState(GameStateFinished)
		the.distributePot()
	}

	return event, nil
}

func (the *TexasHoldemEngine) processCall(player *TexasHoldemPlayer) (*GameEvent, error) {
	callAmount := the.currentBet - player.CurrentBet
	actualAmount := min(callAmount, player.Chips)

	player.Chips -= actualAmount
	player.CurrentBet += actualAmount
	player.TotalBet += actualAmount
	the.pot += actualAmount

	if player.Chips == 0 {
		player.IsAllIn = true
	}

	the.saveHoldemPlayer(player)

	return &GameEvent{
		Type:     "player_called",
		PlayerID: player.ID,
		Data: map[string]interface{}{
			"playerID": player.ID,
			"amount":   actualAmount,
			"pot":      the.pot,
		},
	}, nil
}

func (the *TexasHoldemEngine) processRaise(player *TexasHoldemPlayer, amount int) (*GameEvent, error) {
	totalBet := the.currentBet + amount
	actualAmount := min(totalBet-player.CurrentBet, player.Chips)

	player.Chips -= actualAmount
	player.CurrentBet += actualAmount
	player.TotalBet += actualAmount
	the.pot += actualAmount
	the.currentBet = player.CurrentBet

	if player.Chips == 0 {
		player.IsAllIn = true
	}

	// Reset HasActed for all other players
	for _, p := range the.players {
		holdemPlayer := the.getHoldemPlayer(p.ID)
		if holdemPlayer != nil && holdemPlayer.ID != player.ID && !holdemPlayer.HasFolded && !holdemPlayer.IsAllIn {
			holdemPlayer.HasActed = false
			the.saveHoldemPlayer(holdemPlayer)
		}
	}

	the.saveHoldemPlayer(player)

	return &GameEvent{
		Type:     "player_raised",
		PlayerID: player.ID,
		Data: map[string]interface{}{
			"playerID": player.ID,
			"amount":   amount,
			"totalBet": the.currentBet,
			"pot":      the.pot,
		},
	}, nil
}

func (the *TexasHoldemEngine) processBet(player *TexasHoldemPlayer, amount int) (*GameEvent, error) {
	actualAmount := min(amount, player.Chips)

	player.Chips -= actualAmount
	player.CurrentBet = actualAmount
	player.TotalBet += actualAmount
	the.pot += actualAmount
	the.currentBet = actualAmount

	if player.Chips == 0 {
		player.IsAllIn = true
	}

	the.saveHoldemPlayer(player)

	return &GameEvent{
		Type:     "player_bet",
		PlayerID: player.ID,
		Data: map[string]interface{}{
			"playerID": player.ID,
			"amount":   actualAmount,
			"pot":      the.pot,
		},
	}, nil
}

func (the *TexasHoldemEngine) processCheck(player *TexasHoldemPlayer) (*GameEvent, error) {
	the.saveHoldemPlayer(player)

	return &GameEvent{
		Type:     "player_checked",
		PlayerID: player.ID,
		Data: map[string]interface{}{
			"playerID": player.ID,
		},
	}, nil
}

func (the *TexasHoldemEngine) processAllIn(player *TexasHoldemPlayer) (*GameEvent, error) {
	amount := player.Chips
	player.CurrentBet += amount
	player.TotalBet += amount
	player.Chips = 0
	player.IsAllIn = true
	the.pot += amount

	if player.CurrentBet > the.currentBet {
		the.currentBet = player.CurrentBet
		// Reset HasActed for all other players
		for _, p := range the.players {
			holdemPlayer := the.getHoldemPlayer(p.ID)
			if holdemPlayer != nil && holdemPlayer.ID != player.ID && !holdemPlayer.HasFolded && !holdemPlayer.IsAllIn {
				holdemPlayer.HasActed = false
				the.saveHoldemPlayer(holdemPlayer)
			}
		}
	}

	the.saveHoldemPlayer(player)

	return &GameEvent{
		Type:     "player_all_in",
		PlayerID: player.ID,
		Data: map[string]interface{}{
			"playerID": player.ID,
			"amount":   amount,
			"pot":      the.pot,
		},
	}, nil
}

// IsValidAction checks if an action is valid
func (the *TexasHoldemEngine) IsValidAction(action *GameAction) error {
	if the.GetState() != GameStateInProgress {
		return fmt.Errorf("game is not in progress")
	}

	// Enhanced data validation
	if action.Data == nil {
		return fmt.Errorf("action data is required")
	}

	// Prevent infinite recursion from circular references
	if err := validateDataStructure(action.Data, 0, 10); err != nil {
		return fmt.Errorf("invalid data structure: %v", err)
	}

	player := the.getHoldemPlayer(action.PlayerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.HasFolded {
		return fmt.Errorf("player has folded")
	}

	if player.IsAllIn {
		return fmt.Errorf("player is all-in")
	}

	// Check if it's the player's turn
	currentPlayerID := the.getCurrentActionPlayerID()
	if action.PlayerID != currentPlayerID {
		return fmt.Errorf("not player's turn")
	}

	actionType, ok := action.Data["action"].(string)
	if !ok {
		return fmt.Errorf("action type is required and must be a string")
	}

	// Validate action string format before trimming
	originalAction := actionType
	actionType = strings.TrimSpace(actionType)
	if actionType == "" {
		return fmt.Errorf("action type cannot be empty")
	}

	// Check if trimming changed the string (indicates whitespace issues)
	if originalAction != actionType {
		return fmt.Errorf("action type contains invalid whitespace characters")
	}

	// Check for suspicious characters in action type
	for _, r := range actionType {
		if r < 32 || r > 126 { // Outside printable ASCII range
			return fmt.Errorf("action type contains invalid characters")
		}
	}

	// Validate against known action types (case sensitive)
	validActions := map[string]bool{
		string(ActionFold):  true,
		string(ActionCall):  true,
		string(ActionRaise): true,
		string(ActionBet):   true,
		string(ActionCheck): true,
		string(ActionAllIn): true,
	}

	if !validActions[actionType] {
		return fmt.Errorf("invalid action type: %s", actionType)
	}

	// Check for suspicious or invalid field names in action data
	for key := range action.Data {
		switch key {
		case "action":
			// Valid field
		case "amount":
			// Valid field for bet/raise actions
		default:
			// Check for common typos that indicate field name confusion
			if key == "Action" || key == "type" || key == "Type" {
				return fmt.Errorf("invalid field name '%s', did you mean 'action'?", key)
			}
			// Other fields are ignored rather than rejected to allow extra data
			// that might be used for testing or future extensions
		}
	}

	// Validate specific action
	switch TexasHoldemAction(actionType) {
	case ActionFold:
		// Validate no extra/conflicting data for fold action
		if len(action.Data) > 1 { // Only "action" should be present
			for key := range action.Data {
				if key != "action" {
					return fmt.Errorf("fold action should not contain additional data: %s", key)
				}
			}
		}
		return nil // Always valid if it's player's turn
	case ActionCall:
		if the.currentBet == player.CurrentBet {
			return fmt.Errorf("cannot call when current bet equals player's bet")
		}
		// Validate no conflicting data for call action
		if amount, exists := action.Data["amount"]; exists {
			return fmt.Errorf("call action should not contain amount data: %v", amount)
		}
	case ActionRaise:
		amount, ok := action.Data["amount"]
		if !ok {
			return fmt.Errorf("raise amount is required")
		}
		// Validate amount is a valid number type
		if raiseAmount, ok := amount.(float64); ok {
			if raiseAmount <= 0 {
				return fmt.Errorf("raise amount must be positive")
			}
		} else if raiseAmount, ok := amount.(int); ok {
			if raiseAmount <= 0 {
				return fmt.Errorf("raise amount must be positive")
			}
		} else {
			return fmt.Errorf("raise amount must be a number")
		}
		// Validate no conflicting action type data
		if bet, exists := action.Data["bet"]; exists {
			return fmt.Errorf("raise action should not contain bet data: %v", bet)
		}
	case ActionBet:
		if the.currentBet > 0 {
			return fmt.Errorf("cannot bet when there is already a bet")
		}
		amount, ok := action.Data["amount"]
		if !ok {
			return fmt.Errorf("bet amount is required")
		}
		// Validate amount is a valid number type
		if betAmount, ok := amount.(float64); ok {
			if betAmount <= 0 {
				return fmt.Errorf("bet amount must be positive")
			}
		} else if betAmount, ok := amount.(int); ok {
			if betAmount <= 0 {
				return fmt.Errorf("bet amount must be positive")
			}
		} else {
			return fmt.Errorf("bet amount must be a number")
		}
		// Validate no conflicting action type data
		if raise, exists := action.Data["raise"]; exists {
			return fmt.Errorf("bet action should not contain raise data: %v", raise)
		}
	case ActionCheck:
		if the.currentBet > player.CurrentBet {
			return fmt.Errorf("cannot check when there is a bet to call")
		}
		// Validate no extra data for check action
		if amount, exists := action.Data["amount"]; exists {
			return fmt.Errorf("check action should not contain amount data: %v", amount)
		}
	case ActionAllIn:
		if player.Chips <= 0 {
			return fmt.Errorf("player has no chips to go all-in")
		}
		// Validate no amount data for all-in action
		if amount, exists := action.Data["amount"]; exists {
			return fmt.Errorf("all-in action should not contain amount data: %v", amount)
		}
	default:
		return fmt.Errorf("invalid action type: %s", actionType)
	}

	return nil
}

// GetValidActions returns valid actions for a player
func (the *TexasHoldemEngine) GetValidActions(playerID string) []string {
	player := the.getHoldemPlayer(playerID)
	if player == nil || player.HasFolded || player.IsAllIn {
		return []string{}
	}

	if the.getCurrentActionPlayerID() != playerID {
		return []string{}
	}

	actions := []string{string(ActionFold)}

	if player.Chips > 0 {
		actions = append(actions, string(ActionAllIn))
	}

	if the.currentBet > player.CurrentBet {
		// Player can call
		if player.Chips >= (the.currentBet - player.CurrentBet) {
			actions = append(actions, string(ActionCall))
		}
		// Player can raise
		if player.Chips > (the.currentBet - player.CurrentBet) {
			actions = append(actions, string(ActionRaise))
		}
	} else {
		// No bet to call, player can check or bet
		actions = append(actions, string(ActionCheck))
		if player.Chips > 0 {
			actions = append(actions, string(ActionBet))
		}
	}

	return actions
}

// Helper methods

func (the *TexasHoldemEngine) getHoldemPlayer(playerID string) *TexasHoldemPlayer {
	player, err := the.GetPlayer(playerID)
	if err != nil {
		return nil
	}

	// Convert to TexasHoldemPlayer
	holdemPlayer := &TexasHoldemPlayer{
		Player: player,
		Hand:   NewHand(),
	}

	// Load poker-specific data from player.Data
	if player.Data != nil {
		if chips, ok := player.Data["chips"].(int); ok {
			holdemPlayer.Chips = chips
		} else {
			holdemPlayer.Chips = 1000
		}
		if currentBet, ok := player.Data["currentBet"].(int); ok {
			holdemPlayer.CurrentBet = currentBet
		}
		if totalBet, ok := player.Data["totalBet"].(int); ok {
			holdemPlayer.TotalBet = totalBet
		}
		if hasFolded, ok := player.Data["hasFolded"].(bool); ok {
			holdemPlayer.HasFolded = hasFolded
		}
		if isAllIn, ok := player.Data["isAllIn"].(bool); ok {
			holdemPlayer.IsAllIn = isAllIn
		}
		if hasActed, ok := player.Data["hasActed"].(bool); ok {
			holdemPlayer.HasActed = hasActed
		}
		if handData, ok := player.Data["hand"].([]Card); ok {
			holdemPlayer.Hand.Cards = handData
		}
	} else {
		holdemPlayer.Chips = 1000
	}

	return holdemPlayer
}

// saveHoldemPlayer saves the holdem player data back to the base player
func (the *TexasHoldemEngine) saveHoldemPlayer(holdemPlayer *TexasHoldemPlayer) {
	player, err := the.GetPlayer(holdemPlayer.ID)
	if err != nil {
		return
	}

	if player.Data == nil {
		player.Data = make(map[string]interface{})
	}

	player.Data["chips"] = holdemPlayer.Chips
	player.Data["currentBet"] = holdemPlayer.CurrentBet
	player.Data["totalBet"] = holdemPlayer.TotalBet
	player.Data["hasFolded"] = holdemPlayer.HasFolded
	player.Data["isAllIn"] = holdemPlayer.IsAllIn
	player.Data["hasActed"] = holdemPlayer.HasActed
	player.Data["hand"] = holdemPlayer.Hand.Cards
	player.IsActive = !holdemPlayer.HasFolded
}

func (the *TexasHoldemEngine) getActivePlayers() []*Player {
	activePlayers := make([]*Player, 0)
	for _, player := range the.players {
		holdemPlayer := the.getHoldemPlayer(player.ID)
		if holdemPlayer != nil && !holdemPlayer.HasFolded {
			activePlayers = append(activePlayers, player)
		}
	}

	// Sort by position
	sort.Slice(activePlayers, func(i, j int) bool {
		return activePlayers[i].Position < activePlayers[j].Position
	})

	return activePlayers
}

func (the *TexasHoldemEngine) getCurrentActionPlayerID() string {
	activePlayers := the.getActivePlayers()
	if len(activePlayers) == 0 || the.actionPos >= len(activePlayers) {
		return ""
	}
	return activePlayers[the.actionPos].ID
}

func (the *TexasHoldemEngine) nextPlayer() {
	activePlayers := the.getActivePlayers()
	if len(activePlayers) <= 1 {
		return
	}

	for {
		the.actionPos = (the.actionPos + 1) % len(activePlayers)
		player := the.getHoldemPlayer(activePlayers[the.actionPos].ID)
		if player != nil && !player.HasFolded && !player.IsAllIn {
			break
		}
	}
}

func (the *TexasHoldemEngine) isBettingRoundComplete() bool {
	activePlayers := the.getActivePlayers()

	playersToAct := 0
	for _, player := range activePlayers {
		holdemPlayer := the.getHoldemPlayer(player.ID)
		if holdemPlayer != nil && !holdemPlayer.HasFolded && !holdemPlayer.IsAllIn {
			if !holdemPlayer.HasActed || holdemPlayer.CurrentBet < the.currentBet {
				playersToAct++
			}
		}
	}

	return playersToAct == 0
}

func (the *TexasHoldemEngine) nextBettingRound() error {
	// Reset current bets and hasActed flags
	for _, player := range the.players {
		holdemPlayer := the.getHoldemPlayer(player.ID)
		if holdemPlayer != nil {
			holdemPlayer.CurrentBet = 0
			holdemPlayer.HasActed = false
			the.saveHoldemPlayer(holdemPlayer)
		}
	}
	the.currentBet = 0

	switch the.roundState {
	case PreFlop:
		return the.dealFlop()
	case Flop:
		return the.dealTurn()
	case Turn:
		return the.dealRiver()
	case River:
		return the.showdown()
	default:
		return fmt.Errorf("unknown round state")
	}
}

func (the *TexasHoldemEngine) dealFlop() error {
	// Burn one card
	the.deck.Deal()

	// Deal 3 community cards
	for i := 0; i < 3; i++ {
		card, err := the.deck.Deal()
		if err != nil {
			return err
		}
		the.communityCards.AddCard(card)
	}

	the.roundState = Flop
	the.actionPos = the.smallBlindPos

	the.emitEvent(&GameEvent{
		Type: "flop_dealt",
		Data: map[string]interface{}{
			"communityCards": the.communityCards.Cards,
		},
	})

	return nil
}

func (the *TexasHoldemEngine) dealTurn() error {
	// Burn one card
	the.deck.Deal()

	// Deal 1 community card
	card, err := the.deck.Deal()
	if err != nil {
		return err
	}
	the.communityCards.AddCard(card)

	the.roundState = Turn
	the.actionPos = the.smallBlindPos

	the.emitEvent(&GameEvent{
		Type: "turn_dealt",
		Data: map[string]interface{}{
			"communityCards": the.communityCards.Cards,
		},
	})

	return nil
}

func (the *TexasHoldemEngine) dealRiver() error {
	// Burn one card
	the.deck.Deal()

	// Deal 1 community card
	card, err := the.deck.Deal()
	if err != nil {
		return err
	}
	the.communityCards.AddCard(card)

	the.roundState = River
	the.actionPos = the.smallBlindPos

	the.emitEvent(&GameEvent{
		Type: "river_dealt",
		Data: map[string]interface{}{
			"communityCards": the.communityCards.Cards,
		},
	})

	return nil
}

func (the *TexasHoldemEngine) showdown() error {
	the.roundState = Showdown
	the.determineWinners()
	the.distributePot()
	the.SetState(GameStateFinished)

	the.emitEvent(&GameEvent{
		Type: "showdown",
		Data: map[string]interface{}{
			"winners":        the.winners,
			"communityCards": the.communityCards.Cards,
		},
	})

	return nil
}

func (the *TexasHoldemEngine) determineWinners() {
	activePlayers := the.getActivePlayers()
	playerHands := make(map[string]*PokerHand)

	// Evaluate each player's best hand
	for _, player := range activePlayers {
		holdemPlayer := the.getHoldemPlayer(player.ID)
		if holdemPlayer == nil || holdemPlayer.HasFolded {
			continue
		}

		// Combine hole cards with community cards
		allCards := make([]Card, 0, 7)
		allCards = append(allCards, holdemPlayer.Hand.Cards...)
		allCards = append(allCards, the.communityCards.Cards...)

		// Find best 5-card hand
		bestHand := the.evaluator.FindBestHand(allCards)
		playerHands[player.ID] = bestHand
	}

	// Find winners
	var bestHand *PokerHand
	winners := make([]*TexasHoldemPlayer, 0)

	for playerID, hand := range playerHands {
		if bestHand == nil || hand.Compare(bestHand) > 0 {
			bestHand = hand
			winners = []*TexasHoldemPlayer{the.getHoldemPlayer(playerID)}
		} else if hand.Compare(bestHand) == 0 {
			winners = append(winners, the.getHoldemPlayer(playerID))
		}
	}

	the.winners = winners
}

func (the *TexasHoldemEngine) distributePot() {
	if len(the.winners) == 0 {
		return
	}

	potPerWinner := the.pot / len(the.winners)
	for _, winner := range the.winners {
		winner.Chips += potPerWinner
	}

	the.emitEvent(&GameEvent{
		Type: "pot_distributed",
		Data: map[string]interface{}{
			"winners":      the.winners,
			"potPerWinner": potPerWinner,
			"totalPot":     the.pot,
		},
	})
}

// GetWinners returns the winners of the current hand
func (the *TexasHoldemEngine) GetWinners() []*Player {
	winners := make([]*Player, len(the.winners))
	for i, winner := range the.winners {
		winners[i] = winner.Player
	}
	return winners
}

// IsGameOver checks if the game is over
func (the *TexasHoldemEngine) IsGameOver() bool {
	if the.GetState() == GameStateFinished {
		return true
	}

	// Game is over if only one player has chips
	playersWithChips := 0
	for _, player := range the.players {
		holdemPlayer := the.getHoldemPlayer(player.ID)
		if holdemPlayer != nil && holdemPlayer.Chips > 0 {
			playersWithChips++
		}
	}

	return playersWithChips <= 1
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// validateDataStructure prevents circular references and deep nesting attacks
func validateDataStructure(data interface{}, depth, maxDepth int) error {
	if depth > maxDepth {
		return fmt.Errorf("maximum nesting depth exceeded")
	}

	if data == nil {
		return nil
	}

	visited := make(map[uintptr]bool)
	return validateDataStructureRecursive(data, depth, maxDepth, visited)
}

func validateDataStructureRecursive(data interface{}, depth, maxDepth int, visited map[uintptr]bool) error {
	if depth > maxDepth {
		return fmt.Errorf("maximum nesting depth exceeded")
	}

	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)

	// Check for circular references in pointer types
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Map || v.Kind() == reflect.Slice {
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			ptr := v.Pointer()
			if visited[ptr] {
				return fmt.Errorf("circular reference detected")
			}
			visited[ptr] = true
			defer delete(visited, ptr)
			return validateDataStructureRecursive(v.Elem().Interface(), depth+1, maxDepth, visited)
		}

		if v.Kind() == reflect.Map {
			if !v.IsNil() {
				ptr := v.Pointer()
				if visited[ptr] {
					return fmt.Errorf("circular reference detected")
				}
				visited[ptr] = true
				defer delete(visited, ptr)

				for _, key := range v.MapKeys() {
					if err := validateDataStructureRecursive(key.Interface(), depth+1, maxDepth, visited); err != nil {
						return err
					}
					if err := validateDataStructureRecursive(v.MapIndex(key).Interface(), depth+1, maxDepth, visited); err != nil {
						return err
					}
				}
			}
		}

		if v.Kind() == reflect.Slice {
			if !v.IsNil() {
				ptr := v.Pointer()
				if visited[ptr] {
					return fmt.Errorf("circular reference detected")
				}
				visited[ptr] = true
				defer delete(visited, ptr)

				for i := 0; i < v.Len(); i++ {
					if err := validateDataStructureRecursive(v.Index(i).Interface(), depth+1, maxDepth, visited); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// SetSmallBlind sets the small blind amount for the engine
func (the *TexasHoldemEngine) SetSmallBlind(amount int) {
	the.smallBlind = amount
}

// SetBigBlind sets the big blind amount for the engine
func (the *TexasHoldemEngine) SetBigBlind(amount int) {
	the.bigBlind = amount
}

// GetPublicGameState returns public game state (community cards, pot, etc.)
func (the *TexasHoldemEngine) GetPublicGameState() map[string]interface{} {
	currentPlayerID := ""
	activePlayers := the.getActivePlayers()
	if len(activePlayers) > 0 && the.actionPos < len(activePlayers) {
		currentPlayerID = activePlayers[the.actionPos].ID
	}

	return map[string]interface{}{
		"pot":             the.pot,
		"community_cards": the.communityCards,
		"current_player":  currentPlayerID,
		"round_state":     the.roundState,
		"dealer_position": the.dealerPos,
		"small_blind":     the.smallBlind,
		"big_blind":       the.bigBlind,
	}
}

// GetPlayerState returns private state for a specific player
func (the *TexasHoldemEngine) GetPlayerState(playerID string) map[string]interface{} {
	player, err := the.GetPlayer(playerID)
	if err != nil || player == nil {
		return nil
	}

	holdemPlayer := the.getHoldemPlayer(playerID)
	if holdemPlayer == nil {
		return nil
	}

	return map[string]interface{}{
		"hand":        holdemPlayer.Hand,
		"chips":       holdemPlayer.Chips,
		"current_bet": holdemPlayer.CurrentBet,
		"is_folded":   holdemPlayer.HasFolded,
		"is_all_in":   holdemPlayer.IsAllIn,
		"position":    player.Position,
	}
}
