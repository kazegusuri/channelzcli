package cmd

import (
	"context"
	"os"

	"github.com/kazegusuri/channelzcli/channelz"
	"github.com/spf13/cobra"
)

type ListCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewListCommand(opts *GlobalOptions) *ListCommand {
	c := &ListCommand{
		cmd: &cobra.Command{
			Use:          "list (channel|server)",
			Short:        "list (channel|server)",
			Args:         cobra.ExactArgs(1),
			Aliases:      []string{"ls"},
			SilenceUsage: true,
		},
		opts: opts,
	}
	c.cmd.RunE = c.Run
	return c
}

func (c *ListCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *ListCommand) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	typ := args[0]

	conn, err := newGRPCConnection(ctx, c.opts.Address, c.opts.Insecure)
	if err != nil {
		return err
	}
	defer conn.Close()

	cc := channelz.NewClient(conn)

	switch typ {
	case "channel":
		cc.ListTopChannels(ctx)
	case "server":
		cc.ListServers(ctx)
	default:
		c.cmd.Usage()
		os.Exit(1)
	}

	return nil
}
