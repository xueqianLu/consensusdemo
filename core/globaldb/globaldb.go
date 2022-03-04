package globaldb

import (
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/db"
	"math/big"
)

type GlobalDB interface {
	GetBalance(core.Account) *big.Int
	SubBalance(core.Account, *big.Int) *big.Int
	AddBalance(core.Account, *big.Int) *big.Int
}

func NewGlobalDB(database db.Database) GlobalDB {
	return &memglobaldb{}
}

type memglobaldb struct {
	state db.Database
}

func (m *memglobaldb) setValue(addr *core.Account, value *big.Int) {
	m.state.Set(*addr, value)
}

func (m *memglobaldb) getValue(addr *core.Account) *big.Int {
	if balance, exist := m.state.Get(*addr); exist {
		return balance.(*big.Int)
	} else {
		return new(big.Int)
	}
}

func (m *memglobaldb) GetBalance(addr core.Account) *big.Int {
	return m.getValue(&addr)
}

func (m *memglobaldb) SubBalance(addr core.Account, value *big.Int) *big.Int {
	c := m.getValue(&addr)
	r := new(big.Int).Sub(c, value)
	m.setValue(&addr, r)
	return r
}

func (m *memglobaldb) AddBalance(addr core.Account, value *big.Int) *big.Int {
	c := m.getValue(&addr)
	r := new(big.Int).Add(c, value)
	m.setValue(&addr, r)
	return r
}
