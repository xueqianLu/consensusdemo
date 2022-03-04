package lib

import (
	"github.com/hashrs/consensusdemo/types"
	"golang.org/x/crypto/sha3"
)

type Indexable interface {
	Length() int
	IndexData(idx int) types.Hash
}

func HashSlices(d Indexable) types.Hash {
	var data = make([]byte, 0)
	for i := 0; i < d.Length(); i++ {
		h := d.IndexData(i)
		data = append(data, h.Bytes()...)
	}
	hash := sha3.Sum256(data)

	h := types.Hash{}
	h.SetBytes(hash[:])
	return h
}
