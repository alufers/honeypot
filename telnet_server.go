package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const initialTimeout = time.Second * 15
const afterFirstLineTimeout = time.Minute
const maxContents = 1024 * 128

func RunTelnetServer() {
	// create a tcp server
	log.Printf("Starting telnet server on %v", fmt.Sprintf(":%s", getEnv("TELNET_PORT", "23")))
	server, err := net.Listen("tcp", fmt.Sprintf(":%s", getEnv("TELNET_PORT", "23")))
	if err != nil {
		panic(err)
	}
	defer server.Close()
	listeningProtocolsMutex.Lock()
	listeningProtocols = append(listeningProtocols, "telnet")
	listeningProtocolsMutex.Unlock()
	for {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}
		go handleTelnetConnection(conn)
	}
}

func handleTelnetConnection(conn net.Conn) {
	defer conn.Close()
	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in handleTelnetConnection", r)
		}
	}()

	timeoutTimer := time.NewTimer(initialTimeout)

	var attack *Attack
	defer func() {
		if attack != nil {
			AttackFinished(attack)
		}
	}()
	// read the first line
	reader := bufio.NewReader(conn)

	attack = &Attack{
		Protocol: "telnet",
		SourceIP: strings.Split(conn.RemoteAddr().String(), ":")[0],
		Contents: "",
	}

	go func() {
		<-timeoutTimer.C
		conn.Close()
	}()

	write := func(message string) {
		conn.Write([]byte(message))
		attack.Contents += message
		if err := AttackUpdated(attack); err != nil {
			panic(err)
		}
	}
	log.Printf("cuum")
	if err := AttackStarted(attack); err != nil {
		panic(err)
	}
	write("Papaj2137-XG Broadband Router\r\nVosLogin:")
	username, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	attack.Contents += username
	if err := AttackUpdated(attack); err != nil {
		panic(err)
	}
	write("Password:")
	password, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	attack.Contents += password
	if err := AttackUpdated(attack); err != nil {
		panic(err)
	}
	for {
		write("# ")
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		timeoutTimer.Reset(afterFirstLineTimeout)

		attack.Contents += line
		if len(line) > maxContents {
			attack.Contents = attack.Contents[:maxContents]
			break
		}
		if err := AttackUpdated(attack); err != nil {
			panic(err)
		}

		write("OK\n")
	}

}
