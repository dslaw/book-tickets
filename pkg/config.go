package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	APIVersion         string
	DatabaseURL        string
	Port               string
	CacheURL           string
	TicketHoldDuration time.Duration
	TicketHoldPrefix   string
	SearchURL          string
	SearchUser         string
	SearchPassword     string
	SearchMaxResults   int32
	SearchEventsIndex  string
	SearchVenuesIndex  string
}

func NewConfig() (*Config, bool) {
	databaseURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, false
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		return nil, false
	}

	cacheURL, ok := os.LookupEnv("CACHE_URL")
	if !ok {
		return nil, false
	}

	ticketHoldPrefix, ok := os.LookupEnv("TICKET_HOLD_PREFIX")
	if !ok {
		return nil, false
	}

	ticketHoldDurationString, ok := os.LookupEnv("TICKET_HOLD_DURATION")
	if !ok {
		return nil, false
	}
	ticketHoldDuration, err := time.ParseDuration(ticketHoldDurationString)
	if err != nil {
		return nil, false
	}

	searchURL, ok := os.LookupEnv("SEARCH_URL")
	if !ok {
		return nil, false
	}

	searchUser, ok := os.LookupEnv("SEARCH_USER")
	if !ok {
		return nil, false
	}

	searchPassword, ok := os.LookupEnv("SEARCH_PASSWORD")
	if !ok {
		return nil, false
	}

	searchMaxResultsString, ok := os.LookupEnv("SEARCH_MAX_RESULTS")
	if !ok {
		return nil, false
	}
	searchMaxResultsI64, err := strconv.ParseInt(searchMaxResultsString, 10, 32)
	if err != nil {
		return nil, false
	}
	searchMaxResults := int32(searchMaxResultsI64)

	searchEventsIndex, ok := os.LookupEnv("SEARCH_EVENTS_INDEX")
	if !ok {
		return nil, false
	}

	searchVenuesIndex, ok := os.LookupEnv("SEARCH_VENUES_INDEX")
	if !ok {
		return nil, false
	}

	return &Config{
		APIVersion:         "",
		DatabaseURL:        databaseURL,
		Port:               port,
		CacheURL:           cacheURL,
		TicketHoldPrefix:   ticketHoldPrefix,
		TicketHoldDuration: ticketHoldDuration,
		SearchURL:          searchURL,
		SearchPassword:     searchPassword,
		SearchUser:         searchUser,
		SearchMaxResults:   searchMaxResults,
		SearchEventsIndex:  searchEventsIndex,
		SearchVenuesIndex:  searchVenuesIndex,
	}, true
}
