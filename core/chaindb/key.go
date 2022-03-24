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
	return "hei-chain"
}

func blockKey(number *big.Int) string {
	n := number.Text(10)
	s := make([]byte, len(n)+len(BlockKeyPrefix))
	copy(s, BlockKeyPrefix)
	copy(s[len(BlockKeyPrefix):], n)
	//s = append(s, BlockKeyPrefix...)
	//s = append(s, n...)
	return string(s)
}

func transactionKey(hash types.Hash) string {
	h := hash.String()
	s := make([]byte, len(h)+len(TransactionKeyPrefix))
	copy(s, TransactionKeyPrefix)
	copy(s[len(TransactionKeyPrefix):], h)
	//s = append(s, TransactionKeyPrefix...)
	//s = append(s, h...)
	return string(s)
}

func transactionKey2(hash types.Hash) string {
	h := hash.String()
	s := make([]byte, 0, len(h)+len(TransactionKeyPrefix))
	s = append(s, TransactionKeyPrefix...)
	s = append(s, h...)
	return string(s)
}

func transactionKey3(hash types.Hash) string {
	var builder strings.Builder
	builder.WriteString(TransactionKeyPrefix)
	builder.WriteString(hash.String())

	return builder.String()
}

func receiptKey(hash types.Hash) string {
	h := hash.String()
	s := make([]byte, len(h)+len(ReceiptKeyPrefix))
	copy(s, ReceiptKeyPrefix)
	copy(s[len(ReceiptKeyPrefix):], h)
	//s = append(s, ReceiptKeyPrefix...)
	//s = append(s, h...)
	return string(s)
}
