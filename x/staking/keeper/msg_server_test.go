package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Prevent strconv unused error
var (
	pk1 = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.AccAddress(pk1.Address())
	valAddr1 = sdk.ValAddress(pk1.Address())
	pk2 = ed25519.GenPrivKey().PubKey()
	addr2 = sdk.AccAddress(pk2.Address())
	valAddr2 = sdk.ValAddress(pk2.Address())
	coinRio = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)
	coinRst = sdk.NewInt64Coin("rst", 1000)
	commission1 = types.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func (suite *KeeperTestSuite) TestMsgServerCreateValidator() {
	suite.SetupTest()

	initAccountWithCoins(suite.app, suite.ctx, addr1, sdk.NewCoins(coinRio))
	srv := keeper.NewMsgServerImpl(suite.app.StakingKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	expected, _ := types.NewMsgCreateValidator(valAddr1, pk1, coinRio, types.Description{}, commission1, sdk.OneInt())

	_, err := srv.CreateValidator(wctx, expected)
	suite.Require().NoError(err)

	v, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr1)
	suite.Require().True(found)
	suite.Require().Equal(v.OperatorAddress, expected.ValidatorAddress)
	suite.Require().Equal(v.BondDenom, coinRio.Denom)
	suite.Require().Equal(v.Commission.CommissionRates, expected.Commission)
}

func (suite *KeeperTestSuite) TestMsgServerCreateValidatorInvalidDenom() {
	suite.SetupTest()

	initAccountWithCoins(suite.app, suite.ctx, addr1, sdk.NewCoins(coinRio))
	srv := keeper.NewMsgServerImpl(suite.app.StakingKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	unsupportedDenom := sdk.NewCoin("bitcoin", sdk.NewInt(1))
	expected, _ := types.NewMsgCreateValidator(valAddr1, pk1, unsupportedDenom, types.Description{}, commission1, sdk.OneInt())

	_, err := srv.CreateValidator(wctx, expected)
	suite.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)
}

func (suite *KeeperTestSuite) TestMsgServerCreateValidatorMultipleDenoms() {
	suite.SetupTest()

	params := suite.app.StakingKeeper.GetParams(suite.ctx)
	params.BondDenom = "rio,rst"
	suite.app.StakingKeeper.SetParams(suite.ctx, params)

	// test rst, supported denom
	initAccountWithCoins(suite.app, suite.ctx, addr2, sdk.NewCoins(coinRio, coinRst))
	initAccountWithCoins(suite.app, suite.ctx, addr1, sdk.NewCoins(coinRio, coinRst))
	srv := keeper.NewMsgServerImpl(suite.app.StakingKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	expected, _ := types.NewMsgCreateValidator(valAddr2, pk2, coinRst, types.Description{}, commission1, sdk.OneInt())

	_, err := srv.CreateValidator(wctx, expected)
	suite.Require().NoError(err)

	v, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr2)
	suite.Require().True(found)
	suite.Require().Equal(v.OperatorAddress, expected.ValidatorAddress)
	suite.Require().Equal(v.BondDenom, coinRst.Denom)
	suite.Require().Equal(v.Commission.CommissionRates, expected.Commission)

	// test bitcoin, supported denom
	unsupportedDenom := sdk.NewCoin("bitcoin", sdk.NewInt(1))
	expected, _ = types.NewMsgCreateValidator(valAddr1, pk1, unsupportedDenom, types.Description{}, commission1, sdk.OneInt())
	_, err = srv.CreateValidator(wctx, expected)
	suite.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)

	// test rio, supported denom
	expected, _ = types.NewMsgCreateValidator(valAddr1, pk1, coinRio, types.Description{}, commission1, sdk.OneInt())

	_, err = srv.CreateValidator(wctx, expected)
	suite.Require().NoError(err)

	v, found = suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr1)
	suite.Require().True(found)
	suite.Require().Equal(v.OperatorAddress, expected.ValidatorAddress)
	suite.Require().Equal(v.BondDenom, coinRio.Denom)
	suite.Require().Equal(v.Commission.CommissionRates, expected.Commission)
}