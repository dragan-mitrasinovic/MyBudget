package store

import (
	"context"
	"testing"
)

func createStoreTestUser(t *testing.T, st *Store, sub string) *User {
	t.Helper()

	user, err := st.UpsertUser(context.Background(), sub, sub+"@example.com", sub, "")
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return user
}

func createStoreTestBudget(t *testing.T, st *Store, userID int64, name string) *Budget {
	t.Helper()

	budget, err := st.CreateBudget(context.Background(), userID, name, "EUR", nil)
	if err != nil {
		t.Fatalf("create test budget: %v", err)
	}

	return budget
}
