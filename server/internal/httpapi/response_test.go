package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePathID(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		wantID     int64
		wantOK     bool
		wantStatus int
	}{
		{name: "positive integer", value: "42", wantID: 42, wantOK: true, wantStatus: http.StatusOK},
		{name: "zero", value: "0", wantStatus: http.StatusBadRequest},
		{name: "negative", value: "-1", wantStatus: http.StatusBadRequest},
		{name: "not a number", value: "abc", wantStatus: http.StatusBadRequest},
		{name: "overflow", value: "9223372036854775808", wantStatus: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("id", test.value)
			response := httptest.NewRecorder()

			id, ok := parsePathID(response, request, "id")
			if id != test.wantID || ok != test.wantOK {
				t.Errorf("parsePathID() = (%d, %v), want (%d, %v)", id, ok, test.wantID, test.wantOK)
			}

			if response.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
