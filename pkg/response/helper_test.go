package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := Success(map[string]string{"foo": "bar"})
	Send(c, resp)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.True(t, body["status"].(bool))
	assert.NotNil(t, body["data"])
}

func TestSendError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := InternalServerError("fail")
	SendError(c, resp)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.False(t, body["status"].(bool))
	assert.Equal(t, "Something went wrong", body["msg"])
}

func TestSendSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := Success(map[string]string{"foo": "bar"})
	SendSuccess(c, resp)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.True(t, body["status"].(bool))
	assert.NotNil(t, body["data"])
}
