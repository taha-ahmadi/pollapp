package config

import (
	"fmt"
	"strconv"
)

// parseInt converts a string to an integer
func parseInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int: %w", err)
	}
	return i, nil
}

// parseUint64 converts a string to an uint64
func parseUint64(s string) (uint64, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint64: %w", err)
	}
	return i, nil
}

// parseBool converts a string to a boolean
func parseBool(s string) (bool, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("failed to parse bool: %w", err)
	}
	return b, nil
}
