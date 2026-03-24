package domain

import "context"

type CephStatus struct {
	Stdout string
	Stderr string
}

type CephClient interface {
	Status(ctx context.Context, cluster *Cluster) (*CephStatus, error)
}
