package app

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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

var errEmptyTitle error = errors.New("oops! Looks like the title is missing")

type Title struct {
	input textinput.Model
	item  squad.Article
	err   error
	mode  mode
}

var _ teax.Mode = Title{}

func NewCreateForm(maxWidth int) (Title, tea.Cmd) {
	form := newTitle(create, maxWidth)
	form.input.Reset()
	return form, form.input.Focus()
}

func NewUpdateForm(article squad.Article, maxWidth int) (Title, tea.Cmd) {
	form := newTitle(update, maxWidth)
	form.item = article
	form.input.SetValue(article.Title())
	form.input.CursorEnd()
	return form, form.input.Focus()
}

func (m Title) Update(msg tea.Msg) (teax.Mode, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = nil

		switch {
		case key.Matches(msg, keyset.Ok):
			value := strings.TrimSpace(m.input.Value())

			if value == "" {
				m.err = errEmptyTitle
				return m, nil
			}

			if m.mode == create {
				item := squad.Article{Name: value}
				return m, bubblon.Cmd(teax.NewCreated(item))
			}

			m.item.Name = value
			return m, bubblon.Cmd(teax.NewUpdated(m.item))
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Title) View() string {
	numSections := 1
	if m.err != nil {
		numSections++
	}

	sections := make([]string, 0, numSections)

	if m.err != nil {
		errMsg := style.Error.Render(text.Capitalize(m.err.Error()))
		sections = append(sections, style.SepBot.Render(errMsg))
	}

	sections = append(sections, m.input.View())
	return style.Mid.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func (m Title) Height() int {
	var errorHeight int
	if m.err != nil {
		errorHeight += 1 + style.SepBot.GetVerticalFrameSize()
	}
	return 1 + style.Mid.GetVerticalFrameSize() + errorHeight
}

func (m Title) KeyMap() help.KeyMap {
	return keyset.KeyMaps.Confirm
}

func (m Title) Resize(width, height int) teax.Mode {
	m.input.Width = width - 1
	return m
}

func newTitle(mode mode, width int) Title {
	input := textinput.New()
	input.PromptStyle = style.Highlight
	input.Width = width - 1
	return Title{
		input: input,
		mode:  mode,
	}
}
