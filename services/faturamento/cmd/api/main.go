package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"korp/faturamento/internal/config"
	"korp/faturamento/internal/estoqueclient"
	apphttp "korp/faturamento/internal/http"
	"korp/faturamento/internal/nota"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := connectWithRetry(ctx, cfg.DatabaseURL, 10, 2*time.Second)
	if err != nil {
		slog.Error("não foi possível conectar ao banco de dados", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	estoqueClient := estoqueclient.New(cfg.EstoqueBaseURL)
	repo := nota.NewRepository(pool)
	service := nota.NewService(repo, estoqueClient)
	router := apphttp.NewRouter(service)

	slog.Info("serviço de faturamento iniciado", "port", cfg.Port, "estoqueBaseURL", cfg.EstoqueBaseURL)
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("servidor encerrado com erro", "error", err)
		os.Exit(1)
	}
}

func connectWithRetry(ctx context.Context, databaseURL string, attempts int, delay time.Duration) (*pgxpool.Pool, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				lastErr = pingErr
				pool.Close()
			}
		} else {
			lastErr = err
		}
		slog.Warn("aguardando banco de dados ficar disponível", "tentativa", i+1, "erro", lastErr)
		time.Sleep(delay)
	}
	return nil, lastErr
}
