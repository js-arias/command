// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import "github.com/js-arias/command"

// Cmd is a sub-command that also has its own sub-commands.
var cmd = &command.Command{
	Usage: "cmd <command> [<argument>...]",
	Short: "a collection of commands",
}

func init() {
	cmd.Add(cat)
	cmd.Add(echo)
	cmd.Add(usgErr)
}
