package main

import (
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

var listeningProtocolsMutex = &sync.Mutex{}
var listeningProtocols = []string{}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open(getEnv("HONEYPOT_DB", "honeypot.db")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&Attack{})
	if err != nil {
		panic("failed to migrate database")
	}
	db.Model(&Attack{}).Where("1=1").Update("in_progress", false)
	go RunTelnetServer()
	RunAdminServer()
}
