package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"runtime/debug"
	"strings"
	"time"

	"github.com/alufers/honeypot/fakeshell"
	"github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

type combinedReadWriter struct {
	io.Reader
	io.Writer
}

func RunSSHServer() {
	sshPort := fmt.Sprintf(":%s", getEnv("HONEYPOT_SSH_PORT", "2222"))
	log.Printf("Starting ssh server on %v", sshPort)

	listeningProtocolsMutex.Lock()
	listeningProtocols = append(listeningProtocols, "ssh")
	listeningProtocolsMutex.Unlock()

	ssh.Handle(func(s ssh.Session) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in ssh.Handle", r)
				log.Println("stacktrace from panic: \n" + string(debug.Stack()))
			}
		}()

		timeoutTimer := time.NewTimer(initialTimeout)

		go func() {
			<-timeoutTimer.C
			s.Exit(1)
			s.Close()
		}()

		_, _, isPty := s.Pty()
		if isPty {
			attack := &Attack{
				Protocol:       "ssh",
				SourceIP:       strings.Split(s.RemoteAddr().String(), ":")[0],
				Contents:       "",
				Classification: "authenticated",
			}
			if err := AttackStarted(attack); err != nil {
				panic(err)
			}
			defer AttackFinished(attack)
			term := term.NewTerminal(s, "# ")

			if err := fakeshell.ServiceFakeShellOnTerminal(term, s, attack, func(line string) {
				line = strings.TrimSpace(line)
				if line == "" {
					return
				}
				timeoutTimer.Reset(afterFirstLineTimeout)
				attack.Classification = "command_entered"
			}); err != nil {
				panic(err)
			}
		} else {
			attack := &Attack{
				Protocol:       "ssh",
				SourceIP:       strings.Split(s.RemoteAddr().String(), ":")[0],
				Contents:       "",
				Classification: "ssh_command",
			}
			if err := AttackStarted(attack); err != nil {
				panic(err)
			}
			defer AttackFinished(attack)

			if err, exitCode := fakeshell.ServeSingleCommand(s.RawCommand(), s, s, attack); err != nil {
				panic(err)
			} else {
				s.Exit(int(exitCode))
			}
		}
	})

	log.Fatal(ssh.ListenAndServe(sshPort, nil, ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
		credUsage := &CredentialUsage{
			Protocol: "ssh",
			Username: strings.TrimSpace(ctx.User()),
			Password: strings.TrimSpace(pass),
		}
		if err := db.Create(credUsage).Error; err != nil {
			panic(err)
		}
		rand := rand.Intn(100)
		return rand >= 20
	})))
}
