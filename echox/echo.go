package echox

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

type EchoConfig struct {
	Address    string        `env:"ADDRESS"`
	Port       int           `env:"PORT" envDefault:"8080"`
	JwtSecret  string        `env:"JWT_SECRET"`
	JwtExpire  time.Duration `env:"JWT_EXPIRE" envDefault:"24h"`
	BodyLimit  string        `env:"BODY_LIMIT"`
	UseUptime  bool          `env:"ECHO_UPTIME"`
	UptimePath string        `env:"ECHO_UPTIME_PATH" envDefault:"/uptime"`
}

type defaultValidator struct {
	validator *validator.Validate
}

func (v *defaultValidator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func Run(cfg EchoConfig, setupRoutes func(*echo.Echo)) {
	e := echo.New()
	e.HideBanner = true
	e.Validator = &defaultValidator{
		validator: validator.New(),
	}
	startAt := time.Now()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 健康检查
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if cfg.UseUptime {
		e.GET(cfg.UptimePath, func(c echo.Context) error {
			return c.JSON(http.StatusOK, echo.Map{
				"startAt": startAt,
				"uptime":  time.Since(startAt).String(),
			})
		})
	}

	if cfg.BodyLimit != "" {
		e.Use(middleware.BodyLimit(cfg.BodyLimit))
	}

	if setupRoutes != nil {
		setupRoutes(e)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)))
}
