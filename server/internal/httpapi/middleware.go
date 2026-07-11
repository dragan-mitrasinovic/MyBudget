package httpapi

import (
	"net/http"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

// requireAuthenticatedUser resolves the session user, writing the appropriate error response and
// returning nil when the request is unauthenticated or lookup fails.
func (srv *Server) requireAuthenticatedUser(w http.ResponseWriter, r *http.Request) *store.User {
	user, err := srv.userFromSessionCookie(r)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return nil
	}

	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil
	}

	return user
}
