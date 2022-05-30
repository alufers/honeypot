package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

type FakeBinary func(ctx context.Context, args []string, conn io.ReadWriter) error
type FakeDir int

func makeFS() map[string]interface{} {
	var fs map[string]interface{}
	accessFile := func(path string) string {
		f := fs[path]
		if f == nil {
			return ""
		}
		if _, ok := f.(FakeBinary); ok {
			return fmt.Sprintf("i like cocks %v", path)
		}
		if f, ok := f.(string); ok {
			return f
		}
		return ""
	}

	resolvePath := func(ctx context.Context, path string) string {
		c := interp.HandlerCtx(ctx)
		if strings.HasPrefix(path, "/") {
			return path
		}
		abs, _ := filepath.Abs(filepath.Join(c.Dir, path))
		return abs
	}

	fs = map[string]interface{}{
		"/root": FakeDir(0),
		"/tmp":  FakeDir(1),

		"/bin/echo": "jestem prawdziwą binarką echo",
		"/usr/bin/cat": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) (err error) {
			hadError := false
			for _, arg := range args[1:] {
				resolved := resolvePath(ctx, arg)
				if _, ok := fs[resolved]; !ok {
					hadError = true
					fmt.Fprintf(conn, "cat: %v: No such file or directory\n", resolved)
					continue
				}
				if _, ok := fs[resolved].(FakeDir); ok {
					hadError = true
					fmt.Fprintf(conn, "cat: %v: Is a directory\n ", resolved)
					continue
				}

				fmt.Fprintf(conn, "%v\n", accessFile(resolved))
			}
			if hadError {

				return interp.NewExitStatus(5)
			}
			return nil
		}),
		"/usr/bin/true": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			return nil
		}),
		"/usr/bin/false": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			return interp.NewExitStatus(1)
		}),
		"/usr/bin/cp": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			if len(args) < 3 {
				fmt.Fprintf(conn, "cp: missing file operand\n")
				return interp.NewExitStatus(53)
			}
			src := resolvePath(ctx, args[1])
			dst := resolvePath(ctx, args[2])
			fs[dst] = fs[src]
			return nil
		}),
		"/usr/bin/ls": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) (e error) {
			if len(args) == 1 {
				args = append(args, ".")
			}
			for _, arg := range args[1:] {
				fullPath := resolvePath(ctx, arg)
				foundFiles := make(map[string]bool)
				parentDirExists := false
				for p := range fs {
					log.Printf("Checking %v against %v", p, fullPath)
					if strings.HasPrefix(p, fullPath) {
						stripped := strings.TrimPrefix(p, fullPath)
						stripped = strings.TrimPrefix(stripped, "/")
						stripped = strings.Split(stripped, "/")[0]
						parentDirExists = true
						if strings.Contains(stripped, "/") || stripped == "" {
							continue
						}
						foundFiles[stripped] = true
					}
				}
				if !parentDirExists {
					fmt.Fprintf(conn, "ls: %v: No such file or directory\n", arg)
					e = interp.NewExitStatus(10)
					continue
				}
				toPrint := ""
				for f := range foundFiles {
					toPrint += f + "\n"
				}
				fmt.Fprintf(conn, "%v", toPrint)
			}

			return
		}),
		"/bin/busybox": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			applet := "busybox"
			if len(args) > 1 {
				applet = args[1]
			}
			fmt.Fprintf(conn, "%v: applet not found\n", applet)
			return interp.NewExitStatus(5)
		}),
		"/bin/env": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {

			c := interp.HandlerCtx(ctx)
			if len(args) == 0 {
				return fmt.Errorf("args to short for env")
			}
			if len(args) == 1 {
				c.Env.Each(func(k string, v expand.Variable) bool {
					fmt.Fprintf(conn, "%s=%s\n", k, v)
					return true
				})
				return nil
			}
			binaryToFind := args[1]
			var foundFile interface{}
			if strings.HasPrefix(binaryToFind, "/") {
				foundFile = fs[binaryToFind]
			} else if strings.HasPrefix(binaryToFind, "./") {
				abs, _ := filepath.Abs(filepath.Join(c.Dir, binaryToFind[2:]))
				foundFile = fs[abs]
			} else {
				path := strings.Split(c.Env.Get("PATH").String(), ":")
				for _, p := range path {
					abs, _ := filepath.Abs(filepath.Join(p, binaryToFind))
					if f, ok := fs[abs]; ok {
						foundFile = f
						break
					}
				}
			}
			if foundFile == nil {
				if args[0] == "__CMD" {
					fmt.Fprintf(conn, "-ash: %s: not found\n", binaryToFind)
					return interp.NewExitStatus(5)
				} else {
					fmt.Fprintf(conn, "env: can't execute '%s': No such file or directory\n", binaryToFind)
					return interp.NewExitStatus(5)
				}
			}
			if f, ok := foundFile.(FakeBinary); ok {

				return f(ctx, args[1:], conn)
			}

			fmt.Fprintf(conn, "-ash:  %s: Permission denied\n", binaryToFind)
			return interp.NewExitStatus(5)
		}),
		"/proc/mounts": `/dev/root /rom squashfs ro,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,noatime 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,noatime 0 0
cgroup2 /sys/fs/cgroup cgroup2 rw,nosuid,nodev,noexec,relatime,nsdelegate 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev,noatime 0 0
/dev/mtdblock6 /overlay jffs2 rw,noatime 0 0
overlayfs:/overlay / overlay rw,noatime,lowerdir=/,upperdir=/overlay/upper,workdir=/overlay/work 0 0
tmpfs /dev tmpfs rw,nosuid,relatime,size=512k,mode=755 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,mode=600,ptmxmode=000 0 0
debugfs /sys/kernel/debug debugfs rw,noatime 0 0
none /sys/fs/bpf bpf rw,nosuid,nodev,noexec,noatime,mode=700 0 0`,
	}
	return fs
}
