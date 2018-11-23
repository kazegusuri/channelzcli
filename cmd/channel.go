package cmd

import (
	"github.com/spf13/cobra"
)

type ChannelCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewChannelCommand(opts *GlobalOptions) *ChannelCommand {
	c := &ChannelCommand{
		cmd: &cobra.Command{
			Use:   "channel",
			Short: "",
		},
		opts: opts,
	}

	c.cmd.AddCommand(NewChannelListCommand(c.opts).Command())
	return c
}

func (c *ChannelCommand) Command() *cobra.Command {
	return c.cmd
}
