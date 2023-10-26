package models

import "time"

type Videos struct {
	Date     time.Time `json:"date"`
	Location string    `json:"location"`
}

type InProgress struct {
	Date   time.Time `json:"date"`
	Folder string    `json:"folder"`
}
