package qna

import (
	"slices"
	"unicode/utf8"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"
	"github.com/donderom/sqwat/text"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

type Disambiguated struct {
	start int
}

func Disambiguate(start int) tea.Cmd {
	return bubblon.Cmd(Disambiguated{start: start})
}

type indexItem struct {
	title string
	start int
	end   int
}

func (i indexItem) Title() string       { return i.title }
func (i indexItem) Description() string { return i.Title() }
func (i indexItem) FilterValue() string { return i.Title() }

var _ list.DefaultItem = indexItem{}

func (i indexItem) From() int { return i.start }
func (i indexItem) To() int   { return i.end }
func (i indexItem) IsIn(context []rune) bool {
	return i.end <= len(context)
}

var _ text.Range = indexItem{}

type Ambi struct {
	list    teax.ViewList[indexItem, indexItem]
	context []rune
	indices []int
}

var _ tea.Model = Ambi{}

var (
	ambiKeys []key.Binding = []key.Binding{
		keyset.Ok,
	}

	ambiFullKeys []key.Binding = []key.Binding{
		keyset.Next,
		keyset.Prev,
	}

	ambiDelegate teax.Delegate[indexItem] = teax.Delegate[indexItem]{
		Style:         teax.IdentityStyles[indexItem](),
		ItemName:      "answer",
		ShortHelpKeys: ambiKeys,
		FullHelpKeys:  slices.Concat(ambiKeys, ambiFullKeys),
	}
)

func NewAmbi(title string, context []rune, answer string, indices []int) Ambi {
	items := make([]indexItem, len(indices))
	for i, idx := range indices {
		items[i] = indexItem{
			title: string(context[idx:]),
			start: idx,
			end:   idx + utf8.RuneCountInString(answer),
		}
	}

	list := teax.NewViewList[indexItem](items, title, ambiDelegate)
	list.HighlightPattern(answer)

	return Ambi{
		list:    list,
		context: context,
		indices: indices,
	}
}

func (m Ambi) Init() tea.Cmd {
	return nil
}

func (m Ambi) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyset.Ok):
			return m, tea.Sequence(
				bubblon.Close,
				Disambiguate(m.indices[m.list.GlobalIndex()]),
			)

		case key.Matches(msg, keyset.Esc):
			return m, bubblon.Close

		case key.Matches(msg, keyset.Quit):
			return m, tea.Quit

		case key.Matches(msg, keyset.Next, keyset.Prev):
			m.list.Viewport, cmd = m.list.Viewport.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.list.Resize(msg)
	}

	m.list, cmd = m.list.Update(msg)
	m.updateContext()
	return m, cmd
}

func (m Ambi) View() string {
	helpView := m.list.Help.View(m.list)
	m.list.DecreaseHeight(lipgloss.Height(helpView))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Top.Render(m.list.View()),
		m.list.Viewport.View(),
		style.Bot.Render(helpView),
	)
}

func (m *Ambi) updateContext() {
	selected := m.list.SelectedItem()
	if selected != nil {
		if ii, ok := selected.(indexItem); ok {
			m.list.Viewport.Highlight(m.context, []indexItem{ii}, style.Highlight)
		}
	}
}
