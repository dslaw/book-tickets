package main

import (
	"fmt"
	"log/slog"
	"os"

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

	fmt.Println(config.DatabaseURL)
}
