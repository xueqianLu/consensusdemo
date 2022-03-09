package node

import (
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/controller"
	"github.com/hashrs/consensusdemo/core/chaindb"
	"github.com/hashrs/consensusdemo/core/globaldb"
	"github.com/hashrs/consensusdemo/core/miner"
	"github.com/hashrs/consensusdemo/db/storagedb"
	"github.com/hashrs/consensusdemo/txpool"
	log "github.com/sirupsen/logrus"
)

type Node struct {
	conf   *config.Config
	chain  chaindb.ChainDB
	global globaldb.GlobalDB
	server *controller.APIServer
	worker *miner.Miner
	txp    *txpool.TxPool
}

func NewNode(conf *config.Config) *Node {
	node := &Node{
		conf: conf,
	}
	chainDB := storagedb.NewStorageDB(conf)
	stateDB := storagedb.NewStorageDB(conf)
	node.chain = chaindb.NewChainDB(chainDB)
	node.global = globaldb.NewGlobalDB(stateDB)

	node.server = controller.NewServer(node.chain, node.global)
	node.txp = txpool.NewTxpool(conf)
	node.worker = miner.NewMiner(node.global, node.chain, node.txp, conf)

	return node
}

func (n *Node) Start() {
	log.Info("node start")
	n.txp.Start()
	log.Info("txpool started")
	go n.server.Start()
	log.Info("api server started")
}

func (n *Node) Miner() *miner.Miner {
	return n.worker
}
