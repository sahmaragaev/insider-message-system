package loglevels

import "go.uber.org/zap/zapcore"

type LogLevel string

const (
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
	Fatal LogLevel = "fatal"
)

func (l LogLevel) String() string {
	return string(l)
}

func (l LogLevel) IsValid() bool {
	switch l {
	case Debug, Info, Warn, Error, Fatal:
		return true
	default:
		return false
	}
}

func (l LogLevel) ToZapLevel() zapcore.Level {
	switch l {
	case Debug:
		return zapcore.DebugLevel
	case Info:
		return zapcore.InfoLevel
	case Warn:
		return zapcore.WarnLevel
	case Error:
		return zapcore.ErrorLevel
	case Fatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func FromString(level string) LogLevel {
	ll := LogLevel(level)
	if ll.IsValid() {
		return ll
	}
	return Info
}
