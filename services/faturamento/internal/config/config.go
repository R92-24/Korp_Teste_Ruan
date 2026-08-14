package config

import "os"

type Config struct {
	Port            string
	DatabaseURL     string
	EstoqueBaseURL  string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://korp:korp@localhost:5432/faturamento?sslmode=disable"),
		EstoqueBaseURL: getEnv("ESTOQUE_BASE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
