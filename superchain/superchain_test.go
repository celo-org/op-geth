package superchain

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGetSuperchain(t *testing.T) {
	mainnet, err := GetSuperchain("mainnet")
	require.NoError(t, err)

	require.Equal(t, "Mainnet", mainnet.Name)
	require.Equal(t, common.HexToAddress("0x1b6dEB2197418075AB314ac4D52Ca1D104a8F663"), mainnet.ProtocolVersionsAddr)
	require.EqualValues(t, 1, mainnet.L1.ChainID)

	_, err = GetSuperchain("not a network")
	require.Error(t, err)
}
