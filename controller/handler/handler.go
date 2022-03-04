package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashrs/consensusdemo/controller/model"
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/core/common"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
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
		info := formatTransactionInfo(tx)
		ResponseSuccess(c, info)
	}
}

func (a ApiHandler) GetReceipt(c *gin.Context) {
	strhash := c.Query("hash")
	hash := common.Hex2Hash(strhash)
	receipt := a.chain.GetReceipt(hash)
	if receipt == nil {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("receipt not found"))
	} else {
		info := formatReceiptInfo(receipt)
		ResponseSuccess(c, info)
	}
}

func (a ApiHandler) GetAccount(c *gin.Context) {
	straddr := c.Query("address")
	acc := common.Hex2Account(straddr)
	log.Debug("handler get account ", "address ", straddr, " acc=", acc.String())
	balance := a.global.GetBalance(acc)
	if balance == nil {
		ResponseSuccess(c, 0)
	} else {
		ResponseSuccess(c, balance.Int64())
	}
}

func formatReceiptInfo(receipt *types.Receipt) map[string]interface{} {
	var info = make(map[string]interface{})
	info["from"] = receipt.From.String()
	info["to"] = receipt.To.String()
	info["packedtime"] = receipt.PackedTime
	info["exectime"] = receipt.ExecTime
	info["block"] = receipt.BlockNumber.Text(10)
	info["value"] = receipt.Value.Text(10)
	info["hash"] = receipt.Txhash.String()
	info["status"] = receipt.Status

	return info
}

func formatTransactionInfo(tx *types.FurtherTransaction) map[string]interface{} {
	var info = make(map[string]interface{})
	info["from"] = tx.From.String()
	info["to"] = tx.To().String()
	info["nonce"] = tx.Nonce()
	info["value"] = tx.Value().Text(10)
	info["hash"] = tx.Hash().String()

	return info
}

func formatBlockInfo(block *core.Block) map[string]interface{} {
	var info = make(map[string]interface{})
	info["number"] = block.Header.Number.Int64()
	info["timestamp"] = block.Header.Timestamp
	info["parent"] = block.Header.Parent.String()
	info["txroot"] = block.Header.TxRoot.String()
	info["receiptroot"] = block.Header.ReceiptRoot.String()
	txs := make([]string, 0)
	for _, t := range block.Body.Txs {
		hash := t.Hash()
		txs = append(txs, hash.String())
	}
	info["txs"] = txs
	return info
}

func (a ApiHandler) GetBlock(c *gin.Context) {
	number := c.Query("number")
	bignumber, ok := new(big.Int).SetString(number, 10)
	fmt.Println("api handler got param", " number=", number, "changto bignumber=", ok)
	log.Info("api handler got param", " number=", number, "changto bignumber=", ok)
	if !ok {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("invalid block number"))
		return
	}
	block := a.chain.GetBlock(bignumber)
	if block == nil {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("block not found"))
	} else {
		info := formatBlockInfo(block)
		ResponseSuccess(c, info)
	}
}

func (a ApiHandler) InitAccount(c *gin.Context) {
	var param model.InitAccountParam
	err := c.BindJSON(&param)
	if err != nil {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("invalid param"))
		return
	}
	addr := common.Hex2Account(param.Addr)
	amount, ok := new(big.Int).SetString(param.Amount, 10)
	if !ok {
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("invalid param"))
		return
	}
	a.global.SetBalance(addr, amount)
	ResponseSuccess(c, nil)
}
