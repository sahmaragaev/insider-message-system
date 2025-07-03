package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	err := NewError("CODE", "msg", http.StatusBadRequest)
	assert.Equal(t, "CODE", err.Code)
	assert.Equal(t, "msg", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status)
}

func TestNewErrorWithDetails(t *testing.T) {
	err := NewErrorWithDetails("CODE", "msg", "details", http.StatusInternalServerError)
	assert.Equal(t, "CODE", err.Code)
	assert.Equal(t, "msg", err.Message)
	assert.Equal(t, "details", err.Details)
	assert.Equal(t, http.StatusInternalServerError, err.Status)
}

func TestWrapError(t *testing.T) {
	base := errors.New("base")
	err := WrapError(base, "CODE", "msg", http.StatusConflict)
	assert.Equal(t, "CODE", err.Code)
	assert.Equal(t, "msg", err.Message)
	assert.Equal(t, http.StatusConflict, err.Status)
	assert.Equal(t, "base", err.Details)
}

func TestError_ErrorMethod(t *testing.T) {
	err := NewError("CODE", "msg", http.StatusBadRequest)
	assert.Contains(t, err.Error(), "[CODE] msg")
}
