package router_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"

	. "github.com/m3dev/dsps/server/http/router"
	. "github.com/m3dev/dsps/server/http/testing"
	"github.com/m3dev/dsps/server/http/utils"
)

func TestRouterHTTPMethods(t *testing.T) {
	r := httprouter.New()
	rt := NewRouter(func(r *http.Request, f func(context.Context)) { f(context.Background()) }, r, "/")
	rt.GET("/foo/bar", func(ctx context.Context, args HandlerArgs) {
		utils.SendJSON(ctx, args.W, 200, map[string]interface{}{"ok": "GET"})
	})
	rt.PUT("/foo/bar", func(ctx context.Context, args HandlerArgs) {
		utils.SendJSON(ctx, args.W, 200, map[string]interface{}{"ok": "PUT"})
	})
	rt.POST("/foo/bar", func(ctx context.Context, args HandlerArgs) {
		utils.SendJSON(ctx, args.W, 200, map[string]interface{}{"ok": "POST"})
	})
	rt.DELETE("/foo/bar", func(ctx context.Context, args HandlerArgs) {
		utils.SendJSON(ctx, args.W, 200, map[string]interface{}{"ok": "DELETE"})
	})
	server := httptest.NewServer(r)
	defer server.Close()

	for _, method := range []string{"GET", "PUT", "POST", "DELETE"} {
		res := DoHTTPRequest(t, method, server.URL+"/foo/bar", ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{"ok": method})
	}
}

func TestRouterPathPrefixAndGroup(t *testing.T) {
	r := httprouter.New()
	rt := NewRouter(func(r *http.Request, f func(context.Context)) {
		f(context.Background())
	}, r, "/prefix/bar", AsMiddlewareFunc(func(ctx context.Context, args MiddlewareArgs, next func(context.Context, MiddlewareArgs)) {
		args.W.Header().Add("middleware", "/prefix/bar")
		next(ctx, args)
	}))
	rt.NewGroup("/baz", AsMiddlewareFunc(func(ctx context.Context, args MiddlewareArgs, next func(context.Context, MiddlewareArgs)) {
		args.W.Header().Add("middleware", "/baz")
		next(ctx, args)
	})).GET("/test", func(ctx context.Context, args HandlerArgs) {
		utils.SendJSON(ctx, args.W, 200, map[string]interface{}{"ok": "/"})
	})
	server := httptest.NewServer(r)
	defer server.Close()

	res := DoHTTPRequest(t, "GET", server.URL+"/prefix/bar/baz/test", ``)
	assert.Equal(t, []string{"/prefix/bar", "/baz"}, res.Header.Values("middleware"))
	AssertResponseJSON(t, res, 200, map[string]interface{}{"ok": "/"})
}
