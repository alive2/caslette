package game

import (
	"context"
	"testing"
)

func TestTexasHoldemEngine(t *testing.T) {
	t.Run("NewTexasHoldemEngine", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")
		if engine.smallBlind != 5 {
			t.Errorf("Expected small blind 5, got %d", engine.smallBlind)
		}
		if engine.bigBlind != 10 {
			t.Errorf("Expected big blind 10, got %d", engine.bigBlind)
		}
		if engine.roundState != PreFlop {
			t.Errorf("Expected initial round state %v, got %v", PreFlop, engine.roundState)
		}
	})

	t.Run("Initialize", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")
		config := map[string]interface{}{
			"smallBlind": 25,
			"bigBlind":   50,
		}

		err := engine.Initialize(config)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if engine.smallBlind != 25 {
			t.Errorf("Expected small blind 25 after config, got %d", engine.smallBlind)
		}
		if engine.bigBlind != 50 {
			t.Errorf("Expected big blind 50 after config, got %d", engine.bigBlind)
		}
	})

	t.Run("AddPlayer", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")
		player := &Player{
			ID:   "player1",
			Name: "Player 1",
			Data: map[string]interface{}{
				"chips": 2000,
			},
		}

		err := engine.AddPlayer(player)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		holdemPlayer := engine.getHoldemPlayer("player1")
		if holdemPlayer == nil {
			t.Error("Expected to find holdem player")
		}
		if holdemPlayer.Chips != 2000 {
			t.Errorf("Expected 2000 chips, got %d", holdemPlayer.Chips)
		}
	})

	t.Run("AddTooManyPlayers", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add 10 players
		for i := 0; i < 10; i++ {
			player := &Player{
				ID:   string(rune('a' + i)),
				Name: "Player " + string(rune('A'+i)),
			}
			engine.AddPlayer(player)
		}

		// Try to add 11th player
		player := &Player{
			ID:   "too-many",
			Name: "Too Many",
		}
		err := engine.AddPlayer(player)
		if err == nil {
			t.Error("Expected error when adding more than 10 players")
		}
	})

	t.Run("StartGameInsufficientPlayers", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")
		player := &Player{ID: "player1", Name: "Player 1"}
		engine.AddPlayer(player)

		err := engine.Start()
		if err == nil {
			t.Error("Expected error when starting with only 1 player")
		}
	})

	t.Run("StartGameValidPlayers", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add 2 players
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}

		err := engine.Start()
		if err != nil {
			t.Errorf("Unexpected error starting game: %v", err)
		}

		if engine.GetState() != GameStateInProgress {
			t.Errorf("Expected state %v, got %v", GameStateInProgress, engine.GetState())
		}

		// Check that blinds were posted
		if engine.pot == 0 {
			t.Error("Expected pot to have blinds")
		}

		// Check that hole cards were dealt
		for _, player := range engine.GetPlayers() {
			holdemPlayer := engine.getHoldemPlayer(player.ID)
			if holdemPlayer != nil && holdemPlayer.Hand.Size() != 2 {
				t.Errorf("Expected 2 hole cards for player %s, got %d", player.ID, holdemPlayer.Hand.Size())
			}
		}
	})
}

