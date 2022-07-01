package redis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/domain"
	. "github.com/m3dev/dsps/server/storage/deps/testing"
	"github.com/m3dev/dsps/server/storage/multiplex"
	. "github.com/m3dev/dsps/server/storage/redis"
	. "github.com/m3dev/dsps/server/storage/testing"
)

var storageCtor func(t *testing.T) StorageCtor = func(t *testing.T) StorageCtor {
	return func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
		cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, fmt.Sprintf(`storages: { myRedis: { redis: { singleNode: "%s", timeout: { connect: 500ms }, connection: { max: 10 } } } }`, GetRedisAddr(nil)))
		if err != nil {
			return nil, err
		}
		return NewRedisStorage(
			context.Background(),
			cfg.Storages["myRedis"].Redis,
			systemClock,
			channelProvider,
			EmptyDeps(t),
		)
	}
}

var storageMultiplexCtor func(t *testing.T) StorageCtor = func(t *testing.T) StorageCtor {
	return func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
		redis1, err := storageCtor(t)(ctx, systemClock, channelProvider)
		if err != nil {
			return nil, err
		}
		redis2, err := storageCtor(t)(ctx, systemClock, channelProvider)
		if err != nil {
			return nil, err
		}
		return multiplex.NewStorageMultiplexer(map[domain.StorageID]domain.Storage{
			"redis1": redis1,
			"redis2": redis2,
		})
	}
}

func TestCoreFunction(t *testing.T) {
	CoreFunctionTest(t, storageCtor(t))
}

func TestPubSub(t *testing.T) {
	PubSubTest(t, storageCtor(t))
}

func TestPubSubMultiplex(t *testing.T) {
	// Test with two duplicate storages.
	// It behaves as single storage because operations are idempotent.
	PubSubTest(t, storageMultiplexCtor(t))
}

func TestJwt(t *testing.T) {
	JwtTest(t, storageCtor(t))
}

func TestJwtMultiplex(t *testing.T) {
	// Test with two duplicate storages.
	// It behaves as single storage because operations are idempotent.
	JwtTest(t, storageMultiplexCtor(t))
}
