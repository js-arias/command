// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/js-arias/command"
)

var echo = &command.Command{
	Usage: "echo <argument>...",
	Short: "print its arguments to stdout",
	Long: `
Command echo prints its arguments to stdout in a single line.
	`,
	Run: func(c *command.Command, args []string) error {
		fmt.Fprintf(c.Stdout(), "%s\n", strings.Join(args, " "))
		return nil
	},
}
