package paragraph

import (
	"slices"

	"github.com/donderom/sqwat/answer"
	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/qna"
	"github.com/donderom/sqwat/question"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Inverted struct{}

const invert = teax.Confirmation[Inverted]("Invert question possibility?")

type Item = squad.QA

type Paragraph struct {
	teax.Model[Item]
	viewport  teax.Viewport[squad.Answer]
	context   []rune
	paragraph *squad.Paragraph
}

var _ tea.Model = Paragraph{}

var (
	actions teax.Actions[Item] = teax.DefaultActions[Item]()

	keys []key.Binding = []key.Binding{
		keyset.View,
		keyset.Create,
		keyset.Edit,
		keyset.Delete,
		keyset.Esc,
	}

	fullKeys []key.Binding = []key.Binding{
		keyset.Next,
		keyset.Prev,
		keyset.Add,
		keyset.Invert,
	}
)

func New(
	paragraph *squad.Paragraph,
	title string,
	dataset teax.Dataset,
) Paragraph {
	delegate := teax.Delegate[Item]{
		Style:           questionStyle(paragraph),
		ItemName:        "question",
		ShowDescription: true,
		ShortHelpKeys:   keys,
		FullHelpKeys:    slices.Concat(keys, fullKeys),
	}

	context := []rune(paragraph.Context)

	form := teax.Form[Item]{
		Create: func(maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return qna.NewCreateForm(paragraph.Context, maxDim.Width)
		},
		Edit: func(item Item, maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return qna.NewUpdateForm(paragraph.Context, item, maxDim.Width)
		},
		Delete: teax.Confirmation[teax.Deleted](
			"Delete this question (with all its answers)?",
		),
	}

	return Paragraph{
		Model: teax.Model[Item]{
			List:    teax.NewList(paragraph.QAs, title, delegate),
			Coll:    paragraph,
			Dataset: dataset,
			Form:    form,
			NewModel: func(item *Item) tea.Model {
				return question.New(item, context, dataset)
			},
			Actions: actions,
		},
		viewport:  teax.NewViewport[squad.Answer](),
		paragraph: paragraph,
		context:   context,
	}
}

func (m Paragraph) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg)
		m.updateContext()
		return m, nil

	case teax.Created[squad.Answer]:
		return m.update(func(index int) { m.paragraph.QAs[index].Add(msg.Value) })

	case Inverted:
		return m.update(func(index int) { m.paragraph.Invert(index) })
	}

	if m.Mode == nil && !m.List.Filtering() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keyset.Add):
				if m.List.ItemSelected() {
					index := m.List.GlobalIndex()
					form, cmd := answer.NewCreateForm(
						m.paragraph.Context,
						m.Coll.Get(index),
						m.List.MaxDim().Width,
					)
					m.Mode = form
					return m, cmd
				}

			case key.Matches(msg, keyset.Invert):
				if m.List.ItemSelected() {
					m.Mode = invert
				}
				return m, nil

			case key.Matches(msg, keyset.Next, keyset.Prev):
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd

			default:
			}
		}
	}

	m.Model, cmd = m.Model.Update(msg)
	m.updateContext()
	return m, cmd
}

func (m Paragraph) View() string {
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

func (m *Paragraph) updateContext() {
	if !m.InSync {
		if m.List.ItemSelected() {
			qa := m.Coll.Get(m.List.GlobalIndex())
			if len(qa.Answers()) == 0 {
				m.viewport.Blur()
			}
			m.viewport.Highlight(m.context, qa.Answers(), qa.Highlight())
		} else {
			m.viewport.SetContent(m.paragraph.Context)
		}
	}
}

func (m *Paragraph) update(prepare func(index int)) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.List.ItemSelected() {
		index := m.List.GlobalIndex()
		backup := m.Coll.Get(index)
		prepare(index)
		m.Model, cmd = m.Sync(actions.Update, index, backup)
		return m, cmd
	}

	return m, nil
}

func (m *Paragraph) resize(msg tea.WindowSizeMsg) {
	m.Resize(msg)
	halfHeight := m.List.Height() / 2
	m.List.DecreaseHeight(halfHeight)

	m.viewport.Resize(m.List.Width(), halfHeight)
}

func questionStyle(p *squad.Paragraph) teax.StyleFunc[Item] {
	return teax.StyleFunc[Item](
		func(defaultStyles teax.Styles) teax.ItemStyles[Item] {
			altStyles := defaultStyles
			border := style.Border.Alt
			altStyles.NormalTitle = border.Apply(altStyles.NormalTitle)
			altStyles.SelectedTitle = border.Apply(altStyles.SelectedTitle).
				Foreground(border.Color)
			altStyles.SelectedDesc = altStyles.SelectedDesc.
				BorderForeground(border.Color).
				Foreground(style.Palette.Dark.Blue)

			return func(item Item) teax.Styles {
				styles := defaultStyles
				if item.Impossible {
					styles = altStyles
				}

				if slices.IndexFunc(p.QAs, func(qa squad.QA) bool {
					return qa.Question == item.Question && qa.Id != item.Id
				}) != -1 {
					border := style.Border.Dup
					styles.NormalTitle = border.Apply(styles.NormalTitle)
					styles.SelectedTitle = border.Apply(styles.SelectedTitle)
				}

				emptyAnswers := len(item.Answers()) == 0 && !item.Impossible
				if emptyAnswers || item.OutOfRange([]rune(p.Context)) {
					border := style.Border.Error
					styles.NormalDesc = border.Apply(styles.NormalDesc)
					styles.SelectedDesc = border.Apply(styles.SelectedDesc)
					return styles
				}

				if len(item.Answers()) > 1 {
					border := style.Border.Multi
					styles.NormalDesc = border.Apply(styles.NormalDesc)
					styles.SelectedDesc = border.Apply(styles.SelectedDesc)
				}

				return styles
			}
		})
}
