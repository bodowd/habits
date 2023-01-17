package data

import (
	"errors"
	"fmt"
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

func (d *Database) GetActiveHabits() []Habit {
	var habits []Habit
	d.DB.Where("active = ?", true).Find(&habits)

	return habits
}

func (d *Database) GetAllHabits() []Habit {
	var habits []Habit
	d.DB.Find(&habits)
	return habits
}

func (d *Database) RecordCompletion(habit string) (Completion, error) {
	// Check if the last record for this habitId was the day before
	type Result struct {
		Name   string
		ID     uint
		Streak int
	}
	var result Result
	var streak int
	err := d.DB.Table("habits").
		Select("habits.name, habits.id, completions.streak").
		Joins("inner join completions on completions.habit_id = habits.id").
		Where("habits.name = ? AND completions.recorded_at = ? AND habits.active = true",
			habit, yesterdaysDate()).Find(&result).Error
	if err != nil {
		fmt.Println(err)
	}

	// if a recorded completion from yesterday is not found, streak starts over
	if result.Name == "" {
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
