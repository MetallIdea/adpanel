package command

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

// --- limitedWriter tests ---

func TestLimitedWriter_ExactLimit(t *testing.T) {
	w := newLimitedWriter(10)

	n, err := w.Write([]byte("0123456789"))
	if n != 10 {
		t.Errorf("Write(10 bytes) returned n=%d, want 10", n)
	}
	if err != nil {
		t.Errorf("Write(10 bytes) returned error: %v", err)
	}
	if w.overflow {
		t.Error("overflow should be false at exact limit")
	}
	if len(w.data) != 10 {
		t.Errorf("data length = %d, want 10", len(w.data))
	}
	if w.overflowCount != 0 {
		t.Errorf("overflowCount = %d, want 0", w.overflowCount)
	}
}

func TestLimitedWriter_Overflow(t *testing.T) {
	w := newLimitedWriter(5)

	// Write 3 bytes — fits
	n, _ := w.Write([]byte("abc"))
	if n != 3 || len(w.data) != 3 {
		t.Errorf("after 3 bytes: len=%d, want 3", len(w.data))
	}

	// Write 5 bytes — 2 fit, 3 overflow
	n, _ = w.Write([]byte("01234"))
	if n != 5 {
		t.Errorf("Write(5) returned n=%d, want 5", n)
	}
	if len(w.data) != 5 {
		t.Errorf("data length = %d, want 5", len(w.data))
	}
	if !w.overflow {
		t.Error("overflow should be true after partial overflow")
	}
	if w.overflowCount != 3 {
		t.Errorf("overflowCount = %d, want 3", w.overflowCount)
	}

	// Write 10 more bytes — all discarded
	n, _ = w.Write(bytes.Repeat([]byte("x"), 10))
	if n != 10 {
		t.Errorf("Write(10 more) returned n=%d, want 10", n)
	}
	if w.overflowCount != 13 {
		t.Errorf("overflowCount = %d, want 13", w.overflowCount)
	}
}

func TestLimitedWriter_ZeroLimit(t *testing.T) {
	w := newLimitedWriter(0)

	n, _ := w.Write([]byte("hello"))
	if n != 5 {
		t.Errorf("Write(5) returned n=%d, want 5", n)
	}
	if len(w.data) != 0 {
		t.Errorf("data length = %d, want 0", len(w.data))
	}
	if !w.overflow {
		t.Error("overflow should be true with zero limit")
	}
	if w.overflowCount != 5 {
		t.Errorf("overflowCount = %d, want 5", w.overflowCount)
	}
}

func TestLimitedWriter_NilWrite(t *testing.T) {
	w := newLimitedWriter(100)

	n, err := w.Write(nil)
	if n != 0 {
		t.Errorf("Write(nil) returned n=%d, want 0", n)
	}
	if err != nil {
		t.Errorf("Write(nil) returned error: %v", err)
	}
}

// TestLimitedWriter_ConcurrentSafety verifies that concurrent Writes
// do not corrupt internal state.
func TestLimitedWriter_ConcurrentSafety(t *testing.T) {
	w := newLimitedWriter(1000)
	const goroutines = 10
	const perGoroutine = 100

	done := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < perGoroutine; j++ {
				w.Write([]byte("x"))
			}
			done <- true
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Total written = goroutines * perGoroutine
	totalWritten := goroutines * perGoroutine
	if len(w.data)+int(w.overflowCount) != totalWritten {
		t.Errorf("total bytes (data+overflow) = %d, want %d",
			len(w.data)+int(w.overflowCount), totalWritten)
	}
}

// --- ExecuteCommandWithContext tests ---

func TestExecuteCommandWithContext_InvalidMaxOutput(t *testing.T) {
	ctx := context.Background()

	_, err := ExecuteCommandWithContext(ctx, "pwd", nil, 10*time.Second, 0)
	if err == nil {
		t.Error("expected error for maxOutput=0")
	}

	_, err = ExecuteCommandWithContext(ctx, "pwd", nil, 10*time.Second, -1)
	if err == nil {
		t.Error("expected error for maxOutput=-1")
	}

	_, err = ExecuteCommandWithContext(ctx, "pwd", nil, 10*time.Second, 1024*1024+1)
	if err == nil {
		t.Error("expected error for maxOutput > 1MB")
	}
}

func TestExecuteCommandWithContext_InvalidTimeout(t *testing.T) {
	ctx := context.Background()

	_, err := ExecuteCommandWithContext(ctx, "pwd", nil, 0, 1024)
	if err == nil {
		t.Error("expected error for timeout=0")
	}

	_, err = ExecuteCommandWithContext(ctx, "pwd", nil, -5*time.Second, 1024)
	if err == nil {
		t.Error("expected error for negative timeout")
	}
}

func TestExecuteCommandWithContext_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before calling

	_, err := ExecuteCommandWithContext(ctx, "pwd", nil, 10*time.Second, 1024)
	if err != ErrContextCanceled {
		t.Errorf("expected ErrContextCanceled, got %v", err)
	}
}

func TestExecuteCommandWithContext_UnknownCommand(t *testing.T) {
	ctx := context.Background()

	result, err := ExecuteCommandWithContext(ctx, "nonexistent_binary_xyz", nil, 5*time.Second, 1024)
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
	if result == nil {
		t.Fatal("expected non-nil result even on launch error")
	}
	if result.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1", result.ExitCode)
	}
	if result.LaunchErr == nil {
		t.Error("LaunchErr should be set")
	}
}