func TestTexasHoldemActions(t *testing.T) {
	t.Run("ValidActionChecks", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 3; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		currentPlayerID := engine.getCurrentActionPlayerID()

		// Test fold action (always valid)
		foldAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: currentPlayerID,
			Data: map[string]interface{}{
				"action": "fold",
			},
		}

		err := engine.IsValidAction(foldAction)
		if err != nil {
			t.Errorf("Fold action should be valid: %v", err)
		}

		// Test call action (should be valid if there's a bet to call)
		callAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: currentPlayerID,
			Data: map[string]interface{}{
				"action": "call",
			},
		}

		err = engine.IsValidAction(callAction)
		// Should be valid in preflop with big blind
		if err != nil {
			t.Errorf("Call action should be valid in preflop: %v", err)
		}
	})

	t.Run("InvalidPlayerTurn", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 3; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		// Try action from wrong player
		wrongPlayerAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: "wrong_player",
			Data: map[string]interface{}{
				"action": "fold",
			},
		}

		err := engine.IsValidAction(wrongPlayerAction)
		if err == nil {
			t.Error("Expected error for action from wrong player")
		}
	})

	t.Run("ProcessFoldAction", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 3; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		currentPlayerID := engine.getCurrentActionPlayerID()
		foldAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: currentPlayerID,
			Data: map[string]interface{}{
				"action": "fold",
			},
		}

		event, err := engine.ProcessAction(context.Background(), foldAction)
		if err != nil {
			t.Errorf("Unexpected error processing fold: %v", err)
		}
		if event.Type != "player_folded" {
			t.Errorf("Expected event type 'player_folded', got %s", event.Type)
		}

		// Check that player is no longer active
		holdemPlayer := engine.getHoldemPlayer(currentPlayerID)
		if holdemPlayer != nil && !holdemPlayer.HasFolded {
			t.Error("Player should be marked as folded")
		}
	})

	t.Run("ProcessCallAction", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		initialPot := engine.pot
		currentPlayerID := engine.getCurrentActionPlayerID()

		callAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: currentPlayerID,
			Data: map[string]interface{}{
				"action": "call",
			},
		}

		event, err := engine.ProcessAction(context.Background(), callAction)
		if err != nil {
			t.Errorf("Unexpected error processing call: %v", err)
		}
		if event.Type != "player_called" {
			t.Errorf("Expected event type 'player_called', got %s", event.Type)
		}

		// Check that pot increased
		if engine.pot <= initialPot {
			t.Error("Pot should have increased after call")
		}
	})

	t.Run("ProcessRaiseAction", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		currentPlayerID := engine.getCurrentActionPlayerID()
		raiseAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: currentPlayerID,
			Data: map[string]interface{}{
				"action": "raise",
				"amount": 20,
			},
		}

		event, err := engine.ProcessAction(context.Background(), raiseAction)
		if err != nil {
			t.Errorf("Unexpected error processing raise: %v", err)
		}
		if event.Type != "player_raised" {
			t.Errorf("Expected event type 'player_raised', got %s", event.Type)
		}

		// Check that current bet increased
		if engine.currentBet <= engine.bigBlind {
			t.Error("Current bet should have increased after raise")
		}
	})

	t.Run("GetValidActions", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		currentPlayerID := engine.getCurrentActionPlayerID()
		actions := engine.GetValidActions(currentPlayerID)

		// Should have at least fold and call (preflop with big blind)
		if len(actions) < 2 {
			t.Errorf("Expected at least 2 valid actions, got %d", len(actions))
		}

		// Should contain fold
		foundFold := false
		for _, action := range actions {
			if action == "fold" {
				foundFold = true
				break
			}
		}
		if !foundFold {
			t.Error("Valid actions should include fold")
		}
	})
}

func TestTexasHoldemBettingRounds(t *testing.T) {
	t.Run("BettingRoundProgression", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		if engine.roundState != PreFlop {
			t.Errorf("Expected PreFlop state, got %v", engine.roundState)
		}

		t.Logf("Initial state: roundState=%v, currentBet=%d, pot=%d", engine.roundState, engine.currentBet, engine.pot)

		// Both players act to move to flop
		for i := 0; i < 2; i++ {
			currentPlayerID := engine.getCurrentActionPlayerID()
			if currentPlayerID == "" {
				t.Logf("No current action player at iteration %d", i)
				break
			}

			t.Logf("Iteration %d: currentPlayer=%s", i, currentPlayerID)

			// Check if betting round is already complete
			if engine.isBettingRoundComplete() {
				t.Logf("Betting round complete after %d actions", i)
				break
			}

			// Get the player's current bet to determine correct action
			holdemPlayer := engine.getHoldemPlayer(currentPlayerID)
			var action string
			if holdemPlayer.CurrentBet < engine.currentBet {
				action = "call" // Need to call to match current bet
			} else {
				action = "check" // Already matched current bet, can check
			}

			actionData := &GameAction{
				Type:     "texas_holdem_action",
				PlayerID: currentPlayerID,
				Data: map[string]interface{}{
					"action": action,
				},
			}

			t.Logf("Player %s using action: %s (currentBet=%d, tableBet=%d)", currentPlayerID, action, holdemPlayer.CurrentBet, engine.currentBet)
			engine.ProcessAction(context.Background(), actionData)
			t.Logf("After action %d: roundState=%v, currentBet=%d, pot=%d", i, engine.roundState, engine.currentBet, engine.pot)
		}

		// Should have moved to flop
		if engine.roundState != Flop {
			t.Errorf("Expected Flop state after betting round, got %v", engine.roundState)
		}

		// Should have community cards
		if engine.communityCards.Size() != 3 {
			t.Errorf("Expected 3 community cards in flop, got %d", engine.communityCards.Size())
		}
	})
}

