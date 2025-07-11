package main

import (
	"fmt"
	"os"

	"github.com/donderom/sqwat/splash"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

func main() {
	model, err := model()
	if err != nil {
		fail(err)
	}

	controller, err := bubblon.New(model)
	if err != nil {
		fail(err)
	}

	p := tea.NewProgram(controller, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fail(err)
	}

	if m, ok := m.(bubblon.Controller); ok && m.Err != nil {
		fail(m.Err)
	}
}

func model() (tea.Model, error) {
	if len(os.Args) < 2 {
		return splash.NewPicker("."), nil
	}

	path := os.Args[1]

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return splash.NewPicker(path), nil
	}

	return splash.New(path), nil
}

func fail(err error) {
	fmt.Println("Error running program:", err)
	os.Exit(1)
}
