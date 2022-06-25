package outgoing

import (
	"context"
	"net/http"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/domain"
	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/telemetry"
)

// ClientTemplate is factory object to make Client
type ClientTemplate interface {
	NewClient(tplEnv domain.TemplateStringEnv) (Client, error)
	Close()

	GetFileDescriptorPressure() int // estimated max usage of file descriptors
}

type clientTemplate struct {
	*config.OutgoingWebhookConfig

	h        *http.Client
	maxConns int

	telemetry *telemetry.Telemetry
	sentry    sentry.Sentry
}

// NewClientTemplate returns ClientTemplate instalce
func NewClientTemplate(ctx context.Context, cfg *config.OutgoingWebhookConfig, telemetry *telemetry.Telemetry, sentry sentry.Sentry) (ClientTemplate, error) {
	return &clientTemplate{
		OutgoingWebhookConfig: cfg,

		h:        newHTTPClientFor(ctx, cfg),
		maxConns: *cfg.Connection.Max,

		telemetry: telemetry,
		sentry:    sentry,
	}, nil
}

func (tpl *clientTemplate) NewClient(tplEnv domain.TemplateStringEnv) (Client, error) {
	return newClientImpl(tpl, tplEnv)
}

func (tpl *clientTemplate) Close() {
	tpl.h.CloseIdleConnections()
}

func (tpl *clientTemplate) GetFileDescriptorPressure() int {
	return tpl.maxConns
}
