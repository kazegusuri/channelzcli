package cmd

import (
	"context"
	"os"

	"github.com/kazegusuri/channelzcli/channelz"
	"github.com/spf13/cobra"
)

type TreeCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewTreeCommand(opts *GlobalOptions) *TreeCommand {
	c := &TreeCommand{
		cmd: &cobra.Command{
			Use:          "tree (channel|server)",
			Short:        "tree (channel|server)",
			Args:         cobra.ExactArgs(1),
			SilenceUsage: true,
		},
		opts: opts,
	}
	c.cmd.RunE = c.Run
	return c
}

func (c *TreeCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *TreeCommand) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	typ := args[0]

	conn, err := newGRPCConnection(ctx, c.opts.Address, c.opts.Insecure)
	if err != nil {
		return err
	}
	defer conn.Close()

	cc := channelz.NewClient(conn, c.opts.Output)

	switch typ {
	case "channel":
		cc.TreeTopChannels(ctx)
	case "server":
		cc.TreeServers(ctx)
	default:
		c.cmd.Usage()
		os.Exit(1)
	}

	return nil
}
