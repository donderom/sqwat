package app

import (
	"slices"

	"github.com/donderom/sqwat/article"
	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

type Item = squad.Article

type App struct {
	teax.Model[Item]
}

var _ tea.Model = App{}

var (
	form = teax.Form[Item]{
		Create: func(maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return NewCreateForm(maxDim.Width)
		},
		Edit: func(item Item, maxDim teax.MaxDim) (teax.Mode, tea.Cmd) {
			return NewUpdateForm(item, maxDim.Width)
		},
		Delete: teax.Confirmation[teax.Deleted](
			"Delete this article (with all its paragraphs)?",
		),
	}

	articleStyle teax.StyleFunc[Item] = teax.StyleFunc[Item](
		func(defaultStyles teax.Styles) teax.ItemStyles[Item] {
			return func(item Item) teax.Styles {
				styles := defaultStyles
				if len(item.Paragraphs) > 1 {
					border := style.Border.Multi
					styles.NormalDesc = border.Apply(styles.NormalDesc)
					styles.SelectedDesc = border.Apply(styles.SelectedDesc)
				}

				if len(item.Paragraphs) == 0 {
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
	}

	fullKeys []key.Binding = []key.Binding{
		keyset.Status,
	}

	delegate teax.Delegate[Item] = teax.Delegate[Item]{
		Style:           articleStyle,
		ItemName:        "article",
		ShowDescription: true,
		ShortHelpKeys:   keys,
		FullHelpKeys:    slices.Concat(keys, fullKeys),
	}
)

func New(squad *squad.SQuAD, title string, dataset teax.Dataset) App {
	return App{
		Model: teax.Model[Item]{
			List:    teax.NewList(squad.Articles, title, delegate),
			Coll:    squad,
			Dataset: dataset,
			Form:    form,
			NewModel: func(item *Item) tea.Model {
				return article.New(item, dataset, nil)
			},
			Actions: actions,
		},
	}
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keyset.Esc) && m.List.Unfiltered() {
			m.Mode = nil
			return m, nil
		}
	}

	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m App) View() string {
	numSections := 2

	helpView := m.HelpView()
	m.List.DecreaseHeight(lipgloss.Height(helpView))

	if m.Mode != nil {
		m.List.DecreaseHeight(m.Mode.Height())
		numSections++
	}

	sections := make([]string, 0, numSections)
	sections = append(sections, m.ListView())

	if m.Mode != nil {
		sections = append(sections, m.Mode.View())
	}

	sections = append(sections, style.Bot.Render(helpView))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func typeEnter(l teax.List[Item], _ tea.Cmd) (teax.List[Item], tea.Cmd) {
	return l, bubblon.Cmd(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
}
