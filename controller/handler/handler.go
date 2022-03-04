package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashrs/consensusdemo/core/common"
	"math/big"
	"net/http"
)

type ApiHandler struct {
	chain  ChainReader
	global GlobalReader
}

func NewHandler(chain ChainReader, global GlobalReader) ApiHandler {
	return ApiHandler{chain: chain, global: global}
}

func (a ApiHandler) GetTransaction(c *gin.Context) {
	strhash := c.Query("hash")
	hash := common.Hex2Hash(strhash)
	tx := a.chain.GetTransaction(hash)
	if tx == nil {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("tx not found"))
	} else {
		ResponseSuccess(c, tx)
	}
}

func (a ApiHandler) GetReceipt(c *gin.Context) {
	strhash := c.Query("hash")
	hash := common.Hex2Hash(strhash)
	tx := a.chain.GetReceipt(hash)
	if tx == nil {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("receipt not found"))
	} else {
		ResponseSuccess(c, tx)
	}
}

func (a ApiHandler) GetAccount(c *gin.Context) {
	straddr := c.Query("address")
	acc := common.Hex2Account(straddr)
	balance := a.global.GetBalance(acc)
	if balance == nil {
		ResponseSuccess(c, 0)
	} else {
		ResponseSuccess(c, balance.Int64())
	}
}

func (a ApiHandler) GetBlock(c *gin.Context) {
	number := c.Query("number")
	bignumber, ok := new(big.Int).SetString(number, 10)
	if !ok {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("invalid block number"))
	}
	block := a.chain.GetBlock(bignumber)
	if block == nil {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("block not found"))
	} else {
		ResponseSuccess(c, block)
	}
}
