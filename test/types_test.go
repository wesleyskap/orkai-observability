package test

import (
	"orkai-observability/observability"
	"testing"
)

// TestNewStringField verifies the construction of string-typed logging fields.
func TestNewStringField(t *testing.T) {
	field := observability.NewStringField("mykey", "myval")
	if field.Key != "mykey" {
		t.Fatalf("expected key 'mykey', got %s", field.Key)
	}
	if field.StrValue != "myval" {
		t.Fatalf("expected string value 'myval', got %s", field.StrValue)
	}
	if field.IsInt {
		t.Fatal("expected IsInt to be false")
	}
}

// TestNewIntField verifies the construction of integer-typed logging fields.
func TestNewIntField(t *testing.T) {
	field := observability.NewIntField("mykey", 42)
	if field.Key != "mykey" {
		t.Fatalf("expected key 'mykey', got %s", field.Key)
	}
	if field.IntValue != 42 {
		t.Fatalf("expected int value 42, got %d", field.IntValue)
	}
	if !field.IsInt {
		t.Fatal("expected IsInt to be true")
	}
}
