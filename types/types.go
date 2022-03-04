package types

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"math/big"
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

type RoundInfo struct {
	Timestamp uint64       `json:"time"`
	Txsinfo   []*TxPackage `json:"package"`
}

type Hash struct {
	common.Hash
}

type FurtherTransaction struct {
	types.Transaction
	From common.Address
}

func (t FurtherTransaction) Hash() Hash {
	h := t.Transaction.Hash()
	return Hash{h}
}

type FurtherTransactions []*FurtherTransaction

func (f FurtherTransactions) IndexData(idx int) Hash {
	return f[idx].Hash()
}
func (f FurtherTransactions) Length() int {
	return len(f)
}

type Receipt struct {
	Status      int            `json:"status"`
	Txhash      Hash           `json:"hash"`
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Value       *big.Int       `json:"value"`
	BlockNumber *big.Int       `json:"block"`
	Timestamp   int64          `json:"time"`
}

func (r *Receipt) Data() []byte {
	var b = make([]byte, 0)
	b = append(b, big.NewInt(int64(r.Status)).Bytes()...)
	b = append(b, r.Txhash.Bytes()...)
	b = append(b, r.From.Bytes()...)
	b = append(b, r.To.Bytes()...)
	b = append(b, r.Value.Bytes()...)
	b = append(b, r.BlockNumber.Bytes()...)
	b = append(b, big.NewInt(r.Timestamp).Bytes()...)
	return b
}

func (r *Receipt) Hash() Hash {
	d := r.Data()
	hash := Hash{}
	h := sha3.Sum256(d)
	hash.SetBytes(h[:])
	return hash
}

type Receipts []*Receipt

func (rs Receipts) IndexData(idx int) Hash {
	return rs[idx].Hash()
}
func (rs Receipts) Length() int {
	return len(rs)
}
