package globaldb

import (
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/db"
	"math/big"
)

type GlobalDB interface {
	GetBalance(core.Account) *big.Int
	SetBalance(core.Account, *big.Int)
	SubBalance(core.Account, *big.Int) *big.Int
	AddBalance(core.Account, *big.Int) *big.Int
}

func NewGlobalDB(database db.Database) GlobalDB {
	return &statedb{
		state: database,
	}
}

type statedb struct {
	state db.Database
}

func (m *statedb) setValue(addr core.Account, value *big.Int) {
	v := value.Text(10)
	m.state.Set(addr, v)
}

func (m *statedb) getValue(addr core.Account) *big.Int {
	if v, exist := m.state.Get(addr); exist {
		balance, _ := new(big.Int).SetString(v.(string), 10)
		return balance
	} else {
		return new(big.Int)
	}
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
