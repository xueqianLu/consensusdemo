package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type FurtherTransaction struct {
	types.Transaction
	From common.Address
}
