package cmd

import (
	"context"

	"github.com/kazegusuri/channelzcli/channelz"
	"github.com/spf13/cobra"
)

type ChannelListCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewChannelListCommand(opts *GlobalOptions) *ChannelListCommand {
	c := &ChannelListCommand{
		cmd: &cobra.Command{
			Use:          "list",
			Short:        "List TopChannelz",
			Args:         cobra.ExactArgs(0),
			SilenceUsage: true,
		},
		opts: opts,
	}
	c.cmd.RunE = c.Run
	return c
}

func (c *ChannelListCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *ChannelListCommand) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	conn, err := newGRPCConnection(ctx, c.opts.Address, c.opts.Insecure)
	if err != nil {
		return err
	}
	defer conn.Close()

	cc := channelz.NewClient(conn)
	cc.GetTopChannels(ctx)

	return nil
}
