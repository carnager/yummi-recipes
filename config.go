package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
)

type Config struct {
	Port        string
	DataDir     string
	Secret      string
	OpenAIKey   string
	CreateAdmin string
}

func loadConfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.Port, "port", envOr("YUMMI_PORT", ":8080"), "Server port")
	flag.StringVar(&cfg.DataDir, "data", envOr("YUMMI_DATA_DIR", "./data"), "Data directory")
	flag.StringVar(&cfg.Secret, "secret", envOr("YUMMI_SECRET", ""), "Session secret")
	flag.StringVar(&cfg.OpenAIKey, "openai-key", envOr("YUMMI_OPENAI_KEY", ""), "OpenAI API key")
	flag.StringVar(&cfg.CreateAdmin, "create-admin", "", "Create admin user (username:password)")
	flag.Parse()

	if cfg.Secret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		cfg.Secret = hex.EncodeToString(b)
	}

	return cfg
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
