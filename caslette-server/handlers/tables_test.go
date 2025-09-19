package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func createMockTableHandler() *SecureTableHandler {
	return &SecureTableHandler{
		db:           nil, // No actual DB for unit tests
		validator:    NewSecurityValidator(),
		tableManager: nil, // No actual table manager for unit tests
	}
}

func TestSecureTableHandler_CreateTable_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	tableData := map[string]interface{}{
		"name":        "Test Table",
		"game_type":   "texas_holdem",
		"max_players": 6,
		"buy_in":      100,
		"small_blind": 1,
		"big_blind":   2,
	}

	jsonData, _ := json.Marshal(tableData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTable(c)

	// Should accept valid request (may fail at DB level but security validation should pass)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusInternalServerError)
}

func TestSecureTableHandler_CreateTable_SQLInjectionInName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	tableData := map[string]interface{}{
		"name":        "'; DROP TABLE tables; --",
		"game_type":   "texas_holdem",
		"max_players": 6,
		"buy_in":      100,
		"small_blind": 1,
		"big_blind":   2,
	}

	jsonData, _ := json.Marshal(tableData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_CreateTable_XSSInName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	tableData := map[string]interface{}{
		"name":        "<script>alert('xss')</script>",
		"game_type":   "texas_holdem",
		"max_players": 6,
		"buy_in":      100,
		"small_blind": 1,
		"big_blind":   2,
	}

	jsonData, _ := json.Marshal(tableData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_CreateTable_InvalidMaxPlayers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	testCases := []struct {
		maxPlayers int
		shouldFail bool
		desc       string
	}{
		{-1, true, "negative players"},
		{0, true, "zero players"},
		{1, true, "too few players"},
		{4, false, "valid player count"},
		{10, false, "valid player count"},
		{100, true, "too many players"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tableData := map[string]interface{}{
				"name":        "Test Table",
				"game_type":   "texas_holdem",
				"max_players": tc.maxPlayers,
				"buy_in":      100,
				"small_blind": 1,
				"big_blind":   2,
			}

			jsonData, _ := json.Marshal(tableData)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.CreateTable(c)

			if tc.shouldFail {
				assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject %s", tc.desc)
			} else {
				assert.True(t, w.Code != http.StatusBadRequest, "Should accept %s", tc.desc)
			}
		})
	}
}

func TestSecureTableHandler_CreateTable_InvalidBlinds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	// Test invalid blind structure (big blind should be > small blind)
	tableData := map[string]interface{}{
		"name":        "Test Table",
		"game_type":   "texas_holdem",
		"max_players": 6,
		"buy_in":      100,
		"small_blind": 5, // Invalid: small blind > big blind
		"big_blind":   2,
	}

	jsonData, _ := json.Marshal(tableData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_GetTable_ValidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "123"}}

	req, _ := http.NewRequest("GET", "/tables/123", nil)
	c.Request = req

	handler.GetTable(c)

	// May return not found but should not be bad request
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestSecureTableHandler_GetTable_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "invalid"}}

	req, _ := http.NewRequest("GET", "/tables/invalid", nil)
	c.Request = req

	handler.GetTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_GetTable_SQLInjectionID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "1; DROP TABLE tables; --"}}

	req, _ := http.NewRequest("GET", "/tables/1; DROP TABLE tables; --", nil)
	c.Request = req

	handler.GetTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_JoinTable_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	joinData := map[string]interface{}{
		"table_id": 123,
		"seat":     1,
		"buy_in":   100,
	}

	jsonData, _ := json.Marshal(joinData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables/join", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.JoinTable(c)

	// May fail at business logic but security validation should pass
	assert.True(t, w.Code != http.StatusBadRequest || w.Code == http.StatusUnauthorized)
}

func TestSecureTableHandler_JoinTable_InvalidSeat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	joinData := map[string]interface{}{
		"table_id": 123,
		"seat":     -1, // Invalid seat number
		"buy_in":   100,
	}

	jsonData, _ := json.Marshal(joinData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables/join", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.JoinTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_JoinTable_InvalidBuyIn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	joinData := map[string]interface{}{
		"table_id": 123,
		"seat":     1,
		"buy_in":   -50, // Invalid negative buy-in
	}

	jsonData, _ := json.Marshal(joinData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables/join", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.JoinTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_CreateTable_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	tableData := map[string]interface{}{
		"name":        "", // Empty name
		"game_type":   "texas_holdem",
		"max_players": 6,
		"buy_in":      100,
		"small_blind": 1,
		"big_blind":   2,
	}

	jsonData, _ := json.Marshal(tableData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_CreateTable_InvalidGameType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	tableData := map[string]interface{}{
		"name":        "Test Table",
		"game_type":   "invalid_game_type",
		"max_players": 6,
		"buy_in":      100,
		"small_blind": 1,
		"big_blind":   2,
	}

	jsonData, _ := json.Marshal(tableData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureTableHandler_CreateTable_NegativeBuyIn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockTableHandler()

	tableData := map[string]interface{}{
		"name":        "Test Table",
		"game_type":   "texas_holdem",
		"max_players": 6,
		"buy_in":      -100, // Invalid negative buy-in
		"small_blind": 1,
		"big_blind":   2,
	}

	jsonData, _ := json.Marshal(tableData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTable(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
