package chaindb

import (
	"github.com/hashrs/consensusdemo/types"
	"math/big"
)

const (
	BlockKeyPrefix       = "bl-"
	TransactionKeyPrefix = "tr-"
	ReceiptKeyPrefix     = "re-"
	HeightKeyPrefix      = "hei-"
)

func chainHeightKey() string {
	return "hei-chain"
}

func blockKey(number *big.Int) string {
	n := number.Text(10)
	s := make([]byte, 0, len(n)+len(BlockKeyPrefix))
	s = append(s, BlockKeyPrefix...)
	s = append(s, n...)
	return string(s)
}

func transactionKey(hash types.Hash) string {
	h := hash.String()
	s := make([]byte, 0, len(h)+len(TransactionKeyPrefix))
	s = append(s, TransactionKeyPrefix...)
	s = append(s, h...)
	return string(s)
}

func receiptKey(hash types.Hash) string {
	h := hash.String()
	s := make([]byte, 0, len(h)+len(ReceiptKeyPrefix))
	s = append(s, ReceiptKeyPrefix...)
	s = append(s, h...)
	return string(s)
}
