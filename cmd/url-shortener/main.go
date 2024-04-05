package main

import (
	"flag"
	"fmt"
	"golang_project/internal/config"
	"net/http"

	"golang_project/internal/http-server/handlers/redirect"
	"golang_project/internal/http-server/handlers/url/save"
	mwLogger "golang_project/internal/http-server/middleware/logger"
	"golang_project/internal/storage/memory"

	"golang_project/internal/storage/postgres"

	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal   = "local"
	envDev     = "dev"
	configPath = "config/config.yaml"
	PgCfgPath  = "config/pgConfig.yaml"
)

type strg interface {
	GetURL(alias string) (string, error)
	SaveURL(urlToSave string, alias string) error
}

var usePostgres = flag.Bool("d", false, "Use PostgreSQL for storing URLs")

func main() {
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := setupLogger(cfg.Env)

	log.Info("starting golang project", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	var dbase strg

	if *usePostgres {
		pg_cfg := config.LoadPostgresCfg(PgCfgPath)

		dbase_temp, err := postgres.New(pg_cfg)
		if err != nil {
			fmt.Println("failed to init storage", slog.String("error", string(err.Error())))
			os.Exit(1)
		}
		dbase = strg(dbase_temp)

	} else {
		dbase_temp, err := memory.New()
		if err != nil {
			fmt.Println("failed to init storage", slog.String("error", string(err.Error())))
			os.Exit(1)
		}
		dbase = strg(dbase_temp)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/", save.New(log, dbase))
	router.Get("/{alias}", redirect.New(log, dbase))

	log.Info("starting server", slog.String("address", cfg.Address))

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

	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
