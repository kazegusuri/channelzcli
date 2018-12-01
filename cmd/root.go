package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	Verbose  bool
	Address  string
	Insecure bool
	Input    io.Reader
	Output   io.Writer
}

type RootCommand struct {
	cmd  *cobra.Command
	opts *GlobalOptions
}

func NewRootCommand(r io.Reader, w io.Writer) *RootCommand {
	c := &RootCommand{
		cmd: &cobra.Command{
			Use:   "channelzcli",
			Short: "cli for gRPC channelz",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Help()
			},
		},
		opts: &GlobalOptions{
			Input:  r,
			Output: w,
		},
	}
	c.cmd.PersistentFlags().BoolVarP(&c.opts.Verbose, "verbose", "v", false, "verbose output")
	c.cmd.PersistentFlags().BoolVarP(&c.opts.Insecure, "insecure", "k", false, "with insecure")
	c.cmd.PersistentFlags().StringVar(&c.opts.Address, "addr", "", "address to gRPC server")
	c.cmd.AddCommand(NewListCommand(c.opts).Command())
	c.cmd.AddCommand(NewTreeCommand(c.opts).Command())
	c.cmd.AddCommand(NewDescribeCommand(c.opts).Command())
	return c
}

func (c *RootCommand) Command() *cobra.Command {
	return c.cmd
}
