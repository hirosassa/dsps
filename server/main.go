package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/domain"
	"github.com/m3dev/dsps/server/domain/channel"
	"github.com/m3dev/dsps/server/http"
	httplifecycle "github.com/m3dev/dsps/server/http/lifecycle"
	"github.com/m3dev/dsps/server/logger"
	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/storage"
	"github.com/m3dev/dsps/server/storage/deps"
	"github.com/m3dev/dsps/server/telemetry"
	"github.com/m3dev/dsps/server/unix"
)

// Git commit hash or tag
var buildVersion string

// Distribution name of the build
var buildDist string

// UNIX epoch (e.g. 1605633588)
var buildAt string

func main() {
	err := mainImpl(context.Background(), os.Args[1:], domain.RealSystemClock)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func mainImpl(ctx context.Context, args []string, clock domain.SystemClock) error {
	defer func() { logger.Of(ctx).Debugf(logger.CatServer, "Sever closed.") }()

	var (
		port   = flag.Int("port", 0, "Override http.port configuration item")
		listen = flag.String("listen", "", "Override http.listen configuration item")

		debug      = flag.Bool("debug", false, "Enable debug logs")
		dumpConfig = flag.Bool("dump-config", false, "Dump loaded configuration to stdout (for debug only)")
	)
	if err := flag.CommandLine.Parse(args); err != nil {
		return err
	}
	configFile := flag.Arg(0)
	configOverrides := config.Overrides{
		BuildVersion: buildVersion,
		BuildDist:    buildDist,
		BuildAt:      buildAt,
		Port:         *port,
		Listen:       *listen,
		Debug:        *debug,
	}

	config, err := config.LoadConfigFile(ctx, configFile, configOverrides)
	if err != nil {
		return err
	}
	if *dumpConfig {
		if err := config.DumpConfig(os.Stderr); err != nil {
			return err
		}
	}

	logFilter, err := logger.InitLogger(config.Logging)
	if err != nil {
		return err
	}

	sentry, err := sentry.NewSentry(config.Sentry)
	if err != nil {
		return err
	}
	defer sentry.Shutdown(ctx)

	telemetry, err := telemetry.InitTelemetry(config.Telemetry)
	if err != nil {
		return err
	}
	defer telemetry.Shutdown(ctx)

	channelProvider, err := channel.NewChannelProvider(ctx, &config, channel.ProviderDeps{
		Clock:     clock,
		Telemetry: telemetry,
		Sentry:    sentry,
	})
	if err != nil {
		return err
	}
	defer channelProvider.Shutdown(ctx)

	storage, err := storage.NewStorage(ctx, &config.Storages, clock, channelProvider, deps.StorageDeps{
		Telemetry: telemetry,
		Sentry:    sentry,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := storage.Shutdown(ctx); err != nil {
			logger.Of(ctx).WarnError(logger.CatStorage, "Failed to shutdown storage: %w", err)
		}
	}()

	unix.NotifyUlimit(ctx, unix.UlimitRequirement{
		NoFiles: channelProvider.GetFileDescriptorPressure() + storage.GetFileDescriptorPressure(),
	})

	http.StartServer(ctx, &http.ServerDependencies{
		Config:          &config,
		ChannelProvider: channelProvider,
		Storage:         storage,

		Telemetry:   telemetry,
		Sentry:      sentry,
		LogFilter:   logFilter,
		ServerClose: httplifecycle.NewServerClose(),
	})
	return nil
}
