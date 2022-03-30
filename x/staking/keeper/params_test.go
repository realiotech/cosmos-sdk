package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestDefaultParams() {
	suite.SetupTest()

	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().True(expParams.Equal(resParams))
}

func (suite *KeeperTestSuite) TestSetParamsSetMultiTokenBondDenom() {
	suite.SetupTest()

	//validate for default sdk bond denom
	expParams := types.DefaultParams()
	suite.Require().Equal(expParams.BondDenom, sdk.DefaultBondDenom)

	expParams.BondDenom = "rio,rst"
	suite.app.StakingKeeper.SetParams(suite.ctx, expParams)

	//validate save
	resParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().Equal(resParams.BondDenom, "rio,rst")
}

func (suite *KeeperTestSuite) TestIsBondDenomSupported() {
	suite.SetupTest()

	//validate for default sdk bond denom
	expParams := types.DefaultParams()
	expParams.BondDenom = "rio,rst"
	suite.app.StakingKeeper.SetParams(suite.ctx, expParams)

	res := suite.app.StakingKeeper.IsBondDenomSupported(suite.ctx, "rio")
	suite.Require().True(res)
	suite.Require().True(suite.app.StakingKeeper.IsBondDenomSupported(suite.ctx, "rio"))
	suite.Require().True(suite.app.StakingKeeper.IsBondDenomSupported(suite.ctx, "rst"))
	suite.Require().False(suite.app.StakingKeeper.IsBondDenomSupported(suite.ctx, "stake"))
}
