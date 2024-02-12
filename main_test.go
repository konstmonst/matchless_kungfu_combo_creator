package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCombineTechs(t *testing.T) {
	testCases := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name:     "same",
			a:        "abc",
			b:        "abc",
			expected: "abc",
		},
		{
			name:     "nothing in common",
			a:        "abc",
			b:        "def",
			expected: "abcdef",
		},
		{
			name:     "one char in common",
			a:        "abc",
			b:        "cde",
			expected: "abcde",
		},
		{
			name:     "two chars in common",
			a:        "abc",
			b:        "bcd",
			expected: "abcd",
		},
	}

	for _, tc := range testCases {
		res := combineTechs(tc.a, tc.b)
		if diff := cmp.Diff(tc.expected, res); diff != "" {
			t.Errorf("TestCombineTechs/%s mismatch (-want +got):\n%s", tc.name, diff)
		}
	}
}

func TestFindSmallestCommonString(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		maxChars int
		expected string
	}{
		{
			name:     "same",
			input:    []string{"abc", "abc", "abc"},
			maxChars: 20,
			expected: "abc",
		},
		{
			name:     "nothing in common",
			input:    []string{"abc", "defg", "hi"},
			maxChars: 20,
			expected: "abcdefghi",
		},
		{
			name:     "no combinations",
			input:    []string{"abc", "defg", "hi"},
			maxChars: 6,
			expected: "",
		},
	}

	for _, tc := range testCases {
		res := findSmallestCommonString(tc.input, tc.maxChars)
		if diff := cmp.Diff(tc.expected, res); diff != "" {
			t.Errorf("TestFindSmallestCommonString/%s mismatch (-want +got):\n%s", tc.name, diff)
		}
	}
}
