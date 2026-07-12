package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

func (srv *Server) handleGetDashboardLayout(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	budgetID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	layout, err := srv.store.GetDashboardLayout(r.Context(), user.ID, budgetID)
	if err != nil {
		writeDashboardError(w, err, "load")
		return
	}

	writeJSON(w, http.StatusOK, layout)
}

func (srv *Server) handleUpdateDashboardLayout(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	budgetID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	var layout store.DashboardLayout
	if err := json.NewDecoder(r.Body).Decode(&layout); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	saved, err := srv.store.SaveDashboardLayout(r.Context(), user.ID, budgetID, layout)
	if err != nil {
		writeDashboardError(w, err, "save")
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func writeDashboardError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, store.ErrInvalidDashboardLayout):
		http.Error(w, "invalid dashboard layout", http.StatusBadRequest)
	default:
		http.Error(w, "could not "+action+" dashboard layout", http.StatusInternalServerError)
	}
}
