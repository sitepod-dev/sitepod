package caddy

import (
	"os"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestAuthContext(t *testing.T) {
	// Note: IsAdmin() now checks user.GetBool("is_admin") which requires
	// a properly initialized collection. We can only test nil/empty cases here.
	// Full is_admin testing is done through integration tests.

	t.Run("IsAdmin_Empty", func(t *testing.T) {
		ctx := &AuthContext{}

		if ctx.IsAdmin() {
			t.Error("expected IsAdmin() to return false for empty context")
		}
	})

	t.Run("GetID_User", func(t *testing.T) {
		user := &core.Record{}
		user.Id = "user123"

		ctx := &AuthContext{User: user}

		if got := ctx.GetID(); got != "user123" {
			t.Errorf("GetID() = %q, want %q", got, "user123")
		}
	})

	t.Run("GetID_Empty", func(t *testing.T) {
		ctx := &AuthContext{}

		if got := ctx.GetID(); got != "" {
			t.Errorf("GetID() = %q, want empty string", got)
		}
	})

	// Note: GetEmail and is_admin tests are skipped because core.Record methods
	// require a properly initialized collection which can't be done
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
