package poker

import "caslette-server/poker"

// Message types for poker operations
const (
	// Table Management
	MsgCreateTable     = "poker_create_table"
	MsgListTables      = "poker_list_tables"
	MsgJoinTable       = "poker_join_table"
	MsgLeaveTable      = "poker_leave_table"
	MsgTableUpdate     = "poker_table_update"
	MsgTableListUpdate = "poker_table_list_update"

	// Game Actions
	MsgStartHand       = "poker_start_hand"
	MsgPlayerAction    = "poker_player_action"
	MsgGameAction      = "poker_game_action"
	MsgGameState       = "poker_game_state"
	MsgGetGameState    = "poker_get_game_state"
	MsgHandResult      = "poker_hand_result"
	MsgGameStateUpdate = "poker_game_state_update"
	MsgHandComplete    = "poker_hand_complete"

	// Spectating
	MsgSpectateTable  = "poker_spectate_table"
	MsgStopSpectating = "poker_stop_spectating"

	// Admin Actions
	MsgPauseTable  = "poker_pause_table"
	MsgResumeTable = "poker_resume_table"

	// Notifications
	MsgPlayerDisconnected = "poker_player_disconnected"
	MsgPlayerConnected    = "poker_player_connected"

	// Responses
	MsgError   = "poker_error"
	MsgSuccess = "poker_success"
)

// PokerMessage represents poker-specific WebSocket messages
type PokerMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	UserID  uint        `json:"user_id,omitempty"`
	TableID uint        `json:"table_id,omitempty"`
}

// Request structures
type CreateTableRequest struct {
	Name        string  `json:"name"`
	GameType    string  `json:"game_type"`
	MaxPlayers  int     `json:"max_players"`
	MinBuyIn    int64   `json:"min_buy_in"`
	MaxBuyIn    int64   `json:"max_buy_in"`
	SmallBlind  int64   `json:"small_blind"`
	BigBlind    int64   `json:"big_blind"`
	RakePercent float64 `json:"rake_percent"`
	MaxRake     int64   `json:"max_rake"`
	IsPrivate   bool    `json:"is_private"`
	Password    string  `json:"password"`
}

type JoinTableRequest struct {
	TableID     uint   `json:"table_id"`
	BuyInAmount int64  `json:"buy_in_amount"`
	Password    string `json:"password"`
}

type LeaveTableRequest struct {
	TableID uint `json:"table_id"`
}

type PlayerActionRequest struct {
	TableID uint   `json:"table_id"`
	Action  string `json:"action"` // fold, check, call, bet, raise, all_in
	Amount  int64  `json:"amount"` // For bet/raise
}

type StartHandRequest struct {
	TableID uint `json:"table_id"`
}

// Response structures
type GameStateResponse struct {
	TableID        uint                  `json:"table_id"`
	HandNumber     int                   `json:"hand_number"`
	BettingRound   string                `json:"betting_round"`
	CommunityCards []poker.Card          `json:"community_cards"`
	Pot            int64                 `json:"pot"`
	CurrentBet     int64                 `json:"current_bet"`
	Players        []PlayerStateResponse `json:"players"`
	CurrentPlayer  *PlayerStateResponse  `json:"current_player"`
	DealerPosition int                   `json:"dealer_position"`
}

type PlayerStateResponse struct {
	UserID     uint         `json:"user_id"`
	Username   string       `json:"username"`
	SeatNumber int          `json:"seat_number"`
	ChipCount  int64        `json:"chip_count"`
	CurrentBet int64        `json:"current_bet"`
	TotalBet   int64        `json:"total_bet"`
	Status     string       `json:"status"`
	HoleCards  []poker.Card `json:"hole_cards,omitempty"` // Only sent to the player
	IsInHand   bool         `json:"is_in_hand"`
	HasActed   bool         `json:"has_acted"`
	IsAllIn    bool         `json:"is_all_in"`
	LastAction string       `json:"last_action"`
}

type TableListResponse struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	GameType       string `json:"game_type"`
	MaxPlayers     int    `json:"max_players"`
	MinBuyIn       int64  `json:"min_buy_in"`
	MaxBuyIn       int64  `json:"max_buy_in"`
	SmallBlind     int64  `json:"small_blind"`
	BigBlind       int64  `json:"big_blind"`
	Status         string `json:"status"`
	IsPrivate      bool   `json:"is_private"`
	CreatedBy      string `json:"created_by"`
	PlayerCount    int64  `json:"player_count"`
	AvailableSeats []int  `json:"available_seats"`
}

type HandResultResponse struct {
	TableID    uint           `json:"table_id"`
	HandNumber int            `json:"hand_number"`
	Winners    map[uint]int64 `json:"winners"`
	RakeAmount int64          `json:"rake_amount"`
	Pot        int64          `json:"pot"`
}

// Additional request types
type GameActionRequest struct {
	TableID uint   `json:"table_id"`
	Action  string `json:"action"` // fold, check, call, bet, raise, all_in
	Amount  int    `json:"amount"` // For bet/raise
}

type GetGameStateRequest struct {
	TableID uint `json:"table_id"`
}

type SpectateTableRequest struct {
	TableID uint `json:"table_id"`
}
