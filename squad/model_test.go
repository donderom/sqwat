package squad_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/donderom/sqwat/squad"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		data, err := squad.Load(strings.NewReader("{}"))
		require.NoError(t, err)
		assert.Empty(t, data.Articles)
	})

	t.Run("version alone", func(t *testing.T) {
		t.Parallel()

		data, err := squad.Load(strings.NewReader(`{"version": "v2.0"}`))
		require.NoError(t, err)
		assert.Empty(t, data.Articles)
		assert.Equal(t, "v2.0", data.Version)
	})

	t.Run("empty with version", func(t *testing.T) {
		t.Parallel()

		data, err := squad.Load(strings.NewReader(`{"version": "v2.0", "data": []}`))
		require.NoError(t, err)
		assert.Empty(t, data.Articles)
		assert.Equal(t, "v2.0", data.Version)
	})

	t.Run("ignore other fields", func(t *testing.T) {
		t.Parallel()

		data, err := squad.Load(strings.NewReader(`{"what": "nothing"}`))
		require.NoError(t, err)
		assert.Empty(t, data.Articles)
	})

	t.Run("fail on empty file", func(t *testing.T) {
		t.Parallel()

		_, err := squad.Load(strings.NewReader(""))
		assert.Error(t, err)
	})

	t.Run("fail on mailformed JSON", func(t *testing.T) {
		t.Parallel()

		_, err := squad.Load(strings.NewReader("{"))
		assert.Error(t, err)
	})

	t.Run("empty values", func(t *testing.T) {
		t.Parallel()

		data, err := squad.Load(strings.NewReader(empty))
		assert.Equal(t, emptyData, data)
		require.NoError(t, err)
	})

	t.Run("default data", func(t *testing.T) {
		t.Parallel()

		data, err := squad.Load(strings.NewReader(main))
		assert.Equal(t, mainData(), data)
		require.NoError(t, err)
	})
}

func TestSQuAD(t *testing.T) {
	t.Parallel()

	data := mainData()
	article := squad.Article{Name: "test"}

	// Add
	n := len(data.Articles)
	data.Add(article)
	assert.Len(t, data.Articles, n+1)
	assert.Equal(t, article, data.Articles[n])

	// Insert
	data.Insert(0, article)
	assert.Len(t, data.Articles, n+2)
	assert.Equal(t, article, data.Articles[0])

	// Remove
	data.Remove(0)
	assert.Len(t, data.Articles, n+1)
	assert.NotEqual(t, article, data.Articles[0])

	// Update
	update := squad.Article{Name: "update"}
	data.Update(0, update)
	assert.Len(t, data.Articles, n+1)
	assert.Equal(t, update, data.Articles[0])

	// Get
	assert.Equal(t, update, data.Get(0))
	assert.Equal(t, article, data.Get(len(data.Articles)-1))
}

func TestSave(t *testing.T) {
	t.Parallel()

	var s strings.Builder
	data := mainData()
	require.NoError(t, data.Save(&s))

	assert.Equal(t,
		strings.Join(strings.Fields(main), ""),
		strings.Join(strings.Fields(s.String()), ""),
	)
}

func TestArticle(t *testing.T) {
	t.Parallel()

	article := mainData().Articles[0]
	paragraph := squad.Paragraph{Context: "test"}

	title := "Go (programming language)"
	assert.Equal(t, title, article.Title())
	assert.Equal(t, title, article.FilterValue())
	assert.Equal(t, "2 paragraphs", article.Description())

	// Add
	n := len(article.Paragraphs)
	article.Add(paragraph)
	assert.Len(t, article.Paragraphs, n+1)
	assert.Equal(t, paragraph, article.Paragraphs[n])

	// Insert
	article.Insert(0, paragraph)
	assert.Len(t, article.Paragraphs, n+2)
	assert.Equal(t, paragraph, article.Paragraphs[0])

	// Remove
	article.Remove(0)
	assert.Len(t, article.Paragraphs, n+1)
	assert.NotEqual(t, paragraph, article.Paragraphs[0])

	// Update
	start := mainData().Articles[0].Paragraphs[0].QAs[0].Answers()[0].Start
	prefix := "..."
	update := article.Paragraphs[0]
	update.Context = prefix + update.Context
	assert.Equal(t, start, article.Paragraphs[0].QAs[0].Answers()[0].Start)
	article.Update(0, update)
	assert.Len(t, article.Paragraphs, n+1)
	assert.Equal(t, update, article.Paragraphs[0])
	// Check if answer has been shifted
	assert.Equal(t,
		start+len(prefix),
		article.Paragraphs[0].QAs[0].Answers()[0].Start,
	)

	// Get
	assert.Equal(t, update, article.Get(0))
	assert.Equal(t, paragraph, article.Get(len(article.Paragraphs)-1))
}

