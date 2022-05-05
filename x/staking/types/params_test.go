package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestParamsEqual(t *testing.T) {
	p1 := types.DefaultParams()
	p2 := types.DefaultParams()

	ok := p1.Equal(p2)
	require.True(t, ok)

	p2.UnbondingTime = 60 * 60 * 24 * 2
	p2.BondDenom = "soup"

	ok = p1.Equal(p2)
	require.False(t, ok)
}

func TestValidateBondDenom(t *testing.T) {
	p1 := types.DefaultParams()
	err := types.ValidateBondDenom(p1.BondDenom)
	require.Nil(t, err)

	p1.BondDenom = "stake,rio"
	err = types.ValidateBondDenom(p1.BondDenom)
	require.Nil(t, err)

	p1.BondDenom = "stake,stake,"
	err = types.ValidateBondDenom(p1.BondDenom)
	require.Error(t, err, "invalid denom: stake,stake,")

	p1.BondDenom = "stake,,stake,"
	err = types.ValidateBondDenom(p1.BondDenom)
	require.Error(t, err, "invalid denom: stake,stake,")

	p1.BondDenom = "stake,,"
	err = types.ValidateBondDenom(p1.BondDenom)
	require.Error(t, err, "invalid denom: stake,stake,")

	p1.BondDenom = ",stake"
	err = types.ValidateBondDenom(p1.BondDenom)
	require.Error(t, err, "invalid denom: stake,stake,")
}
