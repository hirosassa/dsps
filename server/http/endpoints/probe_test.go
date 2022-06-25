package endpoints_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	. "github.com/m3dev/dsps/server/http"
	. "github.com/m3dev/dsps/server/http/testing"
)

func TestProbeSuccess(t *testing.T) {
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "GET", baseURL+"/probe/liveness", "")
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 200, res.StatusCode)

		res = DoHTTPRequest(t, "GET", baseURL+"/probe/readiness", "")
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 200, res.StatusCode)
	})
}

func TestProbeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, _, _ := NewMockStorages(ctrl)

	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		storage.EXPECT().Liveness(gomock.Any()).Return(nil, errors.New("mock error"))
		res := DoHTTPRequest(t, "GET", baseURL+"/probe/liveness", "")
		assert.NoError(t, res.Body.Close())
		AssertInternalServerErrorResponse(t, res)

		storage.EXPECT().Readiness(gomock.Any()).Return(nil, errors.New("mock error"))
		res = DoHTTPRequest(t, "GET", baseURL+"/probe/readiness", "")
		assert.NoError(t, res.Body.Close())
		AssertInternalServerErrorResponse(t, res)
	})
}
