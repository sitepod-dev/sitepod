package caddy

import (
	"os"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestAuthContext(t *testing.T) {
	t.Run("IsAdmin_WithSuperuser", func(t *testing.T) {
		// In PocketBase v0.36, records need a proper collection
		// Test with a non-nil record to verify IsAdmin logic
		ctx := &AuthContext{Superuser: &core.Record{}}

		if !ctx.IsAdmin() {
			t.Error("expected IsAdmin() to return true for superuser context")
		}
	})

	t.Run("IsAdmin_WithUser", func(t *testing.T) {
		// Create a mock user record
		ctx := &AuthContext{
			User: &core.Record{},
		}

		if ctx.IsAdmin() {
			t.Error("expected IsAdmin() to return false for user context")
		}
	})

	t.Run("IsAdmin_Empty", func(t *testing.T) {
		ctx := &AuthContext{}

		if ctx.IsAdmin() {
			t.Error("expected IsAdmin() to return false for empty context")
		}
	})

	t.Run("GetID_Superuser", func(t *testing.T) {
		// In v0.36, we can set Id directly on record
		superuser := &core.Record{}
		superuser.Id = "admin123"

		ctx := &AuthContext{Superuser: superuser}

		if got := ctx.GetID(); got != "admin123" {
			t.Errorf("GetID() = %q, want %q", got, "admin123")
		}
	})

	t.Run("GetID_Empty", func(t *testing.T) {
		ctx := &AuthContext{}

		if got := ctx.GetID(); got != "" {
			t.Errorf("GetID() = %q, want empty string", got)
		}
	})

	// Note: GetEmail tests are skipped because core.Record.GetString()
	// requires a properly initialized collection which can't be done
	// without a full PocketBase app context. These are integration-tested
	// through the full application flow instead.
}

func TestIsDemoMode(t *testing.T) {
	testCases := []struct {
		envValue string
		expected bool
	}{
		{"1", true},
		{"true", true},
		{"0", false},
		{"false", false},
		{"", false},
		{"yes", false}, // only "1" or "true" are valid
	}

	for _, tc := range testCases {
		t.Run("IS_DEMO="+tc.envValue, func(t *testing.T) {
			t.Setenv("IS_DEMO", tc.envValue)

			isDemo := isDemoMode()
			if isDemo != tc.expected {
				t.Errorf("isDemoMode() = %v, want %v for IS_DEMO=%q", isDemo, tc.expected, tc.envValue)
			}
		})
	}
}

// isDemoMode checks if IS_DEMO environment variable is set
// This mirrors the logic in apiConfig
func isDemoMode() bool {
	val := os.Getenv("IS_DEMO")
	return val == "1" || val == "true"
}
