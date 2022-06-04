package fakeshell

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/blang/vfs"
	"mvdan.cc/sh/v3/syntax"
)

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
		}
		runner.Run(ctx, file)
		writer.Flush()
	}
}
