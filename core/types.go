package core

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Block struct {
	Header BlockHeader
	Body   BlockBody
}

type BlockBody struct {
	Txs []*FurtherTransaction
}

type BlockHeader struct {
	Parent common.Hash `json:"parent"`
	Hash   common.Hash `json:"hash"`
	Number *big.Int    `json:"number"`
}

type Account common.Address
