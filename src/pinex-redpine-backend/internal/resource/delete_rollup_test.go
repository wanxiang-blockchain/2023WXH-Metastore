package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteUhost(t *testing.T) {
	err := DeleteHost("uhost-n56u63a16yo", "cn-bj2", "cn-bj2-05", "", params.token)
	require.Empty(t, err)
}

func TestDeleteDB(t *testing.T) {
	err := DeleteDB("upgsql-n0s7vfvrlnq", "cn-bj2", "cn-bj2-05", "", params.token)
	require.Empty(t, err)
}
