package core

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashrs/consensusdemo/types"
	"golang.org/x/crypto/sha3"
	"math/big"
)

type Block struct {
	Header *BlockHeader
	Body   *BlockBody
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

func (b *Block) Encode() ([]byte, error) {
	return json.Marshal(b)
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

func (body *BlockBody) Encode() ([]byte, error) {
	return json.Marshal(body)
}

type BlockHeader struct {
	Parent      types.Hash `json:"parent"`
	BlockHash   types.Hash `json:"hash"`
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

func (h *BlockHeader) Encode() ([]byte, error) {
	return json.Marshal(h)
}

type Account struct {
	common.Address
}
