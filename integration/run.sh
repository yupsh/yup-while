#!/bin/sh
# Integration checks for yup-while, run inside a Debian (GNU coreutils) container.
#
# while is a gloo-specific control command: it reads standard input line by line
# and runs COMMAND for each line, piping the line to COMMAND's stdin and
# replacing the line with COMMAND's stdout (one trailing newline trimmed). The
# shell `while` is a builtin, not a program, so there is NO standard Unix `while`
# binary to compare against. This harness is therefore ASSERT-ONLY: every check
# asserts yup-while's own documented behavior. The coreutils tools (tr, sed, rev)
# appear only as the per-line COMMAND body, never as a reference for `while`.
set -eu

fails=0

# assert WANT CMD... — pipe STDIN (via the global $stdin) into yup-while CMD...
# and assert its stdout equals WANT exactly.
assert() {
	want=$1
	shift
	got=$(printf '%s' "$stdin" | yup-while "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  while %s\n' "$*"
	else
		printf 'FAIL  assert  while %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# assert_code WANTCODE CMD... — assert yup-while exits with WANTCODE. The
# pipeline is guarded with `|| gotcode=$?` so a non-zero exit (the very thing
# under test) does not trip `set -e`.
assert_code() {
	wantcode=$1
	shift
	gotcode=0
	printf '%s' "$stdin" | yup-while "$@" >/dev/null 2>&1 || gotcode=$?
	if [ "$gotcode" -eq "$wantcode" ]; then
		printf 'ok    code    while %s -> %s\n' "$*" "$wantcode"
	else
		printf 'FAIL  code    while %s\n        want code: %s\n        got code:  %s\n' "$*" "$wantcode" "$gotcode"
		fails=$((fails + 1))
	fi
}

# Each line is transformed by the body command's stdout. Uppercase via tr.
stdin=$(printf 'alpha\nbeta\n')
assert "$(printf 'ALPHA\nBETA')" tr a-z A-Z

# Reverse each line via rev (coreutils-adjacent; bsdextrautils ships it in the image).
stdin=$(printf 'abc\nxyz\n')
assert "$(printf 'cba\nzyx')" rev

# Body command with arguments: sed substitution applied per line.
stdin=$(printf 'one\ntwo\n')
assert "$(printf 'X\ntwo')" sed 's/one/X/'

# A constant-output body replaces every line with the same text.
stdin=$(printf 'a\nb\nc\n')
assert "$(printf 'z\nz\nz')" echo z

# Empty input emits nothing, regardless of the body command.
stdin=''
assert "" tr a-z A-Z

# Documented contract: a single trailing newline of the body's stdout is
# trimmed, so cat (identity) round-trips each line unchanged.
stdin=$(printf 'keep\nthese\n')
assert "$(printf 'keep\nthese')" cat

# No command operand -> usage error, exit 1. (Invoked with zero operands, so it
# cannot go through assert_code, which always passes at least one operand.)
gotcode=0
printf 'alpha\n' | yup-while >/dev/null 2>&1 || gotcode=$?
if [ "$gotcode" -eq 1 ]; then
	printf 'ok    code    while (no command) -> 1\n'
else
	printf 'FAIL  code    while (no command)\n        want code: 1\n        got code:  %s\n' "$gotcode"
	fails=$((fails + 1))
fi

# A body command that does not exist -> exit 1.
stdin=$(printf 'alpha\n')
assert_code 1 definitely-not-a-real-command-xyz

# An unknown flag -> exit 1.
stdin=''
assert_code 1 --nope

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
