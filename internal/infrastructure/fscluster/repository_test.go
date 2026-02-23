package fscluster_test

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/neatflowcv/ceph-doctor/internal/infrastructure/fscluster"
	"github.com/stretchr/testify/require"
)

func TestNewRepository_UsesXDGStateHome(t *testing.T) {
	// Arrange
	xdgStateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", xdgStateHome)
	t.Setenv("HOME", "/unused-home")

	// Act
	repo, err := fscluster.NewRepository("")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.DirExists(t, filepath.Join(xdgStateHome, "ceph-doctor", "clusters"))
}

func TestNewRepository_UsesHomeFallback(t *testing.T) {
	// Arrange
	home := t.TempDir()
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", home)

	// Act
	repo, err := fscluster.NewRepository("")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.DirExists(t, filepath.Join(home, ".local", "state", "ceph-doctor", "clusters"))
}

func TestNewRepository_ReturnsErrorWhenHomeMissing(t *testing.T) {
	// Arrange
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", "")

	// Act
	repo, err := fscluster.NewRepository("")

	// Assert
	require.Error(t, err)
	require.Nil(t, repo)
	require.ErrorContains(t, err, "HOME is not set")
}

func TestRepository_CreateCluster_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	root := t.TempDir()
	repo, err := fscluster.NewRepository(root)
	require.NoError(t, err)

	cluster, err := domain.NewCluster("cluster-a", "secret", []string{"10.0.0.1"})
	require.NoError(t, err)

	// Act
	err = repo.CreateCluster(t.Context(), cluster)

	// Assert
	require.NoError(t, err)

	targetFile := filepath.Join(root, "clusters", "cluster-a.json")
	stat, err := os.Stat(targetFile)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), stat.Mode().Perm())
}

func TestRepository_CreateCluster_Duplicate(t *testing.T) {
	t.Parallel()

	// Arrange
	root := t.TempDir()
	repo, err := fscluster.NewRepository(root)
	require.NoError(t, err)

	cluster, err := domain.NewCluster("cluster-a", "secret", []string{"10.0.0.1"})
	require.NoError(t, err)
	require.NoError(t, repo.CreateCluster(t.Context(), cluster))

	// Act
	duplicateErr := repo.CreateCluster(t.Context(), cluster)

	// Assert
	require.ErrorIs(t, duplicateErr, domain.ErrClusterAlreadyExists)
}

func TestRepository_UpdateCluster(t *testing.T) {
	t.Parallel()

	// Arrange
	root := t.TempDir()
	repo, err := fscluster.NewRepository(root)
	require.NoError(t, err)

	initial, err := domain.NewCluster("cluster-a", "old-secret", []string{"10.0.0.1"})
	require.NoError(t, err)
	require.NoError(t, repo.CreateCluster(t.Context(), initial))

	updated, err := domain.NewCluster("cluster-a", "new-secret", []string{"10.0.0.2"})
	require.NoError(t, err)

	// Act
	require.NoError(t, repo.UpdateCluster(t.Context(), updated))

	// Assert
	clusters, err := repo.ListClusters(t.Context())
	require.NoError(t, err)
	require.Len(t, clusters, 1)
	require.Equal(t, "cluster-a", clusters[0].Name())
	require.Equal(t, "new-secret", clusters[0].Key())
	require.Equal(t, []string{"10.0.0.2:3300"}, clusters[0].Hosts())
}

func TestRepository_UpdateCluster_NotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	repo, err := fscluster.NewRepository(t.TempDir())
	require.NoError(t, err)

	cluster, err := domain.NewCluster("cluster-a", "secret", []string{"10.0.0.1"})
	require.NoError(t, err)

	// Act
	updateErr := repo.UpdateCluster(t.Context(), cluster)

	// Assert
	require.ErrorIs(t, updateErr, domain.ErrClusterNotFound)
}

