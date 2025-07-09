package text_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/donderom/sqwat/text"
)

func TestFindOverlaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []text.Segment
		expected []text.Segment
	}{
		{
			name: "basic overlap",
			input: []text.Segment{
				{Start: 1, End: 10},
				{Start: 5, End: 15},
				{Start: 8, End: 12},
			},
			expected: []text.Segment{
				{1, 5, text.Original},
				{5, 12, text.Overlap},
				{12, 15, text.Original},
			},
		},
		{
			name: "no overlap",
			input: []text.Segment{
				{Start: 1, End: 3},
				{Start: 4, End: 6},
			},
			expected: []text.Segment{
				{1, 3, text.Original},
				{4, 6, text.Original},
			},
		},
		{
			name: "complete overlap",
			input: []text.Segment{
				{Start: 1, End: 10},
				{Start: 1, End: 10},
				{Start: 1, End: 10},
			},
			expected: []text.Segment{
				{1, 10, text.Overlap},
			},
		},
		{
			name: "touching ranges",
			input: []text.Segment{
				{Start: 1, End: 5},
				{Start: 5, End: 10},
			},
			expected: []text.Segment{
				{1, 10, text.Original},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, text.FindOverlaps(tt.input))
		})
	}
}

func TestIndices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		substr   string
		expected []int
	}{
		{
			name:     "no overlap",
			input:    "hellohellohello",
			substr:   "lo",
			expected: []int{3, 8, 13},
		},
		{
			name:     "overlapping matches",
			input:    "aaaaa",
			substr:   "aa",
			expected: []int{0, 1, 2, 3},
		},
		{
			name:     "consecutive matches",
			input:    "abcabcabc",
			substr:   "abc",
			expected: []int{0, 3, 6},
		},
		{
			name:     "partial overlaps",
			input:    "abcabcabc",
			substr:   "bca",
			expected: []int{1, 4},
		},
		{
			name:     "not found",
			input:    "abcabcabc",
			substr:   "z",
			expected: []int{},
		},
		{
			name:     "empty input",
			input:    "",
			substr:   "a",
			expected: []int{},
		},
		{
			name:     "non-ASCII input",
			input:    "トワイライトプリンセス",
			substr:   "イ",
			expected: []int{2, 4},
		},
		{
			name:     "empty substring",
			input:    "a",
			substr:   "",
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, text.Indices(tt.input, tt.substr))
		})
	}
}
