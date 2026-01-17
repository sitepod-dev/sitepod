package caddy

import (
	"os"
	"testing"

	"github.com/pocketbase/pocketbase/models"
)

func TestAuthContext(t *testing.T) {
	t.Run("IsAdmin_WithAdmin", func(t *testing.T) {
		admin := &models.Admin{
			BaseModel: models.BaseModel{Id: "admin123"},
		}
		admin.Email = "admin@example.com"

		ctx := &AuthContext{Admin: admin}

		if !ctx.IsAdmin() {
			t.Error("expected IsAdmin() to return true for admin context")
		}
	})

	t.Run("IsAdmin_WithUser", func(t *testing.T) {
		// Create a mock user record
		ctx := &AuthContext{
			User: &models.Record{},
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

	t.Run("GetID_Admin", func(t *testing.T) {
		admin := &models.Admin{
			BaseModel: models.BaseModel{Id: "admin123"},
		}

		ctx := &AuthContext{Admin: admin}

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

	t.Run("GetEmail_Admin", func(t *testing.T) {
		admin := &models.Admin{
			BaseModel: models.BaseModel{Id: "admin123"},
		}
		admin.Email = "admin@example.com"

		ctx := &AuthContext{Admin: admin}

		if got := ctx.GetEmail(); got != "admin@example.com" {
			t.Errorf("GetEmail() = %q, want %q", got, "admin@example.com")
		}
	})

	t.Run("GetEmail_Empty", func(t *testing.T) {
		ctx := &AuthContext{}

		if got := ctx.GetEmail(); got != "" {
			t.Errorf("GetEmail() = %q, want empty string", got)
		}
	})
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