func TestRepository_ListClusters_SortsByName(t *testing.T) {
	t.Parallel()

	// Arrange
	repo, err := fscluster.NewRepository(t.TempDir())
	require.NoError(t, err)

	alpha, err := domain.NewCluster("alpha", "secret-a", []string{"10.0.0.2"})
	require.NoError(t, err)
	zeta, err := domain.NewCluster("zeta", "secret-z", []string{"10.0.0.1"})
	require.NoError(t, err)

	require.NoError(t, repo.CreateCluster(t.Context(), zeta))
	require.NoError(t, repo.CreateCluster(t.Context(), alpha))

	// Act
	clusters, err := repo.ListClusters(t.Context())

	// Assert
	require.NoError(t, err)
	require.Len(t, clusters, 2)
	require.Equal(t, "alpha", clusters[0].Name())
	require.Equal(t, "zeta", clusters[1].Name())
}

func TestRepository_ListClusters_InvalidFile(t *testing.T) {
	t.Parallel()

	// Arrange
	root := t.TempDir()
	repo, err := fscluster.NewRepository(root)
	require.NoError(t, err)

	brokenFilePath := filepath.Join(root, "clusters", "broken.json")
	require.NoError(t, os.WriteFile(brokenFilePath, []byte("{invalid"), 0o600))

	// Act
	clusters, listErr := repo.ListClusters(t.Context())

	// Assert
	require.Error(t, listErr)
	require.Nil(t, clusters)
	require.ErrorContains(t, listErr, "decode cluster file")
}

func TestRepository_DeleteCluster_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	root := t.TempDir()
	repo, err := fscluster.NewRepository(root)
	require.NoError(t, err)

	cluster, err := domain.NewCluster("cluster-a", "secret", []string{"10.0.0.1"})
	require.NoError(t, err)
	require.NoError(t, repo.CreateCluster(t.Context(), cluster))

	targetFile := filepath.Join(root, "clusters", "cluster-a.json")

	// Act
	err = repo.DeleteCluster(t.Context(), "cluster-a")

	// Assert
	require.NoError(t, err)
	_, err = os.Stat(targetFile)
	require.Error(t, err)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestRepository_DeleteCluster_NotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	repo, err := fscluster.NewRepository(t.TempDir())
	require.NoError(t, err)

	// Act
	deleteErr := repo.DeleteCluster(t.Context(), "cluster-a")

	// Assert
	require.ErrorIs(t, deleteErr, domain.ErrClusterNotFound)
}

func TestRepository_CreateCluster_EncodesNameInFileName(t *testing.T) {
	t.Parallel()

	// Arrange
	root := t.TempDir()
	repo, err := fscluster.NewRepository(root)
	require.NoError(t, err)

	clusterName := "프로덕션/primary cluster"
	cluster, err := domain.NewCluster(clusterName, "secret", []string{"10.0.0.1"})
	require.NoError(t, err)

	// Act
	err = repo.CreateCluster(t.Context(), cluster)

	// Assert
	require.NoError(t, err)

	encodedName := url.PathEscape(clusterName) + ".json"
	_, err = os.Stat(filepath.Join(root, "clusters", encodedName))
	require.NoError(t, err)
}

func TestRepository_ListClusters_DecodesEscapedFileName(t *testing.T) {
	t.Parallel()

	// Arrange
	root := t.TempDir()
	repo, err := fscluster.NewRepository(root)
	require.NoError(t, err)

	clusterName := "프로덕션/primary cluster"
	record := map[string]any{
		"name":  clusterName,
		"key":   "secret",
		"hosts": []string{"10.0.0.1"},
	}
	payload, err := json.Marshal(record)
	require.NoError(t, err)

	encodedName := url.PathEscape(clusterName) + ".json"
	require.NoError(t, os.WriteFile(filepath.Join(root, "clusters", encodedName), payload, 0o600))

	// Act
	clusters, err := repo.ListClusters(t.Context())

	// Assert
	require.NoError(t, err)
	require.Len(t, clusters, 1)
	require.Equal(t, clusterName, clusters[0].Name())
}

func TestRepository_RespectsCanceledContext(t *testing.T) {
	t.Parallel()

	// Arrange
	repo, err := fscluster.NewRepository(t.TempDir())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	cluster, err := domain.NewCluster("cluster-a", "secret", []string{"10.0.0.1"})
	require.NoError(t, err)

	// Act
	createErr := repo.CreateCluster(ctx, cluster)

	// Assert
	require.Error(t, createErr)
	require.ErrorIs(t, createErr, context.Canceled)
}
