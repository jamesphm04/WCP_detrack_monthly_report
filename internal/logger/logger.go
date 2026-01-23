// internal/logger/logger.go
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a logger that outputs to terminal AND daily log files
func New() (*zap.Logger, error) {
	// Create log directory with today's date: logs/20260125/
	today := time.Now().Format("20060102")
	logDir := filepath.Join("logs", today)
	
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Log file path: logs/20260125/app.log
	logFile := filepath.Join(logDir, "app.log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Console encoder (colorful)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	
	// File encoder (JSON for easy parsing)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Write to both terminal and file
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zapcore.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	return logger, nil
}