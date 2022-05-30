package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"

	"github.com/gofiber/websocket/v2"
)

var shellProviders = []*ShellProvider{}
var shellProvidersMutex = &sync.RWMutex{}

// gets a random shell provider and removes it from the list
func GetNextShellProvider() *ShellProvider {
	if len(shellProviders) == 0 {
		return nil
	}
	idx := rand.Intn(len(shellProviders))
	sp := shellProviders[idx]
	shellProviders = append(shellProviders[:idx], shellProviders[idx+1:]...)
	return sp

}

type ShellProvider struct {
	Hostname     string
	Kind         string
	Conn         websocket.Conn
	notifyClosed chan bool
}

func ServiceConn(attack *Attack, lineReader bufio.Reader, write func(message string)) error {
	var sp *ShellProvider
	write("# ")

	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			return err
		}
		attack.Contents += line

		if err := AttackUpdated(attack); err != nil {
			panic(err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			write("# ")
			continue
		}
		didSendMessage := false
		if sp == nil {
			for {
				sp = GetNextShellProvider()
				if sp == nil {
					return fmt.Errorf("no shell providers available")
				}
				go func() {
					for {
						_, data, err := sp.Conn.ReadMessage()
						if err != nil {
							log.Printf("error reading message from shell provider: %v", err)
							return
						}
						write(string(data))
					}
				}()
				err := sp.Conn.WriteMessage(websocket.TextMessage, []byte(line))
				if err == nil {
					didSendMessage = true
					break
				}
				log.Printf("failed to write to shell provider at %v@%v: %v", sp.Kind, sp.Hostname, err)
			}
			defer func() {
				sp.Conn.Close()
				select {
				case sp.notifyClosed <- true:
				default:
				}
			}()
		}
		if !didSendMessage {
			err := sp.Conn.WriteMessage(websocket.TextMessage, []byte(line))
			if err != nil {
				log.Printf("failed to write to old shell provider at %v@%v: %v", sp.Kind, sp.Hostname, err)
				sp = nil
				break
			}
		}

	}

	return nil
}
