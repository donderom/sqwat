package teax

import (
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Styles = list.DefaultItemStyles

type ItemStyles[T list.Item] func(T) Styles

type Style[T list.Item] interface {
	Transform(styles Styles) ItemStyles[T]
}

type StyleFunc[T list.Item] func(styles Styles) ItemStyles[T]

func (sf StyleFunc[T]) Transform(styles Styles) ItemStyles[T] {
	return sf(styles)
}

func IdentityStyles[T list.Item]() StyleFunc[T] {
	return StyleFunc[T](
		func(styles Styles) ItemStyles[T] {
			return func(a T) Styles {
				return styles
			}
		})
}

type Delegate[T list.Item] struct {
	ShortHelpKeys   []key.Binding
	FullHelpKeys    []key.Binding
	Style           Style[T]
	ItemName        string
	ShowDescription bool
}

type delegate[T list.Item] struct {
	Delegate[T]
	defaultDelegate list.DefaultDelegate
	styles          ItemStyles[T]
}

func NewDelegate[T list.Item](
	defaultDelegate list.DefaultDelegate,
	d Delegate[T],
) delegate[T] {
	return delegate[T]{
		Delegate:        d,
		defaultDelegate: defaultDelegate,
		styles:          d.Style.Transform(defaultDelegate.Styles),
	}
}

func (d delegate[T]) Render(w io.Writer, m list.Model, index int, item list.Item) {
	d.defaultDelegate.Styles = d.styles(item.(T))
	d.defaultDelegate.Render(w, m, index, item)
}

func (d delegate[T]) Height() int {
	return d.defaultDelegate.Height()
}

func (d delegate[T]) Spacing() int {
	return d.defaultDelegate.Spacing()
}

func (d delegate[T]) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return d.defaultDelegate.Update(msg, m)
}
