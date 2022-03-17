// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/js-arias/command"
)

var utf bool
var msg string

var hello = &command.Command{
	Usage: "hello [--utf8] [--message <message>]",
	Short: "print a greeting message",
	Long: `
Command hello prints a greeting "hello, world" message.

Flags are:

	--utf8
		Show an utf message.

	--message <message>
		Use the indicated message instead of "world" message.
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
		// Define flag usage in the Long field
		c.Flags().BoolVar(&utf, "utf8", false, "")
		c.Flags().StringVar(&msg, "message", "world", "")
	},
}
