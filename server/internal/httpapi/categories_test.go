package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

func TestCategoryRequestNormalizeAndValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   categoryRequest
		wantName  string
		wantError bool
	}{
		{
			name:     "normalizes valid request",
			request:  categoryRequest{Name: "  Groceries  "},
			wantName: "Groceries",
		},
		{
			name:     "accepts maximum length name",
			request:  categoryRequest{Name: strings.Repeat("a", maxNameLen)},
			wantName: strings.Repeat("a", maxNameLen),
		},
		{
			name:      "rejects empty name",
			request:   categoryRequest{Name: "  "},
			wantError: true,
		},
		{
			name:      "rejects long name",
			request:   categoryRequest{Name: strings.Repeat("a", maxNameLen+1)},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.normalizeAndValidate()
			if (err != nil) != test.wantError {
				t.Fatalf("normalizeAndValidate() error = %v, wantError %v", err, test.wantError)
			}

			if !test.wantError && test.request.Name != test.wantName {
				t.Errorf("Name = %q, want %q", test.request.Name, test.wantName)
			}
		})
	}
}

func TestCategoryColorRequestNormalizeAndValidate(t *testing.T) {
	customHex := "  #a1b2c3  "
	req := categoryColorRequest{ColorKey: " chart-4 ", ColorHex: &customHex}
	if err := req.normalizeAndValidate(); err != nil {
		t.Fatalf("normalizeAndValidate() error = %v", err)
	}

	if req.ColorKey != "chart-4" || req.ColorHex == nil || *req.ColorHex != "#A1B2C3" {
		t.Errorf("normalized request = %#v", req)
	}

	invalidHex := "#12345"
	if err := (&categoryColorRequest{ColorKey: "chart-1", ColorHex: &invalidHex}).normalizeAndValidate(); err == nil {
		t.Error("invalid colorHex error = nil")
	}

	if err := (&categoryColorRequest{ColorKey: "unknown"}).normalizeAndValidate(); err == nil {
		t.Error("invalid colorKey error = nil")
	}
}

func TestCategoryEndpointsCreateAndMapDomainErrors(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "category-owner")
	budget, err := st.CreateBudget(t.Context(), user.ID, "Household", "EUR")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	requestCategory := func(method, path, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		request.Header.Set("Origin", testAppOrigin)
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	collectionPath := "/api/budgets/" + strconv.FormatInt(budget.ID, 10) + "/categories"
	createResponse := requestCategory(http.MethodPost, collectionPath, `{"name":"  Food  "}`)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.Code, http.StatusCreated)
	}

	var created store.Category
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created category: %v", err)
	}

	if created.Name != "Food" {
		t.Errorf("created category = %#v, want Food category", created)
	}

	duplicateResponse := requestCategory(http.MethodPost, collectionPath, `{"name":"Food"}`)
	if duplicateResponse.Code != http.StatusConflict {
		t.Errorf("duplicate status = %d, want %d", duplicateResponse.Code, http.StatusConflict)
	}

	updatePath := "/api/categories/" + strconv.FormatInt(created.ID, 10)
	updateResponse := requestCategory(http.MethodPatch, updatePath, `{"name":"Meals"}`)
	if updateResponse.Code != http.StatusOK {
		t.Errorf("category update status = %d, want %d", updateResponse.Code, http.StatusOK)
	}
}

