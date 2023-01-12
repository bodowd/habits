package main

import (
	"log"
	"os"

	"github.com/bodowd/habits/cmd"
	tea "github.com/charmbracelet/bubbletea"
)

var models []tea.Model

func main() {
	// p := tea.NewProgram(cmd.InitialModel())
	models = []tea.Model{cmd.NewList()}
	m := models[0]
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
