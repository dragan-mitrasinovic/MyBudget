package store

import (
	"context"
	"testing"
)

func TestUpsertUserRefreshesProfileWithoutChangingIdentity(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	first, err := st.UpsertUser(ctx, "google-sub", "old@example.com", "Old Name", "old-picture")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	updated, err := st.UpsertUser(ctx, "google-sub", "new@example.com", "New Name", "new-picture")
	if err != nil {
		t.Fatalf("update user: %v", err)
	}

	if updated.ID != first.ID {
		t.Errorf("updated user ID = %d, want %d", updated.ID, first.ID)
	}

	if updated.CreatedAt != first.CreatedAt {
		t.Errorf("updated created_at = %q, want %q", updated.CreatedAt, first.CreatedAt)
	}

	if updated.Email != "new@example.com" || updated.Name != "New Name" || updated.Picture != "new-picture" {
		t.Errorf("updated profile = %#v, want refreshed fields", updated)
	}
}
