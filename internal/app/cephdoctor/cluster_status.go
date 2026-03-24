package cephdoctor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
)

var errClusterStatusFailed = errors.New("one or more cluster status checks failed")

func (c *clusterStatusCmd) Run(repo domain.ClusterRepository, cephClient domain.CephClient) error {
	slog.Info("cluster status")

	return runClusterStatus(context.Background(), os.Stdout, repo, cephClient)
}

func runClusterStatus(
	ctx context.Context,
	writer io.Writer,
	repo domain.ClusterRepository,
	cephClient domain.CephClient,
) error {
	clusters, err := repo.ListClusters(ctx)
	if err != nil {
		return fmt.Errorf("list clusters: %w", err)
	}

	if len(clusters) == 0 {
		_, err = fmt.Fprintln(writer, "No clusters registered.")
		if err != nil {
			return fmt.Errorf("write empty cluster status: %w", err)
		}

		return nil
	}

	results := make([]clusterStatusView, 0, len(clusters))
	for _, cluster := range clusters {
		status, statusErr := cephClient.Status(ctx, cluster)
		results = append(results, clusterStatusView{
			cluster: cluster,
			status:  status,
			err:     statusErr,
		})
	}

	err = renderClusterStatusResults(writer, results)
	if err != nil {
		return fmt.Errorf("render cluster status: %w", err)
	}

	for _, result := range results {
		if result.err != nil {
			return errClusterStatusFailed
		}
	}

	return nil
}

func renderClusterStatusResults(writer io.Writer, results []clusterStatusView) error {
	for i, result := range results {
		err := renderClusterStatusResult(writer, i, result)
		if err != nil {
			return fmt.Errorf("render cluster status result: %w", err)
		}
	}

	return nil
}

func renderClusterStatusResult(
	writer io.Writer,
	index int,
	result clusterStatusView,
) error {
	err := writeStatusHeader(writer, index, result.cluster)
	if err != nil {
		return err
	}

	err = writeCephStatusStreams(writer, result.status)
	if err != nil {
		return err
	}

	if result.err != nil {
		_, err = fmt.Fprintf(writer, "[error] %v\n", result.err)
		if err != nil {
			return fmt.Errorf("write status error: %w", err)
		}
	}

	return nil
}

type clusterStatusView struct {
	cluster *domain.Cluster
	status  *domain.CephStatus
	err     error
}

func writeStatusHeader(writer io.Writer, index int, cluster *domain.Cluster) error {
	if index > 0 {
		_, err := fmt.Fprintln(writer)
		if err != nil {
			return fmt.Errorf("write status separator: %w", err)
		}
	}

	_, err := fmt.Fprintf(
		writer,
		"=== %s (%s) ===\n",
		cluster.Name(),
		strings.Join(cluster.Hosts(), ","),
	)
	if err != nil {
		return fmt.Errorf("write status header: %w", err)
	}

	return nil
}

func writeCephStatusStreams(writer io.Writer, status *domain.CephStatus) error {
	if status == nil {
		return nil
	}

	err := writeStatusStream(writer, status.Stdout)
	if err != nil {
		return err
	}

	if status.Stderr == "" {
		return nil
	}

	_, err = fmt.Fprintln(writer, "[stderr]")
	if err != nil {
		return fmt.Errorf("write stderr label: %w", err)
	}

	err = writeStatusStream(writer, status.Stderr)
	if err != nil {
		return err
	}

	return nil
}

func writeStatusStream(writer io.Writer, content string) error {
	if content == "" {
		return nil
	}

	_, err := io.WriteString(writer, content)
	if err != nil {
		return fmt.Errorf("write status stream: %w", err)
	}

	if !strings.HasSuffix(content, "\n") {
		_, err = fmt.Fprintln(writer)
		if err != nil {
			return fmt.Errorf("terminate status stream: %w", err)
		}
	}

	return nil
}
