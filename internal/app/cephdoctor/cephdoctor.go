package cephdoctor

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jedib0t/go-pretty/v6/table"
)

type cli struct {
	Cluster clusterCmd `kong:"cmd,help='Cluster operations.'"`
}

type clusterCmd struct {
	Register   clusterRegisterCmd   `kong:"cmd,help='Register a cluster.'"`
	Unregister clusterUnregisterCmd `kong:"cmd,help='Unregister a cluster.'"`
	List       clusterListCmd       `kong:"cmd,help='List clusters.'"`
}

type clusterRegisterCmd struct {
	Name string `kong:"arg,help='Cluster name.'"`
	Host string `kong:"arg,help='Host in host[:port] format.'"`
	Key  string `kong:"arg,help='Access key.'"`
}

type clusterUnregisterCmd struct {
	Name string `kong:"arg,help='Cluster name.'"`
}

type clusterListCmd struct{}

type exitCodeError int

func (e exitCodeError) Error() string { return "" }

func (e exitCodeError) ExitCode() int { return int(e) }

func (c *clusterRegisterCmd) Run() error {
	slog.Info("cluster register", "name", c.Name, "host", c.Host)

	return exitCodeError(1)
}

func (c *clusterUnregisterCmd) Run() error {
	slog.Info("cluster unregister", "name", c.Name)

	return exitCodeError(1)
}

func (c *clusterListCmd) Run() error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Host"})
	t.Render()

	return exitCodeError(1)
}

func Execute() error {
	var command cli

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	parser, err := kong.New(
		&command,
		kong.Name("cephdoctor"),
		kong.Description("Ceph Doctor CLI"),
	)
	if err != nil {
		return fmt.Errorf("create parser: %w", err)
	}

	if len(os.Args) == 1 {
		ctx, _ := kong.Trace(parser, []string{})
		_ = ctx.PrintUsage(false)

		return exitCodeError(1)
	}

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	err = ctx.Run()
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	return nil
}

func ExitCode(err error) (int, bool) {
	exit, ok := err.(interface{ ExitCode() int })
	if !ok {
		return 0, false
	}

	return exit.ExitCode(), true
}
