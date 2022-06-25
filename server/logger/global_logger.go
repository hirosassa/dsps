package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"

	"github.com/m3dev/dsps/server/config"
)

var rootLoggerLock sync.RWMutex
var rootLogger *loggerImpl

func init() {
	initImpl()
}

func initImpl() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Sampling = nil                 // Disable sampling
	cfg.Level.SetLevel(zap.DebugLevel) // DSPS has own level filtering system

	zap, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(xerrors.Errorf("Failed to initialize zap logger: %w", err))
	}

	rootLoggerLock.Lock()
	defer rootLoggerLock.Unlock()
	rootLogger = &loggerImpl{
		ctx:    context.Background(),
		zap:    zap,
		filter: newDefaultFilter(),
	}
}

// InitLogger initializes Logger
func InitLogger(config *config.LoggingConfig) (*Filter, error) {
	filter, err := NewFilter(config.Category)
	if err != nil {
		return nil, err
	}

	fields := []zap.Field{}
	for key, value := range config.Attributes {
		fields = append(fields, zap.String(key, value))
	}

	rootLoggerLock.Lock()
	defer rootLoggerLock.Unlock()
	rootLogger = rootLogger.withFilter(filter).withAttributes(fields)
	return filter, nil
}
