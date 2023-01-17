package data

import (
	"errors"
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

func (d *Database) RecordCompletion(habitId uint) (Completion, error) {
	// Check if the last record for this habitId was the day before
	var lastCompletion Completion
	var streak int
	err := d.DB.Where("habit_id = ? AND recorded_at = ?",
		habitId, yesterdaysDate()).First(&lastCompletion).Error

	// if a recorded completion from yesterday is not found, streak starts over
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			streak = 1
		}
	} else {
		streak = lastCompletion.Streak + 1
	}

	completion := Completion{RecordedAt: currentDate(), HabitID: habitId, Streak: streak}
	err = d.DB.Create(&completion).Error
	return completion, err
}
