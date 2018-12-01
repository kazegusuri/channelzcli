package cmd

import (
	"context"

	"github.com/kazegusuri/channelzcli/channelz"
	"github.com/spf13/cobra"
)

type ServerDescribeCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewServerDescribeCommand(opts *GlobalOptions) *ServerDescribeCommand {
	c := &ServerDescribeCommand{
		cmd: &cobra.Command{
			Use:          "describe",
			Short:        "Describe server",
			Args:         cobra.ExactArgs(1),
			SilenceUsage: true,
		},
		opts: opts,
	}
	c.cmd.RunE = c.Run
	return c
}

func (c *ServerDescribeCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *ServerDescribeCommand) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	conn, err := newGRPCConnection(ctx, c.opts.Address, c.opts.Insecure)
	if err != nil {
		return err
	}
	defer conn.Close()

	cc := channelz.NewClient(conn)
	cc.DescribeServer(ctx, name)

	return nil
}
