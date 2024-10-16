package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

func leakBucket() gin.HandlerFunc {
	prev := time.Now()
	return func(ctx *gin.Context) {
		now := limit.Take()
		log.Print(color.CyanString("%v", now.Sub(prev)))
		prev = now
	}
}

func ginRun(rps int) {
	limit = ratelimit.New(rps)

	app := gin.Default()
	app.Use(otelgin.Middleware("gin test project"))
	app.Use(leakBucket())

	app.GET("/rate", func(ctx *gin.Context) {
		ctx.JSON(200, "rate limiting test")
	})

	log.Printf(color.CyanString("Current Rate Limit: %v requests/s", rps))
	app.Run(":8081")
}

func main() {
	flag.Parse()

	// Set up OTel
	ctx := context.Background()
	shutdown, err := setupOTelSDK(ctx)
	if err != nil {
		log.Fatalf("Error setting up OTel SDK: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down OTel SDK: %v", err)
		}
	}()

	ginRun(*rps)
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// Set up propagator
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	// Set up trace provider
	traceExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(5*time.Second)),
	)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(3*time.Second))),
	)
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	// Return shutdown function
	return func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		return err
	}, nil
}
