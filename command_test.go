// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// This work is derived from:
//     * Go tool source code
//       available at: https://cs.opensource.google/go/go.
//	 Copyright 2011 The Go Authors.
//     * Cobra source code
//       available at: https://github.com/spf13/cobra.
//       Copyright 2013 Steve Francia.

package command_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/js-arias/command"
)

func TestSimpleCommand(t *testing.T) {
	tests := map[string]struct {
		c    *command.Command
		args []string
		in   string
		out  string
		err  string
	}{
		"redirect stderr": {
			c: &command.Command{
				Usage: "echo <argument>...",
				Run:   echoToStderrRun,
			},
			args: []string{"print", "arguments"},
			err:  "print arguments",
		},
		"redirect input-output": {
			c: &command.Command{
				Usage: "cat",
				Run:   inToOutRun,
			},
			in:  "input\n\nstring",
			out: "input\n\nstring",
		},
		"default flag value": {
			c:   cmdWithFlags(),
			out: "hello, world",
		},
		"bool flag": {
			c:    cmdWithFlags(),
			args: []string{"--utf8"},
			out:  "hello, 世界",
		},
		"flag with value": {
			c:    cmdWithFlags(),
			args: []string{"-message", "flags"},
			out:  "hello, flags",
		},
		"help flag": {
			c:    cmdWithFlags(),
			args: []string{"--help"},
			err:  "usage: hello [--utf8] [--message <message>]",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testExecute(t, test.c, test.args, test.in, test.out, test.err)
		})
	}
}

func testExecute(t testing.TB, c *command.Command, args []string, in, out, errOut string) {
	t.Helper()

	c.SetStdin(strings.NewReader(in))
	var outBuf, errBuf bytes.Buffer
	c.SetStdout(&outBuf)
	c.SetStderr(&errBuf)

	if err := c.Execute(args); err != nil {
		t.Fatalf("args %v: unexpected error: %v", args, err)
	}

	if got := strings.TrimSpace(outBuf.String()); got != out {
		t.Errorf("args %v: stdout: got %q, want %q", args, got, out)
	}
	if got := strings.TrimSpace(errBuf.String()); got != errOut {
		t.Errorf("args %v: stderr: got %q, want %q", args, got, errOut)
	}
}

func TestSimpleError(t *testing.T) {
	tests := map[string]struct {
		c      *command.Command
		args   []string
		errMsg string
	}{
		"command withut run": {
			c: &command.Command{
				Usage: "topic",
			},
			errMsg: "topic: unknown command",
		},
		"error on the run": {
			c: &command.Command{
				Usage: "err",
				Run:   errRun,
			},
			errMsg: "err: an error from a command",
		},
		"invalid flag": {
			c:      cmdWithFlags(),
			args:   []string{"--invalid"},
			errMsg: "hello: flag provided but not defined: -invalid",
		},
		"invalid argument": {
			c: &command.Command{
				Usage: "err <argument>...",
				Run: func(c *command.Command, args []string) error {
					return c.UsageError("expecting arguments")
				},
			},
			errMsg: "err: expecting arguments",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testExecuteError(t, test.c, test.args, test.errMsg)
		})
	}
}

func testExecuteError(t testing.TB, c *command.Command, args []string, msg string) {
	t.Helper()

	err := c.Execute(args)
	if err == nil {
		t.Fatalf("args %v: expecting error %q", args, msg)
	}

	if got := err.Error(); got != msg {
		t.Errorf("args %v: got error %q, want %q", args, got, msg)
	}
}

func echoToStderrRun(c *command.Command, args []string) error {
	fmt.Fprintf(c.Stderr(), "%s\n", strings.Join(args, " "))
	return nil
}

func errRun(c *command.Command, args []string) error {
	return errors.New("an error from a command")
}

func inToOutRun(c *command.Command, args []string) error {
	r := bufio.NewReader(c.Stdin())
	for {
		ln, err := r.ReadString('\n')
		if ln == "" {
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return err
			}
		}
		fmt.Fprintf(c.Stdout(), "%s", ln)
	}
	return nil
}

func cmdWithFlags() *command.Command {
	var utf bool
	var msg string

	return &command.Command{
		Usage: "hello [--utf8] [--message <message>]",
		Short: "print a hello message",
		Run: func(c *command.Command, args []string) error {
			if utf {
				fmt.Fprintf(c.Stdout(), "hello, 世界\n")
				return nil
			}
			fmt.Fprintf(c.Stdout(), "hello, %s\n", msg)
			return nil
		},
		SetFlags: func(c *command.Command) {
			c.Flags().BoolVar(&utf, "utf8", false, "print an utf8 message")
			c.Flags().StringVar(&msg, "message", "world", "sets the greeting message")
		},
	}
}
