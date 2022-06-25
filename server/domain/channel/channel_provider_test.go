package channel

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/domain"
	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/telemetry"
	dspstesting "github.com/m3dev/dsps/server/testing"
)

func TestProvider(t *testing.T) {
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, `channels: [ { regex: "test.+", expire: "1s" } ]`)
	assert.NoError(t, err)
	clock := dspstesting.NewStubClock(t)
	cp, err := NewChannelProvider(context.Background(), &cfg, ProviderDeps{
		Clock:     clock,
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	})
	assert.NoError(t, err)

	test1, err := cp.Get("test1")
	assert.NoError(t, err)
	test1Again, err := cp.Get("test1")
	assert.NoError(t, err)
	assert.NotNil(t, test1)
	assert.Same(t, test1, test1Again)

	notfound, err := cp.Get("not-found")
	assert.Nil(t, notfound)
	dspstesting.IsError(t, domain.ErrInvalidChannel, err)
}

func TestJWTClockSkewLeewayMax(t *testing.T) {
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, strings.ReplaceAll(`channels: [ 
		{ regex: "test.+", expire: "1s", jwt: { iss: [ "https://issuer.example.com/issuer-url" ], keys: { none: [] }, clockSkewLeeway: 5m } },
		{ regex: "test.+", expire: "1s", jwt: { iss: [ "https://issuer.example.com/issuer-url" ], keys: { none: [] }, clockSkewLeeway: 15m } } 
	]`, "\t", "  "))
	assert.NoError(t, err)

	cp, err := NewChannelProvider(context.Background(), &cfg, ProviderDeps{
		Clock:     dspstesting.NewStubClock(t),
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	})
	assert.NoError(t, err)
	assert.Equal(t, domain.Duration{Duration: 15 * time.Minute}, cp.JWTClockSkewLeewayMax())
}

func TestProviderWithInvalidDeps(t *testing.T) {
	_, err := NewChannelProvider(context.Background(), nil, ProviderDeps{})
	assert.Regexp(t, `invalid ProviderDeps`, err.Error())
}

func TestValidateProviderDeps(t *testing.T) {
	valid := ProviderDeps{
		Clock:     domain.RealSystemClock,
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	}
	assert.NoError(t, valid.validateProviderDeps())

	invalid := valid
	invalid.Clock = nil
	assert.Regexp(t, `invalid ProviderDeps: Clock should not be nil`, invalid.validateProviderDeps())

	invalid = valid
	invalid.Telemetry = nil
	assert.Regexp(t, `invalid ProviderDeps: Telemetry should not be nil`, invalid.validateProviderDeps())

	invalid = valid
	invalid.Sentry = nil
	assert.Regexp(t, `invalid ProviderDeps: Sentry should not be nil`, invalid.validateProviderDeps())
}
