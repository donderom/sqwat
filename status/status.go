package status

import (
	"cmp"
	"slices"

	"github.com/donderom/sqwat/app"
	"github.com/donderom/sqwat/article"
	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/paragraph"
	"github.com/donderom/sqwat/question"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"
	"github.com/donderom/sqwat/validation"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

type Item = validation.ValidationResult

type status struct {
	list     teax.List[Item]
	results  []Item
	dataset  teax.Dataset
	filename string
	data     *squad.SQuAD
}

var _ tea.Model = status{}

func NewStatus(
	filename string,
	data *squad.SQuAD,
	results []Item,
	dataset teax.Dataset,
) status {
	slices.SortFunc(results, func(a, b Item) int {
		return cmp.Or(
			cmp.Compare(a.Type, b.Type),
			cmp.Compare(a.Message, b.Message),
		)
	})

	delegate := teax.Delegate[Item]{
		Style:           teax.IdentityStyles[Item](),
		ItemName:        "warning",
		ShowDescription: true,
	}

	return status{
		data:     data,
		list:     teax.NewList(results, "Warnings", delegate),
		results:  results,
		dataset:  dataset,
		filename: filename,
	}
}

func (m status) Init() tea.Cmd {
	return nil
}

func (m status) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.Resize(msg)

	case tea.KeyMsg:
		if m.list.Unfiltered() {
			switch {
			case key.Matches(msg, keyset.Esc):
				return m, bubblon.Close

			case key.Matches(msg, keyset.Quit):
				return m, tea.Quit

			case key.Matches(msg, keyset.View):
				if m.list.ItemSelected() {
					result := m.results[m.list.GlobalIndex()]
					return m, bubblon.ReplaceAll(m.model(result))
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m status) View() string {
	helpView := m.list.Help.View(m.list)
	m.list.DecreaseHeight(lipgloss.Height(helpView))

	return lipgloss.JoinVertical(lipgloss.Left,
		style.Top.Render(m.list.View()),
		style.Bot.Render(helpView),
	)
}

func (s status) model(result validation.ValidationResult) tea.Model {
	a := s.data.At(result.Path.To(validation.Article))
	p := a.At(result.Path.To(validation.Paragraph))

	appModel := func() tea.Model {
		m := app.New(s.data, s.filename, s.dataset)
		m.List.Select(result.Path.To(validation.Article))
		return m
	}
	articleModel := func() tea.Model {
		m := article.New(a, s.dataset, appModel)
		m.List.Select(result.Path.To(validation.Paragraph))
		return m
	}
	paraModel := func() tea.Model {
		m := paragraph.New(p, a.Title(), s.dataset, articleModel)
		m.List.Select(result.Path.To(validation.Question))
		return m
	}

	switch result.Type {
	case validation.Article:
		return appModel()
	case validation.Paragraph:
		return articleModel()
	case validation.Question:
		return paraModel()
	case validation.Answer:
		q := p.At(result.Path.To(validation.Question))
		m := question.New(q, []rune(p.Context), s.dataset, paraModel)
		m.List.Select(result.Path.To(validation.Answer))
		return m
	}

	return nil
}
