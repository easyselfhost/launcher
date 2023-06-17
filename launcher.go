package main

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type Cache struct {
	mu     sync.Mutex
	target *Target
	exp    time.Time
}

func (c *Cache) Get() (Target, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.target == nil || time.Now().After(c.exp) {
		return Target{}, false
	}

	return *c.target, true
}

func (c *Cache) Set(target *Target, exp time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.target = target
	c.exp = exp
}

func (c *Cache) ClearIfSame(t *Target) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if *t.URL == *c.target.URL {
		c.target = nil
	}
}

type Launcher struct {
	cache       *Cache
	lmu         sync.Mutex
	client      InstanceClient
	alarmClient AlarmClient
	cacheTtl    time.Duration
	launchWait  time.Duration
}

func NewLauncerFromConfig(c *Config) *Launcher {
	cli := NewInstanceClientFromConfig(c)
	alarmClient := NewAlarmClientFromConfig(c)

	return &Launcher{
		cache:       new(Cache),
		cacheTtl:    c.CacheTtl,
		client:      cli,
		alarmClient: alarmClient,
		launchWait:  c.WaitTime,
	}
}

func (l *Launcher) HandleProxyError() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t, tok := c.Get("target").(*Target)

			err := next(c)
			if err == nil {
				return nil
			}

			he, ok := err.(*echo.HTTPError)
			if !ok {
				return err
			}

			if he.Code != http.StatusBadGateway && he.Code != http.StatusServiceUnavailable {
				return err
			}

			c.Logger().Debugf("Got HTTP error: %v", he)

			if !tok {
				c.Logger().Debug("Cannot find target in context")
				return err
			}

			ok, err = l.client.CheckInstance(t.Instance)
			if err != nil {
				return err
			}
			if !ok {
				l.cache.ClearIfSame(t)
				return c.Redirect(http.StatusTemporaryRedirect, c.Request().URL.Path)
			}

			return c.Render(http.StatusServiceUnavailable, "RefreshTemplate", RefreshPageParams{
				Title:   "503 Service Unavailable",
				Message: "Sorry, the server may not be ready right now.",
				Emoji:   "😔",
				Seconds: 21,
			})
		}
	}
}

func (l *Launcher) Launch() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var t Target

			t, ok := l.cache.Get()
			if ok {
				c.Set("target", &t)
				return next(c)
			}

			t, created, err := l.getInstance(c)
			if err != nil {
				return err
			}

			err = l.alarmClient.AutoTerminate(c, &t)
			if err != nil {
				return err
			}
			c.Set("target", &t)

			if created && l.launchWait > 0 {
				return c.Render(http.StatusOK, "RefreshTemplate", RefreshPageParams{
					Title:   "Launcher is on it",
					Message: "The server is being initialized.",
					Emoji:   "🤔",
					Seconds: int(l.launchWait / time.Second),
				})
			}

			return next(c)
		}
	}
}

func (l *Launcher) getInstance(c echo.Context) (Target, bool, error) {
	l.lmu.Lock()
	defer l.lmu.Unlock()

	if t, ok := l.cache.Get(); ok {
		return t, false, nil
	}

	t, err := l.client.FindInstance(c)

	if err != nil && !errors.Is(err, errNotFound) {
		return Target{}, false, err
	}

	if err == nil {
		l.cache.Set(t, time.Now().Add(l.cacheTtl))
		return *t, false, nil
	}

	t, err = l.client.LaunchInstance(c)
	if err != nil {
		return Target{}, false, err
	}

	l.cache.Set(t, time.Now().Add(l.cacheTtl))
	return *t, true, nil
}
