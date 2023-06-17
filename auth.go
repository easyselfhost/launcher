package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	authCookieName = "LAUNCHER_AUTH"
	tokenTtl       = 1800 * time.Second
)

func AuthPage(config *Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		redirect := c.QueryParam("redirect_uri")

		path := config.AuthConfig.AuthPath
		if redirect != "" {
			path = fmt.Sprintf("%v?redirect_uri=%v", config.AuthConfig.AuthPath, redirect)
		}
		return c.Render(http.StatusOK, "AuthTemplate", struct {
			Path string
		}{
			Path: path,
		})
	}
}

type TokenService interface {
	NewToken(c echo.Context) (string, error)
	Verify(c echo.Context, token string) (bool, error)
}

type MemTokenService struct {
	mu    sync.Mutex
	store map[string]time.Time
	Clock func() time.Time
}

func NewMemTokenService() *MemTokenService {
	clock := func() time.Time {
		return time.Now()
	}

	return &MemTokenService{
		store: make(map[string]time.Time),
		Clock: clock,
	}
}

func (mts *MemTokenService) NewToken(c echo.Context) (string, error) {
	tokenLen := 32

	token := make([]byte, tokenLen)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	tokenStr := base64.RawURLEncoding.EncodeToString(token)

	mts.mu.Lock()
	defer mts.mu.Unlock()

	mts.store[tokenStr] = mts.Clock().Add(tokenTtl)
	return tokenStr, nil
}

func (mts *MemTokenService) Verify(c echo.Context, token string) (bool, error) {
	mts.mu.Lock()
	defer mts.mu.Unlock()

	exp, ok := mts.store[token]

	if !ok {
		c.Logger().Debugf("Token %v is not found", token)
		return false, nil
	}

	if mts.Clock().After(exp) {
		c.Logger().Debugf("Token %v is not expired after %v", token, exp)
		delete(mts.store, token)
		return false, nil
	}

	mts.store[token] = mts.Clock().Add(tokenTtl)
	return true, nil
}

type Auth struct {
	tokens TokenService
	config *AuthConfig
}

func NewAuthFromConfig(config *Config) *Auth {
	return &Auth{
		tokens: NewMemTokenService(),
		config: &config.AuthConfig,
	}
}

func (a *Auth) redirectToLoginPage(c echo.Context) error {
	redirectUri := c.Request().URL.Path
	if redirectUri != "" && !strings.HasPrefix(redirectUri, "/") {
		redirectUri = "/" + redirectUri
	}
	t := a.config.AuthPath
	if redirectUri != "" {
		t += "?redirect_uri=" + redirectUri
	}
	return c.Redirect(http.StatusTemporaryRedirect, t)
}

func (a *Auth) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(authCookieName)
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					c.Logger().Debug("Cookie is not found")
					return a.redirectToLoginPage(c)
				}
				return err
			}

			ok, err := a.tokens.Verify(c, cookie.Value)
			if err != nil {
				return err
			}

			if !ok {
				return a.redirectToLoginPage(c)
			}

			return next(c)
		}
	}
}

func (a *Auth) Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		redirectUri := c.QueryParam("redirect_uri")

		if username != a.config.Username || password != a.config.Password {
			path := a.config.AuthPath
			if redirectUri != "" {
				path += "?redirect_uri=" + redirectUri
			}
			return c.Render(http.StatusUnauthorized, "AuthTemplate", struct {
				Path string
			}{
				Path: path,
			})
		}

		c.Logger().Infof("Logged in as %v", username)

		c.Set("Username", username)
		token, err := a.tokens.NewToken(c)
		if err != nil {
			return err
		}

		c.SetCookie(&http.Cookie{
			Name:     authCookieName,
			Path:     "/",
			Value:    token,
			HttpOnly: true,
		})
		return c.Redirect(http.StatusFound, redirectUri)
	}
}
