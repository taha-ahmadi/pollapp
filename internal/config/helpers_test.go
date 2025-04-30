package config

import (
	"strings"
	"testing"
)

func TestParseInt(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"-456", -456, false},
		{"abc", 0, true},
		{"123abc", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseInt(tc.input)

			if tc.hasError && err == nil {
				t.Errorf("Expected error for input '%s', but got none", tc.input)
			}

			if !tc.hasError && err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
			}

			if result != tc.expected {
				t.Errorf("Expected %d, got %d for input '%s'", tc.expected, result, tc.input)
			}
		})
	}
}

func TestParseUint64(t *testing.T) {
	testCases := []struct {
		input    string
		expected uint64
		hasError bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"18446744073709551615", 18446744073709551615, false}, // max uint64
		{"-456", 0, true}, // negative numbers aren't valid
		{"abc", 0, true},
		{"123abc", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseUint64(tc.input)

			if tc.hasError && err == nil {
				t.Errorf("Expected error for input '%s', but got none", tc.input)
			}

			if !tc.hasError && err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
			}

			if result != tc.expected {
				t.Errorf("Expected %d, got %d for input '%s'", tc.expected, result, tc.input)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		hasError bool
	}{
		{"true", true, false},
		{"TRUE", true, false},
		{"True", true, false},
		{"1", true, false},
		{"t", true, false},
		{"T", true, false},
		{"false", false, false},
		{"FALSE", false, false},
		{"False", false, false},
		{"0", false, false},
		{"f", false, false},
		{"F", false, false},
		{"yes", false, true},
		{"no", false, true},
		{"", false, true},
		{"abc", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseBool(tc.input)

			if tc.hasError && err == nil {
				t.Errorf("Expected error for input '%s', but got none", tc.input)
			}

			if !tc.hasError && err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
			}

			if result != tc.expected {
				t.Errorf("Expected %v, got %v for input '%s'", tc.expected, result, tc.input)
			}

			// For errors, verify the error message
			if err != nil && !strings.Contains(err.Error(), "failed to parse bool") {
				t.Errorf("Expected error message to contain 'failed to parse bool', got '%s'", err.Error())
			}
		})
	}
}
