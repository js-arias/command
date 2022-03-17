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
//
// The common pattern of that applications is:
// APPNAME COMMAND --FLAG(s) ARGUMENT(s)
//
// Each interaction with the application
// is implemented as a command.
// A command usually runs an action,
// or alternatively
// it can provide multiple children commands,
// or a help topic.
// Flags can be given to modify the command's action
// and arguments usually are the objects
// in which the actions are executed.
//
// No constructor is required to create a new command.
// Just initialize the usage and documentation fields.
// To define the command's action implement the Run function.
// To define the flags used by the command
// implement SetFlag function
// and use method Flags to retrieve the current flag set
// of the command.
// To add children commands use the Add method.
//
// To run a command with a given set of arguments,
// use the method Execute.
package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
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

	parent *Command

	// children commands
	mu       sync.Mutex
	commands map[string]*Command
}

// Add adds a child command to a Command.
// This function panics if the child command is invalid:
//	* because it is nil
//	* because it does not have a name
//	* because there is a child command with the same name
//	* because the child already has a parent
//	* because the command is already a child of the child command
func (c *Command) Add(child *Command) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if child == nil {
		msg := fmt.Sprintf("command %q: adding a nil command", c.longName())
		panic(msg)
	}
	for p := c; p != nil; p = p.parent {
		if p == child {
			msg := fmt.Sprintf("command %q: adding %q: adding a command to itself or its children", c.longName(), child.name())
			panic(msg)
		}
	}

	name := child.name()
	if name == "" {
		msg := fmt.Sprintf("command %q: adding a command without usage", c.longName())
		panic(msg)
	}
	if _, dup := c.commands[name]; dup {
		msg := fmt.Sprintf("command %q: adding %q: command name already in use", c.longName(), name)
		panic(msg)
	}
	if child.parent != nil {
		msg := fmt.Sprintf("command %q: adding %q: command has another parent: %q", c.longName(), name, child.parent.longName())
		panic(msg)
	}

	if c.commands == nil {
		c.commands = make(map[string]*Command)
	}
	c.commands[name] = child
	child.parent = c
}

// Execute executes the Command
// with the arguments after the Command's name.
func (c *Command) Execute(args []string) error {
	// initialize flags
	c.flags = flag.NewFlagSet(c.name(), flag.ContinueOnError)
	c.flags.SetOutput(io.Discard) // do not print flag errors
	c.flags.Usage = func() {}
	if c.SetFlags != nil {
		c.SetFlags(c)
	}

	// parse flags
	err := c.flags.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		if c.hasChildren() {
			help(c.Stderr(), c)
			return nil
		}
		if c.Run == nil {
			help(c.Stdout(), c)
			return nil
		}
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
			return fmt.Errorf("%s: %v", c.longName(), err)
		}
		return nil
	}

	// non runnable command
	if !c.hasChildren() {
		return c.UsageError("unknown command")
	}

	if len(args) == 0 {
		help(c.Stderr(), c)
		return nil
	}
	child, ok := c.child(args[0])
	if !ok {
		if strings.ToLower(args[0]) != "help" {
			return usageError{
				c:   c,
				msg: fmt.Sprintf("%s %s: unknown command", c.longName(), args[0]),
			}
		}
		if err := c.help(args[1:]); err != nil {
			return err
		}
		return nil
	}
	if err := child.Execute(args[1:]); err != nil {
		return err
	}
	return nil
}

//Flags returns the current flag set of the Command.
func (c *Command) Flags() *flag.FlagSet {
	return c.flags
}

