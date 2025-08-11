package teax

import (
	"slices"
	"strings"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/text"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

var viewportStyle lipgloss.Style = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).
	PaddingLeft(1).
	PaddingRight(1).
	MarginLeft(2).
	MarginRight(2)

type Viewport[T text.Range] struct {
	viewport viewport.Model
}

func NewViewport[T text.Range]() Viewport[T] {
	return Viewport[T]{viewport: viewport.New(0, 0)}
}

func (m Viewport[T]) Update(msg tea.Msg) (Viewport[T], tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyset.Next):
			m.viewport.ScrollDown(1)
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd

		case key.Matches(msg, keyset.Prev):
			m.viewport.ScrollUp(1)
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *Viewport[T]) Highlight(
	content []rune,
	answers []T,
	highlight lipgloss.Style,
) {
	var outOfRange []T
	var s strings.Builder
	m.viewport.Style = m.viewport.Style.Faint(len(answers) == 0)
	offset := 0

	for _, r := range answers {
		if !r.IsIn(content) {
			outOfRange = append(outOfRange, r)
		}
	}

	for _, segment := range text.FindOverlaps(answers) {
		s.WriteString(string(content[offset:segment.Start]))
		segText := string(content[segment.Start:segment.End])

		st := highlight
		if i := slices.IndexFunc(outOfRange, func(a T) bool {
			return segment.Start >= a.From() && segment.End <= a.To()
		}); i != -1 {
			st = style.Error
		}

		if segment.Kind == text.Original {
			s.WriteString(st.Render(segText))
		} else {
			s.WriteString(st.Underline(true).Render(segText))
		}

		offset = segment.End
	}

	s.WriteString(string(content[offset:]))
	m.SetContent(s.String())
}

func (m *Viewport[T]) SetContent(content string) {
	wrapped := wordwrap.String(content, m.viewport.Width-1)
	m.viewport.SetContent(wrapped)
	m.viewport.SetYOffset(0)
}

func (m Viewport[T]) View() string {
	return viewportStyle.Render(m.viewport.View())
}

func (m *Viewport[T]) Resize(width, height int) {
	padding := viewportStyle.GetHorizontalPadding()
	m.viewport.Width = width - padding*2

	v := viewportStyle.GetVerticalFrameSize()
	m.viewport.Height = height - v
}

func (m *Viewport[T]) Blur() {
	m.viewport.Style = style.Faint
}

func (m Viewport[T]) Height() int {
	v := viewportStyle.GetVerticalFrameSize()
	return m.viewport.Height + v
}
