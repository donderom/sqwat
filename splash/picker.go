package splash

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/style"
)

var title = lipgloss.NewStyle().
	MarginBottom(1).
	MarginTop(1).
	MarginLeft(2).
	Render("Pick a SQuAD JSON file:")

type picker struct {
	filepicker filepicker.Model
	help       help.Model
	height     int
}

var _ tea.Model = picker{}

func NewPicker(path string) picker {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".json"}
	fp.CurrentDirectory = path

	return picker{
		filepicker: fp,
		help:       help.New(),
	}
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

		if key.Matches(msg, keyset.More) {
			m.help.ShowAll = !m.help.ShowAll
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if selected, path := m.filepicker.DidSelectFile(msg); selected {
		return m, bubblon.Replace(New(path))
	}

	return m, cmd
}

func (m picker) View() string {
	if m.help.ShowAll {
		m.filepicker.SetHeight(m.height - len(m.FullHelp()[0]) - 4)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.filepicker.View(),
		style.Bot.Render(m.HelpView()),
	)
}

func (m picker) ShortHelp() []key.Binding {
	return []key.Binding{
		m.filepicker.KeyMap.Down,
		m.filepicker.KeyMap.Up,
		m.filepicker.KeyMap.Back,
		m.filepicker.KeyMap.Open,
		m.filepicker.KeyMap.Select,
		keyset.More,
	}
}

func (m picker) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		m.ShortHelp()[:5],
		{
			m.filepicker.KeyMap.GoToTop,
			m.filepicker.KeyMap.GoToLast,
			m.filepicker.KeyMap.PageUp,
			m.filepicker.KeyMap.PageDown,
		},
		{
			keyset.CloseMore,
		},
	}
}

func (m picker) HelpView() string {
	return m.help.View(m)
}
