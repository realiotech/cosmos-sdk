package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestDelegate(t *testing.T) {
	// setup the app
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// get pool for checks later
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	moduleBalance := app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), app.StakingKeeper.BondDenom(ctx))

	// accounts
	delAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	validators := app.StakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, len(validators), 1)


	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgDelegate
	}{
		{
			Name:      "invalid coin",
			ExceptErr: true,
			req: types.MsgDelegate{
				DelegatorAddress: delAddrs[0].String(),
				ValidatorAddress: validators[0].OperatorAddress,
				Amount:           sdk.NewCoin("foo_coin", sdk.NewInt(4)),
			},
		},
		{
			Name:      "validator not exists",
			ExceptErr: true,
			req: types.MsgDelegate{
				DelegatorAddress: delAddrs[0].String(),
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4)),
			},
		},
		{
			Name:      "invalid delegator address",
			ExceptErr: true,
			req: types.MsgDelegate{
				DelegatorAddress: "invalid_delegator_addrtess",
				ValidatorAddress: validators[0].OperatorAddress,
				Amount: sdk.NewCoin("foo_coin", sdk.NewInt(4)),

			},
		},

		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgDelegate{
				DelegatorAddress: delAddrs[0].String(),
				ValidatorAddress: validators[0].OperatorAddress,
				Amount:sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(5)),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.Delegate(ctx, &testCase.req)
			if testCase.ExceptErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				balanceForBondedPool := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(bondedPool.GetAddress()), bondDenom)
				require.Equal(t, balanceForBondedPool, moduleBalance.Add(testCase.req.Amount))
			}
		})
	}
}

func TestRedelegate(t *testing.T) {
	// setup the app
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 30)
	startCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, startTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), startCoins))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// accounts
	delAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	addrVals := simapp.ConvertAddrsToValAddrs(delAddrs)
	validators := app.StakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, len(validators), 1)

	//bring in new validator for redelegation
	// create a validator with a self-delegation
	validator := teststaking.NewValidator(t, addrVals[1], PKs[0])
	valTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)

	validators = app.StakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, len(validators), 2)

	// testing unhappy path here as the keeper tests has coverage on success
	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgBeginRedelegate
	}{
		{
			Name:      "invalid coin",
			ExceptErr: true,
			req: types.MsgBeginRedelegate{
				DelegatorAddress: delAddrs[0].String(),
				ValidatorSrcAddress: validators[0].OperatorAddress,
				ValidatorDstAddress: validators[1].OperatorAddress,
				Amount:           sdk.NewCoin("foo_coin", sdk.NewInt(4)),
			},
		},
		{
			Name:      "invalid src validator addr",
			ExceptErr: true,
			req: types.MsgBeginRedelegate{
				DelegatorAddress: delAddrs[0].String(),
				ValidatorSrcAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				ValidatorDstAddress: validators[1].OperatorAddress,
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4)),
			},
		},
		{
			Name:      "dst validator not exists",
			ExceptErr: true,
			req: types.MsgBeginRedelegate{
				DelegatorAddress: delAddrs[0].String(),
				ValidatorSrcAddress: validators[0].OperatorAddress,
				ValidatorDstAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4)),
			},
		},
		{
			Name:      "invalid delegator address",
			ExceptErr: true,
			req: types.MsgBeginRedelegate{
				DelegatorAddress:    "invalid_delegator_addrtess",
				ValidatorSrcAddress: validators[0].OperatorAddress,
				ValidatorDstAddress: validators[1].OperatorAddress,
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4)),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.BeginRedelegate(ctx, &testCase.req)
			if testCase.ExceptErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCancelUnbondingDelegation(t *testing.T) {
	// setup the app
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// set the not bonded pool module account
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 5)

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	moduleBalance := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// accounts
	delAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	validators := app.StakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, len(validators), 1)

	validatorAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)
	delegatorAddr := delAddrs[0]

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(app.StakingKeeper.BondDenom(ctx), 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount,
	)

	// set and retrieve a record
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)

	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgCancelUnbondingDelegation
	}{
		{
			Name:      "invalid height",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid coin",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin("dump_coin", sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "validator not exists",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid delegator address",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid_delegator_addrtess",
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid amount",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Add(sdk.NewInt64Coin(bondDenom, 10)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1))),
				CreationHeight:   10,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(ctx, &testCase.req)
			if testCase.ExceptErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				balanceForNotBondedPool := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(notBondedPool.GetAddress()), bondDenom)
				require.Equal(t, balanceForNotBondedPool, moduleBalance.Sub(testCase.req.Amount))
				moduleBalance = moduleBalance.Sub(testCase.req.Amount)
			}
		})
	}
}
