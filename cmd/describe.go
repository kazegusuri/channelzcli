package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kazegusuri/channelzcli/channelz"
	"github.com/spf13/cobra"
)

type DescribeCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
	addr string
	long bool
	full bool
}

func NewDescribeCommand(opts *GlobalOptions) *DescribeCommand {
	c := &DescribeCommand{
		cmd: &cobra.Command{
			Use:          "describe (channel|server|serversocket) (NAME|ID)",
			Short:        "describe (channel|server|serversocket) (NAME|ID)",
			Aliases:      []string{"desc"},
			Args:         cobra.ExactArgs(2),
			SilenceUsage: true,
		},
		opts: opts,
	}
	c.cmd.RunE = c.Run
	return c
}

func (c *DescribeCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *DescribeCommand) Run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	typ := args[0]
	name := args[1]

	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := newGRPCConnection(dialCtx, c.opts.Address, c.opts.Insecure)
	if err != nil {
		return fmt.Errorf("failed to connect %v: %v", c.opts.Address, err)
	}
	defer conn.Close()

	cc := channelz.NewClient(conn, c.opts.Output)

	switch typ {
	case "channel":
		cc.DescribeChannel(ctx, name)
	case "server":
		cc.DescribeServer(ctx, name)
	case "serversocket":
		cc.DescribeServerSocket(ctx, name)
	default:
		c.cmd.Usage()
		os.Exit(1)
	}

	return nil
}
