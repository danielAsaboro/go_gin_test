package rate_limiter

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
)

var limit ratelimit.Limiter // Global rate limiter instance

// InitRateLimiter sets up the rate limiter with the given requests per second (RPS).
func InitRateLimiter(rps int) {
	limit = ratelimit.New(rps) // Initialize the limiter with the desired RPS
}

// LeakBucket is a middleware for rate limiting using a leaky bucket algorithm.
func LeakBucket() gin.HandlerFunc {
	prev := time.Now() // Tracks the time of the previous request
	return func(ctx *gin.Context) {
		now := limit.Take()                              // Apply rate limiting and wait if needed
		log.Print(color.CyanString("%v", now.Sub(prev))) // Log the time since the last request
		prev = now

		// Proceed to the next handler
		ctx.Next()
	}
}
