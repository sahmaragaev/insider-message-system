package response

import (
	"net/http"

	"go.uber.org/zap/zapcore"
)

const NonAssigned = "N/A"

type (
	Response struct {
		Code int
		Body *Body
		Log  *Log
	}

	Body struct {
		Status bool   `json:"status"`
		Msg    string `json:"msg,omitempty"`
		Data   any    `json:"data,omitempty"`
	}

	Log struct {
		Level zapcore.Level `json:"level" swaggertype:"string" example:"info"`
		Msg   string        `json:"msg"`
		Type  LogType       `json:"type"`
	}

	LogType string
)

const (
	API LogType = "API"
)

func New(code int, body *Body, log *Log) *Response {
	return &Response{
		Code: code,
		Body: body,
		Log:  log,
	}
}

func InternalServerError(msg string) *Response {
	return &Response{
		Code: http.StatusInternalServerError,
		Body: &Body{
			Status: false,
			Msg:    "Something went wrong",
		},
		Log: &Log{
			Level: zapcore.ErrorLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func BadRequest(msg string) *Response {
	return &Response{
		Code: http.StatusBadRequest,
		Body: &Body{
			Status: false,
			Msg:    msg,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func Unauthorized() *Response {
	msg := "Unauthorized"
	return &Response{
		Code: http.StatusUnauthorized,
		Body: &Body{
			Status: false,
			Msg:    msg,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func Forbidden() *Response {
	msg := "Origin not allowed"
	return &Response{
		Code: http.StatusForbidden,
		Body: &Body{
			Status: false,
			Msg:    msg,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func Success(data any) *Response {
	msg := "Request processed successfully"
	return &Response{
		Code: http.StatusOK,
		Body: &Body{
			Status: true,
			Msg:    msg,
			Data:   data,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func Conflict() *Response {
	msg := "Data already exists"
	return &Response{
		Code: http.StatusConflict,
		Body: &Body{
			Status: false,
			Msg:    msg,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func NotFound(msg string) *Response {
	return &Response{
		Code: http.StatusNotFound,
		Body: &Body{
			Status: false,
			Msg:    msg,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func NoContent() *Response {
	return &Response{
		Code: http.StatusNoContent,
		Body: &Body{
			Status: false,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   "no content",
			Type:  API,
		},
	}
}

func SchedulerAlreadyRunning() *Response {
	msg := "Scheduler is already running"
	return &Response{
		Code: http.StatusConflict,
		Body: &Body{
			Status: false,
			Msg:    msg,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func SchedulerNotRunning() *Response {
	msg := "Scheduler is not running"
	return &Response{
		Code: http.StatusBadRequest,
		Body: &Body{
			Status: false,
			Msg:    msg,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func SchedulerStarted() *Response {
	msg := "Scheduler started successfully"
	return &Response{
		Code: http.StatusOK,
		Body: &Body{
			Status: true,
			Msg:    msg,
			Data: map[string]string{
				"status":  "running",
				"message": msg,
			},
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func SchedulerStopped() *Response {
	msg := "Scheduler stopped successfully"
	return &Response{
		Code: http.StatusOK,
		Body: &Body{
			Status: true,
			Msg:    msg,
			Data: map[string]string{
				"status":  "stopped",
				"message": msg,
			},
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func MessageCreated(data any) *Response {
	msg := "Message created successfully"
	return &Response{
		Code: http.StatusCreated,
		Body: &Body{
			Status: true,
			Msg:    msg,
			Data:   data,
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg,
			Type:  API,
		},
	}
}

func ValidationError(details string) *Response {
	msg := "Invalid request parameters"
	return &Response{
		Code: http.StatusBadRequest,
		Body: &Body{
			Status: false,
			Msg:    msg,
			Data: map[string]string{
				"details": details,
			},
		},
		Log: &Log{
			Level: zapcore.InfoLevel,
			Msg:   msg + ": " + details,
			Type:  API,
		},
	}
}
