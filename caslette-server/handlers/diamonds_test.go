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

func createMockDiamondHandler() *SecureDiamondHandler {
	return &SecureDiamondHandler{
		db:        nil, // No actual DB for unit tests
		validator: NewSecurityValidator(),
	}
}

func TestSecureDiamondHandler_GetUserDiamonds_ValidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "userId", Value: "123"}}

	req, _ := http.NewRequest("GET", "/diamonds/user/123", nil)
	c.Request = req

	handler.GetUserDiamonds(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "user_id")
	assert.Contains(t, response, "diamonds")
}

func TestSecureDiamondHandler_GetUserDiamonds_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "userId", Value: "invalid"}}

	req, _ := http.NewRequest("GET", "/diamonds/user/invalid", nil)
	c.Request = req

	handler.GetUserDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureDiamondHandler_GetUserDiamonds_SQLInjection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "userId", Value: "1; DROP TABLE diamonds; --"}}

	req, _ := http.NewRequest("GET", "/diamonds/user/1; DROP TABLE diamonds; --", nil)
	c.Request = req

	handler.GetUserDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureDiamondHandler_AddDiamonds_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  100,
		"reason":  "Purchase bonus",
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/credit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.AddDiamonds(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "user_id")
	assert.Contains(t, response, "amount")
}

func TestSecureDiamondHandler_AddDiamonds_InvalidAmount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	// Test negative amount
	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  -50, // Invalid negative amount
		"reason":  "Test",
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/credit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.AddDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureDiamondHandler_AddDiamonds_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	requestData := map[string]interface{}{
		// user_id missing
		"amount": 100,
		"reason": "Test",
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/credit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.AddDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureDiamondHandler_AddDiamonds_MaliciousReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  100,
		"reason":  "<script>alert('xss')</script>", // XSS attempt
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/credit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.AddDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureDiamondHandler_DeductDiamonds_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  50,
		"reason":  "Purchase item",
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/debit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.DeductDiamonds(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
}

func TestSecureDiamondHandler_DeductDiamonds_SQLInjectionReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  50,
		"reason":  "'; DELETE FROM diamonds WHERE user_id = 123; --",
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/debit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.DeductDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureDiamondHandler_DeductDiamonds_ExcessiveAmount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  999999999, // Excessive amount
		"reason":  "Test",
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/debit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.DeductDiamonds(c)

	// Should accept the request (validation happens at business logic level)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecureDiamondHandler_GetAllTransactions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/diamonds/transactions", nil)
	c.Request = req

	handler.GetAllTransactions(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "transactions")
	assert.Contains(t, response, "total")

	// Should return empty array for placeholder implementation
	transactions := response["transactions"].([]interface{})
	assert.Equal(t, 0, len(transactions))
	assert.Equal(t, float64(0), response["total"])
}

func TestSecureDiamondHandler_ZeroAmountValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  0, // Zero amount should be rejected
		"reason":  "Test",
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/credit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.AddDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecureDiamondHandler_LongReasonValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := createMockDiamondHandler()

	// Create a reason string longer than 200 characters
	longReason := string(make([]byte, 250))
	for i := range longReason {
		longReason = longReason[:i] + "a" + longReason[i+1:]
	}

	requestData := map[string]interface{}{
		"user_id": 123,
		"amount":  100,
		"reason":  longReason,
	}

	jsonData, _ := json.Marshal(requestData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/diamonds/credit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.AddDiamonds(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
