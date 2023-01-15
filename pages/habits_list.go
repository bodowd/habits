package pages

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 15

var (
	titleStyle            = lipgloss.NewStyle().MarginLeft(2)
	itemStyle             = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle     = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle       = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle             = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1).Foreground(lipgloss.Color("241"))
	notificationTextStyle = lipgloss.NewStyle().MarginLeft(2).MarginBottom(1)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

type ListModel struct {
	list        list.Model
	choice      string
	quitting    bool
	numRecorded int
	newEntry    string
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.numRecorded++
				m.newEntry = ""
				m.choice = string(i)
			}
			return m, nil

		case "n":
			textInputModel := NewTextInputModel(m)
			// send nil message because other wise it sends "n"
			// And the text field starts with this letter already in there
			return textInputModel.Update(nil)

			// turn off esc exiting the program
		case "esc":
			return m, nil

		}

	case userSavedMsg:
		m.newEntry = msg.text
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	var s string
	if m.choice != "" {
		s = notificationTextStyle.Render(fmt.Sprintf("Recorded %s", m.choice))
	}

	if m.newEntry != "" {
		s = notificationTextStyle.Render(fmt.Sprintf("Added %s as a new goal to track.", m.newEntry))
	}

	if m.quitting {
		s = notificationTextStyle.Render(fmt.Sprintf("You recorded %d completed goals this session. Goodbye", m.numRecorded))
		return s
	}

	return "\n" + s + "\n\n" + m.list.View() + m.helpView()

}

func (m ListModel) helpView() string {
	return helpStyle.Render("\n ↑/k: up • /j: down • q/ctrl+c: quit • n: create entry \n")
}

func NewList() ListModel {
	items := []list.Item{
		item("Work out"),
		item("DS&Alg"),
		item("Coding"),
		item("Read the Ministry"),
		item("Read the Bible"),
		item("Read current book"),
		item("Study German"),
	}

	const defaultWidth = 200

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "What goal did you complete today?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(false)

	m := ListModel{list: l}
	return m
}
