package main

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
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
	// Extract request data
	method := ctx.Request.Method
	statusCode := 200
	startTime := time.Now()

	// Log request details (optional, for debugging)
	log.Printf("Request received: method=%s, path=%s, time=%s", method, ctx.Request.URL.Path, startTime)

	// Respond with JSON
	ctx.JSON(statusCode, "rate limiting test")
}
