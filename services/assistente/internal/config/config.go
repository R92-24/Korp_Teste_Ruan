package config

import "os"

type Config struct {
	Port           string
	EstoqueBaseURL string

	// AnthropicAPIKey vazia é um estado esperado, não um erro: o serviço sobe,
	// responde às conferências apenas com as regras determinísticas e informa
	// que a análise por IA está indisponível. Isso permite que o projeto seja
	// executado por quem não tem uma chave.
	AnthropicAPIKey string
	AnthropicModel  string
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		EstoqueBaseURL:  getEnv("ESTOQUE_BASE_URL", "http://localhost:8081"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:  getEnv("ANTHROPIC_MODEL", "claude-opus-5"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
