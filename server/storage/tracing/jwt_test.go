package tracing_test

import (
	"context"
	"testing"
	"time"

	"github.com/m3dev/dsps/server/domain"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestJWTTrace(t *testing.T) {
	tr := testTracing(t, func(s domain.Storage) {
		assert.NoError(t, s.AsJwtStorage().RevokeJwt(context.Background(), domain.JwtExp(time.Now()), domain.JwtJti("jti-value")))
		_, err := s.AsJwtStorage().IsRevokedJwt(context.Background(), "jti-value")
		assert.NoError(t, err)
	})
	tr.OT.AssertSpanBy(trace.SpanKindInternal, "DSPS storage RevokeJwt", map[string]interface{}{
		"dsps.storage.id": "test",
		"jwt.jti":         "jti-value",
	})
	tr.OT.AssertSpanBy(trace.SpanKindInternal, "DSPS storage IsRevokedJwt", map[string]interface{}{
		"dsps.storage.id": "test",
		"jwt.jti":         "jti-value",
	})
	tr.OT.AssertSpanBy(trace.SpanKindInternal, "DSPS storage Shutdown", map[string]interface{}{
		"dsps.storage.id": "test",
	})
}
