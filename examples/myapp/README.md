# An example of a `command` application

This small example shows how to use `command`package.

In the main file (see `main.go`)
create a new command to be used as the application,
in the initialization
add commands to them using `Àdd` method.
And in the `main`function just call `Main` method:

```go
// App is the current application
// (i.e. the root command).
var app = &command.Command{
  Usage: "myapp <command> [<argument>...]",
  Short: "a demonstration application for Command package",
}

// Usually commands are added in an init function.
func init() {
  // add commands
  app.Add(hello)
}

// Usually Main is reduced to run Main method of the root command.
func main() {
  app.Main()
}
```

The `Main` method will exit the program
if any error happens.
If you want more control,
you can use `Execute` method directly.

The best practice is to use a file for each command
(they might even be in their own package).
Assign a `Run` function to the command,
and if there are flags to be defined,
define a `SetFlag` function
(using method 'Flags' to retrieve the command's FlagSet).
Document all the arguments and flags
in the `Long` field of the command:

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

Any command can have descendant commands
(see `cmd.go` file).
While it is preferable to restrict command addition
to the main application command,
it is useful in some cases that descendant a command
has its own descendants
(for an example of an application
that use that kind of design see
`go mod` command of go tool
[please note that `go` tool **does not** use `command` package]).

If possible,
provide documentation commands (called *topics*)
to provide additional help information
(see `help.go`).
