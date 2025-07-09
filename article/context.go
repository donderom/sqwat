package article

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"
	"github.com/donderom/sqwat/text"
)

type mode uint8

const (
	create mode = iota
	update
)

var errEmptyContext error = errors.New("oops! Looks like the context is missing")

type Context struct {
	area textarea.Model
	item squad.Paragraph
	err  error
	mode mode
}

var _ teax.Mode = Context{}

func NewUpdateForm(
	paragraph squad.Paragraph,
	maxDim teax.MaxDim,
) (Context, tea.Cmd) {
	form := newContext(update)
	form.item = paragraph
	form.area.SetWidth(maxDim.Width)
	form.area.SetHeight(maxDim.Height)
	form.area.SetValue(paragraph.Context)
	form.area.CursorEnd()
	return form, form.area.Focus()
}

func NewCreateForm(maxDim teax.MaxDim) (Context, tea.Cmd) {
	form := newContext(create)
	form.area.SetWidth(maxDim.Width)
	form.area.SetHeight(maxDim.Height)
	form.area.Reset()
	return form, form.area.Focus()
}

func (m Context) Init() tea.Cmd {
	return nil
}

func (m Context) Update(msg tea.Msg) (teax.Mode, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = nil

		switch {
		case key.Matches(msg, keyset.Save):
			value := strings.TrimSpace(m.area.Value())

			if value == "" {
				m.err = errEmptyContext
				return m, nil
			}

			if m.mode == create {
				item := squad.Paragraph{Context: value}
				return m, bubblon.Cmd(teax.NewCreated(item))
			}

			m.item.Context = value
			return m, bubblon.Cmd(teax.NewUpdated(m.item))
		}
	}

	m.area, cmd = m.area.Update(msg)
	return m, cmd
}

func (m Context) View() string {
	numSections := 1
	if m.err != nil {
		numSections++
	}

	sections := make([]string, 0, numSections)

	if m.err != nil {
		errMsg := style.Error.Render(text.Capitalize(m.err.Error()))
		sections = append(sections, style.SepBot.Render(errMsg))
	}

	sections = append(sections, m.area.View())
	return style.Mid.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func (m Context) Height() int {
	var errorHeight int
	if m.err != nil {
		errorHeight += 1 + style.SepBot.GetVerticalFrameSize()
	}
	return 1 + style.Mid.GetVerticalFrameSize() + errorHeight
}

func (m Context) KeyMap() help.KeyMap {
	return keyset.KeyMaps.Edit
}

func (m Context) Resize(width, height int) teax.Mode {
	m.area.SetWidth(width)
	m.area.SetHeight(height)
	return m
}

func newContext(mode mode) Context {
	input := textarea.New()
	input.Prompt = lipgloss.NormalBorder().Left
	input.FocusedStyle.LineNumber = style.Faint
	input.FocusedStyle.CursorLineNumber = style.Highlight
	return Context{
		area: input,
		mode: mode,
	}
}
