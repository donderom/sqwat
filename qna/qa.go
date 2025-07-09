package qna

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/donderom/sqwat/keyset"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"
	"github.com/donderom/sqwat/text"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

type mode uint8

const (
	Create mode = iota
	Update
)

const (
	inputQuestion = iota
	inputAnswer
)

var (
	sepTop        lipgloss.Style = lipgloss.NewStyle().MarginTop(1)
	empty         lipgloss.Style = lipgloss.NewStyle()
	questionTitle string         = style.Highlight.Render("Question") +
		style.Faint.Render(" (required)")
	answerTitle           string = sepTop.Render(style.Highlight.Render("Answer"))
	errEmptyQuestion      error  = errors.New("question cannot be empty")
	errAnswerOutOfContext error  = errors.New("answer should be a part of context")
)

type QA struct {
	questionStyle lipgloss.Style
	item          squad.QA
	inputs        []textinput.Model
	err           error
	context       string
	focused       int
	navigation    bool
	mode          mode
}

var _ teax.Mode = QA{}

func NewCreateForm(context string, maxWidth int) (QA, tea.Cmd) {
	m := NewQA(context, maxWidth, Create, true)
	cmd := m.Show("", "")
	return m, cmd
}

func NewUpdateForm(context string, qa squad.QA, maxWidth int) (QA, tea.Cmd) {
	m := NewQA(context, maxWidth, Update, true)
	m.item = qa
	var cmd tea.Cmd

	if len(qa.Answers()) == 0 {
		cmd = m.Show(qa.Question, "")
		return m, cmd
	}

	cmd = m.Show(qa.Question, qa.Answers()[0].Text)
	return m, cmd
}

func NewQA(context string, maxWidth int, mode mode, navigation bool) QA {
	q := textinput.New()
	q.Width = maxWidth
	q.Prompt = ""
	q.Validate = validateQuestion

	a := textinput.New()
	a.Width = maxWidth
	a.Prompt = ""
	a.Validate = validateAnswer(context, mode, navigation)

	questionStyle := empty
	focused := inputQuestion
	if !navigation {
		questionStyle = style.Faint
		focused = inputAnswer
	}

	return QA{
		inputs:        []textinput.Model{q, a},
		context:       context,
		questionStyle: questionStyle,
		focused:       focused,
		navigation:    navigation,
		mode:          mode,
	}
}

func (m QA) Update(msg tea.Msg) (teax.Mode, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = nil

		switch {
		case key.Matches(msg, keyset.Tab):
			if m.navigation {
				m.inputs[m.focused].Blur()
				m.focused = (m.focused + 1) % 2
				return m, m.inputs[m.focused].Focus()
			}

		case key.Matches(msg, keyset.Ok):
			if err := m.firstError(); err != nil {
				m.err = err
				return m, nil
			}

			// TODO: it should be is_impossible or not allowed in 1.1
			if m.answer() != "" {
				indices := text.Indices(m.context, m.answer())
				if len(indices) > 1 {
					return m, bubblon.Open(m.newAmbi(indices))
				}
			}

			answerIndex := strings.Index(m.context, m.answer())
			start := utf8.RuneCountInString(m.context[:answerIndex])
			return m, m.newValue(start)
		}

	case Disambiguated:
		return m, m.newValue(msg.start)
	}

	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m QA) View() string {
	sections := make([]string, 0, 5)

	if m.err != nil {
		msg := strings.ToUpper(m.err.Error()[:1]) + m.err.Error()[1:]
		errMsg := style.Error.Render(msg)
		sections = append(sections, style.SepBot.Render(errMsg))
	}

	sections = append(sections,
		questionTitle,
		m.questionStyle.Render(m.inputView(inputQuestion)),
		answerTitle,
		m.inputView(inputAnswer),
	)

	return style.Mid.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func (m QA) Height() int {
	var errorHeight int
	if m.err != nil {
		errorHeight += 2
	}

	return len(m.inputs)*2 +
		sepTop.GetVerticalMargins() +
		style.Mid.GetVerticalFrameSize() +
		errorHeight
}

func (m QA) KeyMap() help.KeyMap {
	if m.navigation {
		return keyset.Bindings(keyset.Ok, keyset.Tab, keyset.Esc)
	}

	return keyset.KeyMaps.Confirm
}

func (m QA) Resize(width, height int) teax.Mode {
	for i := range m.inputs {
		m.inputs[i].Width = width
	}
	return m
}

func (m *QA) Show(q, a string) tea.Cmd {
	m.inputs[inputQuestion].SetValue(q)
	m.inputs[inputAnswer].SetValue(a)

	for i := range m.inputs {
		m.inputs[i].Blur()
		m.inputs[i].CursorEnd()
	}

	return m.inputs[m.focused].Focus()
}

func (m QA) question() string {
	return m.inputs[inputQuestion].Value()
}

func (m QA) answer() string {
	return m.inputs[inputAnswer].Value()
}

func (m QA) inputView(input int) string {
	view := m.inputs[input].View()
	if m.inputs[input].Err != nil {
		return style.Error.Render(view)
	}

	return view
}

func (m QA) firstError() error {
	if err := m.validate(inputQuestion); err != nil {
		return err
	}

	if err := m.validate(inputAnswer); err != nil {
		return err
	}

	return nil
}

func (m QA) newAmbi(indices []int) Ambi {
	return NewAmbi(m.question(), []rune(m.context), m.answer(), indices)
}

func (m QA) newValue(start int) tea.Cmd {
	return func() tea.Msg {
		var a *squad.Answer
		if strings.TrimSpace(m.answer()) != "" {
			a = &squad.Answer{Text: m.answer(), Start: start}
		}

		if m.mode == Create {
			return m.newCreateValue(a)
		}

		if m.mode == Update {
			return m.newUpdateValue(a)
		}

		return nil
	}
}

func (m QA) newCreateValue(answer *squad.Answer) tea.Msg {
	if m.navigation {
		var answers []squad.Answer
		if answer != nil {
			answers = []squad.Answer{*answer}
		}
		item := squad.NewQA(m.question(), answers, false)
		return teax.NewCreated(item)
	}

	if answer == nil {
		return nil
	}

	return teax.NewCreated(*answer)
}

func (m QA) newUpdateValue(answer *squad.Answer) tea.Msg {
	if m.navigation {
		m.item.Question = m.question()
		answers := m.item.Answers()

		if answer == nil && len(answers) != 0 {
			m.item.Remove(0)
		}

		if answer != nil && len(answers) == 0 {
			m.item.Add(*answer)
		}

		if answer != nil && len(answers) > 0 {
			m.item.Update(0, *answer)
		}

		return teax.NewUpdated(m.item)
	}

	return teax.NewUpdated(*answer)
}

func (m QA) validate(input int) error {
	validateFunc := m.inputs[input].Validate
	return validateFunc(m.inputs[input].Value())
}

func validateQuestion(s string) error {
	if strings.TrimSpace(s) == "" {
		return errEmptyQuestion
	}

	return nil
}

func validateAnswer(context string,
	mode mode,
	navigation bool,
) textinput.ValidateFunc {
	return func(s string) error {
		if s == "" && mode == Create && navigation {
			return nil
		}

		if s == "" || !strings.Contains(context, s) {
			return errAnswerOutOfContext
		}

		return nil
	}
}
