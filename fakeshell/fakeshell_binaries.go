package fakeshell

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blang/vfs"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

type FakeBinary func(ctx context.Context, args []string) error

type BetterCtx struct {
	interp.HandlerContext
	FS vfs.Filesystem
}

func UnwrapCtx(ctx context.Context) *BetterCtx {
	return &BetterCtx{
		HandlerContext: interp.HandlerCtx(ctx),
		FS:             ctx.Value("FS").(vfs.Filesystem),
	}
}

func (ctx *BetterCtx) Printf(format string, args ...interface{}) {
	fmt.Fprintf(ctx.Stdout, format, args...)
}
func (ctx *BetterCtx) Printfe(format string, args ...interface{}) {
	fmt.Fprintf(ctx.Stderr, format, args...)
}
func (ctx *BetterCtx) Abs(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	abs, _ := filepath.Abs(filepath.Join(ctx.Dir, p))
	return abs

}

func (ctx *BetterCtx) Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	return io.Copy(dst, src)
}

var fakeshellBinaries map[string]FakeBinary = nil
var busyboxApplets map[string]FakeBinary = nil

func init() {
	fakeshellBinaries = map[string]FakeBinary{
		"/bin/env": func(ctx context.Context, args []string) error {
			log.Printf("/bin/env: %#v", args)
			c := UnwrapCtx(ctx)
			if len(args) == 0 {
				return fmt.Errorf("args to short for env")
			}
			if len(args) == 1 {
				c.Env.Each(func(k string, v expand.Variable) bool {
					c.Printf("%s=%s\n", k, v)
					return true
				})
				return nil
			}
			binaryToFind := args[1]
			var foundFile string
			if strings.HasPrefix(binaryToFind, "/") {
				foundFile = binaryToFind
			} else if strings.HasPrefix(binaryToFind, "./") {
				abs, _ := filepath.Abs(filepath.Join(c.Dir, binaryToFind[2:]))
				foundFile = abs
			} else {
				path := strings.Split(c.Env.Get("PATH").String(), ":")
				for _, p := range path {
					abs, _ := filepath.Abs(filepath.Join(p, binaryToFind))
					if info, err := c.FS.Stat(abs); err == nil && !info.IsDir() {
						foundFile = abs
						break
					}
				}
			}

			if foundFile == "" {
				if args[0] == "__CMD" {
					c.Printfe("-ash: %s: not found\n", binaryToFind)
				} else {
					c.Printfe("env: can't execute '%s': No such file or directory\n", binaryToFind)
				}
				return interp.NewExitStatus(1)
			}
			if bin, ok := fakeshellBinaries[foundFile]; ok {
				return bin(ctx, args[1:])
			}
			isBusyboxSymlink := false
			for _, s := range busyboxSymlinks {
				if foundFile == s {
					isBusyboxSymlink = true
					break
				}
			}
			if isBusyboxSymlink {
				return fakeshellBinaries["/bin/busybox"](ctx, args[1:])
			}

			return nil
		},
		"/bin/busybox": func(ctx context.Context, args []string) error {
			log.Printf("/bin/busybox: %#v", args)
			c := UnwrapCtx(ctx)
			applet := "busybox"
			if len(args) > 0 {
				applet = args[0]
			}
			segs := strings.Split(applet, "/")
			if len(args) > 1 && segs[len(segs)-1] == "busybox" {
				applet = args[1]
			}
			if applet == "busybox" || applet == "/bin/busybox" {
				c.Printf("%v", busyboxHelp)
				return nil
			}
			if applet, ok := busyboxApplets[applet]; ok {
				return applet(ctx, args[1:])
			}
			c.Printf("%s: applet not found\n", applet)
			return interp.NewExitStatus(1)
		},
	}

	busyboxApplets = map[string]FakeBinary{
		"ls": func(ctx context.Context, args []string) (erro error) {
			log.Printf("busybox: ls: %#v", args)
			c := UnwrapCtx(ctx)
			var dirsToList []string = args
			if len(args) == 0 {
				dirsToList = []string{"."}
			}
			for _, dir := range dirsToList {
				if len(dirsToList) > 1 {
					c.Printf("%s:\n", dir)
				}
				dir = c.Abs(dir)
				if info, err := c.FS.Stat(dir); err == nil && info.IsDir() {
					files, err := c.FS.ReadDir(dir)
					if err != nil {
						return err
					}
					for _, f := range files {
						c.Printf("%s\n", f.Name())
					}
				} else {
					if err == nil {
						c.Printfe("ls: %s: Not a directory\n", dir)
						erro = interp.NewExitStatus(1)
					} else {
						c.Printfe("ls : %s: No such file or directory\n", dir)
						erro = interp.NewExitStatus(1)
					}
				}
			}
			return
		},
		"cat": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			if len(args) == 0 {
				c.Copy(c.Stdout, c.Stdin)
				return
			}
			for _, f := range args {
				fullPath := c.Abs(f)
				if info, err := c.FS.Stat(fullPath); err == nil && info.IsDir() {
					c.Printfe("cat: read error: Is a directory\n")
					erro = interp.NewExitStatus(1)
					continue
				}
				file, err := c.FS.OpenFile(fullPath, os.O_RDONLY, 0)
				if err != nil {
					c.Printfe("cat: can't open '%s': %s\n", fullPath, err.Error())
					erro = interp.NewExitStatus(1)
					continue
				}
				defer file.Close()
				io.Copy(c.Stdout, file)
			}
			return
		},
		"cp": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			if len(args) < 2 {
				c.Printfe(busyboxHelps["cat"])
				return
			}
			for i := 0; i < len(args)-1; i++ {
				src := c.Abs(args[i])
				dst := c.Abs(args[i+1])
				if info, err := c.FS.Stat(src); err == nil && info.IsDir() {
					c.Printfe("cp: %s: Is a directory\n", src)
					erro = interp.NewExitStatus(1)
					continue
				}
				if info, err := c.FS.Stat(dst); err == nil && info.IsDir() {
					dst = c.Abs(filepath.Join(dst, filepath.Base(src)))
				}
				srcFile, err := c.FS.OpenFile(src, os.O_RDONLY, 0)
				if err != nil {
					c.Printfe("cp: %s: %s\n", src, err.Error())
					erro = interp.NewExitStatus(1)
					continue
				}
				defer srcFile.Close()
				dstFile, err := c.FS.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
				if err != nil {
					c.Printfe("cp: %s: %s\n", dst, err.Error())
					erro = interp.NewExitStatus(1)
					continue
				}
				defer dstFile.Close()

				if _, err := c.Copy(srcFile, dstFile); err != nil {
					c.Printfe("cp: %s: %s\n", src, err.Error())
					erro = interp.NewExitStatus(1)
					continue
				}
			}
			return
		},
		"echo": func(ctx context.Context, args []string) (erro error) {
			log.Printf("busybox: echo: %#v", args)
			c := UnwrapCtx(ctx)
			flags, rest := ParseBeginningShortFlags(args[1:])
			_, noTrailingNewline := flags["n"]
			_, interpretBackslash := flags["e"]
			if interpretBackslash {
				interpreted := []string{}
				for _, arg := range rest {
					runes := []rune(arg)
					result := ""
					for i := 0; i < len(runes); i++ {
						r := runes[i]
						if r == '\\' && i < len(runes)-1 {
							escape := runes[i+1]
							i++
							switch escape {
							case 'n':
								result += "\n"
							case 't':
								result += "\t"
							case 'r':
								result += "\r"
							case '\\':
								result += "\\"
							case 'x':
								if i+2 < len(runes) {
									hex := string(runes[i+1]) + string(runes[i+2])

									i++
									i++
									if hex, err := strconv.ParseInt(hex, 16, 32); err == nil {

										result += string(rune(hex))
									}
								}
							default:
								result += string(escape)
							}
						} else {
							result += string(r)
						}
					}
					interpreted = append(interpreted, result)
				}
				rest = interpreted
			}

			joined := strings.Join(rest, " ")
			if noTrailingNewline {
				c.Printf(joined)
			} else {
				c.Printf("%s\n", joined)
			}

			return
		},
		"dd": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			flags, bad := ParseFlagsDDStyle(args[1:])
			if bad {
				c.Printfe(busyboxHelps["dd"])
				erro = interp.NewExitStatus(1)
				return
			}
			var input io.Reader = c.Stdin
			var output io.Writer = c.Stdout
			if val, ok := flags["if"]; ok {

				i, err := c.FS.OpenFile(c.Abs(val), os.O_RDONLY, 0)
				if err != nil {
					c.Printfe("dd: can't open '%s': %s\n", val, err.Error())
					erro = interp.NewExitStatus(1)
					return
				}
				defer i.Close()
				input = i
			}
			if val, ok := flags["of"]; ok {
				o, err := c.FS.OpenFile(c.Abs(val), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
				if err != nil {
					c.Printfe("dd: can't open '%s': %s\n", val, err.Error())
					erro = interp.NewExitStatus(1)
					return
				}
				defer o.Close()
				output = o
			}
			var blockSize int64 = 512
			if val, ok := flags["bs"]; ok {
				blockSize, err 

			return
		},
	}

}
