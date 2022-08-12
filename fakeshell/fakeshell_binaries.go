package fakeshell

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

func (ctx *BetterCtx) CopyN(dst io.Writer, src io.Reader, n int64) (written int64, err error) {
	return io.CopyN(dst, src, n)
}

func (ctx *BetterCtx) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	return MapError(ctx.FS.OpenFile(name, flag, perm))
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
		"/bin/nc": func(ctx context.Context, args []string) (erro error) {
			// do nothing and return a success exit status
			return
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
				file, err := c.OpenFile(fullPath, os.O_RDONLY, 0)
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
				srcFile, err := c.OpenFile(src, os.O_RDONLY, 0)
				if err != nil {
					c.Printfe("cp: %s: %s\n", src, err.Error())
					erro = interp.NewExitStatus(1)
					continue
				}
				defer srcFile.Close()
				dstFile, err := c.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
				if err != nil {
					c.Printfe("cp: %s: %s\n", dst, err.Error())
					erro = interp.NewExitStatus(1)
					continue
				}
				defer dstFile.Close()

				if _, err := c.Copy(dstFile, srcFile); err != nil {
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
			if len(args) < 1 {
				c.Printfe("%s\n", busyboxHelps["dd"])
				erro = interp.NewExitStatus(1)
				return
			}
			flags, bad := ParseFlagsDDStyle(args)
			if bad {
				c.Printfe(busyboxHelps["dd"])
				erro = interp.NewExitStatus(1)
				return
			}
			var input io.Reader = c.Stdin
			var output io.Writer = c.Stdout
			var outputFile vfs.File
			if val, ok := flags["if"]; ok {

				i, err := c.OpenFile(c.Abs(val), os.O_RDONLY, 0)
				if err != nil {
					c.Printfe("dd: can't open '%s': %s\n", val, err.Error())
					erro = interp.NewExitStatus(1)
					return
				}
				defer i.Close()
				input = i
			}
			if val, ok := flags["of"]; ok {
				o, err := c.OpenFile(c.Abs(val), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
				if err != nil {
					c.Printfe("dd: can't open '%s': %s\n", val, err.Error())
					erro = interp.NewExitStatus(1)
					return
				}
				defer o.Close()
				output = o
				outputFile = o
			}
			numericFlags := map[string]int64{
				"bs":    512,
				"count": -1,
				"skip":  0,
				"seek":  0,
			}
			for k := range numericFlags {
				if val, ok := flags[k]; ok {
					if val, err := ParseFileSize(val); err != nil {
						c.Printfe("dd: %s\n", err.Error())
						erro = interp.NewExitStatus(1)
						return
					} else {
						numericFlags[k] = val
					}
				}
			}
			log.Printf("busybox: dd flags: %#v", flags)
			log.Printf("busybox: dd numericFlags: %#v", numericFlags)
			// log.Printf("busybox: dd input: %#v", input)
			// log.Printf("busybox: dd output: %#v", output)
			var bs, count, skip, seek = numericFlags["bs"], numericFlags["count"], numericFlags["skip"], numericFlags["seek"]

			if skip > 0 {
				if _, err := io.CopyN(ioutil.Discard, input, skip); err != nil {
					c.Printfe("dd: %s\n", err.Error())
					erro = interp.NewExitStatus(1)
				}
			}

			if seek > 0 {
				if outputFile != nil {
					if _, err := outputFile.Seek(seek, io.SeekStart); err != nil {
						c.Printfe("dd: %s\n", err.Error())
						erro = interp.NewExitStatus(1)
						return
					}
				}
			}

			for count > 0 || count == -1 {
				log.Printf("Copying %v bytes", bs)
				_, err := c.CopyN(output, input, bs)
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					c.Printfe("dd: %s\n", err.Error())
					erro = interp.NewExitStatus(1)
					return
				}
				count--
			}
			c.Printfe("\n")

			return
		},
		"id": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			log.Printf("busybox: id args: %#v", args)
			if len(args) < 1 {
				c.Printf("uid=0(root) gid=0(root)\n")
				return
			}
			argsParsed, _ := ParseBeginningShortFlags(args)
			_, name := argsParsed["n"]
			_, showGroups := argsParsed["g"]
			_, showUser := argsParsed["u"]
			if showGroups || showUser {
				if name {
					c.Printf("root\n")
				} else {
					c.Printf("0\n")
				}
			}
			return
		},
		"uname": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			flags, _, err := ParseBeginningShortFlagsValidated(args, "amnrspvio")
			if err != nil {
				c.Printfe("uname: %s\n", err.Error())
				c.Printfe(busyboxHelps["uname"])
				return interp.NewExitStatus(1)
			}
			_, all := flags["a"]
			result := []string{}
			delete(flags, "a")
			possibleFlags := []string{"m", "n", "r", "s", "p", "v", "i", "o"}
			for _, flag := range possibleFlags {
				if _, ok := flags[flag]; ok || all {
					val := unameData[flag]
					if val == "" {
						val = "unknown"
					}
					result = append(result, val)
				}
			}
			c.Printf("%v\n", strings.Join(result, " "))
			return
		},
		"rm": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			flags, rest, err := ParseBeginningShortFlagsValidated(args, "irRf")
			if len(args) < 1 || err != nil {
				if err != nil {
					c.Printfe("rm: %s\n", err.Error())
				}
				c.Printf(busyboxHelps["rm"])
				return interp.NewExitStatus(1)
			}
			_, recursive := flags["r"]
			queue := []string{}
			for _, arg := range rest {
				queue = append(queue, c.Abs(arg))
				if !recursive {
					if stat, err := c.FS.Stat(c.Abs(arg)); err != nil {
						c.Printf("rm: %s\n", err.Error())
						return interp.NewExitStatus(1)
					} else if stat.IsDir() {
						c.Printf("rm: cannot remove '%s': Is a directory\n", arg)
						return interp.NewExitStatus(1)
					}
				}
			}
			for _, arg := range queue {
				if err := c.FS.Remove(arg); err != nil {
					c.Printf("rm: %s\n", err.Error())
					return interp.NewExitStatus(1)
				}
			}
			return
		},
		"grep": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			if len(args) < 2 {
				c.Printf(busyboxHelps["grep"])
				return interp.NewExitStatus(1)
			}
			reader := bufio.NewReader(c.Stdin)
			pattern := args[0]

			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					c.Printf("grep: %s\n", err.Error())
					return interp.NewExitStatus(1)
				}
				if strings.Contains(line, pattern) {
					c.Printf("%s", line)
				}
			}
			return
		},
		"wc": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			flags, rest, err := ParseBeginningShortFlagsValidated(args, "clwL")
			if len(args) < 1 || err != nil {
				if err != nil {
					c.Printfe("wc: %s\n", err.Error())
				}
				c.Printf(busyboxHelps["wc"])
				return interp.NewExitStatus(1)
			}
			_, countLines := flags["l"]
			_, countWords := flags["w"]
			_, countBytes := flags["c"]
			_, longestLine := flags["L"]
			if !countLines && !countWords && !countBytes && !longestLine {
				countLines = true
				countWords = true
				countBytes = true
			}
			var totalLines, totalWords, totalBytes int
			var longestLineLength int
			if len(rest) == 0 {
				rest = []string{"-"}
			}
			for _, arg := range rest {
				var reader io.Reader
				if arg == "-" {
					reader = c.Stdin
				} else {
					if file, err := c.FS.OpenFile(c.Abs(arg), os.O_RDONLY, 0); err != nil {
						c.Printf("wc: %s\n", err.Error())
						return interp.NewExitStatus(1)
					} else {
						reader = file
					}
				}
				counter := &ScanByteCounter{}
				scanner := bufio.NewScanner(reader)
				scanner.Split(counter.Wrap(bufio.ScanLines))
				for scanner.Scan() {
					totalLines++
					line := scanner.Text()
					if longestLine {
						if len(line) > longestLineLength {
							longestLineLength = len(line)
						}
					}
					if countWords {
						totalWords += len(strings.Fields(line))
					}
				}
				if err := scanner.Err(); err != nil {
					c.Printf("wc: %s\n", err.Error())
					return interp.NewExitStatus(1)
				}
				if countBytes {
					totalBytes += counter.BytesRead
				}
			}
			if countLines {
				c.Printf("%d\n", totalLines)
			}
			if countWords {
				c.Printf("%d\n", totalWords)
			}
			if countBytes {
				c.Printf("%d\n", totalBytes)
			}
			if longestLine {
				c.Printf("%d\n", longestLineLength)
			}
			return
		},
		"which": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			_, rest, err := ParseBeginningShortFlagsValidated(args, "")
			if len(args) < 1 {
				return
			}
			if err != nil {
				c.Printfe("which: %s\n", err.Error())
				c.Printfe(busyboxHelps["which"])
				return interp.NewExitStatus(1)
			}
			cmd := rest[0]

			path := strings.Split(c.Env.Get("PATH").String(), ":")
			var foundFile string
			for _, p := range path {
				abs, _ := filepath.Abs(filepath.Join(p, cmd))
				if info, err := c.FS.Stat(abs); err == nil && !info.IsDir() {
					foundFile = abs
					break
				}
			}
			c.Printf("%s\n", foundFile)

			return
		},
		"w": func(ctx context.Context, args []string) (erro error) {
			c := UnwrapCtx(ctx)
			c.Printf(`11:22:37 up 55 days, 22:28,  3 users,  load average: 0.03, 0.01, 0.00
USER    TTY        LOGIN@   IDLE   JCPU   PCPU WHAT
root    pts/0     11:22    5.00s  0.67s  0.00s w
			`)
			return
		},
	}

}
