package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

type categoryRequest struct {
	Name string `json:"name"`
}

// normalizeAndValidate trims and validates a category payload in place.
func (req *categoryRequest) normalizeAndValidate() error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" || len(req.Name) > maxNameLen {
		return fmt.Errorf("name must be 1-%d characters", maxNameLen)
	}

	return nil
}

func (srv *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	budgetID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	categories, err := srv.store.ListCategories(r.Context(), user.ID, budgetID)
	if err != nil {
		http.Error(w, "could not list categories", http.StatusInternalServerError)
		return
	}

	// Every real budget has at least its Uncategorized category, so an empty result
	// means the budget does not exist or is not the caller's.
	if len(categories) == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

func (srv *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	budgetID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := req.normalizeAndValidate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := srv.store.CreateCategory(r.Context(), user.ID, budgetID, req.Name)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, category)
}

func (srv *Server) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	categoryID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := req.normalizeAndValidate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := srv.store.UpdateCategory(r.Context(), user.ID, categoryID, req.Name)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, category)
}

func (srv *Server) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	categoryID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	if err := srv.store.DeleteCategory(r.Context(), user.ID, categoryID); err != nil {
		writeCategoryError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeCategoryError maps the category data-layer sentinel errors to HTTP responses.
func writeCategoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, store.ErrSystemCategory):
		http.Error(w, "the Uncategorized category cannot be modified", http.StatusForbidden)
	case errors.Is(err, store.ErrDuplicateName):
		http.Error(w, "a category with that name already exists", http.StatusConflict)
	default:
		http.Error(w, "could not save category", http.StatusInternalServerError)
	}
}
