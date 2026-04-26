package config

import "testing"

func TestLoadConfigRejectsDefaultJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "change-me-at-all-costs")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected LoadConfig to reject default JWT secret")
	}
}
