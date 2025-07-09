package squad

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/text"
)

type SQuAD struct {
	Version  string    `json:"version"`
	Articles []Article `json:"data"`
}

type Article struct {
	// It's not Title to not have field and method with the same name
	Name       string      `json:"title"`
	Paragraphs []Paragraph `json:"paragraphs"`
}

var _ list.DefaultItem = Article{}

type Paragraph struct {
	Context string `json:"context"`
	QAs     []QA   `json:"qas"`
}

var _ list.DefaultItem = Paragraph{}

type QA struct {
	Id               string   `json:"id"`
	Question         string   `json:"question"`
	CorrectAnswers   []Answer `json:"answers"`
	PlausibleAnswers []Answer `json:"plausible_answers,omitempty"`
	Impossible       bool     `json:"is_impossible"`
}

var _ list.DefaultItem = QA{}

type Answer struct {
	Text  string `json:"text"`
	Start int    `json:"answer_start"`
}

var _ list.DefaultItem = Answer{}
var _ text.Range = Answer{}

func Load(r io.Reader) (*SQuAD, error) {
	var squad SQuAD
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&squad); err != nil {
		return nil, err
	}

	return &squad, nil
}

func (s *SQuAD) Add(item Article) {
	s.Articles = append(s.Articles, item)
}

func (s *SQuAD) Insert(index int, item Article) {
	s.Articles = slices.Insert(s.Articles, index, item)
}

func (s *SQuAD) Remove(index int) {
	s.Articles = slices.Delete(s.Articles, index, index+1)
}

func (s *SQuAD) Update(index int, item Article) {
	s.Articles[index] = item
}

func (s *SQuAD) Get(index int) Article {
	return s.Articles[index]
}

func (s *SQuAD) At(index int) *Article {
	return &s.Articles[index]
}

func (s *SQuAD) All() []Article {
	return s.Articles
}

func (s *SQuAD) Save(w io.Writer) error {
	jsonData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(jsonData)
	return err
}

func (a *Article) Add(item Paragraph) {
	a.Paragraphs = append(a.Paragraphs, item)
}

func (a *Article) Insert(index int, item Paragraph) {
	a.Paragraphs = slices.Insert(a.Paragraphs, index, item)
}

func (a *Article) Remove(index int) {
	a.Paragraphs = slices.Delete(a.Paragraphs, index, index+1)
}

func (a *Article) Update(index int, item Paragraph) {
	a.Paragraphs[index] = item

	p := &a.Paragraphs[index]
	context := []rune(p.Context)

	// Best effort to fix mismatched answers
	for _, qa := range p.QAs {
		answers := qa.Answers()
		for i := range answers {
			answer := &answers[i]
			if !answer.IsIn(context) {
				indices := text.Indices(p.Context, answer.Text)
				if len(indices) == 1 {
					answer.Start = indices[0]
				}
			}
		}
	}
}

func (a *Article) Get(index int) Paragraph {
	return a.Paragraphs[index]
}

func (a *Article) At(index int) *Paragraph {
	return &a.Paragraphs[index]
}

func (a *Article) All() []Paragraph {
	return a.Paragraphs
}

func (a Article) Title() string { return a.Name }

func (a Article) Description() string {
	return desc(a.Paragraphs, "paragraphs")
}

func (a Article) FilterValue() string { return a.Name }

func (p *Paragraph) Add(item QA) {
	p.QAs = append(p.QAs, item)
}

func (p *Paragraph) Insert(index int, item QA) {
	p.QAs = slices.Insert(p.QAs, index, item)
}

func (p *Paragraph) Remove(index int) {
	p.QAs = slices.Delete(p.QAs, index, index+1)
}

func (p *Paragraph) Update(index int, item QA) {
	p.QAs[index] = item
}

func (p *Paragraph) Get(index int) QA {
	return p.QAs[index]
}

func (p *Paragraph) At(index int) *QA {
	return &p.QAs[index]
}

func (p *Paragraph) All() []QA {
	return p.QAs
}

func (p *Paragraph) Invert(index int) {
	qa := &p.QAs[index]
	if qa.Impossible {
		qa.Impossible = false
		qa.CorrectAnswers = slices.Clone(qa.PlausibleAnswers)
		qa.PlausibleAnswers = nil
	} else {
		qa.Impossible = true
		qa.PlausibleAnswers = slices.Clone(qa.CorrectAnswers)
		qa.CorrectAnswers = nil
	}
}

func (p Paragraph) Title() string { return p.Context }

func (p Paragraph) Description() string {
	return desc(p.QAs, "questions")
}

func (p Paragraph) FilterValue() string { return p.Context }

func NewQA(question string, answers []Answer, impossible bool) QA {
	qa := QA{
		Id:         uuid.New().String(),
		Question:   question,
		Impossible: impossible,
	}
	if impossible {
		qa.PlausibleAnswers = answers
	} else {
		qa.CorrectAnswers = answers
	}
	return qa
}

func (q QA) Answers() []Answer {
	if q.Impossible {
		return q.PlausibleAnswers
	}
	return q.CorrectAnswers
}

func (q *QA) Add(a Answer) {
	if q.Impossible {
		q.PlausibleAnswers = append(q.PlausibleAnswers, a)
	} else {
		q.CorrectAnswers = append(q.CorrectAnswers, a)
	}
}

func (q *QA) Insert(index int, a Answer) {
	if q.Impossible {
		q.PlausibleAnswers = slices.Insert(q.PlausibleAnswers, index, a)
	} else {
		q.CorrectAnswers = slices.Insert(q.CorrectAnswers, index, a)
	}
}

func (q *QA) Update(index int, a Answer) {
	if q.Impossible {
		q.PlausibleAnswers[index] = a
	} else {
		q.CorrectAnswers[index] = a
	}
}

func (q *QA) Remove(index int) {
	if q.Impossible {
		q.PlausibleAnswers = slices.Delete(q.PlausibleAnswers, index, index+1)
	} else {
		q.CorrectAnswers = slices.Delete(q.CorrectAnswers, index, index+1)
	}
}

func (q *QA) Get(index int) Answer {
	return q.Answers()[index]
}

func (q *QA) At(index int) *Answer {
	return &q.Answers()[index]
}

func (q *QA) All() []Answer {
	return q.Answers()
}

func (q QA) Highlight() lipgloss.Style {
	if q.Impossible {
		return style.Alt
	}
	return style.Highlight
}

func (q QA) Modifier() string {
	if q.Impossible {
		return "plausible "
	}
	return ""
}

func (q QA) OutOfRange(context []rune) bool {
	return slices.IndexFunc(q.Answers(), func(a Answer) bool {
		return !a.IsIn(context)
	}) != -1
}

func (q QA) Title() string { return q.Question }

func (q QA) Description() string {
	return desc(q.Answers(), q.Modifier()+"answers")
}

func (q QA) FilterValue() string { return q.Question }

func (a Answer) Title() string       { return a.Text }
func (a Answer) Description() string { return a.Title() }
func (a Answer) FilterValue() string { return a.Title() }

func (a Answer) From() int { return a.Start }
func (a Answer) To() int   { return a.Start + utf8.RuneCountInString(a.Text) }

func (a Answer) IsIn(context []rune) bool {
	end := a.Start + utf8.RuneCountInString(a.Text)
	return string(context[a.Start:end]) == a.Text
}

func desc[T list.DefaultItem](items []T, label string) string {
	num := len(items)
	if num == 0 {
		return "No " + label
	}
	if num == 1 {
		return items[0].Title()
	}
	return fmt.Sprintf("%d %s", num, label)
}
