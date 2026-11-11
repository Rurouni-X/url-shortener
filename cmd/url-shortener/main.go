package main

import (
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	getall "url-shortener/internal/http-server/handlers/url/getAll"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoadConfig()
	log := setupLogger(cfg.Env)
	log.Info("starting url-shortener")
	log.Debug("debug messages are enabled")

	db, err := database.Connect(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer db.Close()

	stor := database.NewStorage(db)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, stor))
	router.Get("/", getall.New(log, stor))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	//TODO handler redirect 1.53 minute
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
