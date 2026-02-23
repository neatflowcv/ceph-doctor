package domain

import (
	"context"
	"errors"
)

var (
	ErrClusterAlreadyExists = errors.New("cluster already exists")
	ErrClusterNotFound      = errors.New("cluster not found")
)

type ClusterRepository interface {
	CreateCluster(ctx context.Context, cluster *Cluster) error
	UpdateCluster(ctx context.Context, cluster *Cluster) error
	ListClusters(ctx context.Context) ([]*Cluster, error)
	DeleteCluster(ctx context.Context, name string) error
}
