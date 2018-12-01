package cmd

import (
	"context"
	"time"

	"github.com/kazegusuri/channelzcli/channelz"
	"github.com/spf13/cobra"
)

type ChannelDescribeCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewChannelDescribeCommand(opts *GlobalOptions) *ChannelDescribeCommand {
	c := &ChannelDescribeCommand{
		cmd: &cobra.Command{
			Use:          "describe",
			Short:        "describe channel",
			Args:         cobra.ExactArgs(1),
			SilenceUsage: true,
		},
		opts: opts,
	}
	c.cmd.RunE = c.Run
	return c
}

func (c *ChannelDescribeCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *ChannelDescribeCommand) Run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	name := args[0]

	conn, err := newGRPCConnection(ctx, c.opts.Address, c.opts.Insecure)
	if err != nil {
		return err
	}
	defer conn.Close()

	cc := channelz.NewClient(conn)
	cc.DescribeChannel(ctx, name)

	return nil
}
