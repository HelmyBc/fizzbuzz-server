package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HelmyBc/fizzbuzz-server/internal/stats"
)

func newTestRouter() http.Handler {
	return NewRouter(stats.New())
}

func TestFizzBuzzHandle_Success(t *testing.T) {
	router := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("fail: GET /fizzbuzz returned status %d, want %d", rec.Code, http.StatusOK)
	}

	var body fizzBuzzResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("fail: Could not unmarshal response %v", err)
	}
	if len(body.Result) != 15 {
		t.Fatalf("expected 15 results, got %d", len(body.Result))
	}
	if body.Result[2] != "fizz" || body.Result[4] != "buzz" || body.Result[14] != "fizzbuzz" {
		t.Fatalf("incorrect values in result")
	}
}

func TestFizzBuzzHandler_MissingParam(t *testing.T) {
	router := newTestRouter()
	// Missing limit parameter, the server should return 400 Bad Request
	req := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=3&int2=5&str1=fizz&str2=buzz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("fail: GET /fizzbuzz returned status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFizzBuzzHandler_InvalidInteger(t *testing.T) {
	router := newTestRouter()
	// int1="a" is not a valid integer, the server should return 400 Bad Request
	req := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=a&int2=5&limit=15&str1=fizz&str2=buzz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("fail: GET /fizzbuzz returned status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFizzBuzzHandler_LimitTooLarge(t *testing.T) {
	router := newTestRouter()
	// limit=1000001 is too large, the server should return 400 Bad Request
	req := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=3&int2=5&limit=1000001&str1=fizz&str2=buzz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("fail: GET /fizzbuzz returned status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFizzBuzz_StrTooLong(t *testing.T) {
	router := newTestRouter()
	strLong := strings.Repeat("a", 101)
	// str1 is too long, the server should return 400 Bad Request
	req := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=3&int2=5&limit=15&str1="+strLong+"&str2=buzz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("fail: GET /fizzbuzz returned status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFizzBuzzHandler_WrongMethod(t *testing.T) {
	router := newTestRouter()
	// POST /fizzbuzz is not allowed, the server should return 405 Method Not Allowed
	req := httptest.NewRequest(http.MethodPost, "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("fail: POST /fizzbuzz returned status %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestStatsHandler_EmptyThenPopulated(t *testing.T) {
	router := newTestRouter()
	rec := httptest.NewRecorder()
	// No records yet -> 404
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/statistics", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("fail: GET /statistics returned status %d, want %d", rec.Code, http.StatusNotFound)
	}
	// Make the same request 3 times, another request once.
	popular := "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz"
	rare := "/fizzbuzz?int1=2&int2=7&limit=20&str1=foo&str2=bar"
	for i := 0; i < 3; i++ {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, popular, nil))
	}
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, rare, nil))

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/statistics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body statisticsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Hits != 3 {
		t.Errorf("Hits = %d, want 3", body.Hits)
	}
	// Assert the structured fields match the popular request's parameters.
	if body.Int1 != 3 || body.Int2 != 5 || body.Limit != 15 {
		t.Errorf("unexpected int params: int1=%d int2=%d limit=%d", body.Int1, body.Int2, body.Limit)
	}
	if body.Str1 != "fizz" || body.Str2 != "buzz" {
		t.Errorf("unexpected str params: str1=%q str2=%q", body.Str1, body.Str2)
	}
}

func TestStatisticsHandler_RequestIDHeader(t *testing.T) {
	router := newTestRouter()

	// Provide a custom request ID, it should be echoed back.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "test-id-123")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "test-id-123" {
		t.Errorf("X-Request-ID = %q, want %q", got, "test-id-123")
	}
}

func TestStatisticsHandler_RequestIDGenerated(t *testing.T) {
	router := newTestRouter()

	// No X-Request-ID header, the server must generate one.
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if id := rec.Header().Get("X-Request-ID"); id == "" {
		t.Error("expected a generated X-Request-ID header, got empty string")
	}
}

func TestHealthHandler(t *testing.T) {
	router := newTestRouter()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestJSONErrorShapes verifies that 404 (unknown path) and 405 (wrong method)
// both return a JSON {"error":"..."} body instead of the stdlib plain-text default.
func TestJSONErrorShapes(t *testing.T) {
	router := newTestRouter()

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"POST /fizzbuzz -> 405 JSON", http.MethodPost, "/fizzbuzz", http.StatusMethodNotAllowed},
		{"DELETE /fizzbuzz -> 405 JSON", http.MethodDelete, "/fizzbuzz", http.StatusMethodNotAllowed},
		{"POST /statistics -> 405 JSON", http.MethodPost, "/statistics", http.StatusMethodNotAllowed},
		{"DELETE /healthz -> 405 JSON", http.MethodDelete, "/healthz", http.StatusMethodNotAllowed},
		{"GET /does-not-exist -> 404 JSON", http.MethodGet, "/does-not-exist", http.StatusNotFound},
		{"GET /fizzbuzz/extra -> 404 JSON", http.MethodGet, "/fizzbuzz/extra", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			// Must be JSON content-type, not text/plain.
			ct := rec.Header().Get("Content-Type")
			if ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
			}
			// Must decode as {"error":"..."}.
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not valid JSON: %v — body: %s", err, rec.Body.String())
			}
			if body.Error == "" {
				t.Errorf("expected non-empty error field, got %s", rec.Body.String())
			}
		})
	}
}
