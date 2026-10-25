package command

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a token bucket algorithm for rate limiting.
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
}

// NewRateLimiter creates a rate limiter with the given tokens per second.
func NewRateLimiter(tokensPerSecond float64) *RateLimiter {
	if tokensPerSecond <= 0 {
		tokensPerSecond = 10
	}

	return &RateLimiter{
		tokens:     tokensPerSecond,
		maxTokens:  tokensPerSecond,
		refillRate: tokensPerSecond,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under the current rate limit.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now

	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

// Handler wraps command execution with rate limiting.
type Handler struct {
	rules       []CommandRule
	rateLimiter *RateLimiter
	timeout     time.Duration
	maxOutput   int
}

// NewHandler creates a new command handler with the given configuration.
func NewHandler(
	rules []CommandRule,
	rateLimit float64,
	timeout time.Duration,
	maxOutput int,
) *Handler {
	return &Handler{
		rules:       rules,
		rateLimiter: NewRateLimiter(rateLimit),
		timeout:     timeout,
		maxOutput:   maxOutput,
	}
}

// Execute handles POST /api/execute.
func (h *Handler) Execute(c *gin.Context) {
	if !h.rateLimiter.Allow() {
		c.JSON(429, gin.H{
			"error": "rate limit exceeded",
		})
		return
	}

	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "invalid JSON request",
		})
		return
	}

	if req.Command != "" {
		h.executeNewAPI(c, req)
		return
	}

	// Legacy API: { "action": "pwd" }.
	output, err := Execute(req.Action, h.rules, h.timeout, h.maxOutput)
	if err != nil {
		// Do not expose internal error details to the client.
		message := "command execution failed"
		if strings.Contains(err.Error(), "unknown action") {
			message = "unknown action"
		} else if strings.Contains(err.Error(), "is required") {
			message = "action is required"
		}
		c.JSON(400, gin.H{
			"error": message,
		})
		return
	}

	c.JSON(200, ExecuteResponse{
		Output: output,
	})
}

// executeNewAPI handles { "command": "pwd", "args": [] }.
func (h *Handler) executeNewAPI(c *gin.Context, req ExecuteRequest) {
	rule := findRule(h.rules, req.Command)
	if rule == nil {
		c.JSON(400, gin.H{
			"error": "unknown command",
		})
		return
	}

	if !rule.ReadOnly {
		c.JSON(403, gin.H{
			"error": "write commands are not allowed",
		})
		return
	}

	if err := ValidateArgs(rule, req.Args); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	cmdTimeout := rule.Timeout
	if cmdTimeout <= 0 {
		cmdTimeout = h.timeout
	}

	result, err := ExecuteCommandWithContext(
		c.Request.Context(),
		rule.Name,
		req.Args,
		cmdTimeout,
		h.maxOutput,
	)

	output := ""
	var exitCode *int

	if result != nil {
		output = string(result.Stdout)

		if result.HasOverflowStdout || result.HasOverflowStderr {
			output += fmt.Sprintf(
				"\n[output truncated, exceeded %d bytes per stream]",
				h.maxOutput,
			)
		}

		if result.ExitCode >= 0 {
			code := result.ExitCode
			exitCode = &code
		}
	}

	if err != nil {
		status := 500

		switch {
		case errors.Is(err, ErrCommandTimeout):
			status = 504
		case errors.Is(err, ErrContextCanceled),
			c.Request.Context().Err() != nil:
			status = 408
		}

		// Do not expose internal process/OS error details to the client.
		message := "command execution failed"
		if status == 504 {
			message = "command timed out"
		} else if status == 408 {
			message = "request canceled"
		}

		c.JSON(status, gin.H{
			"error":     message,
			"output":    output,
			"exit_code": exitCode,
		})
		return
	}

	c.JSON(200, ExecuteResponse{
		Output:   output,
		ExitCode: exitCode,
	})
}
