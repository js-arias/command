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
	"fmt"
	"testing"

	"github.com/js-arias/command"
)

func TestAddPanic(t *testing.T) {
	app := &command.Command{Usage: "other"}
	used := &command.Command{Usage: "used"}
	app.Add(used)

	tests := map[string]struct {
		c   *command.Command
		msg string
	}{
		"adding a nil command": {
			msg: `command "failing-app": adding a nil command`,
		},
		"command without a name": {
			c:   &command.Command{},
			msg: `command "failing-app": adding a command without usage`,
		},
		"command without name (spaces)": {
			c:   &command.Command{Usage: "    "},
			msg: `command "failing-app": adding a command without usage`,
		},
		"repeated command": {
			c:   &command.Command{Usage: "hello"},
			msg: `command "failing-app": adding "hello": command name already in use`,
		},
		"repeated command (caps)": {
			c:   &command.Command{Usage: "HELLO"},
			msg: `command "failing-app": adding "hello": command name already in use`,
		},
		"command with other parent": {
			c:   used,
			msg: `command "failing-app": adding "used": command has another parent: "other"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			msg := appPanic(test.c)
			if msg != test.msg {
				t.Errorf("%s: panic %q, want %q", name, msg, test.msg)
			}
		})
	}
}

func TestAddLoopPanic(t *testing.T) {
	tests := map[string]struct {
		f   func() string
		msg string
	}{
		"adding command to itself": {
			f:   selfPanic,
			msg: `command "loop": adding "loop": adding a command to itself or its children`,
		},
		"adding command to its children": {
			f:   loopPanic,
			msg: `command "parent loop": adding "parent": adding a command to itself or its children`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			msg := test.f()
			if msg != test.msg {
				t.Errorf("%s: panic %q, want %q", name, msg, test.msg)
			}
		})
	}
}

func appPanic(c *command.Command) (msg string) {
	defer func() {
		p := recover()
		msg = capturePanicMessage(p)
	}()

	app := &command.Command{Usage: "failing-app"}
	app.Add(&command.Command{Usage: "hello"})
	app.Add(c)

	return ""
}

func selfPanic() (msg string) {
	defer func() {
		p := recover()
		msg = capturePanicMessage(p)
	}()

	app := &command.Command{Usage: "loop"}
	app.Add(app)

	return ""
}

func loopPanic() (msg string) {
	defer func() {
		p := recover()
		msg = capturePanicMessage(p)
	}()

	app := &command.Command{Usage: "loop"}
	p := &command.Command{Usage: "parent"}
	p.Add(app)

	app.Add(p)
	return ""
}

func capturePanicMessage(p any) string {
	if p == nil {
		return ""
	}

	if msg, ok := p.(string); ok {
		return msg
	}
	return fmt.Sprintf("panic %q", p)
}
