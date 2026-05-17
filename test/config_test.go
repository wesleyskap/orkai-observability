package test

import (
	"testing"

	"github.com/wesleyskap/orkai-observability/observability"
)

// TestValidateConfigValid verifies configuration with correct inputs.
func TestValidateConfigValid(t *testing.T) {
	cfg := observability.Config{
		ServiceName: "user-service",
		Environment: "dev",
		LogLevel:    "debug",
	}
	err := observability.ValidateConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestValidateConfigEmptyService checks validation fails on empty service.
func TestValidateConfigEmptyService(t *testing.T) {
	cfg := observability.Config{
		ServiceName: "",
		Environment: "dev",
		LogLevel:    "debug",
	}
	err := observability.ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty ServiceName, got nil")
	}
}

// TestValidateConfigEmptyEnv checks validation fails on empty environment.
func TestValidateConfigEmptyEnv(t *testing.T) {
	cfg := observability.Config{
		ServiceName: "user-service",
		Environment: "",
		LogLevel:    "debug",
	}
	err := observability.ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty Environment, got nil")
	}
}
