package main

import (
	"fmt"
	"log"
	"os"

	"github.com/neboman11/timelapse-manager/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func open_database() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}

	db.AutoMigrate(&models.InProgress{}, &models.Videos{})
}
