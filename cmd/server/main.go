package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"em_tz_anvar/internal/config"
	"em_tz_anvar/internal/handler"
	"em_tz_anvar/internal/repository"
	"em_tz_anvar/internal/server"
	"em_tz_anvar/internal/service"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title Subscription Aggregator API
// @version 1.0
// @description REST API для агрегации данных об онлайн подписках пользователей

// @host localhost:9090
// @BasePath /api/v1

// @schemes http
func main() {
	//Flag parsing
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	//.env
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found, using environment variables")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	setupLogger(cfg.Logger.Level, cfg.Logger.Format)
	log.Info().Msg("Starting subscription aggregator service")

	//DB connection
	db, err := repository.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Connected to database")

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := server.NewServer(cfg, handlers)

	//Graceful shutdown
	go func() {
		if err := srv.Run(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
	log.Info().Msgf("Server started on port %d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error during server shutdown")
	}

	log.Info().Msg("Server stopped")
}

func setupLogger(level, format string) {
	//level logging
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	if format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
