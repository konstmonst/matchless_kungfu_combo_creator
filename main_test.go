package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMergeStrings(t *testing.T) {
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
		res := mergeStrings(tc.a, tc.b)
		if diff := cmp.Diff(tc.expected, res); diff != "" {
			t.Errorf("TestCombineTechs/%s mismatch (-want +got):\n%s", tc.name, diff)
		}
	}
}

func TestFindSmallestCommonString(t *testing.T) {
	techABC := Tech{V: "abc"}
	techDEFG := Tech{V: "defg"}
	techHI := Tech{V: "hi"}

	testCases := []struct {
		name     string
		input    []Tech
		maxChars int
		expected MergedTechs
	}{
		{
			name:     "same",
			input:    []Tech{techABC, techABC, techABC},
			maxChars: 20,
			expected: MergedTechs{Techs: []*Tech{&techABC, &techABC, &techABC}, Value: techABC.V},
		},
		{
			name:     "nothing is in common",
			input:    []Tech{techABC, techDEFG, techHI},
			maxChars: 20,
			expected: MergedTechs{Techs: []*Tech{&techABC, &techDEFG, &techHI}, Value: "abcdefghi"},
		},
		{
			name:     "no combinations",
			input:    []Tech{techABC, techDEFG, techHI},
			maxChars: 6,
			expected: MergedTechs{},
		},
	}

	for _, tc := range testCases {
		res := findSmallestCommonString(tc.input, tc.maxChars)
		if diff := cmp.Diff(tc.expected, res); diff != "" {
			t.Errorf("TestFindSmallestCommonString/%s mismatch (-want +got):\n%s", tc.name, diff)
		}
	}
}
