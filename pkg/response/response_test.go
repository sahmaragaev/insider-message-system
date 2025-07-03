package response

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInternalServerError(t *testing.T) {
	resp := InternalServerError("fail")
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "Something went wrong", resp.Body.Msg)
}

func TestBadRequest(t *testing.T) {
	resp := BadRequest("bad req")
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "bad req", resp.Body.Msg)
}

func TestUnauthorized(t *testing.T) {
	resp := Unauthorized()
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "Unauthorized", resp.Body.Msg)
}

func TestForbidden(t *testing.T) {
	resp := Forbidden()
	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "Origin not allowed", resp.Body.Msg)
}

func TestSuccess(t *testing.T) {
	data := map[string]string{"foo": "bar"}
	resp := Success(data)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, resp.Body.Status)
	assert.Equal(t, "Request processed successfully", resp.Body.Msg)
	assert.Equal(t, data, resp.Body.Data)
}

func TestConflict(t *testing.T) {
	resp := Conflict()
	assert.Equal(t, http.StatusConflict, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "Data already exists", resp.Body.Msg)
}

func TestNotFound(t *testing.T) {
	resp := NotFound("not found")
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "not found", resp.Body.Msg)
}

func TestNoContent(t *testing.T) {
	resp := NoContent()
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.False(t, resp.Body.Status)
}

func TestSchedulerAlreadyRunning(t *testing.T) {
	resp := SchedulerAlreadyRunning()
	assert.Equal(t, http.StatusConflict, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "Scheduler is already running", resp.Body.Msg)
}

func TestSchedulerNotRunning(t *testing.T) {
	resp := SchedulerNotRunning()
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "Scheduler is not running", resp.Body.Msg)
}

func TestSchedulerStarted(t *testing.T) {
	resp := SchedulerStarted()
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, resp.Body.Status)
	assert.Equal(t, "Scheduler started successfully", resp.Body.Msg)
	assert.NotNil(t, resp.Body.Data)
}

func TestSchedulerStopped(t *testing.T) {
	resp := SchedulerStopped()
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, resp.Body.Status)
	assert.Equal(t, "Scheduler stopped successfully", resp.Body.Msg)
	assert.NotNil(t, resp.Body.Data)
}

func TestMessageCreated(t *testing.T) {
	data := map[string]string{"foo": "bar"}
	resp := MessageCreated(data)
	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.True(t, resp.Body.Status)
	assert.Equal(t, "Message created successfully", resp.Body.Msg)
	assert.Equal(t, data, resp.Body.Data)
}

func TestValidationError(t *testing.T) {
	resp := ValidationError("bad param")
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.False(t, resp.Body.Status)
	assert.Equal(t, "Invalid request parameters", resp.Body.Msg)
	assert.NotNil(t, resp.Body.Data)
}
