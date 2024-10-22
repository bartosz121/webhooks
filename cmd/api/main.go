package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bartosz121/webhooks-api/cmd/api/middleware"
	"github.com/bartosz121/webhooks-api/cmd/api/requestlog"
	"github.com/bartosz121/webhooks-api/cmd/api/resource/discord"
	"github.com/bartosz121/webhooks-api/cmd/api/resource/health"
	http_client_discord "github.com/bartosz121/webhooks-api/cmd/http_client/discord"
	"github.com/bartosz121/webhooks-api/config"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

// TODO: tests

//	@title		Webhooks API
//	@version	1.0

// @basePath	/v1
func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	val := validator.New(validator.WithRequiredStructEnabled())
	c := config.New()
	dc := http_client_discord.New(&c.Discord)
	l := NewLogger(c.Server.Debug)
	discordHandler := discord.NewHandler(l, val, &c.Api, dc)

	mux := http.NewServeMux()
	mux.Handle("GET /health", requestlog.NewHandler(health.Healthcheck, l))
	mux.Handle("POST /discord/gpw-scraper", requestlog.NewHandler(discordHandler.GpwScraperWebhook, l))

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", mux))

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Server.Port),
		Handler:      middleware.RequestId(middleware.ContentTypeJson(mux)),
		ReadTimeout:  c.Server.TimeoutRead,
		WriteTimeout: c.Server.TimeoutWrite,
		IdleTimeout:  c.Server.TimeoutIdle,
	}

	l.Info().Msg("Starting server " + s.Addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Fatal().Err(err).Msg("Server startup failed")
	}
}

func NewLogger(isDebug bool) *zerolog.Logger {
	logLevel := zerolog.InfoLevel

	if isDebug {
		logLevel = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &logger
}
