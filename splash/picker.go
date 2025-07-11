package splash

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"

	"github.com/donderom/sqwat/keyset"
)

var title = lipgloss.NewStyle().
	MarginBottom(1).
	MarginTop(1).
	MarginLeft(2).
	Render("Pick a SQuAD JSON file:")

type picker struct {
	filepicker filepicker.Model
}

var _ tea.Model = picker{}

func NewPicker(path string) picker {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".json"}
	fp.CurrentDirectory = path

	return picker{filepicker: fp}
}

func (m picker) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keyset.Quit) {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if selected, path := m.filepicker.DidSelectFile(msg); selected {
		return m, bubblon.Replace(New(path))
	}

	return m, cmd
}

func (m picker) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.filepicker.View(),
	)
}
