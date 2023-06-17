package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	elog "github.com/labstack/gommon/log"
)

func main() {
	config := ConfigFromEnv()

	e := echo.New()

	if config.LogLevel == "DEBUG" {
		e.Logger.SetLevel(elog.DEBUG)
		e.Logger.Debug("Log level set to DEBUG")
	} else {
		e.Logger.SetLevel(elog.INFO)
	}

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	var auth *Auth

	e.Renderer = NewPageRenderer()
	if config.AuthConfig.EnableAuth {
		auth = NewAuthFromConfig(&config)

		e.GET(config.AuthConfig.AuthPath, AuthPage(&config))
		e.POST(config.AuthConfig.AuthPath, auth.Login())
	}

	launcher := NewLauncerFromConfig(&config)

	pg := e.Group("")
	if config.AuthConfig.EnableAuth {
		pg.Use(auth.Authenticate())
	}
	pg.Use(launcher.Launch())
	pg.Use(launcher.HandleProxyError())
	pg.Use(Proxy)

	e.Logger.Fatal(e.Start(config.Addr()))
}
