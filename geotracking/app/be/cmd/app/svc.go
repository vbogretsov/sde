package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karagenc/fj4echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	"github.com/urfave/cli/v3"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"

	"app/pkg/health"
	"app/pkg/track"
)

func runHTTPServer(ctx context.Context, e *echo.Echo, address string) error {
	f := &fasthttp.Server{
		Handler: fasthttpadaptor.NewFastHTTPHandler(e),
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM)
	defer stop()

	var result error
	go func () {
		if err := f.ListenAndServe(address); err != nil {
			slog.Error("server terminated with error", "err", err)
			result = err
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := f.Shutdown(); err != nil {
		slog.Error("failed to gracefully stop HTTP server", "err", err)
		return err
	}
	
	return result
}

func runSvc(ctx context.Context, _ *cli.Command) error {
	l := slog.New(slog.NewJSONHandler(os.Stdout, cfg.Logging.ToOptions()))
	slog.SetDefault(l)

	nparts, err := getNumPartitions(track.TopicLocation)
	if err != nil {
		return err
	}
	slog.Info("topic partitions", "num", nparts)

	journal := true
	mongo, err := newMongo(cfg.Mongo, &writeconcern.WriteConcern{W: "majority", Journal: &journal})
	if err != nil {
		return err
	}
	defer mongo.Disconnect(ctx)

	producer := newProducer(cfg.Kafka)
	defer producer.Close()

	if err := track.RegisterSchemas(cfg.SchemaRegistry.URL); err != nil {
		return err
	}

	e := echo.New()
	e.JSONSerializer = fj4echo.New()
	e.Validator = NewValidator()

	e.Use(slogecho.New(l))
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	trackAPI := track.New(cfg, producer, mongo)

	e.GET("/healthz", health.Healthz)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet},
	}))

	r := e.Group("/api/track")
	r.POST("/routes", trackAPI.RoutePost)
	r.POST("/locations", trackAPI.LocationPost)
	r.GET("/locations", trackAPI.CourierList)

	slog.Info(
		"Starting Tracker HTTP service",
		slog.Any("host", cfg.Service.Listen.Host),
		slog.Any("port", cfg.Service.Listen.Port),
	)

	slog.Debug("Application configuration", "cfg", cfg)
	return runHTTPServer(ctx, e, cfg.Service.Listen.Address())
}
