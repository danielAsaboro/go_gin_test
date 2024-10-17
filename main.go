package main

import (
	"context"
	"flag"
	"log"
	"sample/rate_limiter"

	"github.com/gin-gonic/gin"
)

var rps = flag.Int("rps", 100, "request per second")

func init() {
	log.SetFlags(0)
	log.SetPrefix("[GIN] ")
	log.SetOutput(gin.DefaultWriter)
}

func main() {
	// Initialize tracer via OpenTelemetry setup
	ctx := context.Background()
	cleanup, err := setupOTelSDK(ctx)
	if err != nil {
		log.Fatalf("Failed to set up OTel SDK: %v", err)
	}
	defer cleanup(ctx)

	// Parse command-line flags
	flag.Parse()

	// Initialize rate limiter
	rate_limiter.InitRateLimiter(*rps)

	// Create a new Gin router
	app := gin.Default()

	// Apply rate limiter middleware
	app.Use(rate_limiter.LeakBucket())

	// Define routes
	app.GET("/rate", func(ctx *gin.Context) {
		ctx.JSON(200, "rate limiting test")
	})

	// Log the current rate limit setting
	log.Printf("Current Rate Limit: %v requests/s", *rps)

	// Run the Gin app on port 8081
	app.Run(":8081")
}
