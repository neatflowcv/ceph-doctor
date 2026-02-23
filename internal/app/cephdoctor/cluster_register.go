package cephdoctor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
)

func (c *clusterRegisterCmd) Run(repo domain.ClusterRepository) error {
	slog.Info("cluster register", "name", c.Name, "host", c.Host)

	cluster, err := domain.NewCluster(c.Name, c.Key, []string{c.Host})
	if err != nil {
		return fmt.Errorf("new cluster: %w", err)
	}

	err = repo.CreateCluster(context.Background(), cluster)
	if err != nil {
		return fmt.Errorf("create cluster: %w", err)
	}

	return nil
}
