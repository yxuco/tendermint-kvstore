package main

import (
	"bytes"

	"github.com/dgraph-io/badger"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

// KVStoreApplication tendermint app
type KVStoreApplication struct {
	db           *badger.DB
	currentBatch *badger.Txn
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

// NewKVStoreApplication tendrmint app constructor
func NewKVStoreApplication(db *badger.DB) *KVStoreApplication {
	return &KVStoreApplication{
		db: db,
	}
}

// Info tendermint ABCI interface
func (app *KVStoreApplication) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

// SetOption tendermint ABCI interface
func (app *KVStoreApplication) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

// DeliverTx tendermint ABCI interface
func (app *KVStoreApplication) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return abcitypes.ResponseDeliverTx{Code: code}
	}

	parts := bytes.Split(req.Tx, []byte("="))
	key, value := parts[0], parts[1]

	err := app.currentBatch.Set(key, value)
	if err != nil {
		panic(err)
	}

	return abcitypes.ResponseDeliverTx{Code: 0}
}

// CheckTx tendermint ABCI interface
func (app *KVStoreApplication) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return abcitypes.ResponseCheckTx{Code: code, GasWanted: 1}
}

// Commit tendermint ABCI interface
func (app *KVStoreApplication) Commit() abcitypes.ResponseCommit {
	app.currentBatch.Commit()

	return abcitypes.ResponseCommit{Data: []byte{}}
}

// Query tendermint ABCI interface
func (app *KVStoreApplication) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	var res abcitypes.ResponseQuery
	res.Key = req.Data
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(req.Data)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			res.Log = "does not exist"
		} else {
			return item.Value(func(val []byte) error {
				res.Log = "exists"
				res.Value = val
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return res
}

// InitChain tendermint ABCI interface
func (app *KVStoreApplication) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

// BeginBlock tendermint ABCI interface
func (app *KVStoreApplication) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.currentBatch = app.db.NewTransaction(true)
	return abcitypes.ResponseBeginBlock{}
}

// EndBlock tendermint ABCI interface
func (app *KVStoreApplication) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

// ListSnapshots tendermint ABCI interface
func (app *KVStoreApplication) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

// OfferSnapshot tendermint ABCI interface
func (app *KVStoreApplication) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

// LoadSnapshotChunk tendermint ABCI interface
func (app *KVStoreApplication) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

// ApplySnapshotChunk tendermint ABCI interface
func (app *KVStoreApplication) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (app *KVStoreApplication) isValid(tx []byte) (code uint32) {
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	key, value := parts[0], parts[1]

	// check if the same key=value already exists
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			return item.Value(func(val []byte) error {
				if bytes.Equal(val, value) {
					code = 2
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return code
}
