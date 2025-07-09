package answer

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/donderom/sqwat/qna"
	"github.com/donderom/sqwat/squad"
)

func NewCreateForm(
	context string,
	qa squad.QA,
	maxWidth int,
) (qna.QA, tea.Cmd) {
	m := qna.NewQA(context, maxWidth, qna.Create, false)
	cmd := m.Show(qa.Question, "")
	return m, cmd
}

func NewUpdateForm(
	context string,
	qa squad.QA,
	answer squad.Answer,
	maxWidth int,
) (qna.QA, tea.Cmd) {
	m := qna.NewQA(context, maxWidth, qna.Update, false)
	cmd := m.Show(qa.Question, answer.Text)
	return m, cmd
}
