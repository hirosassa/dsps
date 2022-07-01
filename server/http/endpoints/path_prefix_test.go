package endpoints_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	. "github.com/m3dev/dsps/server/http"
	. "github.com/m3dev/dsps/server/http/testing"
)

func TestPathPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	WithServer(t, `{ logging: { category: "*": FATAL }, http: { pathPrefix: /foo/bar } }`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		assert.Regexp(t, "/foo/bar$", baseURL)

		res := DoHTTPRequest(t, "GET", baseURL+"/probe/liveness", "")
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 200, res.StatusCode)

		res = DoHTTPRequest(t, "GET", baseURL+"/probe/readiness", "")
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 200, res.StatusCode)
	})
}
