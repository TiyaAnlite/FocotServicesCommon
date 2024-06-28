package echox

import (
	"context"
	"fmt"
	"github.com/TiyaAnlite/FocotServicesCommon/utils"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
	"unsafe"

	"github.com/labstack/echo/v4/middleware"
)

// ResponseWrapper 外层返回Data结构
type ResponseWrapper struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JwtClaims 默认JWT结构
type JwtClaims struct {
	UserId int `json:"u"`
	jwt.RegisteredClaims
}

// NormalResponse 常用返回
func NormalResponse(c echo.Context, data any) error {
	if span := trace.SpanFromContext(c.Request().Context()); span.IsRecording() {
		span.SetAttributes(attribute.Int("http.service_code", http.StatusOK))
	}
	return c.JSON(http.StatusOK, &ResponseWrapper{Code: http.StatusOK, Data: data})
}

// NormalEmptyResponse 空返回
func NormalEmptyResponse(c echo.Context) error {
	if span := trace.SpanFromContext(c.Request().Context()); span.IsRecording() {
		span.SetAttributes(attribute.Int("http.service_code", http.StatusOK))
	}
	return c.JSON(http.StatusOK, &ResponseWrapper{Code: http.StatusOK})
}

// NormalErrorResponse 错误返回
func NormalErrorResponse(c echo.Context, statusCode int, code int, message string) error {
	// echo的封装返回只负责注入业务返回码和业务成功状态，关于错误栈的记录应该由相应的ErrorLog进行
	if span := trace.SpanFromContext(c.Request().Context()); span.IsRecording() {
		span.SetAttributes(attribute.Int("http.service_code", code))
		if statusCode >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(statusCode))
		}
	}
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

func JwtEnabled(cfg *EchoConfig) bool {
	return cfg.JwtSecret != ""
}

func DefaultJwtConfigOld(cfg EchoConfig) middleware.JWTConfig {
	return middleware.JWTConfig{
		SigningKey:  []byte(cfg.JwtSecret),
		TokenLookup: "header:Authorization",
		ErrorHandlerWithContext: func(err error, c echo.Context) error {
			trace.SpanFromContext(c.Request().Context()).SetAttributes(attribute.Int("http.service_code", http.StatusUnauthorized))
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("未授权:%s\n", err.Error()))
		},
	}
}

func DefaultJwtConfig(cfg *EchoConfig) echojwt.Config {
	return echojwt.Config{
		SigningKey:    unsafe.Slice(unsafe.StringData(cfg.JwtSecret), len(cfg.JwtSecret)),
		TokenLookup:   "header:Authorization",
		SigningMethod: "HS256",
		ContextKey:    "user",
		ErrorHandler: func(c echo.Context, err error) error {
			utils.WarningWithCtx(c.Request().Context(), err.Error())
			return NormalErrorResponse(c, http.StatusUnauthorized, http.StatusUnauthorized, "未授权")
		},
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &JwtClaims{}
		},
	}
}

func JwtMiddlewareWithDefaultConfig(cfg *EchoConfig) echo.MiddlewareFunc {
	return echojwt.WithConfig(DefaultJwtConfig(cfg))
}

func JwtExpireTS(cfg *EchoConfig) int64 {
	return time.Now().Add(cfg.JwtExpire).Unix()
}

func MakeJwtToken(cfg *EchoConfig, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JwtSecret))
}

func Tracer() trace.Tracer {
	return otel.Tracer(otelecho.ScopeName)
}

// RootTracer 从上下文中生成追踪器Span，并自动注入返回头部
func RootTracer(c echo.Context, spanName string) (context.Context, trace.Span) {
	var attr []attribute.KeyValue
	if requestId := c.Request().Header.Get(echo.HeaderXRequestID); requestId != "" {
		attr = append(attr, attribute.String("service.api.request_id", requestId))
	}
	rootCtx := c.Request().Context()
	childCtx, span := Tracer().Start(rootCtx, spanName, trace.WithAttributes(attr...))
	respHeader := c.Response().Header()
	rootSpanCtx := trace.SpanFromContext(rootCtx).SpanContext()
	if respHeader.Get("X-Trace-Id") == "" && rootSpanCtx.HasTraceID() {
		respHeader.Set("X-Trace-Id", rootSpanCtx.TraceID().String())
	}
	return childCtx, span
}

// RootTracerNext 自动关闭并打开新追踪器Span上下文的快捷方法
func RootTracerNext(prev trace.Span, c echo.Context, spanName string) (context.Context, trace.Span) {
	prev.End()
	return RootTracer(c, spanName)
}
