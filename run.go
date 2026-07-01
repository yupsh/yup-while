package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	command "github.com/gloo-foo/cmd-while"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "while"

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `while COMMAND [ARG...]

Read standard input line by line, running COMMAND for each line. Each line is
piped to COMMAND's standard input and replaced by COMMAND's standard output
(with one trailing newline trimmed).`

// Error is the package sentinel type; every error the wrapper emits is a const
// of this type, making each path testable with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// ErrNoCommand is returned when no command operands are supplied; without a
// command there is no body to run per line.
const ErrNoCommand Error = "no command given"

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the while CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newCommand(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version string, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "read from stdin and process line by line",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Action:         action(stdin, stdout, fs),
	}
}

func action(stdin io.Reader, stdout io.Writer, _ afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		if c.NArg() == 0 {
			return ErrNoCommand
		}
		source := gloo.ByteReaderSource([]io.Reader{stdin})
		body := lineRunner(c.Args().Slice())
		_, err := gloo.Run(source, gloo.ByteWriteTo(stdout), command.While(body))
		return err
	}
}

// lineRunner returns a body that runs the operand command once per line, piping
// the line to its stdin and using its stdout (less one trailing newline) as the
// transformed line.
func lineRunner(argv []string) func([]byte) ([]byte, error) {
	return func(line []byte) ([]byte, error) {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stdin = bytes.NewReader(line)
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return bytes.TrimSuffix(out, []byte("\n")), nil
	}
}
