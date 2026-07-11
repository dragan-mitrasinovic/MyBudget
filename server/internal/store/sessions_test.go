package store

import (
	"context"
	"testing"
	"time"
)

const (
	testSessionTTL             = 30 * 24 * time.Hour
	testSessionRenewalInterval = 24 * time.Hour
)

func TestGetAndRenewSessionUserDoesNotRenewFreshSession(t *testing.T) {
	st := openTestStore(t)
	user := createSessionTestUser(t, st)
	sessionID, err := st.CreateSession(context.Background(), user.ID, testSessionTTL)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	originalExpiresAt := readSessionExpiresAt(t, st, sessionID)

	gotUser, didRenew, err := st.GetAndRenewSessionUser(
		context.Background(),
		sessionID,
		testSessionTTL,
		testSessionRenewalInterval,
	)
	if err != nil {
		t.Fatalf("get session user: %v", err)
	}

	if gotUser == nil || gotUser.ID != user.ID {
		t.Fatalf("session user = %#v, want user ID %d", gotUser, user.ID)
	}

	if didRenew {
		t.Fatal("didRenew = true, want false for fresh session")
	}

	if expiresAt := readSessionExpiresAt(t, st, sessionID); expiresAt != originalExpiresAt {
		t.Fatalf("expires_at = %q, want unchanged %q", expiresAt, originalExpiresAt)
	}
}

func TestGetAndRenewSessionUserRenewsSessionAfterInterval(t *testing.T) {
	st := openTestStore(t)
	user := createSessionTestUser(t, st)
	sessionID, err := st.CreateSession(context.Background(), user.ID, testSessionTTL)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	oldExpiresAt := time.Now().UTC().Add(testSessionTTL - testSessionRenewalInterval - time.Hour)
	setSessionExpiresAt(t, st, sessionID, oldExpiresAt)
	renewedAfter := time.Now().UTC()

	gotUser, didRenew, err := st.GetAndRenewSessionUser(
		context.Background(),
		sessionID,
		testSessionTTL,
		testSessionRenewalInterval,
	)
	if err != nil {
		t.Fatalf("get session user: %v", err)
	}

	if gotUser == nil || gotUser.ID != user.ID {
		t.Fatalf("session user = %#v, want user ID %d", gotUser, user.ID)
	}

	if !didRenew {
		t.Fatal("didRenew = false, want true after renewal interval")
	}

	expiresAt, err := time.Parse(time.RFC3339, readSessionExpiresAt(t, st, sessionID))
	if err != nil {
		t.Fatalf("parse renewed expiry: %v", err)
	}

	if expiresAt.Before(renewedAfter.Add(testSessionTTL - time.Second)) {
		t.Fatalf("expires_at = %s, want approximately %s", expiresAt, renewedAfter.Add(testSessionTTL))
	}
}

func TestGetAndRenewSessionUserDoesNotReviveExpiredSession(t *testing.T) {
	st := openTestStore(t)
	user := createSessionTestUser(t, st)
	sessionID, err := st.CreateSession(context.Background(), user.ID, testSessionTTL)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	setSessionExpiresAt(t, st, sessionID, time.Now().UTC().Add(-time.Hour))

	gotUser, didRenew, err := st.GetAndRenewSessionUser(
		context.Background(),
		sessionID,
		testSessionTTL,
		testSessionRenewalInterval,
	)
	if err != nil {
		t.Fatalf("get session user: %v", err)
	}

	if gotUser != nil {
		t.Fatalf("session user = %#v, want nil", gotUser)
	}

	if didRenew {
		t.Fatal("didRenew = true, want false for expired session")
	}

	var sessionCount int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE id = ?", sessionID).Scan(&sessionCount); err != nil {
		t.Fatalf("count expired session: %v", err)
	}

	if sessionCount != 0 {
		t.Fatalf("expired session count = %d, want 0", sessionCount)
	}
}

func TestGetAndRenewSessionUserConcurrentRequestsRenewOnce(t *testing.T) {
	const requestCount = 8
	st := openTestStore(t)
	user := createSessionTestUser(t, st)
	sessionID, err := st.CreateSession(context.Background(), user.ID, testSessionTTL)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	oldExpiresAt := time.Now().UTC().Add(testSessionTTL - testSessionRenewalInterval - time.Hour)
	setSessionExpiresAt(t, st, sessionID, oldExpiresAt)

	type result struct {
		user     *User
		didRenew bool
		err      error
	}
	start := make(chan struct{})
	results := make(chan result, requestCount)
	for range requestCount {
		go func() {
			<-start
			gotUser, didRenew, err := st.GetAndRenewSessionUser(
				context.Background(),
				sessionID,
				testSessionTTL,
				testSessionRenewalInterval,
			)
			results <- result{user: gotUser, didRenew: didRenew, err: err}
		}()
	}
	close(start)

	renewalCount := 0
	for range requestCount {
		result := <-results
		if result.err != nil {
			t.Fatalf("get session user: %v", result.err)
		}

		if result.user == nil || result.user.ID != user.ID {
			t.Fatalf("session user = %#v, want user ID %d", result.user, user.ID)
		}

		if result.didRenew {
			renewalCount++
		}
	}

	if renewalCount != 1 {
		t.Fatalf("renewal count = %d, want 1", renewalCount)
	}
}

func createSessionTestUser(t *testing.T, st *Store) *User {
	t.Helper()

	user, err := st.UpsertUser(
		context.Background(),
		"session-test-user",
		"session@example.com",
		"Session Test User",
		"",
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return user
}

func readSessionExpiresAt(t *testing.T, st *Store, sessionID string) string {
	t.Helper()

	var expiresAt string
	if err := st.db.QueryRow("SELECT expires_at FROM sessions WHERE id = ?", sessionID).Scan(&expiresAt); err != nil {
		t.Fatalf("read session expiry: %v", err)
	}

	return expiresAt
}

func setSessionExpiresAt(t *testing.T, st *Store, sessionID string, expiresAt time.Time) {
	t.Helper()

	if _, err := st.db.Exec(
		"UPDATE sessions SET expires_at = ? WHERE id = ?",
		expiresAt.Format(time.RFC3339),
		sessionID,
	); err != nil {
		t.Fatalf("set session expiry: %v", err)
	}
}
