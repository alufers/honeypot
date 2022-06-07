package fakeshell

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/blang/vfs"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

//ServiceFakeshell runs the fake shell on the given reader and writer
func ServiceFakeshell(input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	writer := bufio.NewWriter(output)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "FS", MakeFS())
	runner := createRunner(reader, writer, ctx.Value("FS").(vfs.Filesystem))
	for {
		writer.WriteString("# ")
		writer.Flush()
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		file, err := syntax.NewParser().Parse(strings.NewReader(string(line)), "")
		if err != nil {
			writer.WriteString("-ash: syntax error: " + err.Error() + "\r\n")
			writer.Flush()
			continue
		}
		runner.Run(ctx, file)
		writer.Flush()
	}
}

func ServiceFakeShellOnTerminal(t *term.Terminal, r io.Reader, feedback io.Writer, lineCallback func(string)) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "FS", MakeFS())
	runner := createRunner(r, io.MultiWriter(t, feedback), ctx.Value("FS").(vfs.Filesystem))
	for {

		line, err := t.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if lineCallback != nil {
			lineCallback(line)
		}
		feedback.Write([]byte("# "))
		feedback.Write([]byte(line))
		feedback.Write([]byte("\r\n"))
		file, err := syntax.NewParser().Parse(strings.NewReader(string(line)), "")
		if err != nil {
			t.Write([]byte("-ash: syntax error: " + err.Error() + "\r\n"))
			continue
		}
		runner.Run(ctx, file)

	}
}

func ServeSingleCommand(cmd string, r io.Reader, w io.Writer, feedback io.Writer) (error, uint8) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "FS", MakeFS())
	runner := createRunner(r, io.MultiWriter(w, feedback), ctx.Value("FS").(vfs.Filesystem))
	file, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	feedback.Write([]byte("# "))
	feedback.Write([]byte(cmd))
	feedback.Write([]byte("\r\n"))
	if err != nil {
		feedback.Write([]byte("-ash: syntax error: " + err.Error() + "\r\n"))
		return err, 1
	}
	err = runner.Run(ctx, file)
	if status, ok := interp.IsExitStatus(err); ok {
		return nil, status
	}

	return nil, 0
}
