package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"strings"
	"time"

	"github.com/alufers/honeypot/fakeshell"
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
			log.Println("Recovered in handleTelnetConnection", r)
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
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
		Protocol:       "telnet",
		SourceIP:       strings.Split(conn.RemoteAddr().String(), ":")[0],
		Contents:       "",
		Classification: "empty",
	}

	go func() {
		<-timeoutTimer.C
		conn.Close()
	}()

	write := func(message string) {
		conn.Write([]byte(message))
		attack.Contents += message
		timeoutTimer.Reset(afterFirstLineTimeout)
		if err := AttackUpdated(attack); err != nil {
			panic(err)
		}
	}

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
	attack.Classification = "username_entered"
	write("Password:")
	password, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	attack.Contents += password
	if err := AttackUpdated(attack); err != nil {
		panic(err)
	}
	attack.Classification = "authenticated"
	credUsage := &CredentialUsage{
		Protocol: "telnet",
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}
	if err := db.Create(credUsage).Error; err != nil {
		panic(err)
	}
	timeoutTimer.Reset(afterFirstLineTimeout)
	// attackBuf := bufio.NewWriter(attack)
	// stdout := io.MultiWriter(conn, attackBuf)
	// stdin := io.TeeReader(conn, attackBuf)
	// runner := createRunner(stdin, stdout)
	if err := fakeshell.ServiceFakeshell(WrapConnReaderWriter(attack, conn, conn, func(line string) {
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}
		timeoutTimer.Reset(afterFirstLineTimeout)
		attack.Classification = "command_entered"
	})); err != nil {
		panic(err)
	}

}
