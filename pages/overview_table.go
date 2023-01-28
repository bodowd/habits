package pages

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bodowd/habits/data"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

const defaultWidth = 200

type SelectYearModel struct {
	yearList     list.Model
	listModel    ListModel
	selectedYear string
}

type SelectMonthModel struct {
	selectedYearModel SelectYearModel
	selectedMonth     string
	monthList         list.Model
}

func NewSelectYearModel(listModel ListModel) SelectYearModel {

	years := listModel.db.GetAvailableYears()
	yearItems := make([]list.Item, len(years))
	for count, y := range years {
		yearItems[count] = list.Item(item(y))
	}

	yl := list.New(yearItems, itemDelegate{}, defaultWidth, listHeight)
	yl.Title = "What year are you interested in?"
	yl.SetShowStatusBar(false)
	yl.SetFilteringEnabled(false)
	yl.Styles.Title = titleStyle
	yl.Styles.PaginationStyle = paginationStyle
	yl.Styles.HelpStyle = helpStyle
	yl.SetShowHelp(false)

	return SelectYearModel{yearList: yl, listModel: listModel}
}

func NewSelectMonthModel(sym SelectYearModel) SelectMonthModel {
	months := []list.Item{
		list.Item(item("Jan")),
		list.Item(item("Feb")),
		list.Item(item("Mar")),
		list.Item(item("Apr")),
		list.Item(item("May")),
		list.Item(item("Jun")),
		list.Item(item("Jul")),
		list.Item(item("Aug")),
		list.Item(item("Sep")),
		list.Item(item("Oct")),
		list.Item(item("Nov")),
		list.Item(item("Dec")),
	}

	ml := list.New(months, itemDelegate{}, defaultWidth, 20)
	ml.Title = "What month are you interested in?"
	ml.SetShowStatusBar(false)
	ml.SetFilteringEnabled(false)
	ml.Styles.Title = titleStyle
	ml.Styles.PaginationStyle = paginationStyle
	ml.Styles.HelpStyle = helpStyle
	ml.SetShowHelp(false)

	return SelectMonthModel{monthList: ml, selectedYearModel: sym}
}

func (m SelectYearModel) Init() tea.Cmd {
	return nil
}

func (m SelectYearModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.yearList.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.selectedYear == "" {
				i, ok := m.yearList.SelectedItem().(item)
				if ok {
					m.selectedYear = string(i)
				}
			}
			// show the next list, list of months
			selectMonthModel := NewSelectMonthModel(m)
			return selectMonthModel.Update(nil)
		case tea.KeyCtrlO:
			return m.listModel.Update(nil)
		}
	}

	m.yearList, cmd = m.yearList.Update(msg)
	return m, cmd
}

func (m SelectYearModel) View() string {
	return m.yearList.View() + m.helpView()
}

func (m SelectMonthModel) Init() tea.Cmd {
	return nil
}

func (m SelectMonthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.monthList.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.selectedMonth == "" {
				i, ok := m.monthList.SelectedItem().(item)
				if ok {
					m.selectedMonth = string(i)
				}
			}
			// go to table view
			ntm := NewTableModel(m)
			return ntm.Update(nil)
		case tea.KeyCtrlO:
			return m.selectedYearModel.listModel.Update(nil)
		}

	}

	m.monthList, cmd = m.monthList.Update(msg)
	return m, cmd
}

func (m SelectMonthModel) View() string {
	return m.monthList.View() + m.helpView()
}

func (m SelectYearModel) helpView() string {
	return helpView()
}

func (m SelectMonthModel) helpView() string {
	return helpView()
}

func helpView() string {
	return helpStyle.Render("\n ↑/k: up • ↓/j: down • ctrl+c: quit • ctrl+o: back \n")
}

type TableModel struct {
	table table.Model
	smm   SelectMonthModel
}

func NewTableModel(m SelectMonthModel) TableModel {
	db := m.selectedYearModel.listModel.db
	month := m.selectedMonth
	intMonth := data.MonthToIntMap[month]
	year := m.selectedYearModel.selectedYear
	intYear, err := strconv.Atoi(year)
	if err != nil {
		fmt.Println(err)
	}

	// columns = append(columns, table.Column{Title: h.Habit.Name, Width: 20})
	columns := []table.Column{
		{Title: "Date", Width: 10},
	}

	var rows []table.Row
	// each element in this result has a date and a habit name
	habitsAndCompletions := db.GetActiveHabitsAndCompletions(intMonth, intYear)
	habitsSeen := map[string]int{}
	// use a slice with the habit name at the index to make the columns
	var habitsIndex []string
	habitsIndex = append(habitsIndex, "DUMMY")

	if len(habitsAndCompletions) != 0 {
		count := 1
		for _, h := range habitsAndCompletions {
			// keep track of which index of columns a habit is listed in
			// this index will track where we mark an "x" for a completed habit
			// in the row
			_, ok := habitsSeen[h.Habit.Name]
			if !ok {
				habitsSeen[h.Habit.Name] = count
				habitsIndex = append(habitsIndex, h.Habit.Name)
				count++
			}

		}

		res := make([][]string, 31)

		for i := 0; i < len(res); i++ {
			res[i] = make([]string, len(habitsSeen)+1)
			// first column is the date
			res[i][0] = strconv.Itoa(data.MonthToIntMap[m.selectedMonth]) + "-" + strconv.Itoa(i+1) + "-" + m.selectedYearModel.selectedYear
		}

		for _, h := range habitsAndCompletions {
			date, err := time.Parse("2006-01-02", h.Completion.RecordedAt)
			if err != nil {
				fmt.Println(err)
			}

			// then mark the completion in the corresponding habit column
			res[date.Day()-1][habitsSeen[h.Habit.Name]] = "x"
		}

		for i := 1; i < len(habitsIndex); i++ {
			// double check that the habitsIndex slice are in the correct order
			if habitsSeen[habitsIndex[i]] != i {
				log.Fatal("Column index for habits are wrong")
			}
			columns = append(columns, table.Column{Title: habitsIndex[i], Width: 20})
		}

		for _, r := range res {
			var row table.Row
			for _, i := range r {
				row = append(row, i)
			}
			rows = append(rows, row)
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(35))

	s := table.DefaultStyles()
	t.SetStyles(s)
	tm := TableModel{table: t, smm: m}
	return tm
}

func (m TableModel) Init() tea.Cmd {
	return nil
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlO:
			return m.smm.selectedYearModel.listModel.Update(nil)
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TableModel) View() string {
	return m.table.View() + m.helpView()
}

func (m TableModel) helpView() string {
	return helpView()
}
