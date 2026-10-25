package command

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ExecutionResult holds the outcome of a command execution.
type ExecutionResult struct {
	Stdout            []byte
	Stderr            []byte
	ExitCode          int
	WaitErr           error
	TimeoutErr        bool
	LaunchErr         error
	HasOverflowStdout bool
	HasOverflowStderr bool
}

// limitedWriter collects at most limit bytes and discards the rest.
// It never blocks and never returns an error — overflow is tracked via flags.
type limitedWriter struct {
	mu            sync.Mutex
	data          []byte
	limit         int
	overflow      bool
	overflowCount int64
}

func newLimitedWriter(limit int) *limitedWriter {
	return &limitedWriter{
		data:     make([]byte, 0, limit),
		limit:    limit,
		overflow: false,
	}
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n := len(p)
	remaining := w.limit - len(w.data)
	if remaining <= 0 {
		// Already over limit — discard everything
		w.overflow = true
		w.overflowCount += int64(n)
		return n, nil
	}
	if remaining >= n {
		// Fits entirely
		w.data = append(w.data, p...)
		return n, nil
	}
	// Partially fits
	w.data = append(w.data, p[:remaining]...)
	w.overflow = true
	w.overflowCount += int64(n - remaining)
	return n, nil
}

// ErrCommandTimeout is returned when a command exceeds its time budget.
var ErrCommandTimeout = errors.New("command timed out")

// ErrContextCanceled is returned when the parent context is canceled before timeout.
var ErrContextCanceled = errors.New("context canceled before command completed")

// ExecuteCommandWithContext executes a whitelisted command with strict resource limits.
//
// Parameters:
//   - ctx: parent context (used for cancellation, NOT for timeout — timeout is derived from rule)
//   - cmdName: exact executable name from whitelist
//   - args: command arguments (already validated by ValidateArgs)
//   - timeout: command-specific timeout duration
//   - maxOutput: maximum bytes to retain from stdout/stderr

func ExecuteCommandWithContext(parentCtx context.Context, cmdName string, args []string, timeout time.Duration, maxOutput int) (*ExecutionResult, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}
	if maxOutput <= 0 || maxOutput > 1024*1024 {
		return nil, fmt.Errorf("maxOutput must be between 1 and 1MB")
	}
	if err := parentCtx.Err(); err != nil {
		return nil, ErrContextCanceled
	}

	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	stdoutW := newLimitedWriter(maxOutput)
	stderrW := newLimitedWriter(maxOutput)

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	if err := cmd.Start(); err != nil {
		launchErr := fmt.Errorf("failed to start %q: %w", cmdName, err)
		return &ExecutionResult{
			ExitCode:  -1,
			WaitErr:   launchErr,
			LaunchErr: launchErr,
			Stdout:    stdoutW.data,
			Stderr:    stderrW.data,
		}, launchErr
	}

	waitErr := cmd.Wait()

	exitCode := 0
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	result := &ExecutionResult{
		Stdout:            stdoutW.data,
		Stderr:            stderrW.data,
		ExitCode:          exitCode,
		WaitErr:           waitErr,
		HasOverflowStdout: stdoutW.overflow,
		HasOverflowStderr: stderrW.overflow,
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.TimeoutErr = true
		return result, ErrCommandTimeout
	}
	if parentCtx.Err() != nil {
		return result, ErrContextCanceled
	}

	// A command that ran but returned a non-zero exit code is still a
	// completed execution. The caller can inspect ExitCode and WaitErr.
	return result, nil
}

// Execute is the legacy wrapper that preserves compatibility with the old
// { "action": "pwd" } API. It maps legacy action names to whitelist rules
// and runs them through the same safe executor — no bypass of whitelist,
// timeouts, or output limits.
func Execute(action string, rules []CommandRule, timeout time.Duration, maxOutput int) (string, error) {
	if action == "" {
		return "", fmt.Errorf("action is required")
	}

	rule := findRule(rules, action)
	if rule == nil {
		return "", fmt.Errorf("unknown action: %s", action)
	}

	if !rule.ReadOnly {
		return "", fmt.Errorf("write commands are not allowed")
	}

	// Use the rule's timeout, or fall back to the default.
	cmdTimeout := rule.Timeout
	if cmdTimeout == 0 {
		cmdTimeout = timeout
	}

	result, err := ExecuteCommandWithContext(context.Background(), rule.Name, []string{}, cmdTimeout, maxOutput)

	output := ""
	if result != nil && len(result.Stdout) > 0 {
		output = strings.TrimRight(string(result.Stdout), "\n")
	}

	// Append overflow warning.
	if result != nil && (result.HasOverflowStdout || result.HasOverflowStderr) {
		output += "\n[output truncated, exceeded " + fmt.Sprintf("%d", maxOutput) + " bytes]"
	}

	return output, err
}

// ParseTimeout converts a string duration (e.g. "30s", "5m") to time.Duration.
// Returns an error if the format is invalid.
func ParseTimeout(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: %w", s, err)
	}
	return d, nil
}

// ParseMaxOutput converts a string to a positive integer with a 1MB upper bound.
func ParseMaxOutput(s string) (int, error) {
	if s == "" {
		return 65536, nil // default 64 KiB
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid maxOutput %q: %w", s, err)
	}

	if n <= 0 {
		return 0, fmt.Errorf("maxOutput must be positive, got %d", n)
	}

	if n > 1024*1024 {
		return 0, fmt.Errorf("maxOutput must not exceed 1MB, got %d", n)
	}

	return n, nil
}
