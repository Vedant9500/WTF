package utils

import "testing"

func TestMin(t *testing.T) {
	testCases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a less than b", 1, 2, 1},
		{"b less than a", 5, 3, 3},
		{"equal values", 4, 4, 4},
		{"negative values", -5, -3, -5},
		{"zero and positive", 0, 5, 0},
		{"zero and negative", 0, -5, -5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Min(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Min(%d, %d) = %d; want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	testCases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a greater than b", 5, 2, 5},
		{"b greater than a", 3, 7, 7},
		{"equal values", 4, 4, 4},
		{"negative values", -5, -3, -3},
		{"zero and positive", 0, 5, 5},
		{"zero and negative", 0, -5, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Max(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Max(%d, %d) = %d; want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}
