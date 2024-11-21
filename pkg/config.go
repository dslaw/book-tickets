package main

import "os"

type Config struct {
	DatabaseURL string
	Port        string
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

	return &Config{
		DatabaseURL: databaseURL,
		Port:        port,
	}, true
}
