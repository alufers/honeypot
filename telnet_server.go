package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
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
	for {
		write("# ")
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		timeoutTimer.Reset(afterFirstLineTimeout)

		attack.Contents += line
		attack.Classification = "command_entered"
		if len(line) > maxContents {
			attack.Contents = attack.Contents[:maxContents]
			break
		}
		if err := AttackUpdated(attack); err != nil {
			panic(err)
		}

		r := regexp.MustCompile("cat /proc/mounts; /bin/busybox ([A-Za-z0-9]+)")
		submatches := r.FindStringSubmatch(line)
		if len(submatches) > 0 {
			attack.Classification = "command_executed"
			write(fmt.Sprintf(`/dev/root /rom squashfs ro,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,noatime 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,noatime 0 0
cgroup2 /sys/fs/cgroup cgroup2 rw,nosuid,nodev,noexec,relatime,nsdelegate 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev,noatime 0 0
/dev/mtdblock6 /overlay jffs2 rw,noatime 0 0
overlayfs:/overlay / overlay rw,noatime,lowerdir=/,upperdir=/overlay/upper,workdir=/overlay/work 0 0
tmpfs /dev tmpfs rw,nosuid,relatime,size=512k,mode=755 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,mode=600,ptmxmode=000 0 0
debugfs /sys/kernel/debug debugfs rw,noatime 0 0
none /sys/fs/bpf bpf rw,nosuid,nodev,noexec,noatime,mode=700 0 0
%v: applet not found
`, submatches[1]))
			continue
		} else {
			write("OK\r\n")
		}

	}

}
