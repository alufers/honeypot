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
		"/usr/bin/cat": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			for _, arg := range args[1:] {
				fmt.Fprintf(conn, "%v\n", accessFile(resolvePath(ctx, arg)))
			}
			return nil
		}),
		"/usr/bin/ls": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			if len(args) == 1 {
				args = append(args, ".")
			}
			for _, arg := range args[1:] {
				fullPath := resolvePath(ctx, arg)
				for p := range fs {
					log.Printf("Checking %v against %v", p, fullPath)
					if strings.HasPrefix(p, fullPath) {
						stripped := strings.TrimPrefix(p, fullPath)
						stripped = strings.TrimPrefix(stripped, "/")
						stripped = strings.Split(stripped, "/")[0]
						if strings.Contains(stripped, "/") || stripped == "" {
							continue
						}
						fmt.Fprintf(conn, "%v\n", stripped)
					}
				}
			}

			return nil
		}),
		"/bin/busybox": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			applet := "busybox"
			if len(args) > 1 {
				applet = args[1]
			}
			fmt.Fprintf(conn, "%v: applet not found\n", applet)
			return nil
		}),
		"/bin/env": FakeBinary(func(ctx context.Context, args []string, conn io.ReadWriter) error {
			log.Printf("ev! %#v", args)
			c := interp.HandlerCtx(ctx)
			if len(args) == 0 {
				return fmt.Errorf("args to short")
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
					return fmt.Errorf("ass")
				} else {
					fmt.Fprintf(conn, "env: can't execute '%s': No such file or directory\n", binaryToFind)
					return fmt.Errorf("ass")
				}
			}
			if f, ok := foundFile.(FakeBinary); ok {
				return f(ctx, args[1:], conn)
			}

			fmt.Fprintf(conn, "-ash:  %s: Permission denied\n", binaryToFind)
			return fmt.Errorf("ass")
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
