package pages

import (
	"fmt"
	"io"

	"github.com/bodowd/habits/data"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gorm.io/gorm"
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
	list               list.Model
	choice             string
	numRecorded        int
	db                 data.Database
	errorMessage       string
	streak             int
	StatusMessageFlags StatusMessageFlags
}

type StatusMessageFlags struct {
	newEntry        string
	quitting        bool
	alreadyRecorded bool
	newRecord       bool
	archived        bool
	restoredHabit   string
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) updateHabitsList() ListModel {
	habits := m.db.GetActiveHabits()
	habitItems := itemsToList(habits)
	m.list.SetItems(habitItems)
	return m
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.StatusMessageFlags.quitting = true
			return m, tea.Quit

		case "q":
			return m, nil

		case "enter":
			// reset message flags
			m.StatusMessageFlags = StatusMessageFlags{}
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)

				completion, err := m.db.RecordCompletion(m.choice)
				if err != nil {
					m.StatusMessageFlags.alreadyRecorded = true
					return m, nil
				}
				if err == nil {
					// make sure this flag is set to false so that
					m.numRecorded++
					m.StatusMessageFlags.newRecord = true
				}

				m.streak = completion.Streak
			}
			return m, nil

		case "n":
			// reset message flags
			m.StatusMessageFlags = StatusMessageFlags{}

			textInputModel := NewTextInputModel(m)
			// send nil message because other wise it sends "n"
			// And the text field starts with this letter already in there
			return textInputModel.Update(nil)

			// turn off esc exiting the program
		case "esc":
			return m, nil

		case "a":
			// reset message flags
			m.StatusMessageFlags = StatusMessageFlags{}

			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			err := m.db.ArchiveHabit(m.choice)
			if err != nil {
				fmt.Println(err)
			}
			m.StatusMessageFlags.archived = true
			m = m.updateHabitsList()
			return m, nil

		case "r":
			m.StatusMessageFlags = StatusMessageFlags{}
			// go to restore habits page
			// restoreHabitsModel.Update(nil)
			archivedHabitsModel := NewArchivedHabitsModel(m)
			return archivedHabitsModel.Update(nil)
		case "o":
			// go to overview table page
			selectYearModel := NewSelectYearModel(m)
			return selectYearModel.Update(nil)
		}

	case userSavedMsg:
		// reset message flags
		m.StatusMessageFlags = StatusMessageFlags{}
		m.StatusMessageFlags.newEntry = msg.text
		m = m.updateHabitsList()
		return m, nil

	case restoredHabitMsg:
		m.StatusMessageFlags = StatusMessageFlags{}
		m.StatusMessageFlags.restoredHabit = msg.choice
		m = m.updateHabitsList()
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m ListModel) View() string {
	var s string
	if m.StatusMessageFlags.newRecord {
		s = notificationTextStyle.Render(fmt.Sprintf("Recorded %s. Current streak: %d", m.choice, m.streak))
	}

	if m.StatusMessageFlags.newEntry != "" {
		s = notificationTextStyle.Render(fmt.Sprintf("Added %s as a new goal to track.", m.StatusMessageFlags.newEntry))
	}

	if m.StatusMessageFlags.alreadyRecorded {
		s = notificationTextStyle.Render(fmt.Sprintf("Completion for goal '%s' already recorded for today.", m.choice))
	}

	if m.StatusMessageFlags.archived {
		s = notificationTextStyle.Render(
			fmt.Sprintf(
				"Archived %s",
				m.choice))
	}

	if m.StatusMessageFlags.restoredHabit != "" {
		s = notificationTextStyle.Render(fmt.Sprintf(
			"Restored %s", m.StatusMessageFlags.restoredHabit,
		))
	}

	if m.StatusMessageFlags.quitting {
		s = notificationTextStyle.Render(fmt.Sprintf("You recorded %d completed goals this session. Goodbye", m.numRecorded))
		return s
	}

	return "\n" + s + "\n\n" + m.list.View() + m.helpView()

}

func (m ListModel) helpView() string {
	return helpStyle.Render("\n ↑/k: up • ↓/j: down • ctrl+c: quit • a: archive • n: create entry • o: overview \n")
}

func itemsToList(habits []data.Habit) []list.Item {
	items := make([]list.Item, len(habits))

	for i, h := range habits {
		items[i] = list.Item(item(h.Name))
	}
	return items
}

func NewList(db *gorm.DB) ListModel {
	hdb := data.Database{DB: db}

	habits := hdb.GetActiveHabits()

	items := itemsToList(habits)

	const defaultWidth = 200

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "What goal did you complete today?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(false)

	m := ListModel{list: l, db: hdb}
	return m
}
