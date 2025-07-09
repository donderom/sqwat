package teax

import "github.com/charmbracelet/bubbles/list"

type Saver struct {
	save func() error
}

func NewSaver(save func() error) Saver {
	return Saver{save: save}
}

func (s Saver) Save() error {
	return s.save()
}

type Synced[Item list.DefaultItem] struct {
	Item   Item
	Action Action[Item]
	Err    error
	Index  int
}
