package keeper

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetBondedPool returns the bonded tokens pool's module account
func (k Keeper) GetBondedPool(ctx sdk.Context) (bondedPool authtypes.ModuleAccountI) {
	return k.authKeeper.GetModuleAccount(ctx, types.BondedPoolName)
}

// GetNotBondedPool returns the not bonded tokens pool's module account
func (k Keeper) GetNotBondedPool(ctx sdk.Context) (notBondedPool authtypes.ModuleAccountI) {
	return k.authKeeper.GetModuleAccount(ctx, types.NotBondedPoolName)
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx sdk.Context, coins sdk.Coins) {
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, types.NotBondedPoolName, coins); err != nil {
		panic(err)
	}
}

// notBondedTokensToBonded transfers coins from the not bonded to the bonded pool within staking
func (k Keeper) notBondedTokensToBonded(ctx sdk.Context, coins sdk.Coins) {
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.NotBondedPoolName, types.BondedPoolName, coins); err != nil {
		panic(err)
	}
}

// burnBondedTokens removes coins from the bonded pool module account
func (k Keeper) burnBondedTokens(ctx sdk.Context, coin sdk.Coin) error {
	if !coin.IsPositive() {
		// skip as no coins need to be burned
		return nil
	}

	coins := sdk.NewCoins(coin)

	return k.bankKeeper.BurnCoins(ctx, types.BondedPoolName, coins)
}

// burnNotBondedTokens removes coins from the not bonded pool module account
func (k Keeper) burnNotBondedTokens(ctx sdk.Context, coin sdk.Coin) error {
	if !coin.IsPositive() {
		// skip as no coins need to be burned
		return nil
	}

	coins := sdk.NewCoins(coin)

	return k.bankKeeper.BurnCoins(ctx, types.NotBondedPoolName, coins)
}

// TotalBondedTokens total staking tokens supply which is bonded
func (k Keeper) TotalBondedTokens(ctx sdk.Context) math.Int {
	bondedPool := k.GetBondedPool(ctx)
	bondableTokens := k.BondDenomSlice(ctx)
	balanceSum := sdk.ZeroInt()
	for _, x := range bondableTokens {
		balance := k.bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), x)
		balanceSum = balanceSum.Add(balance.Amount)
	}
	return balanceSum
}

// StakingTokenSupply staking tokens from the total supply
func (k Keeper) StakingTokenSupply(ctx sdk.Context) math.Int {
	bondableTokens := k.BondDenomSlice(ctx)
	totalSum := sdk.ZeroInt()
	for _, x := range bondableTokens {
		sum := k.bankKeeper.GetSupply(ctx, x)
		totalSum = totalSum.Add(sum.Amount)
	}
	return totalSum
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	stakeSupply := k.StakingTokenSupply(ctx)
	if stakeSupply.IsPositive() {
		return sdk.NewDecFromInt(k.TotalBondedTokens(ctx)).QuoInt(stakeSupply)
	}

	return sdk.ZeroDec()
}
