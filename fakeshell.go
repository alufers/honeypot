package main

import (
	"context"
	"io"
	"os"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

type combinedReadWriter struct {
	io.Reader
	io.Writer
}

func createRunner(stdin io.Reader, stdout io.Writer) *interp.Runner {

	fs := makeFS()
	comb := combinedReadWriter{
		Writer: stdout,
		Reader: stdin,
	}
	runner, err := interp.New(
		interp.ExecHandler(func(ctx context.Context, args []string) error {
			return fs["/bin/env"].(FakeBinary)(
				ctx,
				append([]string{"__CMD"}, args...),
				comb,
			)
		}),
		interp.OpenHandler(func(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			return nil, nil
		}),
		interp.StdIO(stdin, stdout, stdout),
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
		panic(err)
	}
	return runner
}
