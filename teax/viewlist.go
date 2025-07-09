package teax

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/text"
)

type ViewList[Item list.DefaultItem, Child text.Range] struct {
	List[Item]
	Viewport Viewport[Child]
}

func NewViewList[Child text.Range, Item list.DefaultItem](
	items []Item,
	title string,
	delegate Delegate[Item],
) ViewList[Item, Child] {
	return ViewList[Item, Child]{
		List:     NewList(items, title, delegate),
		Viewport: NewViewport[Child](),
	}
}

func (m ViewList[Item, Child]) Update(msg tea.Msg) (ViewList[Item, Child], tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyset.Next, keyset.Prev):
			m.Viewport, cmd = m.Viewport.Update(msg)
			return m, cmd
		}
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *ViewList[Item, Child]) Resize(msg tea.WindowSizeMsg) {
	m.List.Resize(msg)
	halfHeight := m.Height() / 2
	m.DecreaseHeight(halfHeight)

	m.Viewport.Resize(m.Width(), halfHeight)
}
