package models

import "time"

type Videos struct {
	Id       uint64    `json:"id" gorm:"primaryKey"`
	Date     time.Time `json:"date"`
	Location string    `json:"location"`
}

type InProgress struct {
	Id     uint64    `json:"id" gorm:"primaryKey"`
	Date   time.Time `json:"date"`
	Folder string    `json:"folder"`
	Status string    `json:"status" gorm:"default:InProgress"`
}
