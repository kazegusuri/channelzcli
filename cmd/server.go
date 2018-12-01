package cmd

import (
	"github.com/spf13/cobra"
)

type ServerCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewServerCommand(opts *GlobalOptions) *ServerCommand {
	c := &ServerCommand{
		cmd: &cobra.Command{
			Use:   "server",
			Short: "",
		},
		opts: opts,
	}

	c.cmd.AddCommand(NewServerListCommand(c.opts).Command())
	c.cmd.AddCommand(NewServerDescribeCommand(c.opts).Command())
	return c
}

func (c *ServerCommand) Command() *cobra.Command {
	return c.cmd
}
