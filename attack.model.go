package main

import (
	"bufio"
	"io"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Attack struct {
	gorm.Model
	Protocol       string `json:"protocol"`
	InProgress     bool   `json:"inProgress"`
	SourceIP       string `json:"sourceIP"`
	Country        string `json:"country"`
	CountryCode    string `json:"countryCode"`
	Location       string `json:"location"`
	ISP            string `json:"isp"`
	Contents       string `json:"contents"`
	Duration       int    `json:"duration"` // milliseconds
	Classification string `json:"classification"`
	Action         string `json:"action"`
}

func (a *Attack) Write(data []byte) (int, error) {
	a.Contents += string(data)
	if err := AttackUpdated(a); err != nil {
		panic(err)
	}
	return len(data), nil
}

var currentAttacks = make(map[uint]*Attack)
var currentAttacksMutex = &sync.Mutex{}

var attacksEventBroadcaster = NewEventBroadcaster[*Attack]()

func AttackStarted(attack *Attack) error {
	log.Printf(" - attack started: %#v", attack)
	AddIpLocationDataToAttack(attack)
	attack.InProgress = true
	if err := db.Save(attack).Error; err != nil {
		log.Printf("failed to save attack: %v", err)
		return err
	}
	currentAttacksMutex.Lock()
	defer currentAttacksMutex.Unlock()

	currentAttacks[attack.ID] = attack

	attacksEventBroadcaster.Broadcast(attack)

	return nil
}

func AttackUpdated(attack *Attack) error {
	attacksEventBroadcaster.Broadcast(attack)
	// if attack updated less than 15 seconds ago, don't update the database
	if attack.UpdatedAt.Add(15 * time.Second).After(time.Now()) {
		return nil
	}
	if err := db.Save(attack).Error; err != nil {
		log.Printf("failed to save attack: %v", err)
		return err
	}
	currentAttacksMutex.Lock()
	defer currentAttacksMutex.Unlock()

	currentAttacks[attack.ID] = attack

	return nil
}

func AttackFinished(attack *Attack) error {
	attack.InProgress = false
	attack.Duration = int(time.Now().Sub(attack.CreatedAt).Milliseconds())
	if err := db.Save(attack).Error; err != nil {
		log.Printf("failed to save attack: %v", err)
		return err
	}
	attack.Action = classifyAction(attack)
	currentAttacksMutex.Lock()
	defer currentAttacksMutex.Unlock()

	delete(currentAttacks, attack.ID)
	attacksEventBroadcaster.Broadcast(attack)
	return nil
}

// WrapConnReaderWriter wraps a reader and a wirter, and extracts the data to be saved in the database
func WrapConnReaderWriter(a *Attack, reader io.Reader, writer io.Writer, lineCallback func(string)) (io.Reader, io.Writer) {
	cbR, cbW := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(cbR)
		for scanner.Scan() {
			line := scanner.Text()
			if lineCallback != nil {
				lineCallback(line)
			}
			a.Write([]byte(line + "\n"))
		}
		if err := scanner.Err(); err != nil {
			log.Printf("error reading from connection: %v", err)
		}
	}()
	outputR, outputW := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(outputR)
		for scanner.Scan() {
			line := scanner.Text()
			a.Write([]byte(line + "\n"))
		}
	}()
	return io.TeeReader(reader, cbW), io.MultiWriter(writer, outputW)
}
