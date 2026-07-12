package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

func TestDashboardEndpointsLoadAndPersistLayout(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "dashboard-owner")
	budget, err := st.CreateBudget(t.Context(), user.ID, "Household", "EUR")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	if _, err := st.CreateCategory(t.Context(), user.ID, budget.ID, "Food"); err != nil {
		t.Fatalf("create category: %v", err)
	}

	path := "/api/budgets/" + strconv.FormatInt(budget.ID, 10) + "/dashboard"
	getRequest := httptest.NewRequest(http.MethodGet, path, nil)
	getRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	getResponse := httptest.NewRecorder()
	handler.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getResponse.Code, http.StatusOK)
	}

	var layout store.DashboardLayout
	if err := json.NewDecoder(getResponse.Body).Decode(&layout); err != nil {
		t.Fatalf("decode layout: %v", err)
	}

	if len(layout.Widgets) != 2 {
		t.Fatalf("widget count = %d, want 2", len(layout.Widgets))
	}

	layout.Widgets[0].Y = 6
	body, err := json.Marshal(layout)
	if err != nil {
		t.Fatalf("encode layout: %v", err)
	}

	putRequest := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	putRequest.Header.Set("Origin", testAppOrigin)
	putRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	putResponse := httptest.NewRecorder()
	handler.ServeHTTP(putResponse, putRequest)

	if putResponse.Code != http.StatusOK {
		t.Fatalf("put status = %d, want %d", putResponse.Code, http.StatusOK)
	}

	getRequest = httptest.NewRequest(http.MethodGet, path, nil)
	getRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	getResponse = httptest.NewRecorder()
	handler.ServeHTTP(getResponse, getRequest)

	var persisted store.DashboardLayout
	if err := json.NewDecoder(getResponse.Body).Decode(&persisted); err != nil {
		t.Fatalf("decode persisted layout: %v", err)
	}

	if persisted.Widgets[0].Y != 6 {
		t.Errorf("persisted widget y = %d, want 6", persisted.Widgets[0].Y)
	}
}

func TestDashboardEndpointsRejectInvalidAndUnownedLayouts(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	owner, ownerSessionID := createHTTPTestUserSession(t, st, "dashboard-owner")
	_, otherSessionID := createHTTPTestUserSession(t, st, "other-user")
	budget, err := st.CreateBudget(t.Context(), owner.ID, "Household", "EUR")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	path := "/api/budgets/" + strconv.FormatInt(budget.ID, 10) + "/dashboard"
	invalidRequest := httptest.NewRequest(
		http.MethodPut,
		path,
		bytes.NewBufferString(`{"version":4,"widgets":[]}`),
	)
	invalidRequest.Header.Set("Origin", testAppOrigin)
	invalidRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ownerSessionID})
	invalidResponse := httptest.NewRecorder()
	handler.ServeHTTP(invalidResponse, invalidRequest)

	if invalidResponse.Code != http.StatusBadRequest {
		t.Errorf("invalid layout status = %d, want %d", invalidResponse.Code, http.StatusBadRequest)
	}

	otherRequest := httptest.NewRequest(http.MethodGet, path, nil)
	otherRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: otherSessionID})
	otherResponse := httptest.NewRecorder()
	handler.ServeHTTP(otherResponse, otherRequest)

	if otherResponse.Code != http.StatusNotFound {
		t.Errorf("other user's status = %d, want %d", otherResponse.Code, http.StatusNotFound)
	}
}
