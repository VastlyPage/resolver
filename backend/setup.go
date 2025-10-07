package backend

import (
	"io"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Logger.SetOutput(io.Discard)

	e.Use(middleware.Gzip(), middleware.Recover())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("startTime", time.Now())
			return next(c)
		}
	})

	return e
}
