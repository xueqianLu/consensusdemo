package chaindb

import (
	"fmt"
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
	return fmt.Sprintf("%s:chain", HeightKeyPrefix)
}

func blockKey(number *big.Int) string {
	return fmt.Sprint("%s:%s", BlockKeyPrefix, number.Text(10))
}

func transactionKey(hash types.Hash) string {
	return fmt.Sprintf("%s:%s", TransactionKeyPrefix, hash.String())
}

func receiptKey(hash types.Hash) string {
	return fmt.Sprintf("%s:%s", ReceiptKeyPrefix, hash.String())
}
