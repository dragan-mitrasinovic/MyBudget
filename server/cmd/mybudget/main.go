package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/httpapi"
	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

func main() {
	isDevelopment, err := strconv.ParseBool(os.Getenv("IS_DEVELOPMENT"))
	if err != nil {
		log.Fatalf("IS_DEVELOPMENT must be true or false: %v", err)
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}

	port := envOrDefault("PORT", "8080")
	dbPath := envOrDefault("DB_PATH", "./mybudget.db")
	sessionTTL := mustParseSessionTTL(envOrDefault("SESSION_TTL", "720h"))

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			log.Printf("close store: %v", err)
		}
	}()

	handler := httpapi.New(st, httpapi.Config{
		GoogleClientID: clientID,
		SessionTTL:     sessionTTL,
		IsDevelopment:  isDevelopment,
	})

	addr := ":" + port
	log.Printf("MyBudget API listening on %s (db=%s, development=%v)", addr, dbPath, isDevelopment)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func mustParseSessionTTL(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("invalid SESSION_TTL %q: %v", s, err)
	}

	return d
}
