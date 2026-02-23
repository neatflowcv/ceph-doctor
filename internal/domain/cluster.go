package domain

import "errors"

// Cluster represents a registered Ceph cluster.
type Cluster struct {
	name  string
	key   string
	hosts *Hosts
}

var (
	ErrEmptyClusterName  = errors.New("cluster name is empty")
	ErrEmptyClusterKey   = errors.New("cluster key is empty")
	ErrEmptyClusterHosts = ErrEmptyHosts
)

func NewCluster(name, key string, hosts []string) (*Cluster, error) {
	if name == "" {
		return nil, ErrEmptyClusterName
	}

	if key == "" {
		return nil, ErrEmptyClusterKey
	}

	clusterHosts, err := NewHosts(hosts)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		name:  name,
		key:   key,
		hosts: clusterHosts,
	}, nil
}

func (c *Cluster) Name() string {
	return c.name
}

func (c *Cluster) Key() string {
	return c.key
}

func (c *Cluster) Hosts() []string {
	return c.hosts.Values()
}
