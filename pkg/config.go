package main

import (
	"os"
	"time"
)

type Config struct {
	APIVersion         string
	DatabaseURL        string
	Port               string
	CacheURL           string
	TicketHoldCacheKey string
	TicketHoldDuration time.Duration
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

	ticketHoldCacheKey, ok := os.LookupEnv("TICKET_HOLD_CACHE_KEY")
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

	return &Config{
		APIVersion:         "",
		DatabaseURL:        databaseURL,
		Port:               port,
		CacheURL:           cacheURL,
		TicketHoldCacheKey: ticketHoldCacheKey,
		TicketHoldDuration: ticketHoldDuration,
	}, true
}
