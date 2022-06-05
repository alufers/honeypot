package fakeshell

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/blang/vfs"
	"golang.org/x/term"
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
