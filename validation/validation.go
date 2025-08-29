package validation

import (
	"context"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/donderom/sqwat/squad"

	"github.com/charmbracelet/bubbles/list"
)

type ItemType uint8

const (
	Article ItemType = iota
	Paragraph
	Question
	Answer
)

func (t ItemType) String() string {
	switch t {
	case Article:
		return "Article"
	case Paragraph:
		return "Paragraph"
	case Question:
		return "Question"
	case Answer:
		return "Answer"
	}
	return "Unknown"
}

type Path map[ItemType]int

func (p Path) To(itemType ItemType) int {
	return p[itemType]
}

type ValidationResult struct {
	Message string
	Path    Path
	Type    ItemType
}

var _ list.DefaultItem = ValidationResult{}

func (vr ValidationResult) Title() string       { return vr.Message }
func (vr ValidationResult) Description() string { return vr.Type.String() }
func (vr ValidationResult) FilterValue() string { return vr.Message }

type ValidationFunc func(
	ctx context.Context,
	article squad.Article,
	index int,
) []ValidationResult

var Validators = []ValidationFunc{
	ValidateEmptyTitle,
	ValidateEmptyParagraphs,
	ValidateEmptyContext,
	ValidateEmptyQAs,
	ValidateEmptyID,
	ValidateEmptyQuestion,
	ValidateNoAnswers,
	ValidateDupQuestions,
	ValidateEmptyAnswer,
	ValidateOutOfContextAnswer,
}

func Run(ctx context.Context, s *squad.SQuAD) []ValidationResult {
	return RunValidations(ctx, s, Validators)
}

func RunValidations(
	ctx context.Context,
	s *squad.SQuAD,
	validators []ValidationFunc,
) []ValidationResult {
	maxWorkers := runtime.NumCPU() * 2
	tasks := genTasks(ctx, s, validators)
	results := validate(ctx, tasks, maxWorkers)

	var collected []ValidationResult
	for {
		select {
		case <-ctx.Done():
			return collected
		case r, ok := <-results:
			if !ok {
				return collected
			}
			collected = append(collected, r)
		}
	}
}

var ValidateEmptyTitle = validateArticle(
	"Empty title",
	func(article squad.Article) bool {
		return strings.TrimSpace(article.Name) == ""
	},
)

var ValidateEmptyParagraphs = validateArticle(
	"No paragraphs",
	func(article squad.Article) bool {
		return len(article.Paragraphs) == 0
	},
)

var ValidateEmptyContext = validateParagraph(
	"Empty context",
	func(paragraph squad.Paragraph) bool {
		return strings.TrimSpace(paragraph.Context) == ""
	},
)

var ValidateEmptyQAs = validateParagraph(
	"No QAs",
	func(paragraph squad.Paragraph) bool {
		return len(paragraph.QAs) == 0
	},
)

var ValidateEmptyID = validateQuestion(
	"Empty ID",
	func(qa squad.QA, _ squad.Paragraph) bool {
		return strings.TrimSpace(qa.Id) == ""
	},
)

var ValidateEmptyQuestion = validateQuestion(
	"Empty question",
	func(qa squad.QA, _ squad.Paragraph) bool {
		return strings.TrimSpace(qa.Question) == ""
	},
)

var ValidateNoAnswers = validateQuestion(
	"No answers",
	func(qa squad.QA, _ squad.Paragraph) bool {
		return !qa.Impossible && len(qa.Answers()) == 0
	},
)

var ValidateDupQuestions = validateQuestion(
	"Duplicate question with impossible counterpart",
	func(qa squad.QA, para squad.Paragraph) bool {
		question := strings.TrimSpace(qa.Question)
		if question == "" {
			return false
		}

		return slices.IndexFunc(para.QAs, func(other squad.QA) bool {
			equestion := question == strings.TrimSpace(other.Question)
			return equestion && !qa.Impossible && other.Impossible
		}) != -1
	},
)

var ValidateEmptyAnswer = validateAnswer(
	"Empty answer",
	func(answer squad.Answer, _ []rune) bool {
		return strings.TrimSpace(answer.Text) == ""
	},
)

