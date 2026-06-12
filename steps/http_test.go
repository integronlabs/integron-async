package steps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunHTTP_Success(t *testing.T) {
	// Setup mock test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		if r.Header.Get("X-Test-Header") != "hello-world" {
			t.Errorf("Expected X-Test-Header value 'hello-world', got '%s'", r.Header.Get("X-Test-Header"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"fact": "Dogs are great."}}`))
	}))
	defer server.Close()

	stepMap := map[string]interface{}{
		"name":   "testRequest",
		"type":   "http",
		"method": "POST",
		"url":    server.URL,
		"headers": map[string]interface{}{
			"X-Test-Header": "$.message.payload.headerVal",
		},
		"body": map[string]interface{}{
			"input": "$.message.payload.data",
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{
				"output": map[string]interface{}{
					"fact": "$.body.data.fact",
				},
				"next": "nextStep",
			},
		},
	}

	stepOutputs := map[string]interface{}{
		"message": map[string]interface{}{
			"payload": map[string]interface{}{
				"headerVal": "hello-world",
				"data":      "some-data",
			},
		},
	}

	output, next, err := runHTTP(context.Background(), stepMap, stepOutputs)
	if err != nil {
		t.Fatalf("runHTTP failed: %v", err)
	}

	if next != "nextStep" {
		t.Errorf("Expected next to be 'nextStep', got '%s'", next)
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected output to be a map")
	}

	if outputMap["fact"] != "Dogs are great." {
		t.Errorf("Expected fact 'Dogs are great.', got '%v'", outputMap["fact"])
	}
}

func TestRunHTTP_NoBody(t *testing.T) {
	// Setup mock test server checking that no body is sent on GET
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got %s", r.Method)
		}
		// Content-Length should be empty or 0 since we sent no body
		if r.ContentLength > 0 {
			t.Errorf("Expected Content-Length to be <= 0, got %d", r.ContentLength)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	stepMap := map[string]interface{}{
		"name":   "testRequestNoBody",
		"type":   "http",
		"method": "GET",
		"url":    server.URL,
		"responses": map[string]interface{}{
			"200": map[string]interface{}{
				"output": map[string]interface{}{
					"status": "$.body.ok",
				},
				"next": "",
			},
		},
	}

	stepOutputs := map[string]interface{}{}

	output, next, err := runHTTP(context.Background(), stepMap, stepOutputs)
	if err != nil {
		t.Fatalf("runHTTP failed: %v", err)
	}

	if next != "" {
		t.Errorf("Expected next to be empty, got '%s'", next)
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected output to be a map")
	}

	if outputMap["status"] != true {
		t.Errorf("Expected status true, got %v", outputMap["status"])
	}
}
