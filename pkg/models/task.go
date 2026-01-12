package models

import (
	"time"

	"github.com/Prysya/go-final-project/pkg/constants"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func (t *Task) ToTime() (time.Time, error) {
	return time.Parse(constants.DateFormat, t.Date)
}

func IsValidDate(date string) bool {
	_, err := time.Parse(constants.DateFormat, date)
	return err == nil
}
