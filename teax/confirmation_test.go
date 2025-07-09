package teax_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/donderom/sqwat/teax"
)

type Message struct{}

func TestConfirmationUpdate(t *testing.T) {
	t.Parallel()

	confirmation := teax.Confirmation[Message]("")
	_, cmd := confirmation.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, Message{}, cmd())
}

func TestConfirmationView(t *testing.T) {
	t.Parallel()

	msg := "message"
	confirmation := teax.Confirmation[Message](msg)
	assert.Contains(t, confirmation.View(), msg)
}
