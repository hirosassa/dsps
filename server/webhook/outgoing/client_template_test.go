package outgoing

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/telemetry"
)

func newClientTemplateByConfig(t *testing.T, channelRegex string, json string) ClientTemplate {
	ctx := context.Background()
	yaml := strings.ReplaceAll(
		fmt.Sprintf(
			`{ channels: [ { regex: "%s", webhooks: [ %s ] } ] }`,
			channelRegex,
			json),
		"\t",
		"  ",
	)

	cfg, err := config.ParseConfig(ctx, config.Overrides{}, yaml)
	assert.NoError(t, err)

	telemetry := telemetry.NewEmptyTelemetry(t)
	tpl, err := NewClientTemplate(ctx, &cfg.Channels[0].Webhooks[0], telemetry, sentry.NewEmptySentry())
	assert.NoError(t, err)
	return tpl
}

func TestNoFilePressure(t *testing.T) {
	tpl := newClientTemplateByConfig(t, `.+`, `{ "url": "http://example.com", "connection": { "max": 1234 } }`)
	assert.Equal(t, 1234, tpl.GetFileDescriptorPressure())
}
