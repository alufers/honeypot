package fakeshell

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/blang/vfs"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

type combinedReadWriter struct {
	io.Reader
	io.Writer
}

func createRunner(stdin io.Reader, stdout io.Writer, fs vfs.Filesystem) *interp.Runner {
	os.MkdirAll("/root", 0755)

	runner, err := interp.New(
		interp.ExecHandler(func(ctx context.Context, args []string) error {
			return fakeshellBinaries["/bin/env"](
				ctx,
				append([]string{"__CMD"}, args...),
			)
		}),
		interp.OpenHandler(func(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			log.Printf("open: %s %d %d", path, flag, perm)
			c := UnwrapCtx(ctx)

			return MapError(fs.(vfs.Filesystem).OpenFile(c.Abs(path), flag, perm))
		}),
		interp.StatHandler(func(ctx context.Context, name string, followSymlinks bool) (os.FileInfo, error) {
			info, err := MapError(fs.(vfs.Filesystem).Stat(name))
			if err != nil {
				return nil, err
			}
			return NewFileInfoWrapper(info), nil
		}),
		interp.ReadDirHandler(func(ctx context.Context, path string) ([]os.FileInfo, error) {
			log.Printf("Read dir ctx: %#v", ctx)
			info, err := MapError(fs.(vfs.Filesystem).ReadDir(path))
			if err != nil {
				return nil, err
			}
			for i := range info {
				info[i] = NewFileInfoWrapper(info[i])
			}
			return info, nil
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
