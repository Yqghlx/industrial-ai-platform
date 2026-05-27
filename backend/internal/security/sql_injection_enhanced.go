package security

import (
	"regexp"
	"strings"
	"unicode"
)

// SEC-MED-05: Enhanced SQL Injection Detection
// This provides comprehensive SQL injection pattern detection beyond simple string matching.

// SQLInjectionConfig holds configuration for SQL injection detection
type SQLInjectionConfig struct {
	// StrictMode enables more aggressive detection (may have false positives)
	StrictMode bool
	// AllowCommonWords allows common words that might trigger false positives
	AllowCommonWords []string
}

// DefaultSQLInjectionConfig returns default configuration
func DefaultSQLInjectionConfig() *SQLInjectionConfig {
	return &SQLInjectionConfig{
		StrictMode: false,
		AllowCommonWords: []string{
			// Common safe words that contain SQL-like patterns
			"selective", "selection", "selected", "selector",
			"insertion", "inserted", "inserting",
			"updated", "updating", "updates",
			"deleted", "deleting", "deletes",
			"executive", "executed", "executing", "execution",
			"union", "reunion", "unionized",
		},
	}
}

// SQL injection detection patterns
var (
	// Basic SQL keywords pattern (case-insensitive)
	sqlKeywordsPattern = regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|EXEC|EXECUTE|UNION|JOIN|WHERE|FROM|INTO|VALUES|SET|ORDER|GROUP|HAVING|LIMIT|OFFSET)\b`)

	// SQL injection specific patterns
	sqlInjectionPatterns = []struct {
		name    string
		pattern *regexp.Regexp
	}{
		// Classic injection patterns
		{"quote_or", regexp.MustCompile(`(?i)'\s*OR\s*`)},
		{"quote_and", regexp.MustCompile(`(?i)'\s*AND\s*`)},
		{"or_condition", regexp.MustCompile(`(?i)\bOR\s+\d+\s*=\s*\d+`)},
		{"and_condition", regexp.MustCompile(`(?i)\bAND\s+\d+\s*=\s*\d+`)},

		// Comment injection
		{"comment_dash", regexp.MustCompile(`--`)},
		{"comment_block", regexp.MustCompile(`(?i)/\*`)},
		{"comment_block_end", regexp.MustCompile(`(?i)\*/`)},
		{"comment_mysql", regexp.MustCompile(`(?i)#\s*$`)},

		// Union injection
		{"union_select", regexp.MustCompile(`(?i)\bUNION\b.*\bSELECT\b`)},
		{"union_all", regexp.MustCompile(`(?i)\bUNION\b.*\bALL\b`)},

		// Stacked queries
		{"stacked_query", regexp.MustCompile(`;\s*(?i)(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|EXEC)`)},

		// Time-based injection
		{"sleep_injection", regexp.MustCompile(`(?i)\bSLEEP\b\s*\(`)},
		{"benchmark_injection", regexp.MustCompile(`(?i)\bBENCHMARK\b\s*\(`)},
		{"waitfor_injection", regexp.MustCompile(`(?i)\bWAITFOR\b\s+DELAY`)},
		{"pg_sleep", regexp.MustCompile(`(?i)\bPG_SLEEP\b\s*\(`)},

		// Function-based injection
		{"load_file", regexp.MustCompile(`(?i)\bLOAD_FILE\b\s*\(`)},
		{"into_outfile", regexp.MustCompile(`(?i)\bINTO\b\s+(?i)OUTFILE\b`)},
		{"into_dumpfile", regexp.MustCompile(`(?i)\bINTO\b\s+(?i)DUMPFILE\b`)},
		{"xp_cmdshell", regexp.MustCompile(`(?i)xp_cmdshell`)},

		// Data manipulation
		{"drop_table", regexp.MustCompile(`(?i)\bDROP\b\s+(?i)TABLE\b`)},
		{"drop_database", regexp.MustCompile(`(?i)\bDROP\b\s+(?i)DATABASE\b`)},
		{"truncate_table", regexp.MustCompile(`(?i)\bTRUNCATE\b\s+(?i)TABLE\b`)},
		{"alter_table", regexp.MustCompile(`(?i)\bALTER\b\s+(?i)TABLE\b`)},

		// Hex/Char encoding attempts
		{"hex_encoding", regexp.MustCompile(`(?i)0x[0-9a-f]+`)},
		{"char_function", regexp.MustCompile(`(?i)\bCHAR\b\s*\(\s*\d+`)},

		// Information schema access
		{"information_schema", regexp.MustCompile(`(?i)information_schema`)},
		{"sys_tables", regexp.MustCompile(`(?i)sys\.(?:tables|columns|objects)`)},

		// Boolean-based injection
		{"boolean_injection", regexp.MustCompile(`(?i)\b(?:TRUE|FALSE)\b.*\b(?:OR|AND)\b`)},

		// String concatenation injection
		{"concat_function", regexp.MustCompile(`(?i)\bCONCAT\b\s*\(`)},
		{"concat_ws", regexp.MustCompile(`(?i)\bCONCAT_WS\b\s*\(`)},

		// Quote-based patterns
		{"quote_escape", regexp.MustCompile(`(?i)\\'`)},
		{"double_quote_escape", regexp.MustCompile(`(?i)\\"`)},
		{"single_quote_sequence", regexp.MustCompile(`''`)},
	}

	// Suspicious characters that often appear in injection
	suspiciousChars = []string{
		"'", "\"", ";", "--", "/*", "*/", "#",
		"=", "<", ">", "(", ")",
	}
)

