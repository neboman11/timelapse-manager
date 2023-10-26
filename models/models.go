package models

import "time"

type Videos struct {
	Date     time.Time `json:"date"`
	Location string    `json:"location"`
}

type InProgress struct {
	Id     uint64    `json:"id"`
	Date   time.Time `json:"date"`
	Folder string    `json:"folder"`
}
