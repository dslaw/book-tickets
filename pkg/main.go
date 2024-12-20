package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	pkgApi "github.com/dslaw/book-tickets/pkg/api"
	"github.com/dslaw/book-tickets/pkg/cache"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/dslaw/book-tickets/pkg/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	config, ok := NewConfig()
	if !ok {
		slog.Error("Unable to read configuration from environment")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		slog.Error("Unable to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	ticketHoldClient, err := cache.NewTicketHoldClientFromURL(
		config.CacheURL,
		config.TicketHoldCacheKey,
	)
	if err != nil {
		slog.Error("Unable to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer ticketHoldClient.Close()

	venuesService := services.NewVenuesService(repos.NewVenuesRepo(pool))
	eventsService := services.NewEventsService(repos.NewEventsRepo(pool))
	ticketsService := services.NewTicketsService(
		repos.NewTicketsRepo(pool),
		ticketHoldClient,
		&services.Time{},
		config.TicketHoldDuration,
	)

	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("API", config.APIVersion))

	pkgApi.RegisterVenuesHandlers(api, venuesService)
	pkgApi.RegisterEventsHandlers(api, eventsService)
	pkgApi.RegisterTicketsHandlers(api, ticketsService)

	address := fmt.Sprintf(":%s", config.Port)
	slog.Info(fmt.Sprintf("Listening on %s", address))
	http.ListenAndServe(address, router)
}
