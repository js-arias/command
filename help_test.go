// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// This work is derived from:
//     * Go tool source code
//       available at: https://github.com/golang/go.
//	 Copyright 2011 The Go Authors.
//     * Cobra source code
//       available at: https://github.com/spf13/cobra.
//       Copyright 2013 Steve Francia.

package command_test

import (
	"testing"
)

var appHelp = `App is an app for testing

Usage:

    app <command> [<argument>...]

The commands are:

    cmd              a collection of commands
    error            always return an error
    hello            print a hello message

Use "app help <command>" for more information about a command.

Additional help topics:

    topic            a help topic

Use "app help <topic>" for more information about that topic.`

var cmdHelp = `A collection of commands

Usage:

    app cmd <command> [<argument>...]

The commands are:

    cat              print stdin
    echo             print its arguments
    error            always return an error

Use "app help cmd <command>" for more information about a command.`

var helloHelp = `Print a hello message

Usage:

    app hello [--utf8] [--message <message>]

Command hello prints the well known "hello, world" message, or if --message
flag is defined, a personalized hello message.`

var catHelp = `Print stdin

Usage:

    app cmd cat

Command cat is used to print the content of the stdin into the stdout.`

var helpTopic = `A help topic

A help topic is a non-runnable command used only for documentation.`

func TestExecuteHelp(t *testing.T) {
	tests := map[string]struct {
		args []string
		err  string
		out  string
	}{
		"app command list (no arguments)": {
			err: appHelp,
		},
		"child command with sub-commands (no arguments)": {
			args: []string{"cmd"},
			err:  cmdHelp,
		},
		"app help": {
			args: []string{"help"},
			out:  appHelp,
		},
		"help on a children command": {
			args: []string{"help", "cmd"},
			out:  cmdHelp,
		},
		"children with commands help": {
			args: []string{"cmd", "help"},
			out:  cmdHelp,
		},
		"help on children": {
			args: []string{"help", "hello"},
			out:  helloHelp,
		},
		"help on grand children": {
			args: []string{"help", "cmd", "cat"},
			out:  catHelp,
		},
		"children with commands help on children": {
			args: []string{"cmd", "help", "cat"},
			out:  catHelp,
		},
		"help topic": {
			args: []string{"help", "topic"},
			out:  helpTopic,
		},
		"help flag on root": {
			args: []string{"-h"},
			err:  appHelp,
		},
		"help flag on a children command with sub-commands": {
			args: []string{"cmd", "-h"},
			err:  cmdHelp,
		},
		"help flag on topic": {
			args: []string{"topic", "--help"},
			out:  helpTopic,
		},
	}

	app := newApp()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testExecute(t, app, test.args, "", test.out, test.err)
		})
	}
}

func TestHelpError(t *testing.T) {
	tests := map[string]struct {
		args   []string
		errMsg string
	}{
		"unknown help topic on root": {
			args:   []string{"help", "unknown"},
			errMsg: `app help unknown: unknown help topic. Run "app help"`,
		},
		"unknown help topic on children": {
			args:   []string{"cmd", "help", "unknown"},
			errMsg: `app help cmd unknown: unknown help topic. Run "app help cmd"`,
		},
		"unknown help topic (on sub-command": {
			args:   []string{"help", "cmd", "unknown"},
			errMsg: `app help cmd unknown: unknown help topic. Run "app help cmd"`,
		},
		"unknown help topic (multiple arguments)": {
			args:   []string{"help", "unknown", "command"},
			errMsg: `app help unknown command: unknown help topic. Run "app help"`,
		},
		" extra arguments in a children command": {
			args:   []string{"help", "hello", "unknown"},
			errMsg: `app help hello unknown: unknown help topic. Run "app help hello"`,
		},
	}

	app := newApp()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testExecuteError(t, app, test.args, test.errMsg)
		})
	}
}