func TestListCategoriesReturnsEmptyArrayForOwnedBudget(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "category-owner")
	_, otherSessionID := createHTTPTestUserSession(t, st, "other-user")
	budget, err := st.CreateBudget(t.Context(), user.ID, "Household", "EUR")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	path := "/api/budgets/" + strconv.FormatInt(budget.ID, 10) + "/categories"
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("list categories status = %d, want %d", response.Code, http.StatusOK)
	}

	var categories []store.Category
	if err := json.NewDecoder(response.Body).Decode(&categories); err != nil {
		t.Fatalf("decode categories: %v", err)
	}

	if categories == nil || len(categories) != 0 {
		t.Fatalf("categories = %#v, want allocated empty array", categories)
	}

	request = httptest.NewRequest(http.MethodGet, path, nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: otherSessionID})
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("unowned list status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestDeleteCategoryEndpointRequiresAndAppliesReassignment(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "category-owner")
	budget, err := st.CreateBudget(t.Context(), user.ID, "Household", "EUR")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	source, err := st.CreateCategory(t.Context(), user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create source category: %v", err)
	}

	target, err := st.CreateCategory(t.Context(), user.ID, budget.ID, "Other")
	if err != nil {
		t.Fatalf("create target category: %v", err)
	}

	otherBudget, err := st.CreateBudget(t.Context(), user.ID, "Travel", "EUR")
	if err != nil {
		t.Fatalf("create other budget: %v", err)
	}

	foreignTarget, err := st.CreateCategory(t.Context(), user.ID, otherBudget.ID, "Travel")
	if err != nil {
		t.Fatalf("create foreign target category: %v", err)
	}

	expense, err := st.CreateExpense(t.Context(), user.ID, budget.ID, source.ID, 1599, "Lunch", "2026-07-11")
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}

	requestDelete := func(query string) *httptest.ResponseRecorder {
		path := "/api/categories/" + strconv.FormatInt(source.ID, 10) + query
		request := httptest.NewRequest(http.MethodDelete, path, nil)
		request.Header.Set("Origin", testAppOrigin)
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	missingTargetResponse := requestDelete("")
	if missingTargetResponse.Code != http.StatusConflict {
		t.Fatalf("missing target status = %d, want %d", missingTargetResponse.Code, http.StatusConflict)
	}

	invalidTargetResponse := requestDelete("?reassignToCategoryId=" + strconv.FormatInt(source.ID, 10))
	if invalidTargetResponse.Code != http.StatusBadRequest {
		t.Fatalf("same target status = %d, want %d", invalidTargetResponse.Code, http.StatusBadRequest)
	}

	invalidQueries := []string{
		"?reassignToCategoryId=",
		"?reassignToCategoryId=abc",
		"?reassignToCategoryId=0",
		"?reassignToCategoryId=999999",
		"?reassignToCategoryId=" + strconv.FormatInt(foreignTarget.ID, 10),
		"?reassignToCategoryId=" + strconv.FormatInt(target.ID, 10) + "&reassignToCategoryId=" + strconv.FormatInt(target.ID, 10),
	}
	for _, query := range invalidQueries {
		response := requestDelete(query)
		if response.Code != http.StatusBadRequest {
			t.Errorf("invalid query %q status = %d, want %d", query, response.Code, http.StatusBadRequest)
		}
	}

	validResponse := requestDelete("?reassignToCategoryId=" + strconv.FormatInt(target.ID, 10))
	if validResponse.Code != http.StatusNoContent {
		t.Fatalf("valid reassignment status = %d, want %d", validResponse.Code, http.StatusNoContent)
	}

	expenses, err := st.ListExpenses(t.Context(), user.ID, budget.ID, store.ExpenseFilter{})
	if err != nil {
		t.Fatalf("list expenses: %v", err)
	}

	if len(expenses) != 1 || expenses[0].ID != expense.ID || expenses[0].CategoryID != target.ID {
		t.Fatalf("reassigned expenses = %#v, want expense %d in category %d", expenses, expense.ID, target.ID)
	}
}

func TestDeleteEmptyCategoryEndpointNeedsNoReassignment(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "category-owner")
	budget, err := st.CreateBudget(t.Context(), user.ID, "Household", "EUR")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	category, err := st.CreateCategory(t.Context(), user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	path := "/api/categories/" + strconv.FormatInt(category.ID, 10)
	request := httptest.NewRequest(http.MethodDelete, path, nil)
	request.Header.Set("Origin", testAppOrigin)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("delete empty category status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestCategoryColorEndpointUpdatesCategory(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "category-owner")
	budget, err := st.CreateBudget(t.Context(), user.ID, "Household", "EUR")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	category, err := st.CreateCategory(t.Context(), user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	path := "/api/categories/" + strconv.FormatInt(category.ID, 10) + "/color"
	request := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(`{"colorKey":"chart-5","colorHex":"#abcdef"}`))
	request.Header.Set("Origin", testAppOrigin)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("update color status = %d, want %d", response.Code, http.StatusOK)
	}

	var updated store.Category
	if err := json.NewDecoder(response.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated category: %v", err)
	}

	if updated.ColorKey != "chart-5" || updated.ColorHex == nil || *updated.ColorHex != "#ABCDEF" {
		t.Errorf("updated category = %#v", updated)
	}
}
