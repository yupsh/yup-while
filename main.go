// Command yup-while is the CLI wrapper around github.com/gloo-foo/cmd-while.
package main

import (
	"bytes"
	"os/exec"

	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-while"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const name = "while"

// Error is the package sentinel type; every error the wrapper emits is a const
// of this type, making each path testable with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// ErrNoCommand is returned when no command operands are supplied; without a
// command there is no body to run per line.
const ErrNoCommand Error = "no command given"

// commandLine is the operand command (program plus its arguments) run once per
// input line.
type commandLine []string

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `while COMMAND [ARG...]

Read standard input line by line, running COMMAND for each line. Each line is
piped to COMMAND's standard input and replaced by COMMAND's standard output
(with one trailing newline trimmed).`

// spec declares the while wrapper: a stdin filter that runs the operand command
// once per input line, replacing each line with the command's output.
var spec = clix.Spec{
	Name:     name,
	Summary:  "read from stdin and process line by line",
	Synopsis: synopsis,
	Build:    build,
}

// build maps the invocation to while's pipeline: standard input feeds while,
// whose per-line body runs the operand command. A bare invocation with no
// command is a usage error.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	argv := commandLine(inv.Args.Args().Slice())
	if len(argv) == 0 {
		return nil, nil, ErrNoCommand
	}
	return clix.Stdin(inv.Stdin), command.While(lineRunner(argv)), nil
}

// lineRunner returns a body that runs the operand command once per line, piping
// the line to its stdin and using its stdout (less one trailing newline) as the
// transformed line.
func lineRunner(argv commandLine) func([]byte) ([]byte, error) {
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

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
