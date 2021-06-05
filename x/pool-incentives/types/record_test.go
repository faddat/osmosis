package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/pool-incentives/types"
)

func TestDistrRecord(t *testing.T) {
	zeroWeight := types.DistrRecord{
		GaugeId: 1,
		Weight:  sdk.NewInt(0),
	}

	require.Error(t, zeroWeight.Validate())

	negativeWeight := types.DistrRecord{
		GaugeId: 1,
		Weight:  sdk.NewInt(-1),
	}

	require.Error(t, negativeWeight.Validate())

	positiveWeight := types.DistrRecord{
		GaugeId: 1,
		Weight:  sdk.NewInt(1),
	}

	require.NoError(t, positiveWeight.Validate())
}
