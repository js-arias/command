// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import "github.com/js-arias/command"

// Usually help topics are in a same file,
// or in files with a common subject.

var topic = &command.Command{
	Usage: "topic",
	Short: "a help topic",
	Long: `
A help topic is a command that does not have any children, and it does not run
any action. It is used to provide online documentation of a particular topic
or subject relevant for the application.
	`,
}
