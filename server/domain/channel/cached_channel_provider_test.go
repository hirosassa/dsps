package channel

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/m3dev/dsps/server/domain"
	"github.com/m3dev/dsps/server/domain/mock"
	dspstesting "github.com/m3dev/dsps/server/testing"
)

func TestPositiveCache(t *testing.T) {
	clock := dspstesting.NewStubClock(t)
	cp := newCachedChannelProvider(dspstesting.ChannelProviderFunc(func(id domain.ChannelID) (domain.Channel, error) {
		return NewChannelByAtomYamls(t, id, []string{`{ regex: ".+", expire: "1s" }`}), nil
	}), clock)
	p := func(id domain.ChannelID) domain.Channel {
		c, err := cp.Get(id)
		if c == nil {
			dspstesting.IsError(t, domain.ErrInvalidChannel, err)
		} else {
			assert.NoError(t, err)
		}
		return c
	}

	test1 := p("test1")
	assert.NotNil(t, test1)
	assert.Same(t, test1, p("test1"))
	assert.Same(t, test1, p("test1"))

	// Enforce GC
	clock.Add(2 * time.Second)
	test2 := p("test2")
	assert.NotNil(t, test2)
	assert.NotSame(t, test1, p("test1")) // GC collected
	test1 = p("test1")
	assert.Same(t, test1, p("test1"))

	// Enforce GC
	for i := 0; i < 10; i++ {
		clock.Add(500 * time.Millisecond)
		assert.NotNil(t, p("test2")) // Keep touching to test2
		p(domain.ChannelID(fmt.Sprintf("test2-add-age-%d", i)))
	}
	assert.Same(t, test2, p("test2"))    // Still alive
	assert.NotSame(t, test1, p("test1")) // GC collected
}

func TestNegativeCache(t *testing.T) {
	clock := dspstesting.NewStubClock(t)
	notFoundCount := 0
	cp := newCachedChannelProvider(dspstesting.ChannelProviderFunc(func(id domain.ChannelID) (domain.Channel, error) {
		if strings.HasPrefix(string(id), "not-found-") {
			notFoundCount++
			return nil, domain.ErrInvalidChannel
		}
		return NewChannelByAtomYamls(t, id, []string{`{ regex: ".+", expire: "1s" }`}), nil
	}), clock)
	p := func(id domain.ChannelID) domain.Channel {
		c, err := cp.Get(id)
		if c == nil {
			dspstesting.IsError(t, domain.ErrInvalidChannel, err)
		} else {
			assert.NoError(t, err)
		}
		return c
	}

	assert.Nil(t, p("not-found-zero"))
	assert.Equal(t, 1, notFoundCount)
	assert.Nil(t, p("not-found-zero"))
	assert.Equal(t, 1, notFoundCount) // Cached

	// Enforce GC
	clock.Add(cachedChannelNegativeCacheExpire + 1*time.Microsecond)
	assert.NotNil(t, p("test-1"))
	assert.NotNil(t, p("test-2"))

	assert.Nil(t, p("not-found-zero"))
	assert.Equal(t, 2, notFoundCount) // Cache evicted
}

func TestCacheChannelError(t *testing.T) {
	clock := dspstesting.NewStubClock(t)
	called := 0
	errToReturn := errors.New("stub error")
	cp := newCachedChannelProvider(dspstesting.ChannelProviderFunc(func(id domain.ChannelID) (domain.Channel, error) {
		called++
		return nil, errToReturn
	}), clock)

	_, err := cp.Get("ch-1")
	assert.Equal(t, 1, called)
	dspstesting.IsError(t, errToReturn, err)

	_, err = cp.Get("ch-1")
	assert.Equal(t, 2, called) // Should not be cached
	dspstesting.IsError(t, errToReturn, err)
}

func TestCachedProviderValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inner := mock.NewMockChannelProvider(ctrl)
	inner.EXPECT().GetFileDescriptorPressure().Return(1234).Times(1)
	leeway := domain.Duration{Duration: 123 * time.Second}
	inner.EXPECT().JWTClockSkewLeewayMax().Return(leeway).Times(1)

	cp := newCachedChannelProvider(inner, domain.RealSystemClock)
	assert.Equal(t, 1234, cp.GetFileDescriptorPressure())
	assert.Equal(t, 1234, cp.GetFileDescriptorPressure())
	assert.Equal(t, leeway, cp.JWTClockSkewLeewayMax())
	assert.Equal(t, leeway, cp.JWTClockSkewLeewayMax())
}
