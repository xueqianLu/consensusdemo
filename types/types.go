package types

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

type TxPackage struct {
	Time      uint64
	Blockhash string
	Tx        *types.Transaction
}

func (t *TxPackage) TimeStamp() uint64 {
	return t.Time
}

func (t *TxPackage) BlockHash() string {
	return t.Blockhash
}
func (t *TxPackage) Hashs() []string {
	unitLen := 32
	var m = make([]string, 0)
	var data = t.Tx.Data()
	for i := 0; len(data) >= (i + unitLen); i += unitLen {
		s := hex.EncodeToString(data[i : i+unitLen])
		m = append(m, s)
	}
	return m
}

type TxWithFrom struct {
	From    []byte `json:"From"`
	TxBytes []byte `json:"TxBytes"`
}

type TxPair struct {
	Hash []byte       `json:"Hash"`
	Txs  []TxWithFrom `json:"Txs"`
}

func (t *TxPair) GetHash() string {
	return hex.EncodeToString(t.Hash)
}

func (t *TxPair) GetTransactions() []*FurtherTransaction {
	var txs = make([]*FurtherTransaction, 0)
	for _, tx := range t.Txs {
		t := FurtherTransaction{}
		err := t.UnmarshalBinary([]byte(tx.TxBytes))
		if err != nil {
			log.Error("decode rlp tx ", "err", err)
			continue
		}
		t.From.SetBytes(tx.From)
		txs = append(txs, &t)
		//log.Info("got transaction with ", "from", t.From, "tx ", t.Transaction)
	}
	return txs
}
