package teax

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode interface {
	Update(msg tea.Msg) (Mode, tea.Cmd)
	View() string
	Height() int
	KeyMap() help.KeyMap
	Resize(width, height int) Mode
}

type NewMode struct {
	Mode Mode
}

func SetMode(mode Mode) tea.Cmd {
	return func() tea.Msg {
		return NewMode{Mode: mode}
	}
}
