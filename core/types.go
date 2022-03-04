package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashrs/consensusdemo/types"
	"golang.org/x/crypto/sha3"
	"math/big"
)

type Block struct {
	Header BlockHeader
	Body   BlockBody
}

func (b *Block) Hash() types.Hash {
	headb := b.Header.Bytes()
	body := b.Body.Bytes()
	data := append(headb, body...)
	hash := types.Hash{}
	h := sha3.Sum256(data)
	hash.SetBytes(h[:])
	return hash
}

type BlockBody struct {
	Txs []*types.FurtherTransaction
}

func (body *BlockBody) Bytes() []byte {
	var b = make([]byte, 0)
	for _, t := range body.Txs {
		b = append(b, t.Data()...)
	}

	return b
}

type BlockHeader struct {
	Parent      types.Hash `json:"parent"`
	Number      *big.Int   `json:"number"`
	Timestamp   uint64     `json:"time"`
	ReceiptRoot types.Hash `json:"receipt"`
	TxRoot      types.Hash `json:"transaction"`
}

func (h *BlockHeader) Bytes() []byte {
	var b = make([]byte, 0)
	b = append(b, h.Parent.Bytes()...)
	b = append(b, h.Number.Bytes()...)
	b = append(b, big.NewInt(int64(h.Timestamp)).Bytes()...)
	b = append(b, h.ReceiptRoot.Bytes()...)
	b = append(b, h.TxRoot.Bytes()...)

	return b
}

type Account struct {
	common.Address
}
