package echox

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4/middleware"
)

// ResponseWrapper 外层返回Data结构
type ResponseWrapper struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NormalResponse 常用返回
func NormalResponse(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, &ResponseWrapper{Code: http.StatusOK, Data: data})
}

// NormalEmptyResponse 空返回
func NormalEmptyResponse(c echo.Context) error {
	return c.JSON(http.StatusOK, &ResponseWrapper{Code: http.StatusOK})
}

// NormalErrorResponse 错误返回
func NormalErrorResponse(c echo.Context, statusCode int, code int, message string) error {
	return c.JSON(statusCode, &ResponseWrapper{Code: code, Message: message})
}

func CheckInput[T any](c echo.Context) (*T, error) {
	var input T
	if err := c.Bind(&input); err != nil {
		return nil, err
	}
	if err := c.Validate(&input); err != nil {
		return nil, err
	}
	return &input, nil
}

func JwtEnabled(cfg EchoConfig) bool {
	return cfg.JwtSecret != ""
}

func DefaultJwtConfig(cfg EchoConfig) middleware.JWTConfig {
	return middleware.JWTConfig{
		SigningKey:  []byte(cfg.JwtSecret),
		TokenLookup: "header:Authorization",
		ErrorHandler: func(err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("未授权:%s\n", err.Error()))
		},
	}
}

func JwtExpireTS(cfg EchoConfig) int64 {
	return time.Now().Add(cfg.JwtExpire).Unix()
}

func MakeJwtToken(cfg EchoConfig, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JwtSecret))
}

func Tracer() trace.Tracer {
	return otel.Tracer(otelecho.ScopeName)
}

// RootTracer 从上下文中生成追踪器Span，并自动注入返回头部
func RootTracer(c echo.Context, spanName string) (context.Context, trace.Span) {
	childCtx, span := Tracer().Start(c.Request().Context(), spanName, trace.WithAttributes(attribute.String("service.api.request_id", c.Request().Header.Get(echo.HeaderXRequestID))))
	respHeader := c.Response().Header()
	spanCtx := span.SpanContext()
	if respHeader.Get("X-Trace-Id") == "" && spanCtx.HasTraceID() {
		respHeader.Set("X-Trace-Id", spanCtx.TraceID().String())
	}
	return childCtx, span
}
