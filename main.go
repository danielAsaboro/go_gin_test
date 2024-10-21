package main

import (
	"flag"
	"log"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/ratelimit"
)

var (
	limit ratelimit.Limiter
	rps   = flag.Int("rps", 100, "request per second")
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("[GIN] ")
	log.SetOutput(gin.DefaultWriter)
}

func ginRun(rps int) {
	limit = ratelimit.New(rps)
	app := gin.Default()
	app.Use(otelgin.Middleware("gin test project"))
	app.Use(leakBucket())
	app.GET("/rate", handleRate)

	log.Printf(color.CyanString("Current Rate Limit: %v requests/s", rps))
	app.Run(":8081")
}

func main() {
	flag.Parse()

	ginRun(*rps)
}
