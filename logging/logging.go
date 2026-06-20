package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	Debug(msg string, attrs ...any)
	Info(msg string, attrs ...any)
	Warn(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}

type stdLogger struct {
	w     io.Writer
	level Level
	mtx   sync.Mutex
}

var (
	mtx           sync.RWMutex
	defaultLogger = NewStdLogger(os.Stderr, LevelWarn)
)

func NewStdLogger(w io.Writer, level Level) Logger {
	if w == nil {
		w = io.Discard
	}
	return &stdLogger{
		w:     w,
		level: level,
	}
}

func SetDefault(logger Logger) {
	if logger == nil {
		logger = NewStdLogger(io.Discard, LevelError)
	}

	mtx.Lock()
	defaultLogger = logger
	mtx.Unlock()
}

func Debug(msg string, attrs ...any) {
	log(LevelDebug, msg, attrs...)
}

func Info(msg string, attrs ...any) {
	log(LevelInfo, msg, attrs...)
}

func Warn(msg string, attrs ...any) {
	log(LevelWarn, msg, attrs...)
}

func Error(msg string, attrs ...any) {
	log(LevelError, msg, attrs...)
}

func log(level Level, msg string, attrs ...any) {
	mtx.RLock()
	logger := defaultLogger
	mtx.RUnlock()

	switch level {
	case LevelDebug:
		logger.Debug(msg, attrs...)
	case LevelInfo:
		logger.Info(msg, attrs...)
	case LevelWarn:
		logger.Warn(msg, attrs...)
	case LevelError:
		logger.Error(msg, attrs...)
	}
}

func (l *stdLogger) Debug(msg string, attrs ...any) {
	l.write(LevelDebug, "debug", msg, attrs...)
}

func (l *stdLogger) Info(msg string, attrs ...any) {
	l.write(LevelInfo, "info", msg, attrs...)
}

func (l *stdLogger) Warn(msg string, attrs ...any) {
	l.write(LevelWarn, "warning", msg, attrs...)
}

func (l *stdLogger) Error(msg string, attrs ...any) {
	l.write(LevelError, "error", msg, attrs...)
}

func (l *stdLogger) write(level Level, label string, msg string, attrs ...any) {
	if level < l.level {
		return
	}

	line := label + ": " + msg
	if formatted := formatAttrs(attrs); formatted != "" {
		line += " (" + formatted + ")"
	}

	l.mtx.Lock()
	fmt.Fprintln(l.w, line)
	l.mtx.Unlock()
}

func formatAttrs(attrs []any) string {
	if len(attrs) == 0 {
		return ""
	}

	parts := make([]string, 0, (len(attrs)+1)/2)
	for i := 0; i < len(attrs); i += 2 {
		key := fmt.Sprint(attrs[i])
		if i+1 >= len(attrs) {
			parts = append(parts, key)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", key, attrs[i+1]))
	}

	return strings.Join(parts, ", ")
}
