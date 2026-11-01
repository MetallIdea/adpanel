package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AllowedArg describes one allowed argument for a whitelisted command.
type AllowedArg struct {
	Exact   string `json:"exact"`
	Pattern string `json:"pattern"`
}

// CommandRule defines a safe execution rule for a single command.
type CommandRule struct {
	Name     string        `json:"name"`
	Args     []AllowedArg  `json:"args"`
	MaxArgs  int           `json:"max_args"`
	Timeout  time.Duration `json:"timeout"`
	ReadOnly bool          `json:"read_only"`
}

// UnmarshalJSON supports timeout as either a duration string (e.g. "10s")
// or an integer representing nanoseconds.
func (r *CommandRule) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name     string          `json:"name"`
		Args     []AllowedArg    `json:"args"`
		MaxArgs  int             `json:"max_args"`
		Timeout  json.RawMessage `json:"timeout"`
		ReadOnly bool            `json:"read_only"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var timeout time.Duration
	if len(raw.Timeout) > 0 && string(raw.Timeout) != "null" {
		var durationString string
		if err := json.Unmarshal(raw.Timeout, &durationString); err == nil {
			parsed, err := time.ParseDuration(durationString)
			if err != nil {
				return fmt.Errorf("invalid timeout %q: %w", durationString, err)
			}
			timeout = parsed
		} else {
			var nanoseconds int64
			if err := json.Unmarshal(raw.Timeout, &nanoseconds); err != nil {
				return fmt.Errorf("timeout must be a duration string or integer nanoseconds: %w", err)
			}
			timeout = time.Duration(nanoseconds)
		}
	}

	r.Name = raw.Name
	r.Args = raw.Args
	r.MaxArgs = raw.MaxArgs
	r.Timeout = timeout
	r.ReadOnly = raw.ReadOnly

	return nil
}

// Default rules used when COMMAND_WHITELIST is not set or empty.
func defaultRules() []CommandRule {
	return []CommandRule{
		{
			Name:     "pwd",
			Timeout:  10 * time.Second,
			ReadOnly: true,
		},
		{
			Name:     "whoami",
			Timeout:  10 * time.Second,
			ReadOnly: true,
		},
		{
			Name:     "hostname",
			Timeout:  10 * time.Second,
			ReadOnly: true,
		},
		{
			Name:     "date",
			Timeout:  10 * time.Second,
			ReadOnly: true,
		},
		{
			Name:     "uptime",
			Timeout:  10 * time.Second,
			ReadOnly: true,
		},
		{
			Name: "ls",
			Args: []AllowedArg{
				{Exact: "-la"},
				{Exact: "-l"},
				{Exact: "-h"},
			},
			MaxArgs:  2,
			Timeout:  10 * time.Second,
			ReadOnly: true,
		},
	}
}

// ValidateConfig checks all configuration values and rules.
func ValidateConfig(
	timeout time.Duration,
	maxOutput int,
	rateLimit float64,
	rules []CommandRule,
) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", timeout)
	}

	if maxOutput <= 0 || maxOutput > 1024*1024 {
		return fmt.Errorf("maxOutput must be between 1 and 1MB, got %d", maxOutput)
	}

	if rateLimit <= 0 {
		return fmt.Errorf("rateLimit must be positive, got %f", rateLimit)
	}

	if len(rules) == 0 {
		return fmt.Errorf("whitelist must not be empty")
	}

	seen := make(map[string]bool, len(rules))
	for i, rule := range rules {
		if err := validateRule(rule); err != nil {
			return fmt.Errorf("rule[%d]: %w", i, err)
		}

		if seen[rule.Name] {
			return fmt.Errorf("duplicate command rule: %q", rule.Name)
		}
		seen[rule.Name] = true
	}

	return nil
}

// validateRule checks a single command rule for invalid values.
func validateRule(rule CommandRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name must not be empty")
	}

	if strings.ContainsAny(rule.Name, "/\\ \t\r\n") {
		return fmt.Errorf("rule name %q must not contain path separators or whitespace", rule.Name)
	}

	if rule.MaxArgs < 0 {
		return fmt.Errorf("max_args must be non-negative, got %d", rule.MaxArgs)
	}

	if rule.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", rule.Timeout)
	}

	for j, arg := range rule.Args {
		if arg.Exact == "" && arg.Pattern == "" {
			return fmt.Errorf("args[%d]: both exact and pattern are empty", j)
		}

		if arg.Pattern != "" {
			if err := validateGlobPattern(arg.Pattern); err != nil {
				return fmt.Errorf("args[%d] pattern: %w", j, err)
			}
		}
	}

	return nil
}

// validateGlobPattern checks that a pattern is absolute and syntactically valid.
func validateGlobPattern(pattern string) error {
	if strings.Contains(pattern, "..") {
		return fmt.Errorf("pattern must not contain '..'")
	}

	if !filepath.IsAbs(pattern) {
		return fmt.Errorf("pattern must be an absolute path, got %q", pattern)
	}

	if _, err := filepath.Match(pattern, ""); err != nil {
		return fmt.Errorf("invalid glob pattern: %w", err)
	}

	return nil
}

// LoadWhitelist reads COMMAND_WHITELIST from the environment.
// If the variable is empty, the default rules are used.
func LoadWhitelist() ([]CommandRule, error) {
	env := os.Getenv("COMMAND_WHITELIST")
	if strings.TrimSpace(env) == "" {
		rules := defaultRules()
		for i, rule := range rules {
			if err := validateRule(rule); err != nil {
				return nil, fmt.Errorf("default rule[%d]: %w", i, err)
			}
		}
		return rules, nil
	}

	var rules []CommandRule
	if err := json.Unmarshal([]byte(env), &rules); err != nil {
		return nil, fmt.Errorf("COMMAND_WHITELIST is not valid JSON: %w", err)
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("COMMAND_WHITELIST JSON array is empty")
	}

	seen := make(map[string]bool, len(rules))
	for i, rule := range rules {
		if err := validateRule(rule); err != nil {
			return nil, fmt.Errorf("rule[%d]: %w", i, err)
		}

		if seen[rule.Name] {
			return nil, fmt.Errorf("duplicate command rule: %q", rule.Name)
		}
		seen[rule.Name] = true
	}

	return rules, nil
}

// findRule looks up a command by name in the whitelist.
func findRule(rules []CommandRule, name string) *CommandRule {
	for i := range rules {
		if rules[i].Name == name {
			return &rules[i]
		}
	}
	return nil
}

// ValidateArgs checks each supplied argument against the rule's allowlist.
func ValidateArgs(rule *CommandRule, args []string) error {
	if rule == nil {
		return fmt.Errorf("command rule must not be nil")
	}

	if rule.MaxArgs > 0 && len(args) > rule.MaxArgs {
		return fmt.Errorf(
			"too many arguments: got %d, max %d",
			len(args),
			rule.MaxArgs,
		)
	}

	if len(rule.Args) == 0 {
		if len(args) > 0 {
			return fmt.Errorf("command %q accepts no arguments", rule.Name)
		}
		return nil
	}

	for _, arg := range args {
		allowedArg := false

		for _, allowed := range rule.Args {
			if allowed.Exact != "" && arg == allowed.Exact {
				allowedArg = true
				break
			}

			if allowed.Pattern != "" {
				matched, err := filepath.Match(allowed.Pattern, arg)
				if err == nil && matched {
					allowedArg = true
					break
				}
			}
		}

		if !allowedArg {
			return fmt.Errorf(
				"argument %q is not allowed for command %q",
				arg,
				rule.Name,
			)
		}
	}

	return nil
}
