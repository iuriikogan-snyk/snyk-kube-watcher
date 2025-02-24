package snyk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

var url string

func TestMonitorImage(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedError  bool
		expectedBody   string
		expectedMethod string
	}{

		{
			name:           "success",
			statusCode:     http.StatusOK,
			expectedError:  false,
			expectedBody:   `{"image":"test-image"}`,
			expectedMethod: http.MethodPost,
		},
		{
			name:           "created",
			statusCode:     http.StatusCreated,
			expectedError:  false,
			expectedBody:   `{"image":"test-image"}`,
			expectedMethod: http.MethodPost,
		},
		{
			name:           "bad request",
			statusCode:     http.StatusBadRequest,
			expectedError:  true,
			expectedBody:   `{"error":"bad request"}`,
			expectedMethod: http.MethodPost,
		},
		{
			name:           "internal error",
			statusCode:     http.StatusInternalServerError,
			expectedError:  true,
			expectedBody:   `{"error":"internal error"}`,
			expectedMethod: http.MethodPost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.expectedMethod {
					t.Errorf("Expected method %s, got %s", tt.expectedMethod, r.Method)
				}
				if r.Header.Get("Authorization") != "token test-token" {
					t.Errorf("Expected Authorization header to be 'token test-token', got %s", r.Header.Get("Authorization"))
				}
				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.expectedBody))
				if err != nil {
					t.Errorf("Error writing response body: %v", err)
				}
			}))
			defer server.Close()
			url = server.URL
			err := MonitorImage(context.Background(), "test-image", "test-org", "test-token")
			if tt.expectedError && err == nil {
				t.Errorf("Expected error, got nil")
			} else if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
