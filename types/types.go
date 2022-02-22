package types

import (
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/core/types"
)

type Md5tx struct {
	Time      uint64
	Blockhash string
	Tx        *types.Transaction
}

func (t *Md5tx) TimeStamp() uint64 {
	return t.Time
}

func (t *Md5tx) BlockHash() string {
	return t.Blockhash
}
func (t *Md5tx) Md5s() []string {
	var m = make([]string, 0)
	var data = t.Tx.Data()
	for i := 0; len(data) >= (i + 16); i += 16 {
		s := hex.EncodeToString(data[i : i+16])
		m = append(m, s)
	}
	return m
}

type TxPair struct {
	Md5 string          `json:"MD5"`
	Tx  json.RawMessage `json:"Tx"`
}

func (t *TxPair) GetMd5() string {
	return t.Md5
}

func (t *TxPair) GetTransaction() *types.Transaction {
	var tx types.Transaction
	err := json.Unmarshal([]byte(t.Tx), &tx)
	if err != nil {
		return nil
	}
	return &tx
}
