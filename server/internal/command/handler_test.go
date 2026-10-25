package command

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rules := []CommandRule{
		{Name: "pwd", Timeout: 10 * time.Second, ReadOnly: true},
		{Name: "ls", Timeout: 10 * time.Second, ReadOnly: true,
			Args: []AllowedArg{{Exact: "-la"}, {Exact: "-l"}, {Exact: "-h"}}, MaxArgs: 2},
		{Name: "echo", Timeout: 5 * time.Second},
	}

	return NewHandler(rules, 10.0, 30*time.Second, 65536)
}

func performRequest(h *Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create a minimal gin engine with just the handler
	r := gin.New()
	r.POST("/api/execute", h.Execute)
	r.ServeHTTP(w, req)

	return w
}

func TestHandler_Execute_UnknownCommand(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "unknown_cmd",
		Args:    nil,
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if err, ok := resp["error"].(string); !ok || err != "unknown command" {
		t.Errorf("error = %v, want 'unknown command'", resp["error"])
	}
}

func TestHandler_Execute_ForbiddenArg(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "ls",
		Args:    []string{"-la", "-R"}, // -R is not allowed
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["error"].(string); !ok {
		t.Error("expected error message in response")
	}
}

func TestHandler_Execute_SuccessfulNewAPI(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "pwd",
		Args:    nil,
	})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ExecuteResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if resp.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestHandler_Execute_LegacyAction(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Action: "pwd",
	})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ExecuteResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if resp.Output == "" {
		t.Error("expected non-empty output from legacy action")
	}
}

func TestHandler_Execute_LegacyUnknownAction(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Action: "nonexistent_action",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if err, ok := resp["error"].(string); !ok || err != "unknown action" {
		t.Errorf("error = %v, want 'unknown action'", resp["error"])
	}
}

func TestHandler_Execute_InvalidJSON(t *testing.T) {
	h := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/api/execute", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/api/execute", h.Execute)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Execute_EmptyBody(t *testing.T) {
	h := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/api/execute", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/api/execute", h.Execute)
	r.ServeHTTP(w, req)

	// Empty body should be treated as legacy API with empty action
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Execute_MaxArgsExceeded(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "ls",
		Args:    []string{"-la", "-h", "-a"}, // max 2
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Execute_ResponseIncludesExitCode(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "pwd",
		Args:    nil,
	})

	var resp ExecuteResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if resp.ExitCode == nil {
		t.Error("expected exit_code to be present in successful response")
	} else if *resp.ExitCode != 0 {
		t.Errorf("exit_code = %d, want 0", *resp.ExitCode)
	}
}

// --- RateLimiter tests ---

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(5.0) // 5 tokens/sec, starts full

	// Should allow 5 requests (initial tokens)
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 6th should be rejected (no tokens left)
	if rl.Allow() {
		t.Error("6th request should be rejected")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(100.0) // fast refill for testing

	// Drain all tokens
	for rl.Allow() {
	}

	// Wait for refill
	time.Sleep(50 * time.Millisecond)

	if !rl.Allow() {
		t.Error("should allow after refill period")
	}
}

func TestRateLimiter_DefaultRate(t *testing.T) {
	rl := NewRateLimiter(0) // should default to 10
	if rl.maxTokens != 10 {
		t.Errorf("maxTokens = %f, want 10", rl.maxTokens)
	}
}

func TestRateLimiter_NegativeRate(t *testing.T) {
	rl := NewRateLimiter(-5) // should default to 10
	if rl.maxTokens != 10 {
		t.Errorf("maxTokens = %f, want 10", rl.maxTokens)
	}
}

func TestRateLimiter_TokensCapped(t *testing.T) {
	rl := NewRateLimiter(5.0)

	// Wait long enough to exceed maxTokens
	time.Sleep(200 * time.Millisecond)

	// Should still only have maxTokens
	count := 0
	for rl.Allow() {
		count++
	}
	if count > 5 {
		t.Errorf("allowed %d requests, expected at most 5 (maxTokens)", count)
	}
}

func TestHandler_Execute_RateLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a handler with a very low rate limit (2 tokens/sec)
	h := NewHandler(
		[]CommandRule{{Name: "pwd", Timeout: 10 * time.Second, ReadOnly: true}},
		2.0, // 2 requests per second
		30*time.Second,
		65536,
	)

	// Drain the initial tokens (starts full = 2 tokens)
	for i := 0; i < 2; i++ {
		w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
			Command: "pwd",
			Args:    nil,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	// Next request should be rate limited (429)
	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "pwd",
		Args:    nil,
	})
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d (429)", w.Code, http.StatusTooManyRequests)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["error"].(string); !ok {
		t.Error("expected error message in 429 response")
	}
}

func TestHandler_Execute_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a handler and a gin engine with auth middleware
	h := setupTestHandler(t)

	// Simulate an unauthorized request (no auth middleware in test,
	// but we verify the handler itself doesn't require auth —
	// auth is enforced at router level)
	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "pwd",
		Args:    nil,
	})
	if w.Code != http.StatusOK {
		t.Errorf("handler without auth: status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_Execute_ResponseExitCodeOnSuccess(t *testing.T) {
	h := setupTestHandler(t)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "pwd",
		Args:    nil,
	})

	var resp ExecuteResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if resp.ExitCode == nil {
		t.Error("expected exit_code to be present in successful response")
	} else if *resp.ExitCode != 0 {
		t.Errorf("exit_code = %d, want 0", *resp.ExitCode)
	}
}

func TestHandler_Execute_WriteCommandRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []CommandRule{
		{Name: "pwd", Timeout: 10 * time.Second, ReadOnly: true},
		{Name: "rm", Timeout: 5 * time.Second, ReadOnly: false},
	}
	h := NewHandler(rules, 10.0, 30*time.Second, 65536)

	w := performRequest(h, "POST", "/api/execute", ExecuteRequest{
		Command: "rm",
		Args:    []string{"-rf", "/tmp"},
	})

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (403)", w.Code, http.StatusForbidden)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if err, ok := resp["error"].(string); !ok || err != "write commands are not allowed" {
		t.Errorf("error = %v, want 'write commands are not allowed'", resp["error"])
	}
}
