package main

import (
	"flag"                // Package for command-line flag parsing
	"log"                 // Package for logging
	"sample/rate_limiter" // Importing the custom rate limiter package

	"github.com/gin-gonic/gin" // Gin web framework package
)

var rps = flag.Int("rps", 100, "request per second") // Defines a command-line flag for rate per second (RPS)

func init() {
	// Initialize log settings
	log.SetFlags(0)                  // Disable time, source file, and line number in logs
	log.SetPrefix("[GIN] ")          // Set a log prefix for all log output
	log.SetOutput(gin.DefaultWriter) // Set the log output to Gin's default writer (console by default)
}

func ginRun(rps int) {
	// Initialize rate limiter with the specified RPS value
	rate_limiter.InitRateLimiter(rps)

	// Create a new Gin router
	app := gin.Default()

	// Apply the rate limiter middleware to the Gin app
	app.Use(rate_limiter.LeakBucket())

	// Define a GET route at /rate to respond with a simple test message
	app.GET("/rate", func(ctx *gin.Context) {
		ctx.JSON(200, "rate limiting test") // Send a JSON response with HTTP status 200
	})

	// Log the current rate limit setting
	log.Printf("Current Rate Limit: %v requests/s", rps)

	// Run the Gin app on port 8081
	app.Run(":8081")
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Start the Gin app with the parsed RPS value
	ginRun(*rps)
}
