package teax

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Collection[Item list.DefaultItem] interface {
	Add(item Item)
	Insert(index int, item Item)
	Remove(index int)
	Update(index int, item Item)
	Get(index int) Item
	At(index int) *Item
	All() []Item
}

type ApplyFunc[Item list.DefaultItem] func(
	list List[Item],
	coll Collection[Item],
	index int,
) (List[Item], tea.Cmd)

type RevertFunc[Item list.DefaultItem] func(
	coll Collection[Item],
	index int,
	item Item,
)

type Action[Item list.DefaultItem] struct {
	Apply  ApplyFunc[Item]
	Revert RevertFunc[Item]
}

type Actions[Item list.DefaultItem] struct {
	Create Action[Item]
	Update Action[Item]
	Delete Action[Item]
}

func (a Action[Item]) WithApply(
	f func(ApplyFunc[Item]) ApplyFunc[Item],
) Action[Item] {
	a.Apply = f(a.Apply)
	return a
}

func (a Action[Item]) WithRevert(
	f func(RevertFunc[Item]) RevertFunc[Item],
) Action[Item] {
	a.Revert = f(a.Revert)
	return a
}

func (a Action[Item]) ApplyAndThen(
	f func(list List[Item], cmd tea.Cmd) (List[Item], tea.Cmd),
) Action[Item] {
	return a.WithApply(func(apply ApplyFunc[Item]) ApplyFunc[Item] {
		return func(
			list List[Item],
			coll Collection[Item],
			index int,
		) (List[Item], tea.Cmd) {
			return f(apply(list, coll, index))
		}
	})
}

type Create[Item list.DefaultItem] struct{}

func (_ Create[Item]) Apply(
	list List[Item],
	coll Collection[Item],
	index int,
) (List[Item], tea.Cmd) {
	cmd := list.Insert(index, coll.Get(index))
	return list, cmd
}

func (_ Create[Item]) Revert(coll Collection[Item], index int, _ Item) {
	coll.Remove(index)
}

func (a Create[Item]) toAction() Action[Item] {
	return Action[Item]{Apply: a.Apply, Revert: a.Revert}
}

type Update[Item list.DefaultItem] struct{}

func (_ Update[Item]) Apply(
	list List[Item],
	coll Collection[Item],
	index int,
) (List[Item], tea.Cmd) {
	cmd := list.SetItem(index, coll.Get(index))
	return list, cmd
}

func (_ Update[Item]) Revert(coll Collection[Item], index int, item Item) {
	coll.Update(index, item)
}

func (a Update[Item]) toAction() Action[Item] {
	return Action[Item]{Apply: a.Apply, Revert: a.Revert}
}

type Delete[Item list.DefaultItem] struct{}

func (_ Delete[Item]) Apply(
	list List[Item],
	_ Collection[Item],
	index int,
) (List[Item], tea.Cmd) {
	cmd := list.Remove(index)
	return list, cmd
}

func (_ Delete[Item]) Revert(coll Collection[Item], index int, item Item) {
	coll.Insert(index, item)
}

func (a Delete[Item]) toAction() Action[Item] {
	return Action[Item]{Apply: a.Apply, Revert: a.Revert}
}

func DefaultActions[Item list.DefaultItem]() Actions[Item] {
	return Actions[Item]{
		Create: Create[Item]{}.toAction(),
		Update: Update[Item]{}.toAction(),
		Delete: Delete[Item]{}.toAction(),
	}
}
