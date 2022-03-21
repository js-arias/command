# Command

Package `command` implements a command line interface
for applications that host multiple children commands
similar to `go` and `git`.

This work is based on the [go tool command interface](https://cs.opensource.google/go/go)
and the [cobra package](https://github.com/spf13/cobra).

## Getting started

The common pattern of that applications like `go` and `git` is:

`APPNAME COMMAND --FLAG(s) ARGUMENT(s)`

Each interaction with the application
is implemented as a command.
A command usually runs an action,
or alternatively
it can provide multiple children commands,
or a help topic.
Flags can be given to modify the command's action
and arguments usually are the objects
in which the actions are executed.

In package `command` the commands are implemented
by the type `Command`.
No constructor is required to create a new command.
Just initialize the usage and documentation fields.
To define the command's action implement the `Run` function.
To define the flags used by the command
implement `SetFlag` function
and use method `Flags` to retrieve the current flag set
of the command.

Here is an example of a command initialization:

```go
var hello = &command.Command{
  Usage: "hello [--utf8] [--message <message>]",
  Short: "print a greeting message",
  Long: `
Command hello prints a greeting "hello, world" message.

Flags are:

  -utf8
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
```

To add children commands use the `Add` method.

To run a command with a given set of arguments,
use the method `Execute`.
This method provide flag parsing,
and default help for the command.

Usually,
a root command is executed using the arguments
from the command line.
Use the method `Main`:

```go
// App is the current application
// (i.e. the root command).
var app = &command.Command{
  Usage: "myapp <command> [<argument>...]",
  Short: "a demonstration application for Command package",
}

func init() {
  // add commands
  app.Add(hello)
}

func main() {
  app.Main()
}
```

If a command does not have defined
a `Run` function,
and does not have any children,
it is called a *help topic*,
and is used only for documentation.

See directory [examples/myapp](https://github.com/js-arias/command/tree/main/examples/myapp)
for a demonstration application.

Package `command` is intentionally simple,
for a very detailed package see the [cobra package](https://github.com/spf13/cobra).

## Authorship and license

Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>.
All rights reserved.
Distributed under BSD2 licenses that can be found in the LICENSE file.

This work is based on:

* Go tool source code.
  Available at <https://cs.opensource.google/go/go>.
  Copyright 2011 The Go Authors.
* Cobra package source code.
  Available at <https://github.com/spf13/cobra>.
  Copyright 2013 Steve Francia.
