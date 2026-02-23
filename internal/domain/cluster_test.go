package domain_test

import (
	"testing"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewCluster_PropagatesHostValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty host", func(t *testing.T) {
		t.Parallel()

		// Arrange
		name := "cluster-a"
		key := "secret"
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
		name := "cluster-a"
		key := "secret"
		hosts := []string{"10.0.0.1", "10.0.0.1:3300"}

		// Act
		cluster, err := domain.NewCluster(name, key, hosts)

		// Assert
		require.ErrorIs(t, err, domain.ErrDuplicateHost)
		require.Nil(t, cluster)
	})
}
