package httpresponse

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	Health(responseRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d", responseRecorder.Code)
	}
	if responseRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q", responseRecorder.Header().Get("Content-Type"))
	}
	if strings.TrimSpace(responseRecorder.Body.String()) != `{"status":"UP"}` {
		t.Fatalf("body = %s", responseRecorder.Body.String())
	}
}

func TestDecodeJSON(t *testing.T) {
	for _, testCase := range []struct {
		name string
		body string
		want bool
	}{
		{name: "valid object", body: `{"name":"Keyboard"}`, want: true},
		{name: "unknown field", body: `{"unexpected":true}`},
		{name: "multiple values", body: `{"name":"Keyboard"}{"name":"Mouse"}`},
		{name: "malformed", body: `{`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(testCase.body))
			var destination struct {
				Name string `json:"name"`
			}
			errorValue := DecodeJSON(httptest.NewRecorder(), request, &destination)
			if (errorValue == nil) != testCase.want {
				t.Fatalf("DecodeJSON() error = %v", errorValue)
			}
		})
	}
}
