package question

import (
	"slices"

	"github.com/donderom/sqwat/answer"
	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item = squad.Answer

type Question struct {
	teax.Model[Item]
	viewport teax.Viewport[squad.Answer]
	context  []rune
	qa       *squad.QA
}

var _ tea.Model = Question{}

var (
	actions teax.Actions[Item] = teax.DefaultActions[Item]()

	keys []key.Binding = []key.Binding{
		keyset.Create,
		keyset.Edit,
		keyset.Delete,
	}

	fullKeys []key.Binding = []key.Binding{
		keyset.Next,
		keyset.Prev,
	}
)

func New(qa *squad.QA, context []rune, saver teax.Saver) Question {
	delegate := teax.Delegate[Item]{
		Style:         answerStyle(context, qa.Impossible),
		ItemName:      qa.Modifier() + "answer",
		ShortHelpKeys: keys,
		FullHelpKeys:  slices.Concat(keys, fullKeys),
	}

	form := teax.Form[Item]{
		Create: func(maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return answer.NewCreateForm(string(context), *qa, maxDim.Width)
		},
		Edit: func(item Item, maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return answer.NewUpdateForm(string(context), *qa, item, maxDim.Width)
		},
		Delete: teax.Confirmation[teax.Deleted]("Delete this answer?"),
	}

	return Question{
		Model: teax.Model[Item]{
			List:     teax.NewList(qa.Answers(), qa.Question, delegate),
			Coll:     qa,
			Saver:    saver,
			Form:     form,
			NewModel: func(_ *Item) tea.Model { return nil },
			Actions:  actions,
		},
		viewport: teax.NewViewport[squad.Answer](),
		qa:       qa,
		context:  context,
	}
}

func (m Question) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg)
		halfHeight := m.List.Height() / 2
		m.List.DecreaseHeight(halfHeight)

		m.viewport.Resize(m.List.Width(), halfHeight)

		m.updateContext()
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Mode == nil && key.Matches(msg, keyset.Next, keyset.Prev) {
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	m.Model, cmd = m.Model.Update(msg)
	m.updateContext()
	return m, cmd
}

func (m Question) View() string {
	numSections := 3

	helpView := m.HelpView()
	m.List.DecreaseHeight(lipgloss.Height(helpView))

	if m.Mode != nil {
		m.List.DecreaseHeight(m.Mode.Height())
		numSections++
	}

	sections := make([]string, 0, numSections)
	sections = append(sections, m.ListView())

	if m.Mode != nil {
		m.viewport.Blur()
	}
	sections = append(sections, m.viewport.View())

	if m.Mode != nil {
		sections = append(sections, m.Mode.View())
	}

	sections = append(sections, style.Bot.Render(helpView))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Question) updateContext() {
	if !m.InSync {
		if m.List.ItemSelected() {
			index := m.List.GlobalIndex()
			answer := m.qa.Answers()[index : index+1]
			m.viewport.Highlight(m.context, answer, m.qa.Highlight())
		} else {
			m.viewport.Blur()
			m.viewport.SetContent(string(m.context))
		}
	}
}

func answerStyle(context []rune, impossible bool) teax.StyleFunc[Item] {
	return teax.StyleFunc[Item](
		func(defaultStyles teax.Styles) teax.ItemStyles[Item] {
			altStyles := defaultStyles
			border := style.Border.Alt
			altStyles.SelectedTitle = border.Apply(altStyles.SelectedTitle).
				Foreground(border.Color)

			return func(item Item) teax.Styles {
				styles := defaultStyles

				if impossible {
					styles = altStyles
				}

				if !item.IsIn(context) {
					border := style.Border.Error
					styles.NormalTitle = border.Apply(styles.NormalTitle)
					styles.SelectedTitle = border.Apply(styles.SelectedTitle)
				}

				return styles
			}
		})
}
