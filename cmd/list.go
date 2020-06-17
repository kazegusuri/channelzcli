package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

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
			Use:          "list (channel|server|serversocket)",
			Short:        "list (channel|server|serversocket)",
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	typ := args[0]

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
		cc.ListTopChannels(ctx)
	case "server":
		cc.ListServers(ctx)
	case "serversocket":
		cc.ListServerSockets(ctx)
	default:
		c.cmd.Usage()
		os.Exit(1)
	}

	return nil
}
