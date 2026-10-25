package command

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidateArgs(t *testing.T) {
	rule := &CommandRule{
		Name: "ls",
		Args: []AllowedArg{
			{Exact: "-la"},
			{Exact: "-l"},
			{Exact: "-h"},
		},
		MaxArgs: 2,
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no arguments", nil, false},
		{"one allowed argument", []string{"-la"}, false},
		{"two allowed arguments", []string{"-la", "-h"}, false},
		{"unknown argument", []string{"-la", "/etc"}, true},
		{"too many arguments", []string{"-la", "-l", "-h"}, true},
		{"combined argument is rejected", []string{"-la -h"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArgs(rule, tt.args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateArgs(%v) error = %v, wantErr %v",
					tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestCommandRuleTimeoutParsing(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:  "duration string",
			input: `{"name":"pwd","timeout":"10s"}`,
			want:  10 * time.Second,
		},
		{
			name:  "duration number in nanoseconds",
			input: `{"name":"pwd","timeout":10000000000}`,
			want:  10 * time.Second,
		},
		{
			name:    "invalid duration",
			input:   `{"name":"pwd","timeout":"invalid"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule CommandRule
			err := json.Unmarshal([]byte(tt.input), &rule)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && rule.Timeout != tt.want {
				t.Fatalf("Timeout = %v, want %v", rule.Timeout, tt.want)
			}
		})
	}
}

// --- Additional ValidateArgs edge cases ---

func TestValidateArgs_MaxArgsZero(t *testing.T) {
	// MaxArgs == 0 means no limit on argument count
	rule := &CommandRule{
		Name:    "cmd",
		MaxArgs: 0,
		Args:    []AllowedArg{{Exact: "-v"}},
	}

	// With MaxArgs=0 and Args defined, each arg must match the pattern
	err := ValidateArgs(rule, []string{"-v", "-v"})
	if err != nil {
		t.Errorf("MaxArgs=0 with matching args should pass: %v", err)
	}

	// But non-matching args should still fail
	err = ValidateArgs(rule, []string{"-v", "-x"})
	if err == nil {
		t.Error("expected error for non-matching arg even with MaxArgs=0")
	}
}

func TestValidateArgs_MaxArgsZeroNoArgs(t *testing.T) {
	// MaxArgs == 0, no Args defined — len(rule.Args)==0 means no args allowed
	rule := &CommandRule{
		Name:    "cmd",
		MaxArgs: 0,
		Args:    nil,
	}

	err := ValidateArgs(rule, []string{"anything"})
	if err == nil {
		t.Error("expected error: empty Args means no arguments allowed regardless of MaxArgs")
	}
}

func TestValidateArgs_NilRule(t *testing.T) {
	err := ValidateArgs(nil, []string{"x"})
	if err == nil {
		t.Error("expected error for nil rule")
	}
}

func TestValidateArgs_EmptyRuleName(t *testing.T) {
	rule := &CommandRule{
		Name: "",
		Args: []AllowedArg{{Exact: "-a"}},
	}
	err := ValidateArgs(rule, []string{"-a"})
	if err != nil {
		t.Errorf("empty rule name should not affect arg validation: %v", err)
	}
}

func TestValidateArgs_PatternMatch(t *testing.T) {
	rule := &CommandRule{
		Name: "cat",
		Args: []AllowedArg{{Pattern: "/etc/hosts"}},
	}

	err := ValidateArgs(rule, []string{"/etc/hosts"})
	if err != nil {
		t.Errorf("pattern match should succeed: %v", err)
	}

	err = ValidateArgs(rule, []string{"/etc/passwd"})
	if err == nil {
		t.Error("pattern should not match /etc/passwd")
	}
}

func TestValidateArgs_ArgsEmptyBothFields(t *testing.T) {
	// An AllowedArg with both Exact and Pattern empty does NOT match anything.
	// validateRule should catch this, but ValidateArgs doesn't re-validate.
	rule := &CommandRule{
		Name: "cmd",
		Args: []AllowedArg{{Exact: "", Pattern: ""}},
	}

	err := ValidateArgs(rule, []string{"-a"})
	if err == nil {
		t.Error("expected error: empty exact+pattern does not match any argument")
	}
}

// --- LoadWhitelist tests ---

func TestLoadWhitelist_DefaultRules(t *testing.T) {
	// Ensure COMMAND_WHITELIST env is empty
	t.Setenv("COMMAND_WHITELIST", "")

	rules, err := LoadWhitelist()
	if err != nil {
		t.Fatalf("LoadWhitelist() error: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("expected default rules, got none")
	}

	// Check that pwd is in defaults
	found := false
	for _, r := range rules {
		if r.Name == "pwd" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'pwd' in default rules")
	}
}

func TestLoadWhitelist_Environment(t *testing.T) {
	jsonConfig := `[{"name":"echo","timeout":"5s","read_only":true}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	rules, err := LoadWhitelist()
	if err != nil {
		t.Fatalf("LoadWhitelist() error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "echo" {
		t.Errorf("rule name = %q, want %q", rules[0].Name, "echo")
	}
	if !rules[0].ReadOnly {
		t.Error("expected ReadOnly=true")
	}
}

func TestLoadWhitelist_InvalidJSON(t *testing.T) {
	t.Setenv("COMMAND_WHITELIST", "not json")

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadWhitelist_EmptyArray(t *testing.T) {
	t.Setenv("COMMAND_WHITELIST", "[]")

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for empty JSON array")
	}
}

func TestLoadWhitelist_DuplicateNames(t *testing.T) {
	jsonConfig := `[
		{"name":"pwd","timeout":"5s"},
		{"name":"pwd","timeout":"10s"}
	]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for duplicate rule names")
	}
}

func TestLoadWhitelist_EmptyRuleName(t *testing.T) {
	jsonConfig := `[{"name":"","timeout":"5s"}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for empty rule name")
	}
}

func TestLoadWhitelist_NegativeTimeout(t *testing.T) {
	jsonConfig := `[{"name":"pwd","timeout":"-5s"}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for negative timeout")
	}
}

func TestLoadWhitelist_ZeroTimeout(t *testing.T) {
	// Zero timeout should be rejected — it would fail at execution time
	// with a confusing error message
	jsonConfig := `[{"name":"pwd","timeout":"0s"}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for zero timeout")
	}
}

func TestLoadWhitelist_NegativeMaxArgs(t *testing.T) {
	jsonConfig := `[{"name":"pwd","max_args":-1}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for negative max_args")
	}
}

func TestLoadWhitelist_InvalidGlobPattern(t *testing.T) {
	// Pattern with ".." should be rejected
	jsonConfig := `[{"name":"cat","args":[{"pattern":"../etc/passwd"}]}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for pattern with '..'")
	}
}

func TestLoadWhitelist_RelativePattern(t *testing.T) {
	// Non-absolute pattern should be rejected
	jsonConfig := `[{"name":"cat","args":[{"pattern":"relative/path"}]}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for relative pattern")
	}
}

func TestLoadWhitelist_CommandNameWithWhitespace(t *testing.T) {
	jsonConfig := `[{"name":"ls /tmp","timeout":"5s"}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for command name with whitespace")
	}
}

func TestLoadWhitelist_CommandNameWithPathSeparator(t *testing.T) {
	jsonConfig := `[{"name":"etc/passwd","timeout":"5s"}]`
	t.Setenv("COMMAND_WHITELIST", jsonConfig)

	_, err := LoadWhitelist()
	if err == nil {
		t.Error("expected error for command name with path separator")
	}
}

// --- ValidateConfig tests ---

func TestValidateConfig(t *testing.T) {
	rules := []CommandRule{{Name: "pwd", Timeout: 10 * time.Second}}

	tests := []struct {
		name      string
		timeout   time.Duration
		maxOutput int
		rateLimit float64
		rules     []CommandRule
		wantErr   bool
	}{
		{"valid config", 30 * time.Second, 65536, 10.0, rules, false},
		{"zero timeout", 0, 65536, 10.0, rules, true},
		{"negative timeout", -5 * time.Second, 65536, 10.0, rules, true},
		{"zero maxOutput", 30 * time.Second, 0, 10.0, rules, true},
		{"negative maxOutput", 30 * time.Second, -1, 10.0, rules, true},
		{"maxOutput > 1MB", 30 * time.Second, 1024*1024 + 1, 10.0, rules, true},
		{"zero rateLimit", 30 * time.Second, 65536, 0, rules, true},
		{"negative rateLimit", 30 * time.Second, 65536, -1, rules, true},
		{"empty rules", 30 * time.Second, 65536, 10.0, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.timeout, tt.maxOutput, tt.rateLimit, tt.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- findRule tests ---

func TestFindRule(t *testing.T) {
	rules := []CommandRule{
		{Name: "pwd", Timeout: 10 * time.Second},
		{Name: "ls", Timeout: 5 * time.Second},
	}

	rule := findRule(rules, "pwd")
	if rule == nil || rule.Name != "pwd" {
		t.Error("expected to find 'pwd' rule")
	}

	if findRule(rules, "unknown") != nil {
		t.Error("expected nil for unknown command")
	}

	if findRule(nil, "pwd") != nil {
		t.Error("expected nil for nil rules slice")
	}
}
