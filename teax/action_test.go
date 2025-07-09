package teax_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/donderom/sqwat/teax"
)

type Item struct {
	title string
	desc  string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

var (
	delegate = teax.Delegate[Item]{Style: teax.IdentityStyles[Item]()}
	testItem = Item{title: "test"}
	fillItem = Item{title: "fill"}
	newItem  = Item{title: "new"}
)

type Coll struct{ items []Item }

func (c *Coll) Add(item Item)               { c.items = append(c.items, item) }
func (c *Coll) Insert(index int, item Item) { c.items = slices.Insert(c.items, index, item) }
func (c *Coll) Remove(index int)            { c.items = slices.Delete(c.items, index, index+1) }
func (c *Coll) Update(index int, item Item) { c.items[index] = item }
func (c *Coll) Get(index int) Item          { return c.items[index] }
func (c *Coll) At(index int) *Item          { return &c.items[index] }
func (c *Coll) All() []Item                 { return c.items }

func TestCreateAction(t *testing.T) {
	t.Parallel()

	coll := &Coll{[]Item{testItem, fillItem}}
	list := teax.NewList(coll.items, "", delegate)
	action := teax.DefaultActions[Item]().Create

	// Apply
	assert.Len(t, list.Items(), len(coll.items))
	coll.Add(newItem)
	assert.NotEqual(t, len(coll.items), len(list.Items()))
	lastIndex := len(coll.items) - 1
	list, _ = action.Apply(list, coll, lastIndex)
	assert.Len(t, list.Items(), len(coll.items))
	assert.Equal(t, newItem, list.Items()[lastIndex])

	// Revert
	assert.Len(t, coll.items, 3)
	action.Revert(coll, lastIndex, Item{})
	assert.Len(t, coll.items, 2)
}

func TestUpdateAction(t *testing.T) {
	t.Parallel()

	coll := &Coll{[]Item{testItem, fillItem}}
	list := teax.NewList(coll.items, "", delegate)
	action := teax.DefaultActions[Item]().Update

	// Apply
	coll.Update(0, newItem)
	assert.Equal(t, testItem, list.Items()[0])
	list, _ = action.Apply(list, coll, 0)
	assert.Len(t, list.Items(), len(coll.items))
	assert.Equal(t, newItem, list.Items()[0])

	// Revert
	assert.Len(t, coll.items, 2)
	action.Revert(coll, 0, testItem)
	assert.Len(t, coll.items, 2)
	assert.Equal(t, testItem, coll.items[0])
}

func TestDeleteAction(t *testing.T) {
	t.Parallel()

	coll := &Coll{[]Item{testItem, fillItem}}
	list := teax.NewList(coll.items, "", delegate)
	action := teax.DefaultActions[Item]().Delete

	// Apply
	assert.Len(t, list.Items(), len(coll.items))
	coll.Remove(0)
	assert.NotEqual(t, len(coll.items), len(list.Items()))
	list, _ = action.Apply(list, coll, 0)
	assert.Len(t, list.Items(), len(coll.items))

	// Revert
	assert.Len(t, coll.items, 1)
	assert.NotEqual(t, testItem, coll.Get(0))
	action.Revert(coll, 0, testItem)
	assert.Len(t, coll.items, 2)
	assert.Equal(t, testItem, coll.Get(0))
}
