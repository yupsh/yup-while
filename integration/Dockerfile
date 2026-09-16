# Integration harness: build the yup-while binary for Linux, then run it inside a
# Debian image (GNU coreutils) so its flags, inputs, and outputs can be checked.
# while is a gloo-specific control command with NO standard Unix equivalent (the
# shell `while` is a builtin, not a program), so the harness asserts yup-while's
# own documented behavior — it uses coreutils tools (tr, sed, …) only as the
# per-line COMMAND body. Invoked by `make integration`.
FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/yup-while .

FROM debian:stable-slim
# rev (used as a per-line body command) ships in bsdextrautils, not coreutils.
RUN apt-get update && apt-get install -y --no-install-recommends bsdextrautils && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/yup-while /usr/local/bin/yup-while
COPY integration/run.sh /run.sh
ENTRYPOINT ["/bin/sh", "/run.sh"]
