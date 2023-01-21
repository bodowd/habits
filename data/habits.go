package data

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Habit struct {
	gorm.Model
	Name      string `gorm:"unique;not null"`
	CreatedAt string
	Active    bool
	Records   []Completion
}

type Completion struct {
	gorm.Model
	RecordedAt string
	Streak     int
	HabitID    uint
}

type Database struct {
	DB *gorm.DB
}

func currentDate() string {
	return time.Now().Format("2006-01-02")
}

func yesterdaysDate() string {
	return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
}

func NewHabit(name string) *Habit {
	return &Habit{
		Name:   name,
		Active: true,
	}
}

func (d *Database) CreateHabit(name string) (Habit, error) {
	hab := Habit{Name: name, CreatedAt: currentDate(), Active: true}
	if err := d.DB.Create(&hab).Error; err != nil {
		return hab, err
	}
	return hab, nil
}

type HabitAndCompletion struct {
	Habit
	Completion
}

func (d *Database) GetActiveHabitsAndCompletions(month, year int) []HabitAndCompletion {
	var habitsAndStreak []HabitAndCompletion

	firstDayOfMonth, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-01", year, month))
	if err != nil {
		fmt.Println(err)
	}

	lastDayOfMonth, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-31", year, month))
	if err != nil {
		fmt.Println(err)
	}

	d.DB.Table("habits").
		Select("habits.*, completions.*").
		Joins("INNER JOIN completions ON completions.habit_id=habits.id").
		Where("habits.active = ? AND completions.recorded_at BETWEEN ? AND ?",
			true, firstDayOfMonth, lastDayOfMonth).
		Find(&habitsAndStreak)

	return habitsAndStreak
}

func (d *Database) GetAvailableYears() []string {
	var years []string
	d.DB.Raw("SELECT DISTINCT STRFTIME('%Y', recorded_at) FROM completions").Scan(&years)
	fmt.Println(years)
	return years
}

func (d *Database) getHabits(activeFlag bool) []Habit {
	var habits []Habit
	d.DB.Where("active = ?", activeFlag).Find(&habits)

	return habits

}

func (d *Database) GetActiveHabits() []Habit {
	return d.getHabits(true)
}

func (d *Database) GetInactiveHabits() []Habit {
	return d.getHabits(false)
}

func (d *Database) GetAllHabits() []Habit {
	var habits []Habit
	d.DB.Find(&habits)
	return habits
}

func (d *Database) getHabitByName(habit string) (Habit, error) {
	var h Habit
	if err := d.DB.Where("name = ?", habit).First(&h).Error; err != nil {
		return h, err
	}
	return h, nil
}

type Result struct {
	Name   string
	ID     uint
	Streak int
}

func (d *Database) getCompletionAtTime(habit, day string) (Result, error) {
	// Check if the last record for this habitId was the day before
	var result Result
	err := d.DB.Table("habits").
		Select("habits.name, habits.id, completions.streak").
		Joins("inner join completions on completions.habit_id = habits.id").
		Where("habits.name = ? AND completions.recorded_at = ? AND habits.active = true",
			habit, day).First(&result).Error
	if err != nil {
		return result, err
	}
	return result, nil
}

type AlreadyRecordedTodayError struct{}

func (e *AlreadyRecordedTodayError) Error() string {
	return "Already recorded completion for today"
}

func (d *Database) RecordCompletion(habit string) (Completion, error) {
	// don't allow more completions if completion already recorded today
	_, err := d.getCompletionAtTime(habit, currentDate())
	if err == nil {
		return Completion{}, &AlreadyRecordedTodayError{}
	}

	var streak int

	// if a recorded completion from yesterday is not found, streak starts over
	result, err := d.getCompletionAtTime(habit, yesterdaysDate())
	if err != nil {
		// if the record is not found, streak starts over
		if strings.Contains("record not found", err.Error()) {
			streak = 1

			// find the habit id
			var h Habit
			err := d.DB.Table("habits").Where("name = ? AND active=true", habit, yesterdaysDate()).First(&h).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return Completion{}, err
				}
			}
			result.ID = h.ID
		}
	} else {
		streak = result.Streak + 1
	}

	completion := Completion{
		RecordedAt: currentDate(),
		HabitID:    result.ID,
		Streak:     streak,
	}
	err = d.DB.Create(&completion).Error
	return completion, err
}

func (d *Database) ArchiveHabit(habit string) error {
	err := d.DB.Model(&Habit{}).
		Where("name = ? AND active = true", habit).
		Update("active", false).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) RestoreHabit(habit string) error {
	err := d.DB.Model(&Habit{}).
		Where("name = ? AND active = false", habit).
		Update("active", true).Error
	if err != nil {
		return err
	}
	return nil

}
