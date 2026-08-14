package main

import (
	"log/slog"

	"korp/assistente/internal/config"
	"korp/assistente/internal/conferencia"
	"korp/assistente/internal/estoqueclient"
	apphttp "korp/assistente/internal/http"
)

func main() {
	cfg := config.Load()

	estoqueClient := estoqueclient.New(cfg.EstoqueBaseURL)
	analisador := conferencia.NovoAnalisador(cfg.AnthropicAPIKey, cfg.AnthropicModel)
	service := conferencia.NewService(estoqueClient, analisador)
	router := apphttp.NewRouter(service, analisador.Disponivel())

	slog.Info("serviço assistente iniciado",
		"port", cfg.Port,
		"estoqueBaseURL", cfg.EstoqueBaseURL,
		"iaDisponivel", analisador.Disponivel(),
	)
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("servidor encerrado com erro", "error", err)
	}
}
