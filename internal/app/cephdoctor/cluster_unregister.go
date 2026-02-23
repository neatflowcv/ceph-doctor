package cephdoctor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
)

func (c *clusterUnregisterCmd) Run(repo domain.ClusterRepository) error {
	slog.Info("cluster unregister", "name", c.Name)

	err := repo.DeleteCluster(context.Background(), c.Name)
	if err != nil {
		return fmt.Errorf("delete cluster: %w", err)
	}

	return nil
}