func TestTexasHoldemGameEnd(t *testing.T) {
	t.Run("GameEndByFolding", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players and start game
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		// First player folds, leaving only one active player
		currentPlayerID := engine.getCurrentActionPlayerID()
		foldAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: currentPlayerID,
			Data: map[string]interface{}{
				"action": "fold",
			},
		}

		engine.ProcessAction(context.Background(), foldAction)

		// Game should be finished
		if engine.GetState() != GameStateFinished {
			t.Errorf("Expected game to be finished, got state %v", engine.GetState())
		}

		// Should have a winner
		winners := engine.GetWinners()
		if len(winners) != 1 {
			t.Errorf("Expected 1 winner, got %d", len(winners))
		}
	})
}

func TestTexasHoldemPositions(t *testing.T) {
	t.Run("HeadsUpPositions", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add 2 players
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		// In heads up, dealer is small blind
		if engine.smallBlindPos != engine.dealerPos {
			t.Error("In heads up, dealer should be small blind")
		}
		if engine.bigBlindPos != (engine.dealerPos+1)%2 {
			t.Error("In heads up, big blind should be opposite dealer")
		}
	})

	t.Run("MultiWayPositions", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add 4 players
		for i := 1; i <= 4; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		// In multi-way, small blind is left of dealer
		expectedSB := (engine.dealerPos + 1) % 4
		expectedBB := (engine.dealerPos + 2) % 4

		if engine.smallBlindPos != expectedSB {
			t.Errorf("Expected small blind position %d, got %d", expectedSB, engine.smallBlindPos)
		}
		if engine.bigBlindPos != expectedBB {
			t.Errorf("Expected big blind position %d, got %d", expectedBB, engine.bigBlindPos)
		}
	})
}

func TestTexasHoldemChipManagement(t *testing.T) {
	t.Run("BlindPosting", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players with specific chip amounts
		for i := 1; i <= 2; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
				Data: map[string]interface{}{
					"chips": 1000,
				},
			}
			engine.AddPlayer(player)
		}
		engine.Start()

		// Check that blinds were deducted from chips
		activePlayers := engine.getActivePlayers()
		sbPlayer := engine.getHoldemPlayer(activePlayers[engine.smallBlindPos].ID)
		bbPlayer := engine.getHoldemPlayer(activePlayers[engine.bigBlindPos].ID)

		if sbPlayer.Chips != 1000-engine.smallBlind {
			t.Errorf("Small blind player should have %d chips, got %d", 1000-engine.smallBlind, sbPlayer.Chips)
		}
		if bbPlayer.Chips != 1000-engine.bigBlind {
			t.Errorf("Big blind player should have %d chips, got %d", 1000-engine.bigBlind, bbPlayer.Chips)
		}

		// Check pot
		expectedPot := engine.smallBlind + engine.bigBlind
		if engine.pot != expectedPot {
			t.Errorf("Expected pot %d, got %d", expectedPot, engine.pot)
		}
	})

	t.Run("AllInScenario", func(t *testing.T) {
		engine := NewTexasHoldemEngine("holdem-game")

		// Add players with limited chips
		player1 := &Player{
			ID:   "1",
			Name: "Player 1",
			Data: map[string]interface{}{
				"chips": 50, // Can cover big blind but not much more
			},
		}
		player2 := &Player{
			ID:   "2",
			Name: "Player 2",
			Data: map[string]interface{}{
				"chips": 1000,
			},
		}

		engine.AddPlayer(player1)
		engine.AddPlayer(player2)
		engine.Start()

		// Find player with enough chips to go all-in
		currentPlayerID := engine.getCurrentActionPlayerID()

		allInAction := &GameAction{
			Type:     "texas_holdem_action",
			PlayerID: currentPlayerID,
			Data: map[string]interface{}{
				"action": "all_in",
			},
		}

		event, err := engine.ProcessAction(context.Background(), allInAction)
		if err != nil {
			t.Errorf("Unexpected error processing all-in: %v", err)
		}
		if event.Type != "player_all_in" {
			t.Errorf("Expected event type 'player_all_in', got %s", event.Type)
		}

		// Check that player is marked as all-in
		holdemPlayer := engine.getHoldemPlayer(currentPlayerID)
		if holdemPlayer != nil && !holdemPlayer.IsAllIn {
			t.Error("Player should be marked as all-in")
		}
		if holdemPlayer != nil && holdemPlayer.Chips != 0 {
			t.Error("All-in player should have 0 chips")
		}
	})
}
