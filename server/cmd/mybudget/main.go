package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}

	port := getenv("PORT", "8080")
	dbPath := getenv("DB_PATH", "./mybudget.db")
	secureCookies := getenv("SECURE_COOKIES", "false") == "true"
	sessionTTL := parseTTL(getenv("SESSION_TTL", "720h"))

	store, err := OpenStore(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	srv := &Server{
		store:          store,
		googleClientID: clientID,
		secureCookies:  secureCookies,
		sessionTTL:     sessionTTL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/google", srv.handleGoogleLogin)
	mux.HandleFunc("GET /api/me", srv.handleMe)
	mux.HandleFunc("POST /api/auth/logout", srv.handleLogout)

	addr := ":" + port
	log.Printf("MyBudget API listening on %s (db=%s, secureCookies=%v)", addr, dbPath, secureCookies)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func parseTTL(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("invalid SESSION_TTL %q: %v", s, err)
	}

	return d
}
