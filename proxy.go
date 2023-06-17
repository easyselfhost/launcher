package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Proxy(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		target, ok := ctx.Get("target").(*Target)

		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "proxy target not set")
		}

		ctx.Logger().Debugf("proxy to address %s", target.URL.Host)
		targets := []*middleware.ProxyTarget{
			{
				URL: target.URL,
			},
		}

		proxyFunc := middleware.Proxy(middleware.NewRoundRobinBalancer(targets))

		return proxyFunc(next)(ctx)
	}
}
