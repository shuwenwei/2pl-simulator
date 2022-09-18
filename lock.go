package homework

import "sync"

type LockType int

const (
	S = LockType(iota)
	X
)

type LockDesc struct {
	mu         sync.Mutex
	cond       *sync.Cond
	lockStatus LockType
	holders    map[uint64]struct{}
}

func NewLockDesc() *LockDesc {
	lockDesc := new(LockDesc)
	lockDesc.holders = map[uint64]struct{}{}
	lockDesc.cond = sync.NewCond(&lockDesc.mu)
	return lockDesc
}
