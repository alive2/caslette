package poker

import (
	"caslette-server/models"
	pokerengine "caslette-server/poker"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// GameManager handles poker game logic and state
type GameManager struct {
	db           *gorm.DB
	tableLocks   map[uint]*sync.Mutex
	locksMutex   sync.RWMutex
	gameEngines  map[uint]*pokerengine.GameEngine
	enginesMutex sync.RWMutex
}

func NewGameManager(db *gorm.DB) *GameManager {
	return &GameManager{
		db:          db,
		tableLocks:  make(map[uint]*sync.Mutex),
		gameEngines: make(map[uint]*pokerengine.GameEngine),
	}
}

func (gm *GameManager) GetTableLock(tableID uint) *sync.Mutex {
	gm.locksMutex.Lock()
	defer gm.locksMutex.Unlock()

	if lock, exists := gm.tableLocks[tableID]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	gm.tableLocks[tableID] = lock
	return lock
}

func (gm *GameManager) InitializeTable(tableID uint) {
	gm.enginesMutex.Lock()
	defer gm.enginesMutex.Unlock()

	// Table will be initialized when first game starts
	gm.gameEngines[tableID] = nil
}

func (gm *GameManager) GetEngine(tableID uint) (*pokerengine.GameEngine, bool) {
	gm.enginesMutex.RLock()
	defer gm.enginesMutex.RUnlock()

	engine, exists := gm.gameEngines[tableID]
	return engine, exists && engine != nil
}

func (gm *GameManager) CheckAndStartGame(tableID uint) {
	tableLock := gm.GetTableLock(tableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// Get table and players
	var table models.PokerTable
	if err := gm.db.First(&table, tableID).Error; err != nil {
		log.Printf("Error fetching table %d: %v", tableID, err)
		return
	}

	var players []models.TablePlayer
	if err := gm.db.Preload("User").Where("table_id = ?", tableID).Find(&players).Error; err != nil {
		log.Printf("Error fetching players for table %d: %v", tableID, err)
		return
	}

	// Check if we have enough players to start
	if len(players) < 2 {
		return
	}

	// Update table status to playing
	if err := gm.db.Model(&table).Update("status", "playing").Error; err != nil {
		log.Printf("Error updating table status: %v", err)
		return
	}

	// Convert players slice to pointer slice
	playerPtrs := make([]*models.TablePlayer, len(players))
	for i := range players {
		playerPtrs[i] = &players[i]
	}

	// Create new game engine for this table
	engine := pokerengine.NewGameEngine(&table, playerPtrs)
	gm.enginesMutex.Lock()
	gm.gameEngines[tableID] = engine
	gm.enginesMutex.Unlock()

	// Start new hand
	gm.startNewHand(tableID, engine)
}

func (gm *GameManager) startNewHand(tableID uint, engine *pokerengine.GameEngine) {
	// Start new hand in engine
	if err := engine.StartNewHand(); err != nil {
		log.Printf("Error starting new hand for table %d: %v", tableID, err)
		return
	}

	// Create hand record in database with correct field names
	hand := models.GameHand{
		TableID:        tableID,
		HandNumber:     engine.HandNumber,
		DealerPosition: engine.DealerPosition,
		SmallBlindSeat: engine.SmallBlindPosition,
		BigBlindSeat:   engine.BigBlindPosition,
		Status:         "preflop",
		PotAmount:      engine.Pot,
		RakeAmount:     0,
		StartedAt:      time.Now(),
	}

	if err := gm.db.Create(&hand).Error; err != nil {
		log.Printf("Error creating hand record: %v", err)
		return
	}

	// Create player hands with correct field names
	for _, player := range engine.Players {
		if player.IsInHand {
			holeCardsJSON, _ := json.Marshal(player.HoleCards)
			ph := models.PlayerHand{
				GameHandID: hand.ID,
				UserID:     player.UserID,
				SeatNumber: player.SeatNumber,
				HoleCards:  string(holeCardsJSON),
				Status:     "active",
				TotalBet:   0,
			}

			if err := gm.db.Create(&ph).Error; err != nil {
				log.Printf("Error creating player hand: %v", err)
			}
		}
	}

	// Broadcast game start to all players at table
	gm.broadcastGameState(tableID, &hand, engine)
}

func (gm *GameManager) HandleGameAction(client Client, msg PokerMessage) {
	var req GameActionRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		client.SendError("Invalid game action request")
		return
	}

	tableLock := gm.GetTableLock(req.TableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// Get engine
	engine, exists := gm.GetEngine(req.TableID)
	if !exists {
		client.SendError("Game not found")
		return
	}

	// Find player position
	var playerPos int = -1
	for i, player := range engine.Players {
		if player.UserID == client.GetUserID() {
			playerPos = i
			break
		}
	}

	if playerPos == -1 {
		client.SendError("Player not found in game")
		return
	}

	// Process action through engine
	if err := engine.ProcessPlayerAction(playerPos, req.Action, int64(req.Amount)); err != nil {
		client.SendError(fmt.Sprintf("Invalid action: %v", err))
		return
	}

	// Record the action in database
	gm.recordPlayerAction(req.TableID, client.GetUserID(), req.Action, req.Amount)

	// Check if betting round is complete
	if engine.IsBettingRoundComplete() {
		gm.advanceGameState(req.TableID, engine)
	} else {
		// Broadcast updated game state
		gm.broadcastGameState(req.TableID, nil, engine)
	}
}

func (gm *GameManager) recordPlayerAction(tableID, userID uint, action string, amount int) {
	// Get current hand
	var hand models.GameHand
	if err := gm.db.Where("table_id = ? AND status != 'finished'", tableID).First(&hand).Error; err != nil {
		log.Printf("Error finding current hand: %v", err)
		return
	}

	// Record bet if applicable with correct field names
	if amount > 0 {
		bet := models.Bet{
			GameHandID:   hand.ID,
			UserID:       userID,
			BettingRound: hand.Status,
			Action:       action,
			Amount:       int64(amount),
			TotalBet:     int64(amount), // Simplified for now
			Sequence:     1,             // Simplified for now
			CreatedAt:    time.Now(),
		}

		if err := gm.db.Create(&bet).Error; err != nil {
			log.Printf("Error recording bet: %v", err)
		}
	}

	// Update player hand status if folded
	if action == "fold" {
		if err := gm.db.Model(&models.PlayerHand{}).
			Where("game_hand_id = ? AND user_id = ?", hand.ID, userID).
			Update("status", "folded").Error; err != nil {
			log.Printf("Error updating player hand status: %v", err)
		}
	}
}

func (gm *GameManager) advanceGameState(tableID uint, engine *pokerengine.GameEngine) {
	// Advance to next betting round or complete hand
	switch engine.BettingRound {
	case "preflop":
		engine.DealFlop()
		engine.BettingRound = "flop"
	case "flop":
		engine.DealTurn()
		engine.BettingRound = "turn"
	case "turn":
		engine.DealRiver()
		engine.BettingRound = "river"
	case "river":
		gm.completeHand(tableID, engine)
		return
	}

	// Broadcast updated game state
	gm.broadcastGameState(tableID, nil, engine)
}

func (gm *GameManager) completeHand(tableID uint, engine *pokerengine.GameEngine) {
	// Evaluate hands and determine winners
	engine.EvaluateHands()
	winners := engine.DetermineWinners()
	winnings, rakeAmount := engine.DistributePot()

	// Get current hand
	var hand models.GameHand
	if err := gm.db.Where("table_id = ? AND status != 'finished'", tableID).First(&hand).Error; err != nil {
		log.Printf("Error finding current hand: %v", err)
		return
	}

	// Update hand as completed with correct field names
	communityCardsJSON, _ := json.Marshal(engine.CommunityCards)
	updates := map[string]interface{}{
		"status":          "finished",
		"pot_amount":      engine.Pot,
		"rake_amount":     rakeAmount,
		"community_cards": string(communityCardsJSON),
		"finished_at":     time.Now(),
	}

	if len(winners) == 1 {
		updates["winner_user_id"] = winners[0].UserID
	}

	if err := gm.db.Model(&hand).Updates(updates).Error; err != nil {
		log.Printf("Error completing hand: %v", err)
		return
	}

	// Update player chip counts
	for playerID, winAmount := range winnings {
		if err := gm.db.Model(&models.TablePlayer{}).
			Where("table_id = ? AND user_id = ?", tableID, playerID).
			Update("chip_count", gorm.Expr("chip_count + ?", winAmount)).Error; err != nil {
			log.Printf("Error updating player chips: %v", err)
		}
	}

	// Broadcast hand completion
	gm.broadcastHandComplete(tableID, winners, winnings)

	// Start next hand after delay
	go func() {
		time.Sleep(5 * time.Second)
		gm.startNewHand(tableID, engine)
	}()
}

func (gm *GameManager) broadcastGameState(tableID uint, hand *models.GameHand, engine *pokerengine.GameEngine) {
	// This will be implemented when we connect the broadcasting system
	log.Printf("Broadcasting game state for table %d", tableID)
}

func (gm *GameManager) broadcastHandComplete(tableID uint, winners []*pokerengine.GamePlayer, winnings map[uint]int64) {
	// This will be implemented when we connect the broadcasting system
	log.Printf("Broadcasting hand completion for table %d", tableID)
}

func (gm *GameManager) GetGameState(tableID uint, userID uint) (*GameStateResponse, error) {
	// Get current hand
	var hand models.GameHand
	if err := gm.db.Preload("PlayerHands.User").Where("table_id = ? AND status != 'finished'", tableID).First(&hand).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &GameStateResponse{
				TableID:      tableID,
				BettingRound: "waiting",
			}, nil
		}
		return nil, err
	}

	// Get engine state
	engine, exists := gm.GetEngine(tableID)
	if !exists {
		return nil, fmt.Errorf("game engine not found")
	}

	currentPlayer := engine.GetCurrentPlayer()

	// Parse community cards from JSON string
	var communityCards []pokerengine.Card
	if hand.CommunityCards != "" {
		json.Unmarshal([]byte(hand.CommunityCards), &communityCards)
	}

	// Build response
	response := &GameStateResponse{
		TableID:        tableID,
		HandNumber:     hand.HandNumber,
		BettingRound:   hand.Status,
		DealerPosition: hand.DealerPosition,
		CommunityCards: communityCards,
		Pot:            hand.PotAmount,
		CurrentBet:     engine.CurrentBet,
		Players:        make([]PlayerStateResponse, 0),
	}

	// Set current player if exists
	if currentPlayer != nil {
		for _, ph := range hand.PlayerHands {
			if ph.UserID == currentPlayer.UserID {
				currentPlayerState := &PlayerStateResponse{
					UserID:     ph.UserID,
					Username:   ph.User.Username,
					SeatNumber: ph.SeatNumber,
					ChipCount:  currentPlayer.ChipCount,
					Status:     ph.Status,
				}
				response.CurrentPlayer = currentPlayerState
				break
			}
		}
	}

	// Add player states
	for _, ph := range hand.PlayerHands {
		// Find current chip count from engine
		var currentChips int64
		for _, player := range engine.Players {
			if player.UserID == ph.UserID {
				currentChips = player.ChipCount
				break
			}
		}

		playerState := PlayerStateResponse{
			UserID:     ph.UserID,
			Username:   ph.User.Username,
			SeatNumber: ph.SeatNumber,
			ChipCount:  currentChips,
			Status:     ph.Status,
		}

		// Add hole cards only for the requesting player
		if ph.UserID == userID {
			var holeCards []pokerengine.Card
			if ph.HoleCards != "" {
				json.Unmarshal([]byte(ph.HoleCards), &holeCards)
			}
			playerState.HoleCards = holeCards
		}

		response.Players = append(response.Players, playerState)
	}

	return response, nil
}
