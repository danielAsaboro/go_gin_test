package main

import (
	"context"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	// add an otel trace here
	tracer := otel.Tracer("rate-limiter-service")
	_, span := tracer.Start(context.Background(), "rate-limiting")
	defer span.End()

	// Optionally, add some attributes to the span
	span.SetAttributes(attribute.String("method", ctx.Request.Method))
	span.SetAttributes(attribute.String("path", ctx.FullPath()))

	ctx.JSON(200, "rate limiting test")
}
