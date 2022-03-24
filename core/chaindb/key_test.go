package chaindb

import (
	"github.com/hashrs/consensusdemo/core/common"
	"github.com/hashrs/consensusdemo/types"
	"testing"
)

func benchmark(b *testing.B, f func(hash types.Hash) string) {
	var h = common.Hex2Hash("0x4daf6f018bdc33e7f5548382a995b89a6eb809d09fdf386a750d10ed5b88b134")
	for i := 0; i < b.N; i++ {
		f(h)
	}
}

func BenchmarkMemChaindb_TransactionKey1(b *testing.B) {
	benchmark(b, transactionKey)
}
