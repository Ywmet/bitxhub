package executor

import (
	"github.com/meshplus/bitxhub-core/agency"
	"github.com/meshplus/bitxhub-kit/types"
	"github.com/meshplus/bitxhub-model/pb"
)

type SerialExecutor struct {
	normalTxs         []*types.Hash
	interchainCounter map[string][]uint64
	applyTxFunc       agency.ApplyTxFunc
	boltContracts     map[string]agency.Contract
}

func NewSerialExecutor(f1 agency.ApplyTxFunc, f2 agency.RegisterContractFunc) agency.TxsExecutor {
	return &SerialExecutor{
		applyTxFunc:   f1,
		boltContracts: f2(),
	}
}

func init() {
	agency.RegisterExecutorConstructor("serial", NewSerialExecutor)
}

func (se *SerialExecutor) ApplyTransactions(txs []*pb.Transaction) []*pb.Receipt {
	se.interchainCounter = make(map[string][]uint64)
	se.normalTxs = make([]*types.Hash, 0)
	receipts := make([]*pb.Receipt, 0, len(txs))

	for i, tx := range txs {
		receipts = append(receipts, se.applyTxFunc(i, tx, nil))
	}

	return receipts
}

func (se *SerialExecutor) GetBoltContracts() map[string]agency.Contract {
	return se.boltContracts
}

func (se *SerialExecutor) AddNormalTx(hash *types.Hash) {
	se.normalTxs = append(se.normalTxs, hash)
}

func (se *SerialExecutor) GetNormalTxs() []*types.Hash {
	return se.normalTxs
}

func (se *SerialExecutor) AddInterchainCounter(to string, index uint64) {
	se.interchainCounter[to] = append(se.interchainCounter[to], index)
}

func (se *SerialExecutor) GetInterchainCounter() map[string][]uint64 {
	return se.interchainCounter
}
