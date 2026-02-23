package cephdoctor

import (
	"context"
	"fmt"
	"os"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
)

func (c *clusterListCmd) Run(repo domain.ClusterRepository) error {
	clusters, err := repo.ListClusters(context.Background())
	if err != nil {
		return fmt.Errorf("list clusters: %w", err)
	}

	renderClusterTable(os.Stdout, clusters)

	return nil
}
