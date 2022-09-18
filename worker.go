package homework

import (
	"math/rand"
	"sync"
)

type TxnState int

const (
	Normal = TxnState(iota)
	Abort
)

var wg sync.WaitGroup

type worker struct {
	id        int
	curTxnId  uint64
	holdLocks []int
	wg        *sync.WaitGroup
	*resource
}

func NewWorker(id int, resource *resource, wg *sync.WaitGroup) *worker {
	w := new(worker)
	w.id = id
	w.holdLocks = make([]int, 0)
	w.resource = resource
	w.wg = wg
	return w
}

func (w *worker) run() {
	for i := 0; i < 10000; i++ {
		w.curTxnId = transactionManger.GetTxnId()
		w.op()
	}
	w.wg.Done()
}

func (w *worker) op() {
	idx := rand.Intn(N)
	idx1 := (idx + 1) % N
	idx2 := (idx + 2) % N
	j := rand.Intn(N)
	for {
		if w.acquireLock(idx, S) == Abort {
			continue
		} else {
			w.holdLocks = append(w.holdLocks, idx)
		}

		if w.acquireLock(idx1, S) == Abort {
			w.releaseAllLocks()
			continue
		} else {
			w.holdLocks = append(w.holdLocks, idx1)
		}

		if w.acquireLock(idx2, S) == Abort {
			w.releaseAllLocks()
			continue
		} else {
			w.holdLocks = append(w.holdLocks, idx2)
		}

		if w.acquireLock(j, X) == Abort {
			w.releaseAllLocks()
			continue
		} else {
			w.holdLocks = append(w.holdLocks, j)
		}
		//log.Printf("txn-%d acquire all lock success\n", w.curTxnId)
		w.data[j] = w.data[idx] + w.data[idx1] + w.data[idx2]

		w.releaseAllLocks()
		return
	}
}

func (w *worker) releaseAllLocks() {
	for _, i := range w.holdLocks {
		w.releaseLock(i)
	}
	w.holdLocks = make([]int, 0)
}

func (w *worker) acquireLock(i int, requestLockType LockType) (state TxnState) {
	Dprintf("txn-%d acquire lock-%d\n", w.curTxnId, i)
	state = Normal

	w.resourceLock.Lock()
	lockDesc, ok := w.lockMap[i]
	if !ok {
		lockDesc = NewLockDesc()
		w.lockMap[i] = lockDesc
		w.refCountMap[i] = 0
	}
	refCount := w.refCountMap[i]
	w.refCountMap[i] = refCount + 1
	w.resourceLock.Unlock()

	for {
		lockDesc.mu.Lock()
		// 没有其他worker获取锁
		holdersCount := len(lockDesc.holders)
		if holdersCount == 0 {
			lockDesc.lockStatus = requestLockType
			lockDesc.holders[w.curTxnId] = struct{}{}
			lockDesc.mu.Unlock()
			return
		}
		// 自己已经持有S或X锁，且当前锁没有其他事务，自己重复获取的情况
		if holdersCount == 1 {
			if _, exist := lockDesc.holders[w.curTxnId]; exist {
				if lockDesc.lockStatus == X || requestLockType == X {
					lockDesc.lockStatus = X
				}
				lockDesc.mu.Unlock()
				return
			}
		}
		// 被其他worker持有S锁
		if lockDesc.lockStatus == S {
			// S和S
			if requestLockType == S {
				lockDesc.holders[w.curTxnId] = struct{}{}
				lockDesc.mu.Unlock()
				return
			}
		}

		// S和X 或 X和X

		// 死锁预防
		state = w.avoidDeadLock(lockDesc)
		if state == Abort {
			lockDesc.mu.Unlock()
			return state
		}

		lockDesc.cond.Wait()
		lockDesc.mu.Unlock()
	}
}

// 使用wait-die的方式
// 允许老的事务等待新的事务,新的事务要获取老的事务的lock时abort
func (w *worker) avoidDeadLock(lockDesc *LockDesc) TxnState {
	for txnId := range lockDesc.holders {
		if txnId < w.curTxnId {
			Dprintf("txn-%d abort due to wait-die policy\n", w.curTxnId)
			return Abort
		}
	}
	return Normal
}

func (w *worker) releaseLock(i int) {
	w.resourceLock.Lock()
	defer w.resourceLock.Unlock()

	lockDesc := w.lockMap[i]

	lockDesc.mu.Lock()

	delete(lockDesc.holders, w.curTxnId)
	// 减少计数
	refCount := w.refCountMap[i]
	w.refCountMap[i] = refCount - 1
	// 如果该lock没有出现在其他地方
	if refCount == 1 {
		delete(w.lockMap, i)
	}
	lockDesc.cond.Broadcast()
	lockDesc.mu.Unlock()
}
