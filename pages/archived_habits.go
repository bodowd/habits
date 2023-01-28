package pages

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ArchivedHabitsModel struct {
	list      list.Model
	listModel ListModel
	choice    string
}

func NewArchivedHabitsModel(listModel ListModel) ArchivedHabitsModel {
	archivedHabits := listModel.db.GetInactiveHabits()

	items := itemsToList(archivedHabits)

	const defaultWidth = 200

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "What habit do you want to restore and track again?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(false)

	m := ArchivedHabitsModel{
		list:      l,
		listModel: listModel,
	}

	return m
}

func (m ArchivedHabitsModel) Init() tea.Cmd {
	return nil
}

func (m ArchivedHabitsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}

			err := m.listModel.db.RestoreHabit(m.choice)
			if err != nil {
				fmt.Println(err)
			}
			restored := restoredHabitMsg{
				choice: m.choice,
			}
			return m.listModel.Update(restored)
		case tea.KeyCtrlO:
			return m.listModel.Update(nil)

		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ArchivedHabitsModel) View() string {
	return m.list.View() + m.helpView()
}

func (m ArchivedHabitsModel) helpView() string {
	return helpStyle.Render("\n ↑/k: up • ↓/j: down • ctrl+c: quit • ctrl+o: back • enter: restore\n")
}

type restoredHabitMsg struct {
	choice string
}
