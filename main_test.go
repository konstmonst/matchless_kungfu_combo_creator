package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCalcMergePos(t *testing.T) {
	testCases := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{
			name:     "same",
			a:        "abc",
			b:        "abc",
			expected: 0,
		},
		{
			name:     "nothing in common",
			a:        "abc",
			b:        "def",
			expected: 3,
		},
		{
			name:     "one char in common",
			a:        "abc",
			b:        "cde",
			expected: 2,
		},
		{
			name:     "two chars in common",
			a:        "abc",
			b:        "bcd",
			expected: 1,
		},
		{
			name:     "bug1",
			a:        "ONAOOA",
			b:        "NAOOAN",
			expected: 1,
		},
	}

	for _, tc := range testCases {
		res := calcMergePos([]byte(tc.a), []byte(tc.b))
		if diff := cmp.Diff(tc.expected, res); diff != "" {
			t.Errorf("TestCombineTechs/%s mismatch (-want +got):\n%s", tc.name, diff)
		}
	}
}

func TestMergeInners(t *testing.T) {
	innerABC := &Inner{Bytes: []byte("abc")}
	innerDEFG := &Inner{Bytes: []byte("defg")}
	innerHI := &Inner{Bytes: []byte("hi")}
	innerNAOOAN := &Inner{Bytes: []byte("NAOOAN")}
	innerONAOOA := &Inner{Bytes: []byte("ONAOOA")}

	testCases := []struct {
		name     string
		input    []*Inner
		maxChars int
		expected MergedInners
	}{
		{
			name:     "bug1",
			input:    []*Inner{innerNAOOAN, innerONAOOA},
			maxChars: 20,
			expected: MergedInners{
				InnerIndices: []int{1, 0},
				MergePos:     []int{0, 1},
				CachedValue:  []byte("ONAOOAN"),
			},
		},
		{
			name:     "no combinations",
			input:    []*Inner{innerABC, innerDEFG, innerHI},
			maxChars: 6,
			expected: MergedInners{},
		},
		{
			name:     "nothing is in common",
			input:    []*Inner{innerABC, innerDEFG, innerHI},
			maxChars: 20,
			expected: MergedInners{
				InnerIndices: []int{0, 1, 2},
				MergePos:     []int{0, 3, 7},
				CachedValue:  []byte("abcdefghi"),
			},
		},
		{
			name:     "same",
			input:    []*Inner{innerABC, innerABC, innerABC},
			maxChars: 20,
			expected: MergedInners{
				InnerIndices: []int{0, 1, 2},
				MergePos:     []int{0, 0, 0},
				CachedValue:  innerABC.Bytes,
			},
		},
	}

	for _, tc := range testCases {
		mergeCache = map[uint]int{}
		res := mergeInners(tc.input, tc.maxChars)
		if diff := cmp.Diff(tc.expected, res); diff != "" {
			t.Errorf("TestFindSmallestCommonString/%s mismatch (-want +got):\n%s", tc.name, diff)
		}
	}
}