func TestParagraph(t *testing.T) {
	t.Parallel()

	paragraph := mainData().Articles[0].Paragraphs[0]
	qa := squad.QA{Id: "test"}

	title := "Go is a high-level general purpose programming language that is statically typed and compiled."
	assert.Equal(t, title, paragraph.Title())
	assert.Equal(t, title, paragraph.FilterValue())
	assert.Equal(t, "2 questions", paragraph.Description())

	// Add
	n := len(paragraph.QAs)
	paragraph.Add(qa)
	assert.Len(t, paragraph.QAs, n+1)
	assert.Equal(t, qa, paragraph.QAs[n])

	// Insert
	paragraph.Insert(0, qa)
	assert.Len(t, paragraph.QAs, n+2)
	assert.Equal(t, qa, paragraph.QAs[0])

	// Remove
	paragraph.Remove(0)
	assert.Len(t, paragraph.QAs, n+1)
	assert.NotEqual(t, qa, paragraph.QAs[0])

	// Update
	update := squad.QA{Id: "update"}
	paragraph.Update(0, update)
	assert.Len(t, paragraph.QAs, n+1)
	assert.Equal(t, update, paragraph.QAs[0])

	// Get
	assert.Equal(t, update, paragraph.Get(0))
	assert.Equal(t, qa, paragraph.Get(len(paragraph.QAs)-1))
}

func TestInvertQuestion(t *testing.T) {
	t.Parallel()

	paragraph := mainData().Articles[0].Paragraphs[0]
	qa := squad.QA{
		Id:             "test",
		CorrectAnswers: []squad.Answer{{Text: "x", Start: 2025}},
	}
	paragraph.Add(qa)

	lastIndex := len(paragraph.QAs) - 1
	paragraph.Invert(lastIndex)
	assert.True(t, paragraph.Get(lastIndex).Impossible)
	assert.Empty(t, paragraph.Get(lastIndex).CorrectAnswers)
	assert.Len(t, paragraph.Get(lastIndex).PlausibleAnswers, len(qa.Answers()))

	paragraph.Invert(lastIndex)
	assert.Equal(t, qa, paragraph.Get(lastIndex))
}

func TestQA(t *testing.T) {
	t.Parallel()

	t.Run("correct", func(t *testing.T) {
		t.Parallel()

		qa := mainData().Articles[0].Paragraphs[0].QAs[0]
		testQA(t, qa)
	})

	t.Run("plausible", func(t *testing.T) {
		t.Parallel()

		qa := mainData().Articles[0].Paragraphs[0].QAs[0]
		testQA(t, squad.NewQA(qa.Question, slices.Clone(qa.CorrectAnswers), true))
	})
}

func TestAnswer(t *testing.T) {
	t.Parallel()

	answer := mainData().Articles[0].Paragraphs[0].QAs[0].Answers()[0]
	title := "high-level general purpose programming language"
	assert.Equal(t, title, answer.Title())
	assert.Equal(t, title, answer.FilterValue())
	assert.Equal(t, title, answer.Description())

	assert.Equal(t, 8, answer.From())
	assert.Equal(t, len(title)+8, answer.To())
}

