package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerContextKeyType int

const loggerContextKey = loggerContextKeyType(1)

// Of returns or creates Logger instance associated to the context.
func Of(ctx context.Context) Logger {
	return of(ctx)
}

func of(ctx context.Context) *loggerImpl {
	if logger, ok := ctx.Value(loggerContextKey).(*loggerImpl); ok {
		return logger
	}

	rootLoggerLock.RLock()
	defer rootLoggerLock.RUnlock()
	return rootLogger
}

// WithAttributes returns builder to create child context that holds child logger
func WithAttributes(ctx context.Context) ContextLoggerBuilder {
	return &contextLoggerBuilder{
		ctx:        ctx,
		baseLogger: of(ctx),
		fields:     make([]zap.Field, 0, 16),
	}
}

// PinLoggerContext creates context-specific logger along with the new context.
// By default, logger.Of(Context) returns same logger instance if no configuration changed.
// Some operation (e.g. global handlers) looks top-most context that logger belongs to.
func PinLoggerContext(ctx context.Context) context.Context {
	logger := of(ctx).copy()
	ctx = context.WithValue(ctx, loggerContextKey, logger)
	logger.ctx = ctx
	return ctx
}

// ContextLoggerBuilder is an interface to create child context that holds child logger
type ContextLoggerBuilder interface {
	Build() context.Context

	WithStr(key string, value string) ContextLoggerBuilder
	WithInt(key string, value int) ContextLoggerBuilder
	WithInt64(key string, value int64) ContextLoggerBuilder
	WithBool(key string, value bool) ContextLoggerBuilder
}

type contextLoggerBuilder struct {
	ctx context.Context

	baseLogger *loggerImpl
	fields     []zap.Field
}

func (b *contextLoggerBuilder) Build() context.Context {
	newLogger := b.baseLogger.withAttributes(b.fields)
	ctx := context.WithValue(b.ctx, loggerContextKey, newLogger)
	newLogger.ctx = ctx
	return ctx
}

func (b *contextLoggerBuilder) WithStr(key string, value string) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.String(key, value))
	return b
}

func (b *contextLoggerBuilder) WithInt(key string, value int) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Int(key, value))
	return b
}

func (b *contextLoggerBuilder) WithInt64(key string, value int64) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Int64(key, value))
	return b
}

func (b *contextLoggerBuilder) WithBool(key string, value bool) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Bool(key, value))
	return b
}
