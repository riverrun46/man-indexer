package main

import (
	"encoding/json"
	"fmt"
	"manindexer/adapter/bitcoin"
	"manindexer/common"
	"manindexer/database"
	"manindexer/database/mongodb"
	"manindexer/man"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

func TestGetBlock(t *testing.T) {

	chain := &bitcoin.BitcoinChain{}
	block, err := chain.GetBlock(1)
	fmt.Println(err)
	b := block.(*wire.MsgBlock)
	fmt.Println(b.Header.BlockHash().String())

	txret, err := chain.GetTransaction("798a14129d9697906908046998431ee9e97293bc6c5e8d9d3418f1d944913608")
	fmt.Println(err)
	tx := txret.(*btcutil.Tx)
	fmt.Println("HasWitness", tx.HasWitness())
	for _, out := range tx.MsgTx().TxOut {
		fmt.Println(out.Value)
	}

	indexer := &bitcoin.Indexer{ChainParams: &chaincfg.TestNet3Params}
	pins := indexer.CatchPinsByTx(tx.MsgTx(), 123, 11123232, "", "", 0)
	fmt.Println(len(pins))
	for _, pin := range pins {
		fmt.Println("----------------")
		fmt.Printf("%+v\n", pin)
		//fmt.Println("-----------------\ncontent:", string(inscription.Pin.ContentBody))
		//fmt.Println(hex.EncodeToString(inscription.Pin.ContentBody))
	}
}
func TestGetPin(t *testing.T) {
	txId := "793e32472f85e94cae3ea552c320c362137a84b864d6cda6f342864375f4dbcf"
	chain := &bitcoin.BitcoinChain{}
	txret, err := chain.GetTransaction(txId)
	if err != nil {
		return
	}
	tx := txret.(*btcutil.Tx)
	fmt.Println("HasWitness", tx.HasWitness())
	indexer := &bitcoin.Indexer{ChainParams: &chaincfg.TestNet3Params}
	pins := indexer.CatchPinsByTx(tx.MsgTx(), 0, 0, "", "", 0)
	fmt.Println(pins)
	for _, pin := range pins {
		fmt.Println(string(pin.ContentBody))
	}
}
func TestAddMempoolPin(t *testing.T) {
	dbAdapter := &mongodb.Mongodb{}
	pin, err := dbAdapter.GetPinByNumberOrId("2")
	fmt.Println(err, pin.Address)
	err = dbAdapter.AddMempoolPin(pin)
	fmt.Println(err)
}
func TestDelMempoolPin(t *testing.T) {
	man.InitAdapter("btc", "mongo", "1", "1")
	man.DeleteMempoolData(2572919)
}
func TestConfig(t *testing.T) {
	config := common.Config
	fmt.Println(config.Protocols)
}

func TestGetDbPin(t *testing.T) {
	mg := mongodb.Mongodb{}
	p, err := mg.GetPinByNumberOrId("64")
	fmt.Println(err)
	//fmt.Println(p.ContentBody)
	contentType := common.DetectContentType(&p.ContentBody)
	fmt.Println(contentType)
}
func TestMongoGeneratorFind(t *testing.T) {
	jsonData := `
	{"collection":"pins","action":"sum","filterRelation":"or","field":["number"],
	"filter":[{"operator":"=","key":"number","value":1},{"operator":"=","key":"number","value":2}],
	"cursor":0,"limit":1,"sort":["number","desc"]
	}
	`
	var g database.Generator
	err := json.Unmarshal([]byte(jsonData), &g)
	fmt.Println(err)
	fmt.Println(g.Action)
	mg := mongodb.Mongodb{}
	ret, err := mg.GeneratorFind(g)
	fmt.Println(err, len(ret))
	if err == nil {
		jsonStr, err1 := json.Marshal(ret)
		if err1 != nil {
			fmt.Println("Error marshalling JSON:", err)
		}
		fmt.Println(string(jsonStr))
	}
}
func TestGetSaveData(t *testing.T) {
	man.InitAdapter("btc", "mongo", "2", "1")
	a, _, _, _, _, _, _ := man.GetSaveData(471)
	//mg := mongodb.Mongodb{}
	//err := mg.BatchUpsertMetaIdInfo(d)
	fmt.Println(a)
}