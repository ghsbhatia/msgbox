package ctxlog

import (
	"context"

	"github.com/go-kit/kit/log"
)

// CtxLogger satisfies log.Logger and is designed to be constructed into a context
// via New. Components can retrieve it from the context via From, and use the
// Log method to append keyvals. The entrypoint (e.g. an http.Handler) should
// Flush at the end of its lifecycle (e.g. the end of the request).
type CtxLogger struct {
	keyvals []interface{}
}

// Log implements go-kit/kit/log.Logger, buffering keyvals into memory.
// Use Keyvals at the end of the lifecycle to log to a concrete logger.
// Log always succeeds.
func (logger *CtxLogger) Log(keyvals ...interface{}) error {
	logger.keyvals = append(logger.keyvals, keyvals...)
	return nil
}

// Keyvals returns the keyvals that have been collected by the Logger, and can
// be passed to the Log method of a concrete logger.
func (logger *CtxLogger) Keyvals() []interface{} {
	return logger.keyvals
}

// NewLogger is a helper function to create a new Logger, log the initial set of
// keyvals, inject it into a context, and return everything, all in one motion.
func NewLogger(ctx context.Context, initialKeyvals ...interface{}) (context.Context, *CtxLogger) {
	logger := &CtxLogger{}
	logger.Log(initialKeyvals...)
	return context.WithValue(ctx, keyvalue, logger), logger
}

// Logger is a helper function to extract a Logger from a context.
// If no ctxlog.Logger exists in the context, a NopLogger is returned.
func Logger(ctx context.Context) log.Logger {
	v := ctx.Value(keyvalue)
	if v == nil {
		return log.NewNopLogger()
	}
	logger, ok := v.(*CtxLogger)
	if !ok {
		return log.NewNopLogger()
	}
	return logger
}

type keytype struct{}

var keyvalue = keytype{}
