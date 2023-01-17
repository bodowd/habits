package pages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gorm.io/gorm"
)

type (
	errMsg error
)

type TextInputModel struct {
	textInput textinput.Model
	text      string
	err       error
	listModel ListModel
	db        *gorm.DB
}

func NewTextInputModel(listModel ListModel) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter habit"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return TextInputModel{
		textInput: ti,
		err:       nil,
		listModel: listModel,
	}
}

type DuplicateError struct{}

func (e *DuplicateError) Error() string {
	return "This entry already exists"
}

type userSavedMsg struct {
	text string
}

func (m TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlO:
			return m.listModel.Update(msg)
		case tea.KeyEnter:
			m.text = m.textInput.Value()
			_, err := m.listModel.db.CreateHabit(m.text)
			if err != nil {
				m.text = ""
				if strings.Contains(err.Error(), "UNIQUE") {
					m.err = &DuplicateError{}
				}
				return m, nil
			}
			saved := userSavedMsg{
				text: m.text,
			}
			// switch view back to list
			return m.listModel.Update(saved)
		}

		// handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m TextInputModel) View() string {
	var s string = ""
	s = "What new goal do you want to track?\n\n"
	s += fmt.Sprintf(
		"%s\n\n%s",
		m.textInput.View(),
		helpStyle.Render("\n ctrl+c: quit • ctrl+o: back • enter: save entry\n"),
	) + "\n"

	if m.text != "" {
		s += fmt.Sprintf("Saved %s!", m.text)
	}

	if m.err != nil {
		s += fmt.Sprintf("Error: %s", m.err.Error())
	}

	return s
}
