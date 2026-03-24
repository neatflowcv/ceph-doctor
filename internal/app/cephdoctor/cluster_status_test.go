//nolint:testpackage // Command rendering is tested through unexported helpers.
package cephdoctor

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/stretchr/testify/require"
)

var errExecFailed = errors.New("exec failed")
var errNotImplemented = errors.New("not implemented")

func TestRunClusterStatus_EmptyRepository(t *testing.T) {
	t.Parallel()

	repo := &fakeClusterRepository{clusters: nil, err: nil}
	cephClient := &fakeCephClient{statuses: nil, errs: nil, called: false, clusters: nil}

	var output bytes.Buffer

	err := runClusterStatus(t.Context(), &output, repo, cephClient)

	require.NoError(t, err)
	require.Equal(t, "No clusters registered.\n", output.String())
	require.False(t, cephClient.called)
}

func TestRunClusterStatus_RendersResults(t *testing.T) {
	t.Parallel()

	alpha, err := domain.NewCluster("alpha", "secret-a", []string{"10.0.0.2"})
	require.NoError(t, err)

	zeta, err := domain.NewCluster("zeta", "secret-z", []string{"10.0.0.1:4400"})
	require.NoError(t, err)

	repo := &fakeClusterRepository{
		clusters: []*domain.Cluster{alpha, zeta},
		err:      nil,
	}
	cephClient := &fakeCephClient{
		statuses: map[*domain.Cluster]*domain.CephStatus{
			alpha: {
				Stdout: "cluster:\n  id: 1\n",
				Stderr: "",
			},
			zeta: {
				Stdout: "cluster:\n  id: 2",
				Stderr: "warn",
			},
		},
		errs:     map[*domain.Cluster]error{},
		called:   false,
		clusters: nil,
	}

	var output bytes.Buffer

	err = runClusterStatus(t.Context(), &output, repo, cephClient)

	require.NoError(t, err)
	require.True(t, cephClient.called)
	require.Equal(t, []*domain.Cluster{alpha, zeta}, cephClient.clusters)
	require.Equal(
		t,
		"=== alpha (10.0.0.2:3300) ===\ncluster:\n  id: 1\n\n" +
			"=== zeta (10.0.0.1:4400) ===\ncluster:\n  id: 2\n[stderr]\nwarn\n",
		output.String(),
	)
}

func TestRunClusterStatus_ReturnsErrorWhenAnyClusterFails(t *testing.T) {
	t.Parallel()

	alpha, err := domain.NewCluster("alpha", "secret-a", []string{"10.0.0.2"})
	require.NoError(t, err)

	zeta, err := domain.NewCluster("zeta", "secret-z", []string{"10.0.0.1"})
	require.NoError(t, err)

	repo := &fakeClusterRepository{
		clusters: []*domain.Cluster{alpha, zeta},
		err:      nil,
	}
	cephClient := &fakeCephClient{
		statuses: map[*domain.Cluster]*domain.CephStatus{
			alpha: {
				Stdout: "ok\n",
				Stderr: "",
			},
			zeta: {
				Stdout: "still-ran\n",
				Stderr: "",
			},
		},
		errs: map[*domain.Cluster]error{
			alpha: errExecFailed,
		},
		called:   false,
		clusters: nil,
	}

	var output bytes.Buffer

	err = runClusterStatus(t.Context(), &output, repo, cephClient)

	require.ErrorIs(t, err, errClusterStatusFailed)
	require.Contains(t, output.String(), "=== alpha (10.0.0.2:3300) ===")
	require.Contains(t, output.String(), "[error] exec failed")
	require.Contains(t, output.String(), "=== zeta (10.0.0.1:3300) ===")
	require.Contains(t, output.String(), "still-ran")
}

type fakeClusterRepository struct {
	clusters []*domain.Cluster
	err      error
}

func (f *fakeClusterRepository) CreateCluster(context.Context, *domain.Cluster) error {
	return errNotImplemented
}

func (f *fakeClusterRepository) UpdateCluster(context.Context, *domain.Cluster) error {
	return errNotImplemented
}

func (f *fakeClusterRepository) ListClusters(context.Context) ([]*domain.Cluster, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.clusters, nil
}

func (f *fakeClusterRepository) DeleteCluster(context.Context, string) error {
	return errNotImplemented
}

type fakeCephClient struct {
	statuses map[*domain.Cluster]*domain.CephStatus
	errs     map[*domain.Cluster]error
	called   bool
	clusters []*domain.Cluster
}

func (f *fakeCephClient) Status(_ context.Context, cluster *domain.Cluster) (*domain.CephStatus, error) {
	f.called = true
	f.clusters = append(f.clusters, cluster)

	return f.statuses[cluster], f.errs[cluster]
}
