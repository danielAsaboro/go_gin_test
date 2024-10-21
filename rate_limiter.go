package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func leakBucket() gin.HandlerFunc {
	prev := time.Now()
	return func(ctx *gin.Context) {
		now := limit.Take()
		log.Print(color.CyanString("%v", now.Sub(prev)))
		prev = now
	}
}

func handleRate(ctx *gin.Context) {
	// Initialize OpenTelemetry tracer
	tracer := otel.Tracer("rate-limiter-service")
	_, span := tracer.Start(ctx, "rate-limiting")
	defer span.End()

	// Extract request data
	method := ctx.Request.Method
	scheme := "http"
	statusCode := 200
	host := ctx.Request.Host
	port := ctx.Request.URL.Port()
	if port == "" {
		port = "8081"
	}
	startTime := time.Now()

	// Set span status
	span.SetStatus(codes.Ok, "")

	// Add semantic conventions for HTTP request attributes
	span.SetAttributes(
		semconv.HTTPMethodKey.String(method),
		semconv.HTTPSchemeKey.String(scheme),
		semconv.HTTPStatusCodeKey.Int(statusCode),
		semconv.HTTPTargetKey.String(ctx.Request.URL.Path),
		semconv.HTTPURLKey.String(ctx.Request.URL.String()),
		semconv.HTTPHostKey.String(host),
		semconv.NetHostPortKey.String(port),
		semconv.HTTPUserAgentKey.String(ctx.Request.UserAgent()),
		semconv.HTTPRequestContentLengthKey.Int64(ctx.Request.ContentLength),
		semconv.NetPeerIPKey.String(ctx.ClientIP()),
	)

	// Custom attributes for OpenTelemetry span
	span.SetAttributes(
		attribute.String("created_at", startTime.Format(time.RFC3339Nano)),
		attribute.Float64("duration_ns", float64(time.Since(startTime).Nanoseconds())),
		attribute.String("referer", ctx.Request.Referer()),
		attribute.String("request_type", "Incoming"),
		attribute.String("sdk_type", "go-gin"),
		attribute.String("service_version", "1.0.0"), // Add your service version here
		attribute.StringSlice("tags", []string{}),
	)

	// Additional request-specific attributes (path params, query params, headers, body)
	span.SetAttributes(
		attribute.String("query_params", ctx.Request.URL.RawQuery),
		attribute.String("request_body", "{}"), // Assuming empty body for demonstration
		attribute.String("request_headers", fmt.Sprintf("%v", ctx.Request.Header)),
		attribute.String("response_body", "{}"),
		attribute.String("response_headers", "{}"),
	)

	// Respond with JSON
	ctx.JSON(statusCode, "rate limiting test")
}
