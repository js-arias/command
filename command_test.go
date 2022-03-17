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

func TestCommand(t *testing.T) {
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
		"children command": {
			c:    newApp(),
			args: []string{"hello"},
			out:  "hello, world",
		},
		"children command (caps)": {
			c:    newApp(),
			args: []string{"HELLO", "-message", "caps"},
			out:  "hello, caps",
		},
		"grand children command": {
			c:    newApp(),
			args: []string{"cmd", "echo", "print", "arguments"},
			err:  "print arguments",
		},
		"redirecting io": {
			c:    newApp(),
			args: []string{"cmd", "cat"},
			in:   "input\nstring",
			out:  "input\nstring",
		},
		"flag in children": {
			c:    newApp(),
			args: []string{"hello", "-utf8"},
			out:  "hello, 世界",
		},
		"children usage": {
			c:    newApp(),
			args: []string{"cmd", "cat", "-h"},
			err:  "usage: app cmd cat",
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

func TestError(t *testing.T) {
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
		"unknown command": {
			c:      newApp(),
			args:   []string{"unknown"},
			errMsg: "app unknown: unknown command",
		},
		"running a help topic": {
			c:      newApp(),
			args:   []string{"topic"},
			errMsg: "app topic: unknown command",
		},
		"an error from a command": {
			c:      newApp(),
			args:   []string{"error"},
			errMsg: "app error: an error from a command",
		},
		"an error from a command (invalid arguments)": {
			c:      newApp(),
			args:   []string{"cmd", "error"},
			errMsg: "app cmd error: expecting arguments",
		},
		"undefined flag": {
			c:      newApp(),
			args:   []string{"hello", "--undef"},
			errMsg: "app hello: flag provided but not defined: -undef",
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
		Long: `
Command hello prints the well known "hello, world" message, or if --message
flag is defined, a personalized hello message.
		`,
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

func newApp() *command.Command {
	app := &command.Command{
		Usage: "app <command> [<argument>...]",
		Short: "app is an app for testing",
	}

	app.Add(cmdWithFlags())

	errCmd := &command.Command{
		Usage: "error",
		Short: "always return an error",
		Run:   errRun,
	}
	app.Add(errCmd)

	topic := &command.Command{
		Usage: "topic",
		Short: "a help topic",
		Long:  "A help topic is a non-runnable command used only for documentation.",
	}
	app.Add(topic)

	cmd := &command.Command{
		Usage: "cmd <command> [<argument>...]",
		Short: "a collection of commands",
	}
	app.Add(cmd)

	echo := &command.Command{
		Usage: "echo <argument>...",
		Short: "print its arguments",
		Run:   echoToStderrRun,
	}
	cmd.Add(echo)

	cat := &command.Command{
		Usage: "cat",
		Short: "print stdin",
		Long:  "Command cat is used to print the content of the stdin into the stdout.",
		Run:   inToOutRun,
	}
	cmd.Add(cat)

	errCmd = &command.Command{
		Usage: "error <argument>...",
		Short: "always return an error",
		Run: func(c *command.Command, args []string) error {
			return c.UsageError("expecting arguments")
		},
	}
	cmd.Add(errCmd)

	return app
}
