package echox

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/klog/v2"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type EchoConfig struct {
	Address              string        `env:"ADDRESS"`
	Port                 int           `env:"PORT" envDefault:"8080"`
	JwtSecret            string        `env:"JWT_SECRET"`
	JwtExpire            time.Duration `env:"JWT_EXPIRE" envDefault:"24h"`
	BodyLimit            string        `env:"BODY_LIMIT"`
	UseUptime            bool          `env:"ECHO_UPTIME"`
	UseHealthCheck       bool          `env:"ECHO_HEALTH" envDefault:"true"`
	UseTelemetry         bool          `env:"ECHO_TELEMETRY" envDefault:"true"`
	UseLogger            bool          `env:"ECHO_LOGGER" envDefault:"true"`
	UseRecover           bool          `env:"ECHO_RECOVER" envDefault:"true"`
	UseRequestIdInjector bool          `env:"ECHO_REQUEST_ID_INJECTOR" envDefault:"true"`
	UptimePath           string        `env:"ECHO_UPTIME_PATH" envDefault:"/uptime"`
	TelemetryHostName    string        `env:"ECHO_TELEMETRY_HOSTNAME" envDefault:"Echo.dev"`

	server *echo.Echo
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

// Run 启动服务端
func Run(cfg *EchoConfig, setupRoutes func(*echo.Echo)) {
	e := echo.New()
	e.HideBanner = true
	e.Validator = &defaultValidator{
		validator: validator.New(),
	}
	startAt := time.Now()

	if cfg.UseTelemetry {
		klog.Info("Echo telemetry on")
		e.Use(otelecho.Middleware(cfg.TelemetryHostName))
	}
	if cfg.UseLogger {
		e.Use(middleware.Logger())
	}
	if cfg.UseRecover {
		e.Use(PanicRecover)
	}
	if cfg.UseRequestIdInjector {
		e.Use(RequestIdInjector)
	}

	// 健康检查
	if cfg.UseHealthCheck {
		e.GET("/", func(c echo.Context) error {
			return NormalEmptyResponse(c)
		})
	}

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

	cfg.server = e

	ret := e.Start(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port))
	if !errors.Is(ret, http.ErrServerClosed) {
		e.Logger.Fatal(ret)
	}
}

// Shutdown 优雅关闭服务端
func Shutdown(cfg *EchoConfig) {
	if cfg.server != nil {
		_ = cfg.server.Shutdown(context.Background())
	}
}

// PanicRecover 提供对崩溃的处理
func PanicRecover(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				if r == http.ErrAbortHandler {
					panic(r)
				}
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				stack := make([]byte, 4<<10) // default stack length: 4kb
				length := runtime.Stack(stack, false)
				klog.Errorf("[PANIC RECOVER] %v\n%s", err, stack[:length])
				// msg := fmt.Sprintf("[PANIC RECOVER] %v\n%s", err, stack[:length])
				// c.Logger().Error(msg)
				_ = NormalErrorResponse(c, http.StatusInternalServerError, http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		return next(c)
	}
}

// RequestIdInjector 支持腾讯云API网关和Postman中RequestID的注入
func RequestIdInjector(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqHeaders := c.Request().Header
		requestId := reqHeaders.Get("X-Api-Requestid")
		if requestId != "" {
			// via api gateway
			reqHeaders.Set(echo.HeaderXRequestID, requestId) // X-Api-Requestid -> X-Request-Id
		} else if strings.HasPrefix(c.Request().UserAgent(), "PostmanRuntime") {
			// via postman tools
			requestId = fmt.Sprintf("Postman-%s", reqHeaders.Get("Postman-Token"))
			reqHeaders.Set(echo.HeaderXRequestID, requestId)
		}

		if requestId != "" {
			c.Response().Header().Add(echo.HeaderXRequestID, requestId)
			span := trace.SpanFromContext(c.Request().Context())
			span.SetAttributes(attribute.String("http.request_id", requestId))
		}
		return next(c)
	}
}
