package text

import (
	"cmp"
	"slices"
	"strings"
	"unicode/utf8"
)

type Kind uint

const (
	Original Kind = iota
	Overlap
)

type Range interface {
	From() int
	To() int
	IsIn(context []rune) bool
}

type Segment struct {
	Start int
	End   int
	Kind  Kind
}

func (s Segment) From() int { return s.Start }
func (s Segment) To() int   { return s.End }
func (s Segment) IsIn(context []rune) bool {
	return s.End <= len(context)
}

type Event struct {
	Pos   int
	Delta int
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func Indices(s, substr string) []int {
	indices := make([]int, 0)

	if substr == "" {
		return indices
	}

	i := 0
	for {
		if pos := strings.Index(s[i:], substr); pos != -1 {
			indices = append(indices, utf8.RuneCountInString(s[:i+pos]))
			i += pos + 1
		} else {
			break
		}
	}
	return indices
}

func FindOverlaps[T Range](ranges []T) []Segment {
	if len(ranges) == 0 {
		return nil
	}

	events := make([]Event, 0, len(ranges)*2)
	for _, r := range ranges {
		events = append(events, Event{r.From(), 1}, Event{r.To(), -1})
	}

	slices.SortFunc(events, func(a, b Event) int {
		if a.Pos != b.Pos {
			return a.Pos - b.Pos
		}
		return a.Delta - b.Delta
	})

	segments := make([]Segment, 0, len(events))
	current := 0
	prevPos := events[0].Pos

	for _, e := range events {
		if e.Pos != prevPos && current > 0 {
			kind := Original
			if current > 1 {
				kind = Overlap
			}
			segments = append(segments, Segment{
				Start: prevPos,
				End:   e.Pos,
				Kind:  kind,
			})
		}
		current += e.Delta
		prevPos = e.Pos
	}

	return mergeSegments(segments)
}

func mergeSegments(segments []Segment) []Segment {
	n := len(segments)
	if n <= 1 {
		return segments
	}

	// Sort in-place by Start
	slices.SortFunc(segments, func(a, b Segment) int {
		return cmp.Compare(a.Start, b.Start)
	})

	// In-place compaction to avoid allocating new slice
	write := 0
	for read := 1; read < n; read++ {
		curr := segments[read]
		last := &segments[write]

		if last.Kind == curr.Kind && last.End == curr.Start {
			last.End = curr.End
		} else {
			write++
			segments[write] = curr
		}
	}

	return segments[:write+1]
}
