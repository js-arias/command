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

// Package command implements a command line interface
// for applications that host multiple children commands
// similar to go and git.
package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// A Command is a command in an application
// like 'run' in 'go run'.
//
// When creating a Command always set up the Command's usage field.
// To provide help messages
// define a short and long description of the Command.
type Command struct {
	// Usage is the usage message of the Command
	// including flags and arguments.
	// (do not include any parent Command).
	//
	// Recommended syntax is as follows:
	//	[]  indicates an optional flag or argument.
	//	<>  indicates a value to be set by the user,
	//	    for example <file>.
	//	... indicates that multiple values
	//	    of the previous argument can be specified.
	//
	// The first word of the usage message
	// is taken to be the Command's name.
	Usage string

	// Short is a short description
	// (on a single line)
	// of the Command.
	Short string

	// Long is a long description
	// of the Command.
	Long string

	// Run runs the Command.
	// The args are the unparsed arguments.
	Run func(c *Command, args []string) error

	// SetFlags is the function used
	// to define the flags specific to the command.
	// Use method Flags to retrieve
	// the FlagSet of the command.
	SetFlags func(c *Command)

	flags *flag.FlagSet

	// Stdin specifies the Command's standard input
	stdin io.Reader

	// Stdout and stderr
	// specifies the Command's standard output and error.
	stdout io.Writer
	stderr io.Writer
}

// Execute executes the Command
// with the arguments after the Command's name.
func (c *Command) Execute(args []string) error {
	// initialize flags
	c.flags = flag.NewFlagSet(c.name(), flag.ContinueOnError)
	c.flags.Usage = func() {}
	if c.SetFlags != nil {
		c.SetFlags(c)
	}

	// parse flags
	err := c.flags.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		c.usage(c.Stderr())
		return nil
	}
	if err != nil {
		return c.UsageError(err.Error())
	}
	args = c.flags.Args()

	// run the command
	if c.Run != nil {
		err := c.Run(c, args)
		if errors.Is(err, usageError{}) {
			return err
		}
		if err != nil {
			return fmt.Errorf("%s: %v", c.name(), err)
		}
		return nil
	}
	return fmt.Errorf("%s: unknown command", c.name())
}

//Flags returns the current flag set of the Command.
func (c *Command) Flags() *flag.FlagSet {
	return c.flags
}

// SetStderr sets the Command's standard error.
func (c *Command) SetStderr(w io.Writer) {
	c.stderr = w
}

// SetStdin sets the Command's standard input.
func (c *Command) SetStdin(r io.Reader) {
	c.stdin = r
}

// SetStdout sets the Command's standard output.
func (c *Command) SetStdout(w io.Writer) {
	c.stdout = w
}

// Stderr returns the Command's standard error.
// By default returns os.Stderr.
func (c *Command) Stderr() io.Writer {
	if c.stderr != nil {
		return c.stderr
	}
	return os.Stderr
}

// Stdin returns the Command's standard input.
// By default returns os.Stdin.
func (c *Command) Stdin() io.Reader {
	if c.stdin != nil {
		return c.stdin
	}
	return os.Stdin
}

// Setout returns the Command's standard output.
// By default returns os.Stdout.
func (c *Command) Stdout() io.Writer {
	if c.stdout != nil {
		return c.stdout
	}
	return os.Stdout
}

// UsageError should be returned by Run function
// when an error on an argument is found.
func (c *Command) UsageError(msg string) error {
	return usageError{
		c:   c,
		msg: fmt.Sprintf("%s: %s", c.name(), msg),
	}
}

// Name returns the Command's name.
func (c *Command) name() string {
	f := strings.Fields(c.Usage)
	if len(f) == 0 {
		return ""
	}
	return strings.ToLower(f[0])
}

// Usage prints the Command's usage.
func (c *Command) usage(w io.Writer) {
	if c.Run == nil {
		return
	}
	fmt.Fprintf(w, "usage: %s\n", c.Usage)
}

type usageError struct {
	c   *Command
	msg string
}

func (e usageError) Error() string {
	return e.msg
}

func (e usageError) Is(target error) bool {
	if _, ok := target.(usageError); ok {
		return true
	}
	return false
}
