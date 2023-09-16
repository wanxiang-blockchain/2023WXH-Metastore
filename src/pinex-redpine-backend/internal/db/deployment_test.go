package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const testAddressId = "pinex_123"

func TestListDeployments(t *testing.T) {
	deps, err := ListDeployment(nil, testAddressId)
	require.Empty(t, err)
	require.NotEmpty(t, deps)
}

func TestDelete(t *testing.T) {
	deps, err := ListDeployment(nil, testAddressId)
	require.Empty(t, err)
	require.NotEmpty(t, deps)
	for _, dep := range deps {
		err := DeleteDeployment(dep.ID)
		require.Empty(t, err)
	}
}
