package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"math"
	"math/big"
)

func SufficientAmount(have, required *big.Int) (remainder *big.Int, neg bool) {
	remainder = have.Sub(have, required)
	return remainder, remainder.Cmp(big.NewInt(0)) < 0
}

// ParseAmount parse float amount to withdraw
// 123.123 = 123,123_000_000_000_000_000 / 1_000_000_000 = 123,123_000_000
func ParseAmount(amount string, precision int) (*big.Int, error) {
	decAmount, err := sdk.NewDecFromStr(amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dec")
	}
	return decAmount.QuoInt64(int64(math.Pow10(precision))).BigInt(), nil
}
