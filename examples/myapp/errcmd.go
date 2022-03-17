// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"errors"

	"github.com/js-arias/command"
)

var errCmd = &command.Command{
	Usage: "error",
	Short: "always return an error",
	Long: `
Command error always return an error. It demonstrates errors produced from a
command.
	`,
	Run: func(c *command.Command, args []string) error {
		return errors.New("an error from a command")
	},
}
