package chaindb

import (
	"github.com/hashrs/consensusdemo/types"
	"math/big"
	"strings"
)

const (
	BlockKeyPrefix       = "bl-"
	TransactionKeyPrefix = "tr-"
	ReceiptKeyPrefix     = "re-"
	HeightKeyPrefix      = "hei-"
)

func chainHeightKey() string {
	return "hei-:chain"
}

func blockKey(number *big.Int) string {
	var builder strings.Builder
	builder.WriteString(BlockKeyPrefix)
	builder.WriteString(":")
	builder.WriteString(number.Text(10))

	return builder.String()
}

func transactionKey(hash types.Hash) string {
	var builder strings.Builder
	builder.WriteString(TransactionKeyPrefix)
	builder.WriteString(":")
	builder.WriteString(hash.String())

	return builder.String()
}

func receiptKey(hash types.Hash) string {
	var builder strings.Builder
	builder.WriteString(ReceiptKeyPrefix)
	builder.WriteString(":")
	builder.WriteString(hash.String())
	return builder.String()
}
