// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import "github.com/js-arias/command"

var usgErr = &command.Command{
	Usage: "error <argument>",
	Short: "always return an usage error",
	Long: `
Command error always return an usage error. It demonstrates errors produced
when parsing flags and arguments.
	`,
	Run: func(c *command.Command, args []string) error {
		return c.UsageError("expecting arguments")
	},
}