var ValidateOutOfContextAnswer = validateAnswer(
	"Answer is out of context",
	func(answer squad.Answer, context []rune) bool {
		return strings.TrimSpace(answer.Text) != "" && !answer.IsIn(context)
	},
)

func genTasks(
	ctx context.Context,
	s *squad.SQuAD,
	validators []ValidationFunc,
) <-chan task {
	tasks := make(chan task)

	go func() {
		defer close(tasks)
		for _, validator := range validators {
			for i, article := range s.Articles {
				select {
				case <-ctx.Done():
					return
				case tasks <- newTask(validator, article, i):
				}
			}
		}
	}()

	return tasks
}

func validate(
	ctx context.Context,
	tasks <-chan task,
	maxWorkers int,
) <-chan ValidationResult {
	results := make(chan ValidationResult)

	var wg sync.WaitGroup
	for range maxWorkers {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-tasks:
					if !ok {
						return
					}
					for _, r := range task.run(ctx) {
						select {
						case <-ctx.Done():
							return
						case results <- r:
						}
					}
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

type task struct {
	validator ValidationFunc
	article   squad.Article
	index     int
}

func (t task) run(ctx context.Context) []ValidationResult {
	return t.validator(ctx, t.article, t.index)
}

func newTask(validator ValidationFunc, article squad.Article, index int) task {
	return task{
		validator: validator,
		article:   article,
		index:     index,
	}
}

func validateArticle(
	message string,
	cond func(squad.Article) bool,
) ValidationFunc {
	return func(
		ctx context.Context,
		article squad.Article,
		index int,
	) []ValidationResult {
		if ctx.Err() != nil {
			return nil
		}

		if cond(article) {
			return []ValidationResult{{
				Message: message,
				Type:    Article,
				Path: Path{
					Article: index,
				},
			}}
		}

		return nil
	}
}

func validateParagraph(
	message string,
	cond func(squad.Paragraph) bool,
) ValidationFunc {
	return func(
		ctx context.Context,
		article squad.Article,
		index int,
	) []ValidationResult {
		var results []ValidationResult

		if ctx.Err() != nil {
			return results
		}

		for i, para := range article.Paragraphs {
			if ctx.Err() != nil {
				return results
			}

			if cond(para) {
				results = append(results, ValidationResult{
					Message: message,
					Type:    Paragraph,
					Path: Path{
						Article:   index,
						Paragraph: i,
					},
				})
			}
		}

		return results
	}
}

func validateQuestion(
	message string,
	cond func(qa squad.QA, para squad.Paragraph) bool,
) ValidationFunc {
	return func(
		ctx context.Context,
		article squad.Article,
		index int,
	) []ValidationResult {
		var results []ValidationResult

		if ctx.Err() != nil {
			return results
		}

		for i, para := range article.Paragraphs {
			if ctx.Err() != nil {
				return results
			}

			for j, qa := range para.QAs {
				if ctx.Err() != nil {
					return results
				}

				if cond(qa, para) {
					results = append(results, ValidationResult{
						Message: message,
						Type:    Question,
						Path: Path{
							Article:   index,
							Paragraph: i,
							Question:  j,
						},
					})
				}
			}
		}

		return results
	}
}

func validateAnswer(
	message string,
	cond func(answer squad.Answer, context []rune) bool,
) ValidationFunc {
	return func(
		ctx context.Context,
		article squad.Article,
		index int,
	) []ValidationResult {
		var results []ValidationResult

		if ctx.Err() != nil {
			return results
		}

		for i, para := range article.Paragraphs {
			if ctx.Err() != nil {
				return results
			}

			context := []rune(para.Context)

			for j, qa := range para.QAs {
				if ctx.Err() != nil {
					return results
				}

				for k, answer := range qa.Answers() {
					if ctx.Err() != nil {
						return results
					}

					if cond(answer, context) {
						results = append(results, ValidationResult{
							Message: message,
							Type:    Answer,
							Path: Path{
								Article:   index,
								Paragraph: i,
								Question:  j,
								Answer:    k,
							},
						})
					}
				}
			}
		}

		return results
	}
}
