package splash

import (
	"context"
	"fmt"
	"os"

	"github.com/donderom/sqwat/app"
	"github.com/donderom/sqwat/squad"
	"github.com/donderom/sqwat/status"
	"github.com/donderom/sqwat/style"
	"github.com/donderom/sqwat/teax"
	"github.com/donderom/sqwat/validation"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

type loaded struct {
	dataset *squad.SQuAD
}

type failed struct {
	err error
}

type Splash struct {
	spinner  spinner.Model
	filename string
	width    int
	height   int
}

var _ tea.Model = Splash{}

func New(filename string) Splash {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = style.Highlight

	return Splash{
		spinner:  s,
		filename: filename,
	}
}

func (m Splash) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.load())
}

func (m Splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case failed:
		return m, bubblon.Fail(msg.err)

	case loaded:
		dataset := dataset{data: msg.dataset, filename: m.filename}
		model := app.New(msg.dataset, m.filename, dataset)
		return m, bubblon.Replace(model)

	case tea.WindowSizeMsg:
		h, v := style.App.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Splash) View() string {
	return style.Center(m.width, m.height).Render(
		fmt.Sprintf("%s Loading file %s...", m.spinner.View(), m.filename),
	)
}

func (m Splash) load() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(m.filename)
		if err != nil {
			return failed{err: err}
		}

		dataset, err := squad.Load(file)
		if err != nil {
			return failed{err: err}
		}

		if err = file.Close(); err != nil {
			return failed{err: err}
		}

		return loaded{dataset: dataset}
	}
}

type dataset struct {
	data     *squad.SQuAD
	filename string
}

var _ teax.Dataset = dataset{}

func (d dataset) Save() error {
	file, err := os.Create(d.filename)
	if err != nil {
		return err
	}

	if err = d.data.Save(file); err != nil {
		return err
	}

	if err = file.Close(); err != nil {
		return err
	}

	return nil
}

func (d dataset) Status(ctx context.Context) tea.Model {
	results := validation.Run(ctx, d.data)
	return status.NewStatus(d.filename, d.data, results, d)
}
