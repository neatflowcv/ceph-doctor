package domain_test

import (
	"testing"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/stretchr/testify/require"
)

const (
	testClusterName = "cluster-a"
	testClusterKey  = "secret"
)

func TestNewCluster_PropagatesHostValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty host", func(t *testing.T) {
		t.Parallel()

		// Arrange
		name := testClusterName
		key := testClusterKey
		hosts := []string{"", "10.0.0.1"}

		// Act
		cluster, err := domain.NewCluster(name, key, hosts)

		// Assert
		require.ErrorIs(t, err, domain.ErrEmptyHost)
		require.Nil(t, cluster)
	})

	t.Run("duplicate hosts", func(t *testing.T) {
		t.Parallel()

		// Arrange
		name := testClusterName
		key := testClusterKey
		hosts := []string{"10.0.0.1", "10.0.0.1:3300"}

		// Act
		cluster, err := domain.NewCluster(name, key, hosts)

		// Assert
		require.ErrorIs(t, err, domain.ErrDuplicateHost)
		require.Nil(t, cluster)
	})
}

func TestNewCluster_ValidatesNameAndKeyErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()

		// Arrange
		name := ""
		key := testClusterKey
		hosts := []string{"10.0.0.1"}

		// Act
		cluster, err := domain.NewCluster(name, key, hosts)

		// Assert
		require.ErrorIs(t, err, domain.ErrEmptyClusterName)
		require.Nil(t, cluster)
	})

	t.Run("empty key", func(t *testing.T) {
		t.Parallel()

		// Arrange
		name := testClusterName
		key := ""
		hosts := []string{"10.0.0.1"}

		// Act
		cluster, err := domain.NewCluster(name, key, hosts)

		// Assert
		require.ErrorIs(t, err, domain.ErrEmptyClusterKey)
		require.Nil(t, cluster)
	})
}
