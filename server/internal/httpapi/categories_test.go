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

func TestCategoryEndpointsCreateAndMapDomainErrors(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "category-owner")
	budget, err := st.CreateBudget(t.Context(), user.ID, "Household", "EUR", nil)
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

	if created.Name != "Food" || created.IsSystem {
		t.Errorf("created category = %#v, want custom Food category", created)
	}

	duplicateResponse := requestCategory(http.MethodPost, collectionPath, `{"name":"Food"}`)
	if duplicateResponse.Code != http.StatusConflict {
		t.Errorf("duplicate status = %d, want %d", duplicateResponse.Code, http.StatusConflict)
	}

	categories, err := st.ListCategories(t.Context(), user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	var systemCategoryID int64
	for _, category := range categories {
		if category.IsSystem {
			systemCategoryID = category.ID
			break
		}
	}

	updatePath := "/api/categories/" + strconv.FormatInt(systemCategoryID, 10)
	systemResponse := requestCategory(http.MethodPatch, updatePath, `{"name":"Other"}`)
	if systemResponse.Code != http.StatusForbidden {
		t.Errorf("system category update status = %d, want %d", systemResponse.Code, http.StatusForbidden)
	}
}
