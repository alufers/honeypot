package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fasthttp/websocket"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var key string
var numInstances int = 1
var wsUrl string

type combinedReadWriter struct {
	io.Reader
	io.Writer
}

func runInstance() error {
	log.Printf("Dialing %v", wsUrl)
	c, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return fmt.Errorf("error dialing websocket: %v", err)
	}
	log.Printf("Connected to %v", wsUrl)
	defer c.Close()

	fs := makeFS()
	stdoutReader, stdoutWriter := io.Pipe()
	stdinReader, _ := io.Pipe()
	comb := combinedReadWriter{
		Writer: stdoutWriter,
		Reader: stdinReader,
	}
	runner, err := interp.New(
		interp.ExecHandler(func(ctx context.Context, args []string) error {
			log.Printf("CHUJ! %#v", args)
			return fs["/bin/env"].(FakeBinary)(
				ctx,
				append([]string{"__CMD"}, args...),
				comb,
			)
		}),
		interp.StdIO(stdinReader, stdoutWriter, stdoutWriter),
		interp.Dir("/root"),
		interp.Env(expand.Environ(expand.ListEnviron(
			"USER=root",
			"SHLVL=1",
			"HOME=/root",
			"SSH_TTY=/dev/pts/1",
			`PS1=\[\e]0;\u@\h: \w\a\]\u@\h:\w\$ `,
			"ENV=/etc/shinit",
			"LOGNAME=root",
			"TERM=xterm-256color",
			"PATH=/usr/sbin:/usr/bin:/sbin:/bin",
			"SHELL=/bin/ash",
			"PWD=/root",
		))),
	)
	if err != nil {
		return fmt.Errorf("error creating interpreter: %v", err)
	}
	runner = runner
	go func() {
		for {
			var data []byte = make([]byte, 1024)
			_, err := stdoutReader.Read(data)
			if err != nil {
				log.Printf("error reading from shell: %v", err)
				return
			}
			err = c.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("error writing to websocket: %v", err)
				return
			}
		}
	}()
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("error reading message from websocket: %v", err)
		}
		log.Printf("Got message: %v", string(data))
		file, _ := syntax.NewParser().Parse(strings.NewReader(string(data)), "")
		runner.Run(context.Background(), file)
		stdoutWriter.Write([]byte("# "))

	}
}

func runInstanceLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v", r)
		}
	}()
	for {

		err := runInstance()
		log.Printf("error running instance: %v", err)
		time.Sleep(time.Second)

	}
}

func main() {
	flag.StringVar(&key, "key", "devkey", "The shell provider key set by SHELL_PROVIDER_KEY on the server")
	flag.IntVar(&numInstances, "instances", 1, "The number of instances to spin up")
	flag.Parse()
	rawUrl := flag.Arg(0)
	parsed, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatalf("error parsing url: %v", err)
	}
	if parsed.Scheme == "http" {
		parsed.Scheme = "ws"
	} else if parsed.Scheme == "https" {
		parsed.Scheme = "wss"
	}
	hn, _ := os.Hostname()
	pq := parsed.Query()
	pq.Set("key", key)
	pq.Set("hostname", hn)
	pq.Set("kind", "fakeshell")
	parsed.RawQuery = pq.Encode()
	if parsed.Host == "" {
		log.Fatalf("error parsing url: missing host")
	}
	wsUrl = parsed.String()
	log.Printf("Running %v instances connected to %v", numInstances, wsUrl)
	for i := 0; i < numInstances; i++ {
		go runInstanceLoop()
	}
	c := make(chan int)
	<-c // block forever
}
