package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		stdin      string
		wantOut    string
		wantErrSub string
		args       []string
		wantCode   int
	}{
		{
			name:    "uppercase each line",
			args:    []string{"while", "tr", "a-z", "A-Z"},
			stdin:   "alpha\nbeta\n",
			wantOut: "ALPHA\nBETA\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"while", "--version"},
			wantOut: "while version 1.2.3\n",
		},
		{
			name:    "empty input emits nothing",
			args:    []string{"while", "tr", "a-z", "A-Z"},
			stdin:   "",
			wantOut: "",
		},
		{
			name:       "body command failure exits 1",
			args:       []string{"while", "definitely-not-a-real-command-xyz"},
			stdin:      "alpha\n",
			wantCode:   1,
			wantErrSub: "while:",
		},
		{
			name:       "missing command operand exits 1",
			args:       []string{"while"},
			stdin:      "alpha\n",
			wantCode:   1,
			wantErrSub: "while:",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"while", "--nope"},
			wantCode:   1,
			wantErrSub: "while:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, afero.NewMemMapFs())

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
