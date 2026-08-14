package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"korp/estoque/internal/config"
	apphttp "korp/estoque/internal/http"
	"korp/estoque/internal/produto"
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

	repo := produto.NewRepository(pool)
	service := produto.NewService(repo)
	router := apphttp.NewRouter(service)

	slog.Info("serviço de estoque iniciado", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("servidor encerrado com erro", "error", err)
		os.Exit(1)
	}
}

// connectWithRetry tenta conectar ao Postgres repetidamente: no docker
// compose o container do banco pode ainda estar subindo quando este serviço
// inicia, então uma falha imediata de conexão não deve derrubar o processo.
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