// Main executes a Command
// using the IS command line arguments.
// If an error happens when executing the Command,
// it will print the error
// in the programs' standard error,
// and finish the application.
//
// Main will panic if the Command is not a root Command.
func (c *Command) Main() {
	if c.parent != nil {
		msg := fmt.Sprintf("command %q: running Main in a command with parent", c.longName())
		panic(msg)
	}

	err := c.Execute(os.Args[1:])
	if errors.Is(err, usageError{}) {
		fmt.Fprintf(c.Stderr(), "%v\n", err)
		from := err.(usageError).c
		from.usage(c.Stderr())
		fmt.Fprintf(c.Stderr(), "Run %q for details.\n", from.helpPath())
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(c.Stderr(), "%v.\n", err)
		os.Exit(1)
	}
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
// By default returns its parent stderr
// or os.Stderr if parent is nil.
func (c *Command) Stderr() io.Writer {
	if c.stderr != nil {
		return c.stderr
	}
	if c.parent != nil {
		return c.parent.Stderr()
	}
	return os.Stderr
}

// Stdin returns the Command's standard input.
// By default returns its parent stdin
// or os.Stdin if parent is nil.
func (c *Command) Stdin() io.Reader {
	if c.stdin != nil {
		return c.stdin
	}
	if c.parent != nil {
		return c.parent.Stdin()
	}
	return os.Stdin
}

// Stdout returns the Command's standard output.
// By default returns its parent stdout
// or os.Stdout if parent is nil.
func (c *Command) Stdout() io.Writer {
	if c.stdout != nil {
		return c.stdout
	}
	if c.parent != nil {
		return c.parent.Stdout()
	}
	return os.Stdout
}

// UsageError should be returned by Run function
// when an error on an argument is found.
func (c *Command) UsageError(msg string) error {
	return usageError{
		c:   c,
		msg: fmt.Sprintf("%s: %s", c.longName(), msg),
	}
}

// Child returns a child Command
// with the given name.
func (c *Command) child(name string) (*Command, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	name = strings.ToLower(name)
	if name == "" {
		return nil, false
	}
	child, ok := c.commands[name]
	return child, ok
}

// Children returns the names
// of the children Commands.
func (c *Command) children() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var children []string
	for _, c := range c.commands {
		children = append(children, c.name())
	}
	sort.Strings(children)
	return children
}

// Help prints the help message of the Command.
func (c *Command) help(args []string) error {
	if len(args) == 0 {
		help(c.Stdout(), c)
		return nil
	}

	child, ok := c.child(args[0])
	if !ok {
		return fmt.Errorf("%s %s: unknown help topic. Run %q", c.helpPath(), strings.Join(args, " "), c.helpPath())
	}
	return child.help(args[1:])
}

// LongName returns the Command's long name,
// i.e. the name of the Command and all of its parents.
func (c *Command) longName() string {
	name := c.name()
	for p := c.parent; p != nil; p = p.parent {
		name = fmt.Sprintf("%s %s", p.name(), name)
	}
	return name
}

// LongUsage returns the Command's full usage line,
// i.e. the usage line including all of its parents.
func (c *Command) longUsage() string {
	usage := c.Usage
	for p := c.parent; p != nil; p = p.parent {
		usage = fmt.Sprintf("%s %s", p.name(), usage)
	}
	return usage
}

// Name returns the Command's name.
func (c *Command) name() string {
	f := strings.Fields(c.Usage)
	if len(f) == 0 {
		return ""
	}
	return strings.ToLower(f[0])
}

// HasChildren returns true if the command
// has at least one child.
func (c *Command) hasChildren() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.commands) > 0
}

// HelpPath returns the help path of the Command.
func (c *Command) helpPath() string {
	var path []string
	for p := c; p != nil; p = p.parent {
		path = append([]string{p.name()}, path...)
	}
	path = append([]string{path[0], "help"}, path[1:]...)
	return strings.Join(path, " ")
}

// Usage prints the Command's usage.
func (c *Command) usage(w io.Writer) {
	if c.Run == nil {
		return
	}
	fmt.Fprintf(w, "usage: %s\n", c.longUsage())
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

// Help prints the help of a command on w.
func help(w io.Writer, c *Command) {
	fmt.Fprintf(w, "%s\n\n", toTitle(c.Short))
	if c.Run != nil || c.hasChildren() {
		fmt.Fprintf(w, "Usage:\n\n    %s\n\n", c.longUsage())
	}

	if long := strings.TrimSpace(c.Long); long != "" {
		fmt.Fprintf(w, "%s\n\n", long)
	}

	if !c.hasChildren() {
		return
	}

	children := c.children()
	topics := false
	fmt.Fprintf(w, "The commands are:\n\n")
	for _, n := range children {
		cmd, ok := c.child(n)
		if !ok {
			continue
		}
		if cmd.Run == nil && !cmd.hasChildren() {
			topics = true
			continue
		}
		fmt.Fprintf(w, "    %-16s %s\n", cmd.name(), cmd.Short)
	}
	hp := c.helpPath()
	fmt.Fprintf(w, "\nUse %q for more information about a command.\n\n", hp+" <command>")

	if !topics {
		return
	}
	fmt.Fprintf(w, "Additional help topics:\n\n")
	for _, n := range children {
		t, ok := c.child(n)
		if !ok {
			continue
		}
		if t.Run != nil || t.hasChildren() {
			continue
		}
		fmt.Fprintf(w, "    %-16s %s\n", t.name(), t.Short)
	}
	fmt.Fprintf(w, "\nUse %q for more information about that topic.\n\n", hp+" <topic>")
}

func toTitle(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}
	r, i := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[i:]
}
