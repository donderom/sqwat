package teax

import (
	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Confirmation[M tea.Msg] string

func (c Confirmation[M]) Update(msg tea.Msg) (Mode, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyset.Ok):
			return c, func() tea.Msg { return *new(M) }
		}
	}
	return c, nil
}

func (c Confirmation[M]) View() string {
	return style.Mid.Inherit(style.Highlight).Render(string(c))
}

func (c Confirmation[M]) Height() int {
	return 1 + style.Mid.GetVerticalFrameSize()
}

func (c Confirmation[M]) KeyMap() help.KeyMap {
	return keyset.KeyMaps.Confirm
}

func (c Confirmation[M]) Resize(width, height int) Mode {
	return c
}
