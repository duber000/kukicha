package cast

import (
	"encoding/json"
	"testing"
)

func TestSmartInt(t *testing.T) {
	tests := []struct {
		input    any
		expected int
		wantErr  bool
	}{
		{42, 42, false},
		{int64(42), 42, false},
		{42.5, 42, false},
		{"42", 42, false},
		{json.Number("42"), 42, false},
		{true, 1, false},
		{false, 0, false},
		{"abc", 0, true},
		{nil, 0, true},
	}

	for _, tt := range tests {
		got, err := SmartInt(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("SmartInt(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("SmartInt(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSmartFloat64(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
		wantErr  bool
	}{
		{42.5, 42.5, false},
		{float32(42.5), 42.5, false},
		{42, 42.0, false},
		{"42.5", 42.5, false},
		{json.Number("42.5"), 42.5, false},
		{true, 1.0, false},
		{false, 0.0, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		got, err := SmartFloat64(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("SmartFloat64(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("SmartFloat64(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSmartBool(t *testing.T) {
	tests := []struct {
		input    any
		expected bool
		wantErr  bool
	}{
		{true, true, false},
		{false, false, false},
		{1, true, false},
		{0, false, false},
		{1.0, true, false},
		{0.0, false, false},
		{"true", true, false},
		{"false", false, false},
		{"1", true, false},
		{"0", false, false},
		{"abc", false, true},
	}

	for _, tt := range tests {
		got, err := SmartBool(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("SmartBool(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("SmartBool(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSmartString(t *testing.T) {
	tests := []struct {
		input    any
		expected string
		wantErr  bool
	}{
		{"hello", "hello", false},
		{42, "42", false},
		{42.5, "42.5", false},
		{true, "true", false},
		{[]byte("bytes"), "bytes", false},
		{nil, "", false},
	}

	for _, tt := range tests {
		got, err := SmartString(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("SmartString(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("SmartString(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
