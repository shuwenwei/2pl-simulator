package homework

import "sync"

const (
	N = 100000
)

type resource struct {
	data         [N]int
	resourceLock sync.Mutex
	lockMap      map[int]*LockDesc
	refCountMap  map[int]int
}

func GenerateResource(data [N]int) *resource {
	return &resource{
		data:        data,
		lockMap:     map[int]*LockDesc{},
		refCountMap: map[int]int{},
	}
}
