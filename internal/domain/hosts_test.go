package domain_test

import (
	"testing"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewHosts_EmptyHosts(t *testing.T) {
	t.Parallel()

	// Arrange
	input := []string(nil)

	// Act
	hosts, err := domain.NewHosts(input)

	// Assert
	require.ErrorIs(t, err, domain.ErrEmptyHosts)
	require.Nil(t, hosts)
}

func TestNewHosts_DefaultPort(t *testing.T) {
	t.Parallel()

	// Arrange
	input := []string{"10.0.0.1", "10.0.0.2:4400"}
	want := []string{"10.0.0.1:3300", "10.0.0.2:4400"}

	// Act
	hosts, err := domain.NewHosts(input)

	// Assert
	require.NoError(t, err)
	require.Equal(t, want, hosts.Values())
}

func TestNewHosts_DuplicateAfterNormalization(t *testing.T) {
	t.Parallel()

	// Arrange
	input := []string{"10.0.0.1", "10.0.0.1:3300"}

	// Act
	hosts, err := domain.NewHosts(input)

	// Assert
	require.ErrorIs(t, err, domain.ErrDuplicateHost)
	require.Nil(t, hosts)
}

func TestNewHosts_EmptyHost(t *testing.T) {
	t.Parallel()

	// Arrange
	input := []string{"", "10.0.0.1"}

	// Act
	hosts, err := domain.NewHosts(input)

	// Assert
	require.ErrorIs(t, err, domain.ErrEmptyHost)
	require.Nil(t, hosts)
}

func TestHosts_ValuesReturnsCopy(t *testing.T) {
	t.Parallel()

	// Arrange
	hosts, err := domain.NewHosts([]string{"10.0.0.1"})
	require.NoError(t, err)

	want := []string{"10.0.0.1:3300"}

	// Act
	values := hosts.Values()
	values[0] = "changed:3300"

	// Assert
	require.Equal(t, want, hosts.Values())
}
