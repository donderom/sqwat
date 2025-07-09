package main

import (
	"fmt"
	"os"

	"github.com/donderom/sqwat/splash"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

const help = `Usage: sqwat <input>

Arguments:
  input        The input SQuAD JSON file

Example:
  sqwat train-v2.0.json`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(help)
		os.Exit(1)
	}

	controller, err := bubblon.New(splash.New(os.Args[1]))
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

func fail(err error) {
	fmt.Println("Error running program:", err)
	os.Exit(1)
}
