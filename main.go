package main

import (
	"log"
	"os"

	"github.com/bodowd/habits/pages"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(pages.NewList())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
