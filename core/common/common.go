package common

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/types"
)

func Hex2Account(h string) core.Account {
	addr := common.HexToAddress(h)
	return core.Account{addr}
}

func Hex2Hash(h string) types.Hash {
	hash := common.HexToHash(h)
	return types.Hash{
		hash,
	}
}
