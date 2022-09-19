package homework

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	workerNum := 100
	datas := [N]int{}
	for i := 0; i < len(datas); i++ {
		datas[i] = rand.Intn(100)
	}
	resources := GenerateResource(datas)
	wg := sync.WaitGroup{}
	wg.Add(workerNum)
	for i := 0; i < workerNum; i++ {
		w := NewWorker(i, resources, &wg)
		go w.run()
	}
	wg.Wait()
	if len(resources.lockMap) != 0 {
		t.FailNow()
	}
	if len(resources.refCountMap) != 0 {
		t.FailNow()
	}
}

func TestLock(t *testing.T) {
	datas := [N]int{}
	for i := 0; i < len(datas); i++ {
		datas[i] = rand.Intn(100)
	}
	resources := GenerateResource(datas)
	wg := sync.WaitGroup{}
	fmt.Println("---------------TestLockUpgrade-----------------")
	w1 := NewWorker(1, resources, &wg)
	w2 := NewWorker(2, resources, &wg)
	finishCh := make(chan struct{})
	go func(finishCh chan struct{}) {
		w1.curTxnId = transactionManger.GetTxnId()
		w1.acquireLock(1, S)
		w1.acquireLock(1, X)
		w1.releaseLock(1)
		finishCh <- struct{}{}
	}(finishCh)

	select {
	case <-finishCh:
		break
	case <-time.After(2 * time.Second):
		t.FailNow()
	}

	fmt.Println("---------------TestDeadLock-----------------")
	w1.curTxnId = transactionManger.GetTxnId()
	w2.curTxnId = transactionManger.GetTxnId()
	wg.Add(2)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		if w1.acquireLock(1, X) == Abort {
			log.Printf("w1 abort\n")
			return
		}
		time.Sleep(2*time.Second)
		if w1.acquireLock(2, X) == Abort {
			w1.releaseLock(1)
			log.Printf("w1 abort\n")
			return
		}
		log.Printf("w1 finished\n")
	}()
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		if w2.acquireLock(2, X) == Abort {
			log.Printf("w2 abort\n")
			return
		}
		time.Sleep(2 * time.Second)
		if w2.acquireLock(1, X) == Abort {
			log.Printf("w2 abort\n")
			w2.releaseLock(2)
			return
		}
		log.Printf("w2 finished\n")
	}()
	go func() {
		wg.Wait()
		finishCh<- struct{}{}
	}()
	select {
	case <-finishCh:
		break
	case <-time.After(6 * time.Second):
		t.FailNow()
	}
}