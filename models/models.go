package models

import "time"

type Timelapse struct {
	Id         uint64    `json:"id" gorm:"primaryKey"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	Folder     string    `json:"folder"`
	VideoFile  string    `json:"videoFile"`
	Status     string    `json:"status" gorm:"default:InProgress"`
	ImageCount uint32    `json:"imageCount" gorm:"default:0"`
}
