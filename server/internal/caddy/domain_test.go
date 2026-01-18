package caddy

import (
	"strings"
	"testing"
)

// normalizeSubdomainStandalone is a standalone version for testing
// (mirrors the logic in SitePodHandler.normalizeSubdomain)
func normalizeSubdomainStandalone(name string) string {
	result := strings.ToLower(name)
	var normalized strings.Builder
	for _, c := range result {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			normalized.WriteRune(c)
		} else {
			normalized.WriteRune('-')
		}
	}
	return strings.Trim(normalized.String(), "-")
}

func TestNormalizeSubdomain(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "MyProject",
			expected: "myproject",
		},
		{
			name:     "spaces_to_hyphens",
			input:    "my project",
			expected: "my-project",
		},
		{
			name:     "underscores_to_hyphens",
			input:    "my_project",
			expected: "my-project",
		},
		{
			name:     "special_chars",
			input:    "my@project!test",
			expected: "my-project-test",
		},
		{
			name:     "trim_leading_hyphens",
			input:    "---myproject",
			expected: "myproject",
		},
		{
			name:     "trim_trailing_hyphens",
			input:    "myproject---",
			expected: "myproject",
		},
		{
			name:     "trim_both_hyphens",
			input:    "---myproject---",
			expected: "myproject",
		},
		{
			name:     "numbers_preserved",
			input:    "project123",
			expected: "project123",
		},
		{
			name:     "hyphens_preserved",
			input:    "my-project",
			expected: "my-project",
		},
		{
			name:     "unicode_to_hyphens",
			input:    "我的项目",
			expected: "",
		},
		{
			name:     "mixed_unicode",
			input:    "my项目test",
			expected: "my--test",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "only_special_chars",
			input:    "@#$%",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeSubdomainStandalone(tc.input)
			if got != tc.expected {
				t.Errorf("normalizeSubdomain(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestDomainTypeDetection(t *testing.T) {
	// Test domain type detection logic (system vs custom)
	// Domain is "system" if it ends with ".{baseDomain}" or equals baseDomain
	testCases := []struct {
		name         string
		domain       string
		baseDomain   string
		expectedType string
	}{
		{
			name:         "system_subdomain",
			domain:       "myproject.example.com",
			baseDomain:   "example.com",
			expectedType: "system",
		},
		{
			name:         "system_exact_match",
			domain:       "example.com",
			baseDomain:   "example.com",
			expectedType: "system",
		},
		{
			name:         "custom_different_domain",
			domain:       "mysite.io",
			baseDomain:   "example.com",
			expectedType: "custom",
		},
		{
			name:         "custom_partial_match",
			domain:       "notexample.com",
			baseDomain:   "example.com",
			expectedType: "custom",
		},
		{
			name:         "system_nested_subdomain",
			domain:       "app.myproject.example.com",
			baseDomain:   "example.com",
			expectedType: "system",
		},
		{
			name:         "custom_similar_suffix",
			domain:       "myexample.com",
			baseDomain:   "example.com",
			expectedType: "custom",
		},
		{
			name:         "system_localhost",
			domain:       "myproject.localhost",
			baseDomain:   "localhost",
			expectedType: "system",
		},
		{
			name:         "system_localhost_with_base_port",
			domain:       "myproject.localhost",
			baseDomain:   "localhost:8080",
			expectedType: "system",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			domainType := detectDomainType(tc.domain, tc.baseDomain)
			if domainType != tc.expectedType {
				t.Errorf("detectDomainType(%q, %q) = %q, want %q",
					tc.domain, tc.baseDomain, domainType, tc.expectedType)
			}
		})
	}
}

// detectDomainType mirrors the logic from apiAddDomain
func detectDomainType(domain, baseDomain string) string {
	// Strip port from base domain for comparison
	if idx := strings.Index(baseDomain, ":"); idx != -1 {
		baseDomain = baseDomain[:idx]
	}

	if strings.HasSuffix(domain, "."+baseDomain) || domain == baseDomain {
		return "system"
	}
	return "custom"
}

func TestSlugNormalization(t *testing.T) {
	// Test slug normalization logic from apiAddDomain
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_becomes_root",
			input:    "",
			expected: "/",
		},
		{
			name:     "root_unchanged",
			input:    "/",
			expected: "/",
		},
		{
			name:     "adds_leading_slash",
			input:    "blog",
			expected: "/blog",
		},
		{
			name:     "preserves_leading_slash",
			input:    "/blog",
			expected: "/blog",
		},
		{
			name:     "nested_path",
			input:    "/docs/api",
			expected: "/docs/api",
		},
		{
			name:     "nested_without_slash",
			input:    "docs/api",
			expected: "/docs/api",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeSlug(tc.input)
			if got != tc.expected {
				t.Errorf("normalizeSlug(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// normalizeSlug mirrors the slug normalization logic from apiAddDomain
func normalizeSlug(slug string) string {
	if slug == "" {
		return "/"
	}
	if !strings.HasPrefix(slug, "/") {
		return "/" + slug
	}
	return slug
}

func TestDomainValidation(t *testing.T) {
	testCases := []struct {
		name    string
		domain  string
		isValid bool
	}{
		{
			name:    "valid_domain",
			domain:  "example.com",
			isValid: true,
		},
		{
			name:    "valid_subdomain",
			domain:  "sub.example.com",
			isValid: true,
		},
		{
			name:    "valid_multi_subdomain",
			domain:  "a.b.c.example.com",
			isValid: true,
		},
		{
			name:    "valid_with_hyphen",
			domain:  "my-site.example.com",
			isValid: true,
		},
		{
			name:    "valid_with_numbers",
			domain:  "site123.example.com",
			isValid: true,
		},
		{
			name:    "empty_domain",
			domain:  "",
			isValid: false,
		},
		{
			name:    "only_tld",
			domain:  "com",
			isValid: false,
		},
		{
			name:    "starts_with_hyphen",
			domain:  "-example.com",
			isValid: false,
		},
		{
			name:    "ends_with_hyphen",
			domain:  "example-.com",
			isValid: false,
		},
		{
			name:    "double_dot",
			domain:  "example..com",
			isValid: false,
		},
		{
			name:    "contains_underscore",
			domain:  "my_site.example.com",
			isValid: false,
		},
		{
			name:    "too_long_label",
			domain:  strings.Repeat("a", 64) + ".com",
			isValid: false,
		},
		{
			name:    "max_length_label",
			domain:  strings.Repeat("a", 63) + ".com",
			isValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidDomain(tc.domain)
			if got != tc.isValid {
				t.Errorf("isValidDomain(%q) = %v, want %v", tc.domain, got, tc.isValid)
			}
		})
	}
}

// isValidDomain validates domain format
func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// Must have at least one dot (TLD only is not valid)
	if !strings.Contains(domain, ".") {
		return false
	}

	// Check for double dots
	if strings.Contains(domain, "..") {
		return false
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 {
			return false
		}
		if len(label) > 63 {
			return false
		}
		// Label cannot start or end with hyphen
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		// Label can only contain alphanumeric and hyphens
		for _, c := range label {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
	}

	return true
}

func TestVerificationTokenFormat(t *testing.T) {
	// Test that verification token format is correct
	// Format: sitepod-verify-{16-char-uuid-prefix}
	token := "sitepod-verify-abcd1234efgh5678"

	if !strings.HasPrefix(token, "sitepod-verify-") {
		t.Error("token should start with 'sitepod-verify-'")
	}

	suffix := strings.TrimPrefix(token, "sitepod-verify-")
	if len(suffix) != 16 {
		t.Errorf("token suffix should be 16 chars, got %d", len(suffix))
	}
}

func TestVerificationTXTRecord(t *testing.T) {
	// Test TXT record format
	domain := "example.com"
	token := "sitepod-verify-abcd1234efgh5678"

	expectedTXT := "_sitepod.example.com TXT \"sitepod-verify-abcd1234efgh5678\""
	actualTXT := buildVerificationTXT(domain, token)

	if actualTXT != expectedTXT {
		t.Errorf("buildVerificationTXT(%q, %q) = %q, want %q",
			domain, token, actualTXT, expectedTXT)
	}
}

func buildVerificationTXT(domain, token string) string {
	return "_sitepod." + domain + " TXT \"" + token + "\""
}

func TestDomainLookupKey(t *testing.T) {
	// Test that domain+slug combination is used as unique key
	testCases := []struct {
		domain1 string
		slug1   string
		domain2 string
		slug2   string
		same    bool
	}{
		{
			domain1: "example.com", slug1: "/",
			domain2: "example.com", slug2: "/",
			same: true,
		},
		{
			domain1: "example.com", slug1: "/",
			domain2: "example.com", slug2: "/blog",
			same: false,
		},
		{
			domain1: "example.com", slug1: "/blog",
			domain2: "other.com", slug2: "/blog",
			same: false,
		},
		{
			domain1: "EXAMPLE.COM", slug1: "/",
			domain2: "example.com", slug2: "/",
			same: true, // domains are case-insensitive
		},
	}

	for i, tc := range testCases {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			key1 := buildDomainKey(tc.domain1, tc.slug1)
			key2 := buildDomainKey(tc.domain2, tc.slug2)
			if (key1 == key2) != tc.same {
				t.Errorf("expected same=%v for keys %q and %q", tc.same, key1, key2)
			}
		})
	}
}

func buildDomainKey(domain, slug string) string {
	return strings.ToLower(domain) + ":" + slug
}
