package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/httpapi"
	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
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
	devAuth := getenv("DEV_AUTH", "false") == "true"

	// The dummy-login shortcut is a full auth bypass; never let it run in a production
	// configuration. SECURE_COOKIES=true is our signal for "this is production".
	if devAuth && secureCookies {
		log.Fatal("DEV_AUTH must not be enabled with SECURE_COOKIES=true")
	}

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
		SecureCookies:  secureCookies,
		SessionTTL:     sessionTTL,
		DevAuth:        devAuth,
	})

	addr := ":" + port
	log.Printf("MyBudget API listening on %s (db=%s, secureCookies=%v)", addr, dbPath, secureCookies)
	if err := http.ListenAndServe(addr, handler); err != nil {
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
