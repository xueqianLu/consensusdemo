package globaldb

import (
	"fmt"
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/db"
	"github.com/hashrs/consensusdemo/db/memdb"
	"math/big"
	"sync"
)

type GlobalDB interface {
	Commit() error
	GetBalance(core.Account) *big.Int
	SetBalance(core.Account, *big.Int)
	SubBalance(core.Account, *big.Int) *big.Int
	AddBalance(core.Account, *big.Int) *big.Int
}

func NewGlobalDB(database db.Database) GlobalDB {
	return &statedb{
		cache: memdb.NewMemDB(),
		state: database,
	}
}

type statedb struct {
	cache db.CacheKV
	dirty sync.Map
	state db.Database
}

func keyAccount(addr core.Account) string {
	return fmt.Sprintf("ac-%s", addr.String())
}

func (m *statedb) commit() error {
	m.dirty.Range(func(key, value interface{}) bool {
		addr := key.(core.Account)
		balan := value.(*big.Int)
		k := keyAccount(addr)
		v := balan.Text(10)
		m.state.Set(k, []byte(v))
		m.cache.Set(key, value)
		m.dirty.Delete(key)
		return true
	})
	return nil
}

func (m *statedb) setValue(addr core.Account, value *big.Int) {
	m.dirty.Store(addr, value)
}

func (m *statedb) getValue(addr core.Account) *big.Int {
	if v, exist := m.dirty.Load(addr); exist {
		return v.(*big.Int)
	}
	if v, exist := m.cache.Get(addr); exist {
		balance := v.(*big.Int)
		return balance
	}

	k := keyAccount(addr)
	if v, exist := m.state.Get(k); exist {
		balance, _ := new(big.Int).SetString(string(v), 10)
		m.cache.Set(addr, balance)
		return balance
	}
	return new(big.Int)
}

func (m *statedb) GetBalance(addr core.Account) *big.Int {
	return m.getValue(addr)
}

func (m *statedb) SubBalance(addr core.Account, value *big.Int) *big.Int {
	c := m.getValue(addr)
	r := new(big.Int).Sub(c, value)
	m.setValue(addr, r)
	return r
}

func (m *statedb) AddBalance(addr core.Account, value *big.Int) *big.Int {
	c := m.getValue(addr)
	r := new(big.Int).Add(c, value)
	m.setValue(addr, r)
	return r
}

func (m *statedb) SetBalance(addr core.Account, value *big.Int) {
	r := new(big.Int).Set(value)
	m.setValue(addr, r)
}

func (m *statedb) Commit() error {
	return m.commit()
}
