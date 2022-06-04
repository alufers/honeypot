package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/alufers/honeypot/fakeshell"
	"github.com/gliderlabs/ssh"
)

func RunSSHServer() {
	sshPort := fmt.Sprintf(":%s", getEnv("HONEYPOT_SSH_PORT", "2222"))
	log.Printf("Starting ssh server on %v", sshPort)

	listeningProtocolsMutex.Lock()
	listeningProtocols = append(listeningProtocols, "ssh`")
	listeningProtocolsMutex.Unlock()

	ssh.Handle(func(s ssh.Session) {
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

			if err := fakeshell.ServiceFakeshell(WrapConnReaderWriter(attack, s, s, func(line string) {
				line = strings.TrimSpace(line)
				if line == "" {
					return
				}
				// timeoutTimer.Reset(afterFirstLineTimeout)
				attack.Classification = "command_entered"
			})); err != nil {
				panic(err)
			}
		} else {
			s.Write([]byte("AAAAAH"))
			s.Exit(1)
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
