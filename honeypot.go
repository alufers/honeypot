package main

import (
	"log"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

var listeningProtocolsMutex = &sync.Mutex{}
var listeningProtocols = []string{}

func migrateClassifications() {
	rows, err := db.Model(&Attack{}).Where("classification = \"\" OR classification IS NULL").Rows()
	if err != nil {
		panic(err)
	}

	attacksToSave := []*Attack{}
	func() {
		defer rows.Close()
		for rows.Next() {
			var attack Attack
			if err := db.ScanRows(rows, &attack); err != nil {
				panic(err)
			}

			emptyValues := []string{
				"Papaj2137-XG Broadband Router\r\nVosLogin:",
				"",
			}
			var isEmptyValue bool
			for _, emptyValue := range emptyValues {
				if attack.Contents == emptyValue {
					isEmptyValue = true
					break
				}
			}
			if isEmptyValue {
				attack.Classification = "empty"
			} else {
				attack.Classification = "authenticated"
			}
			attacksToSave = append(attacksToSave, &attack)
		}

	}()
	log.Printf("Migrating %d attacks", len(attacksToSave))
	for _, attack := range attacksToSave {
		if err := db.Save(attack).Error; err != nil {
			panic(err)
		}
	}

}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open(getEnv("HONEYPOT_DB", "honeypot.db")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&Attack{}, &CredentialUsage{})
	if err != nil {
		panic("failed to migrate database")
	}
	migrateClassifications()
	db.Model(&Attack{}).Where("1=1").Update("in_progress", false)
	go RunTelnetServer()
	go RunSSHServer()
	RunAdminServer()
}
