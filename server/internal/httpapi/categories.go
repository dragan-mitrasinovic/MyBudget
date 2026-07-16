package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

type categoryRequest struct {
	Name string `json:"name"`
}

type categoryColorRequest struct {
	ColorKey string  `json:"colorKey"`
	ColorHex *string `json:"colorHex"`
}

var categoryColorHexPattern = regexp.MustCompile(`^#[0-9A-F]{6}$`)

// normalizeAndValidate trims and validates a category payload in place.
func (req *categoryRequest) normalizeAndValidate() error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" || len(req.Name) > maxNameLen {
		return fmt.Errorf("name must be 1-%d characters", maxNameLen)
	}

	return nil
}

func (req *categoryColorRequest) normalizeAndValidate() error {
	req.ColorKey = strings.TrimSpace(req.ColorKey)
	if !store.IsValidCategoryColorKey(req.ColorKey) {
		return errors.New("invalid colorKey")
	}

	if req.ColorHex == nil {
		return nil
	}

	colorHex := strings.ToUpper(strings.TrimSpace(*req.ColorHex))
	if !categoryColorHexPattern.MatchString(colorHex) {
		return errors.New("colorHex must be null or a six-digit hex color")
	}

	req.ColorHex = &colorHex
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
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, "could not list categories", http.StatusInternalServerError)
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

func (srv *Server) handleUpdateCategoryColor(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	categoryID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	var req categoryColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := req.normalizeAndValidate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := srv.store.UpdateCategoryColor(r.Context(), user.ID, categoryID, req.ColorKey, req.ColorHex)
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

	reassignToCategoryID, ok := parseOptionalReassignmentCategoryID(w, r)
	if !ok {
		return
	}

	if err := srv.store.DeleteCategory(r.Context(), user.ID, categoryID, reassignToCategoryID); err != nil {
		writeCategoryError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseOptionalReassignmentCategoryID(w http.ResponseWriter, r *http.Request) (*int64, bool) {
	values, exists := r.URL.Query()["reassignToCategoryId"]
	if !exists {
		return nil, true
	}

	if len(values) != 1 || values[0] == "" {
		http.Error(w, "invalid reassignToCategoryId", http.StatusBadRequest)
		return nil, false
	}

	categoryID, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || categoryID <= 0 {
		http.Error(w, "invalid reassignToCategoryId", http.StatusBadRequest)
		return nil, false
	}

	return &categoryID, true
}

// writeCategoryError maps the category data-layer sentinel errors to HTTP responses.
func writeCategoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, store.ErrDuplicateName):
		http.Error(w, "a category with that name already exists", http.StatusConflict)
	case errors.Is(err, store.ErrCategoryNotEmpty):
		http.Error(w, "category has expenses; reassignToCategoryId is required", http.StatusConflict)
	case errors.Is(err, store.ErrInvalidReassignment):
		http.Error(w, "reassignment category must be different and belong to the same budget", http.StatusBadRequest)
	default:
		http.Error(w, "could not save category", http.StatusInternalServerError)
	}
}
