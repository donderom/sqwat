package teax

import (
	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/style"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

type Dataset interface {
	Save() error
}

type Synced[Item list.DefaultItem] struct {
	Item   Item
	Action Action[Item]
	Err    error
	Index  int
}

type Model[Item list.DefaultItem] struct {
	List     List[Item]
	Actions  Actions[Item]
	Form     Form[Item]
	Coll     Collection[Item]
	Mode     Mode
	Dataset  Dataset
	NewModel func(*Item) tea.Model
	InSync   bool
}

func (m Model[Item]) Init() tea.Cmd {
	if len(m.Coll.All()) == 0 {
		mode, cmd := m.Form.Create(m.List.MaxDim())
		return tea.Batch(cmd, SetMode(mode))
	}
	return nil
}

func (m Model[Item]) Update(msg tea.Msg) (Model[Item], tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.InSync {
			return m, nil
		}

		if key.Matches(msg, keyset.Esc) {
			if m.Mode == nil && m.List.Unfiltered() {
				return m, bubblon.Close
			}
			m.Mode = nil
		}

	case NewMode:
		m.Mode = msg.Mode
		return m, tea.WindowSize()

	case tea.WindowSizeMsg:
		m.List.Resize(msg)
		if m.Mode != nil {
			maxDim := m.List.MaxDim()
			m.Mode = m.Mode.Resize(maxDim.Width, maxDim.Height)
		}

	case Created[Item]:
		index := len(m.Coll.All())
		m.Coll.Add(msg.Value)
		return m.Sync(m.Actions.Create, index, *new(Item))

	case Updated[Item]:
		if m.List.ItemSelected() {
			index := m.List.GlobalIndex()
			backup := m.Coll.Get(index)
			m.Coll.Update(index, msg.Value)
			return m.Sync(m.Actions.Update, index, backup)
		}

	case Deleted:
		if m.List.ItemSelected() {
			index := m.List.GlobalIndex()
			backup := m.Coll.Get(index)
			m.Coll.Remove(index)
			return m.Sync(m.Actions.Delete, index, backup)
		}

	case Synced[Item]:
		if msg.Err != nil {
			msg.Action.Revert(m.Coll, msg.Index, msg.Item)
			m.InSync = false
			return m, tea.Batch(
				m.List.ToggleSpinner(),
				m.List.NewStatus(style.Error.Render(msg.Err.Error())),
			)
		}

		m.List, cmd = msg.Action.Apply(m.List, m.Coll, msg.Index)
		m.InSync = false
		return m, tea.Batch(m.List.ToggleSpinner(), cmd)

	case bubblon.Closed:
		if m.List.ItemSelected() {
			index := m.List.GlobalIndex()
			m.List.SetItem(index, m.Coll.Get(index))
		}
	}

	if m.Mode != nil {
		m.Mode, cmd = m.Mode.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.List.Filtering() {
			break
		}

		switch {
		case key.Matches(msg, keyset.Quit):
			return m, tea.Quit

		case key.Matches(msg, keyset.View):
			if m.List.ItemSelected() {
				return m, tea.Sequence(
					m.List.ClearStatus(),
					bubblon.Open(m.NewModel(m.Coll.At(m.List.GlobalIndex()))),
				)
			}

		case key.Matches(msg, keyset.Create):
			m.Mode, cmd = m.Form.Create(m.List.MaxDim())
			return m, cmd

		case key.Matches(msg, keyset.Edit):
			if m.List.ItemSelected() {
				index := m.List.GlobalIndex()
				m.Mode, cmd = m.Form.Edit(m.Coll.Get(index), m.List.MaxDim())
				return m, cmd
			}

		case key.Matches(msg, keyset.Delete):
			if m.List.ItemSelected() {
				m.Mode = m.Form.Delete
			}
			return m, nil
		}
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model[Item]) View() string {
	return ""
}

func (m Model[Item]) ListView() string {
	mainStyle := style.Top.Render

	if m.Mode != nil || m.InSync {
		return mainStyle(style.Faint.Render(m.List.View()))
	}

	return mainStyle(m.List.View())
}

func (m Model[Item]) HelpView() string {
	helpView := m.List.Help.View

	if m.Mode != nil {
		return helpView(m.Mode.KeyMap())
	}

	if len(m.Coll.All()) == 0 {
		if _, ok := m.List.keyHelp[keyset.Esc.Help()]; ok {
			return helpView(keyset.Bindings(keyset.Create, keyset.Esc, keyset.Quit))
		}
		return helpView(keyset.Bindings(keyset.Create, keyset.Quit))
	}

	return helpView(m.List)
}

func (m Model[Item]) Sync(
	action Action[Item],
	index int,
	item Item,
) (Model[Item], tea.Cmd) {
	m.Mode = nil
	m.InSync = true

	return m, tea.Batch(
		m.List.StartSpinner(),
		func() tea.Msg {
			return Synced[Item]{
				Action: action,
				Index:  index,
				Item:   item,
				Err:    m.Dataset.Save(),
			}
		},
	)
}

func (m *Model[Item]) Resize(msg tea.WindowSizeMsg) {
	m.List.Resize(msg)
	if m.Mode != nil {
		maxDim := m.List.MaxDim()
		m.Mode = m.Mode.Resize(maxDim.Width, maxDim.Height)
	}
}
