// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// MyApp is an demonstrative application
// to show the common usage of command package.
package main

import (
	"github.com/js-arias/command"
)

// App is the current application
// (i.e. the root command).
var app = &command.Command{
	Usage: "myapp <command> [<argument>...]",
	Short: "a demonstration application for Command package",
}

// Usually commands are added in an init function,
// and each command is defined in its own file.
func init() {
	app.Add(cmd)
	app.Add(errCmd)
	app.Add(hello)
	app.Add(topic)
}

// Usually Main is reduced to run Main method of the root command.
func main() {
	app.Main()
}
