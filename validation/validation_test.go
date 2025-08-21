package validation_test

import (
	"context"
	"testing"

	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEmptyTitle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		article := squad.Article{Name: "Test"}

		results := validation.ValidateEmptyTitle(ctx, article, 0)
		assert.Empty(t, results)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		assertValidationResult(t,
			squad.Article{Name: ""},
			validation.ValidateEmptyTitle,
			"Empty title",
			validation.Article,
		)
	})
}

func TestValidateEmptyParagraphs(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		article := squad.Article{
			Paragraphs: []squad.Paragraph{
				{
					Context: "",
				},
			},
		}

		results := validation.ValidateEmptyParagraphs(ctx, article, 0)
		assert.Empty(t, results)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		assertValidationResult(t,
			squad.Article{},
			validation.ValidateEmptyParagraphs,
			"No paragraphs",
			validation.Article,
		)
	})
}

func TestValidateEmptyContext(t *testing.T) {
	t.Parallel()

	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				Context: "",
			},
			{
				Context: "Context",
			},
		},
	}

	assertValidationResult(t,
		article,
		validation.ValidateEmptyContext,
		"Empty context",
		validation.Paragraph,
	)
}

func TestValidateEmptyQAs(t *testing.T) {
	t.Parallel()

	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				QAs: nil,
			},
			{
				QAs: []squad.QA{{}},
			},
		},
	}

	assertValidationResult(t,
		article,
		validation.ValidateEmptyQAs,
		"No QAs",
		validation.Paragraph,
	)
}

func TestValidateEmptyID(t *testing.T) {
	t.Parallel()

	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				QAs: []squad.QA{
					{
						Id: "",
					},
					{
						Id: "id",
					},
				},
			},
		},
	}

	assertValidationResult(t,
		article,
		validation.ValidateEmptyID,
		"Empty ID",
		validation.Question,
	)
}

func TestValidateEmptyQuestion(t *testing.T) {
	t.Parallel()

	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				QAs: []squad.QA{
					{
						Question: "",
					},
					{
						Question: "question",
					},
				},
			},
		},
	}

	assertValidationResult(t,
		article,
		validation.ValidateEmptyQuestion,
		"Empty question",
		validation.Question,
	)
}

func TestValidateNoAnswers(t *testing.T) {
	t.Parallel()

	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				QAs: []squad.QA{
					{},
					{
						CorrectAnswers: []squad.Answer{{}},
					},
					{
						Impossible:       true,
						PlausibleAnswers: []squad.Answer{{}},
					},
				},
			},
		},
	}

	assertValidationResult(t,
		article,
		validation.ValidateNoAnswers,
		"No answers",
		validation.Question,
	)
}

func TestValidateDupQuestions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	q := "wat?"
	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				QAs: []squad.QA{
					{
						Question: q,
					},
					{
						Question: q,
					},
					{
						Question:   q,
						Impossible: true,
					},
				},
			},
		},
	}
	index := 0
	path := validation.Path{
		validation.Article:   index,
		validation.Paragraph: index,
		validation.Question:  index,
	}

	results := validation.ValidateDupQuestions(ctx, article, index)
	require.Len(t, results, 2)

	for i, result := range results {
		path[validation.Question] = i
		assert.Equal(t, "Duplicate question with a different type", result.Message)
		assert.Equal(t, validation.Question, result.Type)
		assert.Equal(t, path, result.Path)
	}
}

func TestValidateEmptyAnswer(t *testing.T) {
	t.Parallel()

	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				QAs: []squad.QA{
					{
						CorrectAnswers: []squad.Answer{{Text: ""}},
					},
					{
						CorrectAnswers: []squad.Answer{{Text: "Answer"}},
					},
				},
			},
		},
	}

	assertValidationResult(t,
		article,
		validation.ValidateEmptyAnswer,
		"Empty answer",
		validation.Answer,
	)
}

func TestValidateOutOfContextAnswer(t *testing.T) {
	t.Parallel()

	article := squad.Article{
		Paragraphs: []squad.Paragraph{
			{
				Context: "Just when I thought I was out, they pull me back in!",
				QAs: []squad.QA{
					{
						CorrectAnswers: []squad.Answer{
							{
								Text:  "out",
								Start: 25,
							},
							{
								Text:  "in",
								Start: 49,
							},
						},
					},
				},
			},
		},
	}

	assertValidationResult(t,
		article,
		validation.ValidateOutOfContextAnswer,
		"Answer is out of context",
		validation.Answer,
	)
}

func assertValidationResult(
	t *testing.T,
	article squad.Article,
	validator validation.ValidationFunc,
	message string,
	itemType validation.ItemType,
) {
	t.Helper()

	ctx := context.Background()
	index := 0
	path := validation.Path{validation.Article: index}
	if itemType > validation.Article {
		path[validation.Paragraph] = index
	}
	if itemType > validation.Paragraph {
		path[validation.Question] = index
	}
	if itemType > validation.Question {
		path[validation.Answer] = index
	}

	results := validator(ctx, article, index)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, message, result.Message)
	assert.Equal(t, itemType, result.Type)
	assert.Equal(t, path, result.Path)
}
