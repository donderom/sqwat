package teax

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
)

type Mode interface {
	Update(msg tea.Msg) (Mode, tea.Cmd)
	View() tea.View
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
