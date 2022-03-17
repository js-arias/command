// Copyright © 2021 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/js-arias/command"
)

var stderr bool

var cat = &command.Command{
	Usage: "cat [--stderr]",
	Short: "print stdin into stdout",
	Long: `
Command cat demonsatrates basic IO redirection. By default it will print the
contents of stdin into stdout. If flag --stderr is defined, it will output in
the stderr.
	`,
	Run: func(c *command.Command, args []string) error {
		r := bufio.NewReader(c.Stdin())
		out := c.Stdout()
		if stderr {
			out = c.Stderr()
		}

		for {
			ln, err := r.ReadString('\n')
			if ln == "" {
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return err
				}
			}
			fmt.Fprintf(out, "%s", ln)
		}
		return nil
	},
	SetFlags: func(c *command.Command) {
		c.Flags().BoolVar(&stderr, "stderr", false, "")
	},
}
