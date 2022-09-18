package homework

import "sync"

var transactionManger = txnManager{}

type txnManager struct {
	mu    sync.Mutex
	txnId uint64
}

func (tm *txnManager) GetTxnId() uint64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.txnId++
	return tm.txnId
}
