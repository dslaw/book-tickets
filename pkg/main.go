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
	"github.com/dslaw/book-tickets/pkg/search"
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
		config.TicketHoldPrefix,
	)
	if err != nil {
		slog.Error("Unable to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer ticketHoldClient.Close()

	searchClient, err := search.NewSearchClient(
		config.SearchURL,
		config.SearchUser,
		config.SearchPassword,
		config.SearchEventsIndex,
		config.SearchVenuesIndex,
	)
	if err != nil {
		slog.Error("Unable to create an OpenSearch client", "error", err)
		os.Exit(1)
	}

	venuesService := services.NewVenuesService(repos.NewVenuesRepo(pool))
	eventsService := services.NewEventsService(repos.NewEventsRepo(pool))
	ticketsService := services.NewTicketsService(
		repos.NewTicketsRepo(pool),
		ticketHoldClient,
		config.TicketHoldDuration,
	)
	searchService, err := services.NewSearchService(searchClient, config.SearchMaxResults)
	if err != nil {
		slog.Error("Unable to create a search service", "error", err)
		os.Exit(1)
	}

	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("API", config.APIVersion))

	pkgApi.RegisterVenuesHandlers(api, venuesService)
	pkgApi.RegisterEventsHandlers(api, eventsService)
	pkgApi.RegisterTicketsHandlers(api, ticketsService)
	pkgApi.RegisterSearchHandlers(api, searchService)

	address := fmt.Sprintf(":%s", config.Port)
	slog.Info(fmt.Sprintf("Listening on %s", address))
	http.ListenAndServe(address, router)
}
