package keyset

import "github.com/charmbracelet/bubbles/key"

var (
	Quit key.Binding = key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	)

	View key.Binding = NewEnter("view")
	Ok   key.Binding = NewEnter("ok")

	Edit key.Binding = key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	)

	Delete key.Binding = key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	)

	Create key.Binding = key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	)

	Add key.Binding = key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add answer"),
	)

	NextPage key.Binding = key.NewBinding(
		key.WithKeys("right", "l", "pgdown", "f"),
		key.WithHelp("→/l/pgdn", "next page"),
	)

	Next key.Binding = key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "scroll down"),
	)

	Prev key.Binding = key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "scroll up"),
	)

	Esc key.Binding = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	)

	More key.Binding = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "more"),
	)

	CloseMore key.Binding = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close more"),
	)

	Save key.Binding = key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	)

	Tab key.Binding = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch input"),
	)

	Invert key.Binding = key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "invert possibility"),
	)
)

func NewEnter(desc string) key.Binding {
	return key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↩", desc),
	)
}

type KeyMap struct {
	bindings []key.Binding
}

func (km KeyMap) ShortHelp() []key.Binding {
	return km.bindings
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

func Bindings(bindings ...key.Binding) KeyMap {
	return KeyMap{bindings: bindings}
}

var KeyMaps = struct {
	Confirm KeyMap
	Edit    KeyMap
}{
	Confirm: Bindings(Ok, Esc),
	Edit:    Bindings(Save, Esc),
}
