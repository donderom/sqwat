package article

import (
	"slices"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/paragraph"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

type Item = squad.Paragraph

type Article struct {
	teax.Model[Item]
	viewport teax.Viewport[squad.Answer]
}

var _ tea.Model = Article{}

var (
	form = teax.Form[Item]{
		Create: func(maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return NewCreateForm(maxDim)
		},
		Edit: func(item Item, maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return NewUpdateForm(item, maxDim)
		},
		Delete: teax.Confirmation[teax.Deleted](
			"Delete this paragraph (with all its questions)?",
		),
	}

	paragraphStyle teax.StyleFunc[Item] = teax.StyleFunc[Item](
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

				if len(item.QAs) > 1 {
					border := style.Border.Multi
					styles.NormalDesc = border.Apply(styles.NormalDesc)
					styles.SelectedDesc = border.Apply(styles.SelectedDesc)
				}

				if len(item.QAs) == 0 {
					border := style.Border.Error
					styles.NormalDesc = border.Apply(styles.NormalDesc)
					styles.SelectedDesc = border.Apply(styles.SelectedDesc)
				}
				return styles
			}
		})

	defaultActions teax.Actions[Item] = teax.DefaultActions[Item]()
	actions        teax.Actions[Item] = teax.Actions[Item]{
		Create: defaultActions.Create.ApplyAndThen(typeEnter),
		Update: defaultActions.Update,
		Delete: defaultActions.Delete,
	}

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
	}

	delegate teax.Delegate[Item] = teax.Delegate[Item]{
		Style:           paragraphStyle,
		ItemName:        "paragraph",
		ShowDescription: true,
		ShortHelpKeys:   keys,
		FullHelpKeys:    slices.Concat(keys, fullKeys),
	}
)

func New(article *squad.Article, saver teax.Saver) Article {
	return Article{
		Model: teax.Model[Item]{
			List:  teax.NewList(article.Paragraphs, article.Title(), delegate),
			Coll:  article,
			Saver: saver,
			Form:  form,
			NewModel: func(item *Item) tea.Model {
				return paragraph.New(item, article.Title(), saver)
			},
			Actions: actions,
		},
		viewport: teax.NewViewport[squad.Answer](),
	}
}

func (m Article) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg)
		m.updateContext()
		return m, nil

	case tea.KeyMsg:
		if m.Mode == nil && !m.List.Filtering() {
			switch {
			case key.Matches(msg, keyset.Next, keyset.Prev):
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}
	}

	m.Model, cmd = m.Model.Update(msg)
	m.updateContext()
	return m, cmd
}

func (m Article) View() string {
	numSections := 3

	helpView := m.HelpView()
	m.List.DecreaseHeight(lipgloss.Height(helpView))

	switch m.Mode.(type) {
	case teax.Confirmation[teax.Deleted]:
		m.List.DecreaseHeight(m.Mode.Height())
		numSections++
	}

	sections := make([]string, 0, numSections)
	if len(m.Coll.All()) == 0 && (m.Mode == nil || m.InSync) {
		m.List.IncreaseHeight(m.viewport.Height())
	}
	sections = append(sections, m.ListView())

	switch m.Mode.(type) {
	case Context:
		sections = append(sections, m.Mode.View())
	default:
		if m.Mode != nil || m.InSync {
			m.viewport.Blur()
		}
		if len(m.Coll.All()) > 0 {
			sections = append(sections, m.viewport.View())
		}
	}

	switch m.Mode.(type) {
	case teax.Confirmation[teax.Deleted]:
		sections = append(sections, m.Mode.View())
	}

	sections = append(sections, style.Bot.Render(helpView))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Article) updateContext() {
	if m.List.ItemSelected() && !m.InSync {
		p := m.Coll.Get(m.List.GlobalIndex())
		if len(p.QAs) == 1 {
			qa := p.QAs[0]
			m.viewport.Highlight([]rune(p.Context), qa.Answers(), qa.Highlight())
		} else {
			m.viewport.SetContent(p.Context)
		}
	}
}

func (m *Article) resize(msg tea.WindowSizeMsg) {
	m.List.Resize(msg)
	halfHeight := m.List.Height() / 2
	m.List.DecreaseHeight(halfHeight)

	m.viewport.Resize(m.List.Width(), halfHeight)

	if m.Mode != nil {
		maxDim := m.List.MaxDim()
		m.Mode = m.Mode.Resize(maxDim.Width, maxDim.Height)
	}
}

func typeEnter(l teax.List[Item], _ tea.Cmd) (teax.List[Item], tea.Cmd) {
	return l, bubblon.Cmd(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
}