func testQA(t *testing.T, qa squad.QA) {
	t.Helper()

	title := "What is Go?"
	assert.Equal(t, title, qa.Title())
	assert.Equal(t, title, qa.FilterValue())
	assert.Equal(t, "high-level general purpose programming language", qa.Description())

	answer := squad.Answer{Text: "test"}

	// Add
	n := len(qa.Answers())
	qa.Add(answer)
	assert.Len(t, qa.Answers(), n+1)
	assert.Equal(t, answer, qa.Answers()[n])

	// Insert
	qa.Insert(0, answer)
	assert.Len(t, qa.Answers(), n+2)
	assert.Equal(t, answer, qa.Answers()[0])

	// Remove
	qa.Remove(0)
	assert.Len(t, qa.Answers(), n+1)
	assert.NotEqual(t, answer, qa.Answers()[0])

	// Update
	update := squad.Answer{Text: "update"}
	qa.Update(0, update)
	assert.Len(t, qa.Answers(), n+1)
	assert.Equal(t, update, qa.Answers()[0])

	// Get
	assert.Equal(t, update, qa.Get(0))
	assert.Equal(t, answer, qa.Get(len(qa.Answers())-1))
}

var empty = `{
		"data": [
			{
				"title": "",
				"paragraphs": [
					{
						"context": "",
						"qas": [
							{
								"id": "",
								"question": "",
								"answers": [
									{
										"text": "",
										"answer_start": 0
									}
								]
							}
						]
					}
				]
			}
		]
	}`

var emptyData = &squad.SQuAD{
	Version: "",
	Articles: []squad.Article{
		{
			Name: "",
			Paragraphs: []squad.Paragraph{
				{
					Context: "",
					QAs: []squad.QA{
						{
							Id:       "",
							Question: "",
							CorrectAnswers: []squad.Answer{
								{
									Text:  "",
									Start: 0,
								},
							},
						},
					},
				},
			},
		},
	},
}

var main = `{
		"version": "v2.0",
		"data": [
			{
				"title": "Go (programming language)",
				"paragraphs": [
					{
						"context": "Go is a high-level general purpose programming language that is statically typed and compiled.",
						"qas": [
							{
								"id": "1",
								"question": "What is Go?",
								"answers": [
									{
										"text": "high-level general purpose programming language",
										"answer_start": 8
									}
								],
								"is_impossible": false
							},
							{
								"id": "2",
								"question": "Is Go dynamic?",
								"answers": [
									{
										"text": "statically typed",
										"answer_start": 64
									}
								],
								"is_impossible": false
							}
						]
					},
					{
						"context": "It is syntactically similar to C",
						"qas": [
							{
								"id": "3",
								"question": "What is it similar to?",
								"answers": [
									{
										"text": "C",
										"answer_start": 31
									}
								],
								"is_impossible": false
							}
						]
					}
				]
			},
			{
				"title": "Rust (programming language)",
				"paragraphs": [
					{
						"context": "Rust is a general-purpose programming language emphasizing performance, type safety, and concurrency.",
						"qas": [
							{
								"id": "4",
								"question": "What does Rust emphasize that Go do not?",
								"answers": [
									{
										"text": "type safety",
										"answer_start": 72
									}
								],
								"is_impossible": false
							}
						]
					}
				]
			}
		]
	}`

func mainData() *squad.SQuAD {
	return &squad.SQuAD{
		Version: "v2.0",
		Articles: []squad.Article{
			{
				Name: "Go (programming language)",
				Paragraphs: []squad.Paragraph{
					{
						Context: "Go is a high-level general purpose programming language that is statically typed and compiled.",
						QAs: []squad.QA{
							{
								Id:       "1",
								Question: "What is Go?",
								CorrectAnswers: []squad.Answer{
									{
										Text:  "high-level general purpose programming language",
										Start: 8,
									},
								},
							},
							{
								Id:       "2",
								Question: "Is Go dynamic?",
								CorrectAnswers: []squad.Answer{
									{
										Text:  "statically typed",
										Start: 64,
									},
								},
							},
						},
					},
					{
						Context: "It is syntactically similar to C",
						QAs: []squad.QA{
							{
								Id:       "3",
								Question: "What is it similar to?",
								CorrectAnswers: []squad.Answer{
									{
										Text:  "C",
										Start: 31,
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "Rust (programming language)",
				Paragraphs: []squad.Paragraph{
					{
						Context: "Rust is a general-purpose programming language emphasizing performance, type safety, and concurrency.",
						QAs: []squad.QA{
							{
								Id:       "4",
								Question: "What does Rust emphasize that Go do not?",
								CorrectAnswers: []squad.Answer{
									{
										Text:  "type safety",
										Start: 72,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
