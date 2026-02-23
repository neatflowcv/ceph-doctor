package cephdoctor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/neatflowcv/ceph-doctor/internal/infrastructure/fscluster"
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

func (c *clusterRegisterCmd) Run(repo domain.ClusterRepository) error {
	slog.Info("cluster register", "name", c.Name, "host", c.Host)

	cluster, err := domain.NewCluster(c.Name, c.Key, []string{c.Host})
	if err != nil {
		return fmt.Errorf("new cluster: %w", err)
	}

	err = repo.CreateCluster(context.Background(), cluster)
	if errors.Is(err, domain.ErrClusterAlreadyExists) {
		return exitCodeError(1)
	}

	if err != nil {
		return fmt.Errorf("create cluster: %w", err)
	}

	return nil
}

func (c *clusterUnregisterCmd) Run(repo domain.ClusterRepository) error {
	slog.Info("cluster unregister", "name", c.Name)

	err := repo.DeleteCluster(context.Background(), c.Name)
	if errors.Is(err, domain.ErrClusterNotFound) {
		return exitCodeError(1)
	}

	if err != nil {
		return fmt.Errorf("delete cluster: %w", err)
	}

	return nil
}

func (c *clusterListCmd) Run(repo domain.ClusterRepository) error {
	clusters, err := repo.ListClusters(context.Background())
	if err != nil {
		return fmt.Errorf("list clusters: %w", err)
	}

	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.AppendHeader(table.Row{"Name", "Hosts"})

	for _, cluster := range clusters {
		tableWriter.AppendRow(table.Row{
			cluster.Name(),
			strings.Join(cluster.Hosts(), ","),
		})
	}

	tableWriter.Render()

	return nil
}

func Execute() error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	repo, err := fscluster.NewRepository("")
	if err != nil {
		return fmt.Errorf("new repository: %w", err)
	}

	var command cli

	parser, err := kong.New(
		&command,
		kong.Name("cephdoctor"),
		kong.Description("Ceph Doctor CLI"),
		kong.BindTo(repo, (*domain.ClusterRepository)(nil)),
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
