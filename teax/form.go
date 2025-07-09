package teax

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Created[Item list.DefaultItem] struct {
	Value Item
}

func NewCreated[Item list.DefaultItem](value Item) Created[Item] {
	return Created[Item]{Value: value}
}

type Updated[Item list.DefaultItem] struct {
	Value Item
}

func NewUpdated[Item list.DefaultItem](value Item) Updated[Item] {
	return Updated[Item]{Value: value}
}

type Deleted struct{}

type MaxDim struct {
	Width  int
	Height int
}

type Form[Item list.DefaultItem] struct {
	Create func(maxDim MaxDim) (Mode, tea.Cmd)
	Edit   func(item Item, maxDim MaxDim) (Mode, tea.Cmd)
	Delete Confirmation[Deleted]
}
