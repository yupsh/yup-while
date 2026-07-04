package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// parse runs args through a bare command and returns the parsed accessor.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func invocation(t *testing.T, args ...string) clix.Invocation {
	return clix.Invocation{Args: parse(t, args...), Stdin: strings.NewReader(""), Fs: afero.NewMemMapFs()}
}

func TestBuild_NoCommandIsUsageError(t *testing.T) {
	src, filter, err := build(invocation(t, name))
	if !errors.Is(err, ErrNoCommand) {
		t.Fatalf("err=%v, want ErrNoCommand", err)
	}
	if src != nil || filter != nil {
		t.Fatalf("src=%v filter=%v, want both nil on error", src, filter)
	}
	if err.Error() != string(ErrNoCommand) {
		t.Fatalf("message=%q, want %q", err.Error(), string(ErrNoCommand))
	}
}

func TestBuild_PassesOperandsToWhile(t *testing.T) {
	src, filter, err := build(invocation(t, name, "cat"))
	if err != nil || src == nil || filter == nil {
		t.Fatalf("build: src=%v filter=%v err=%v", src, filter, err)
	}
}

func TestLineRunner_TransformsLine(t *testing.T) {
	got, err := lineRunner(commandLine{"cat"})([]byte("hello\n"))
	if err != nil {
		t.Fatalf("lineRunner: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("output=%q, want %q", got, "hello")
	}
}

func TestLineRunner_CommandError(t *testing.T) {
	_, err := lineRunner(commandLine{"this-command-does-not-exist-xyz"})([]byte("x"))
	if err == nil {
		t.Fatal("lineRunner: want error for missing command, got nil")
	}
}

func Test_main(t *testing.T) {
	orig := runMain
	t.Cleanup(func() { runMain = orig })
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
}