// ContainsSQLInjection performs comprehensive SQL injection detection
// Returns true if input contains potential SQL injection patterns
func ContainsSQLInjection(input string, config *SQLInjectionConfig) bool {
	if config == nil {
		config = DefaultSQLInjectionConfig()
	}

	// Normalize input
	normalized := normalizeInput(input)

	// Check against allowed common words (avoid false positives)
	for _, word := range config.AllowCommonWords {
		if strings.Contains(strings.ToLower(normalized), strings.ToLower(word)) {
			// Remove the allowed word from input for checking
			normalized = strings.ReplaceAll(strings.ToLower(normalized), strings.ToLower(word), "")
		}
	}

	// Check for SQL injection patterns
	for _, pattern := range sqlInjectionPatterns {
		if pattern.pattern.MatchString(normalized) {
			return true
		}
	}

	// In strict mode, check for SQL keywords
	if config.StrictMode {
		if sqlKeywordsPattern.MatchString(normalized) {
			return true
		}
	}

	// Check for suspicious character combinations
	if hasSuspiciousCharCombo(normalized) {
		return true
	}

	return false
}

// ContainsSQLInjectionSimple provides simple SQL injection detection
// This is the enhanced version of the original containsSQLInjectionPattern
func ContainsSQLInjectionSimple(input string) bool {
	patterns := []string{
		"' OR ",
		"' AND ",
		" OR ",  // Without quotes
		" AND ", // Without quotes
		"UNION",
		"SELECT",
		"DROP",
		"INSERT",
		"UPDATE",
		"DELETE",
		"--",
		";",
		"/*",
		"*/",
		"EXEC",
		"xp_cmdshell",
		"SLEEP",
		"BENCHMARK",
		"WAITFOR",
		"information_schema",
		"LOAD_FILE",
		"INTO OUTFILE",
		"INTO DUMPFILE",
		"PG_SLEEP",
	}

	upperInput := strings.ToUpper(input)
	for _, pattern := range patterns {
		if strings.Contains(upperInput, strings.ToUpper(pattern)) {
			// Check for false positive with common words
			if isFalsePositive(input, pattern) {
				continue
			}
			return true
		}
	}
	return false
}

// normalizeInput normalizes the input string for detection
func normalizeInput(input string) string {
	// Remove extra whitespace
	var result strings.Builder
	lastWasSpace := false
	for _, r := range input {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				result.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			result.WriteRune(r)
			lastWasSpace = false
		}
	}
	return result.String()
}

// hasSuspiciousCharCombo checks for suspicious character combinations
func hasSuspiciousCharCombo(input string) bool {
	// Check for quote followed by semicolon (classic injection)
	if strings.Contains(input, "'") && strings.Contains(input, ";") {
		// Check if they appear close together
		idx := strings.Index(input, "'")
		if idx >= 0 && idx < len(input)-10 {
			substr := input[idx:min(idx+10, len(input))]
			if strings.Contains(substr, ";") {
				return true
			}
		}
	}

	// Check for quote followed by equals (injection attempt)
	if strings.Contains(input, "'") && strings.Contains(input, "=") {
		idx := strings.Index(input, "'")
		if idx >= 0 && idx < len(input)-5 {
			substr := input[idx:min(idx+5, len(input))]
			if strings.Contains(substr, "=") {
				return true
			}
		}
	}

	return false
}

// isFalsePositive checks if the detected pattern is likely a false positive
func isFalsePositive(input, pattern string) bool {
	lowerInput := strings.ToLower(input)

	// Common safe words to check
	safeWords := []string{
		"selective", "selection", "selected",
		"insertion", "inserted",
		"updated", "updating", "updates",
		"deleted", "deleting",
		"executive", "executed", "execution",
	}

	for _, word := range safeWords {
		if strings.Contains(lowerInput, word) {
			// Check if the pattern appears as part of a safe word
			lowerPattern := strings.ToLower(pattern)
			if strings.Contains(word, lowerPattern) {
				return true
			}
		}
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
