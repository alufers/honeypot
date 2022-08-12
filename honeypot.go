package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

var listeningProtocolsMutex = &sync.Mutex{}
var listeningProtocols = []string{}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

func migrateActions() {

	log.Printf("Migrating actions!\n")
	var totalUnmigrated int64
	if err := db.Model(&Attack{}).Where("action IS NULL OR action = ''").Count(&totalUnmigrated).Error; err != nil {
		log.Fatalf("failed to count unmigrated attacks: %v", err)
	}
	var unmigratedLeft int64 = totalUnmigrated
	const batchSize = 3000
	const workerCount = 10

	// workerJobs := make(chan *Attack, 4)
	// workerCount := make(chan *Attack, 4)

	for {

		timeStart := time.Now()
		var attacks []Attack
		db.Select([]string{"contents", "id", "classification"}).Where("action IS NULL OR action = ''").Limit(batchSize).Find(&attacks)
		if len(attacks) == 0 {
			break
		}
		updates := map[string][]uint{}
		for _, attack := range attacks {

			attack.Action = classifyAction(&attack)

			// if err := db.
			// 	Model(&attack).
			// 	Where("id = ?", attack.ID).
			// 	Update("action", attack.Action).Error; err != nil {
			// 	log.Fatalf("failed to update attack action: %v", err)
			// }
			// if err := db.Save(&attack).Error; err != nil {
			// 	log.Fatalf("failed to save attack: %v", err)
			// }
			if updates[attack.Action] == nil {
				updates[attack.Action] = []uint{}
			}
			updates[attack.Action] = append(updates[attack.Action], attack.ID)
		}
		for action, ids := range updates {
			if err := db.Model(&Attack{}).Where("id IN (?)", ids).Update("action", action).Error; err != nil {
				log.Fatalf("failed to update attack action batch: %v", err)
			}
		}

		timeEnd := time.Now()
		duration := timeEnd.Sub(timeStart)
		rate := float64(batchSize) / float64(duration.Seconds())
		eta := time.Duration(float64(unmigratedLeft) / rate * float64(time.Second))
		unmigratedLeft -= int64(len(attacks))
		log.Printf(
			"migrated %d/%d percent: %.2f%%    Rate: %.2f ETA: %v",
			totalUnmigrated-unmigratedLeft,
			totalUnmigrated,
			float64(totalUnmigrated-unmigratedLeft)/float64(totalUnmigrated)*100,
			rate,
			eta,
		)
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
		log.Fatalf("failed to migrate database: %v", err)
	}
	migrateActions()
	db.Model(&Attack{}).Where("in_progress = ?", true).Update("in_progress", false)
	go RunTelnetServer()
	go RunSSHServer()
	RunAdminServer()
}
