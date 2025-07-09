package teax

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/style"
)

const statusMessageLifetime = 5 * time.Second

type model = list.Model

type List[T list.DefaultItem] struct {
	model
}

func NewList[T list.DefaultItem](
	items []T,
	title string,
	delegate Delegate[T],
) List[T] {
	listItems := make([]list.Item, len(items))
	for idx, item := range items {
		listItems[idx] = item
	}

	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.ShowDescription = delegate.ShowDescription
	d := NewDelegate(defaultDelegate, delegate)

	list := list.New(listItems, d, 0, 0)
	list.Title = title
	list.SetStatusBarItemName(d.ItemName, d.ItemName+"s")
	list.StatusMessageLifetime = statusMessageLifetime
	list.SetSpinner(spinner.MiniDot)
	list.KeyMap.NextPage = keyset.NextPage
	list.AdditionalShortHelpKeys = func() []key.Binding { return d.ShortHelpKeys }
	list.AdditionalFullHelpKeys = func() []key.Binding { return d.FullHelpKeys }
	list.SetShowHelp(false)
	list.KeyMap.Quit = keyset.Quit

	return List[T]{model: list}
}

func (m List[T]) Update(msg tea.Msg) (List[T], tea.Cmd) {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

func (m *List[T]) HighlightPattern(pattern string) {
	m.Filter = list.UnsortedFilter
	m.SetFilterText(pattern)
	m.KeyMap.ClearFilter.SetEnabled(false)
	m.KeyMap.Filter.SetEnabled(false)
}

func (m List[T]) ItemSelected() bool {
	return m.SelectedItem() != nil
}

func (m List[T]) Filtering() bool {
	return m.FilterState() == list.Filtering
}

func (m List[T]) Unfiltered() bool {
	return m.FilterState() == list.Unfiltered
}

func (m List[T]) MaxDim() MaxDim {
	return MaxDim{
		Width:  m.Width(),
		Height: m.Height() - 1,
	}
}

func (m *List[T]) Remove(index int) tea.Cmd {
	m.RemoveItem(index)
	return nil
}

func (m *List[T]) Insert(index int, item T) tea.Cmd {
	cmd := m.InsertItem(index, item)
	m.Select(index)
	return cmd
}

func (m *List[T]) DecreaseHeight(v int) {
	m.SetHeight(m.Height() - v)
}

func (m *List[T]) IncreaseHeight(v int) {
	m.SetHeight(m.Height() + v)
}

func (m *List[T]) Resize(msg tea.WindowSizeMsg) {
	h, v := style.Top.GetFrameSize()
	m.SetSize(msg.Width-h, msg.Height-v)
}

func (m *List[T]) ClearStatus() tea.Cmd {
	return m.setStatus("", 0)
}

func (m *List[T]) NewStatus(status string) tea.Cmd {
	return m.setStatus(status, statusMessageLifetime)
}

func (m *List[T]) setStatus(status string, timeout time.Duration) tea.Cmd {
	m.StatusMessageLifetime = timeout
	return m.NewStatusMessage(status)
}