func TestExecuteCommandWithContext_SuccessfulPwd(t *testing.T) {
	ctx := context.Background()

	result, err := ExecuteCommandWithContext(ctx, "pwd", nil, 10*time.Second, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if len(result.Stdout) == 0 {
		t.Error("expected non-empty stdout from pwd")
	}
	if result.TimeoutErr {
		t.Error("TimeoutErr should be false for successful command")
	}
}

func TestExecuteCommandWithContext_NonZeroExitCode(t *testing.T) {
	ctx := context.Background()

	// false always exits with code 1
	result, err := ExecuteCommandWithContext(ctx, "false", nil, 10*time.Second, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
	if result.WaitErr == nil {
		t.Error("WaitErr should be set for non-zero exit")
	}
	if result.TimeoutErr {
		t.Error("TimeoutErr should be false for normal exit")
	}
}

func TestExecuteCommandWithContext_Timeout(t *testing.T) {
	ctx := context.Background()

	// sleep 10s with 100ms timeout — should time out
	result, err := ExecuteCommandWithContext(ctx, "sleep", []string{"10"}, 100*time.Millisecond, 1024)
	if err != ErrCommandTimeout {
		t.Errorf("expected ErrCommandTimeout, got %v", err)
	}
	if !result.TimeoutErr {
		t.Error("TimeoutErr should be true")
	}
	if result.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1 for timed-out command", result.ExitCode)
	}
}

func TestExecuteCommandWithContext_OutputLimit(t *testing.T) {
	ctx := context.Background()

	// dd if=/dev/urandom bs=1024 count=2 produces ~2KB
	result, err := ExecuteCommandWithContext(
		ctx,
		"dd",
		[]string{"if=/dev/urandom", "bs=1024", "count=2"},
		10*time.Second,
		1024, // 1KB limit
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Stdout) > 1024 {
		t.Errorf("stdout length = %d, expected <= 1024", len(result.Stdout))
	}
	if !result.HasOverflowStdout {
		t.Error("HasOverflowStdout should be true when output exceeds limit")
	}
}

// TestExecuteCommandWithContext_ParentContextCancelMidway tests that
// canceling the parent context during execution returns ErrContextCanceled.
func TestExecuteCommandWithContext_ParentContextCancelMidway(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan *ExecutionResult, 1)
	go func() {
		result, _ := ExecuteCommandWithContext(ctx, "sleep", []string{"10"}, 30*time.Second, 1024)
		done <- result
	}()

	// Let the command start, then cancel
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case result := <-done:
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for goroutine to finish")
	}
}

// TestExecuteCommandWithContext_LegacyExecute tests the Execute wrapper.
func TestExecuteCommandWithContext_LegacyExecute(t *testing.T) {
	rules := []CommandRule{
		{Name: "pwd", Timeout: 10 * time.Second, ReadOnly: true},
	}

	output, err := Execute("pwd", rules, 30*time.Second, 65536)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output from pwd")
	}
}

func TestExecuteCommandWithContext_LegacyExecute_UnknownAction(t *testing.T) {
	rules := []CommandRule{
		{Name: "pwd", Timeout: 10 * time.Second},
	}

	_, err := Execute("nonexistent", rules, 30*time.Second, 65536)
	if err == nil {
		t.Error("expected error for unknown action")
	}
}

func TestExecuteCommandWithContext_LegacyExecute_EmptyAction(t *testing.T) {
	rules := []CommandRule{
		{Name: "pwd", Timeout: 10 * time.Second},
	}

	_, err := Execute("", rules, 30*time.Second, 65536)
	if err == nil {
		t.Error("expected error for empty action")
	}
}

func TestExecuteCommandWithContext_LegacyExecute_WriteCommandRejected(t *testing.T) {
	rules := []CommandRule{
		{Name: "pwd", Timeout: 10 * time.Second, ReadOnly: true},
		{Name: "rm", Timeout: 5 * time.Second, ReadOnly: false},
	}

	_, err := Execute("rm", rules, 30*time.Second, 65536)
	if err == nil {
		t.Fatal("expected error for write command via legacy API")
	}
}

// --- ParseTimeout and ParseMaxOutput tests ---

func TestParseTimeout(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"10s", 10 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"1h", time.Hour, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseTimeout(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeout(%q) error = %v, wantErr %v",
					tt.input, err, tt.wantErr)
				return
			}
			if err == nil && d != tt.want {
				t.Errorf("ParseTimeout(%q) = %v, want %v",
					tt.input, d, tt.want)
			}
		})
	}
}

func TestParseMaxOutput(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"", 65536, false}, // default
		{"1024", 1024, false},
		{"1048576", 1048576, false}, // 1MB exact
		{"1048577", 0, true},        // exceeds 1MB
		{"0", 0, true},              // zero
		{"-1", 0, true},             // negative
		{"abc", 0, true},            // non-numeric
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			n, err := ParseMaxOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMaxOutput(%q) error = %v, wantErr %v",
					tt.input, err, tt.wantErr)
				return
			}
			if err == nil && n != tt.want {
				t.Errorf("ParseMaxOutput(%q) = %d, want %d",
					tt.input, n, tt.want)
			}
		})
	}
}

// --- io.Writer interface compliance ---

func TestLimitedWriter_IoWriterInterface(t *testing.T) {
	var w io.Writer = newLimitedWriter(50)
	if _, err := w.Write(bytes.Repeat([]byte("x"), 100)); err != nil {
		t.Errorf("io.Writer.Write returned error: %v", err)
	}
	lw := w.(*limitedWriter)
	if len(lw.data) > 50 {
		t.Errorf("data length = %d, want <= 50", len(lw.data))
	}
}
