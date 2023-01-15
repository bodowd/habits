package main

import (
	"log"
	"os"

	"github.com/bodowd/habits/pages"
	tea "github.com/charmbracelet/bubbletea"
)

var models []tea.Model

type Model struct {
	GoToNewEntryPage bool
	ListPage         pages.ListModel
	NewEntryPage     pages.TextInputModel
}

func NewModel() Model {
	m := Model{
		GoToNewEntryPage: false,
		ListPage:         pages.NewList(),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.ListPage.Update(msg)
}

func (m Model) View() string {
	return m.ListPage.View()
}

func main() {
	// p := tea.NewProgram(cmd.InitialModel())
	// models = []tea.Model{pages.NewTextInputModel(), pages.NewList()}
	// m := models[0]
	p := tea.NewProgram(NewModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
