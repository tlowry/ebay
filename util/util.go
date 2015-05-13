package util

import (
	"appengine"
	"strings"
	"time"
)

type TierConf struct {
	Tiers     [][]string
	OutputDir string
}

func SanitizeNum(str string) string {
	ret := strings.Replace(str, ",", "", -1)
	ret = strings.Replace(ret, " ", "", -1)
	return ret
}

type MakeNew func() interface{}

type OnReturn func(item interface{})

type Pool struct {
	maxSize    int
	available  int
	In         chan interface{}
	Out        chan interface{}
	closeChan  chan bool
	MakeFunc   MakeNew
	ReturnFunc OnReturn
	running    bool
	ctx        appengine.Context
}

func NewPool(maxSize int, ctx appengine.Context) *Pool {
	pool := Pool{}
	pool.In = make(chan interface{}, maxSize)
	pool.Out = make(chan interface{}, maxSize)
	pool.closeChan = make(chan bool, 1)
	pool.maxSize = maxSize
	pool.running = true
	pool.MakeFunc = nil
	pool.ReturnFunc = nil
	pool.ctx = ctx
	go pool.serve()

	return &pool
}

func (pool *Pool) serve() {

	pool.available = pool.maxSize
	// Fill up the pool with instances

	for i := 0; i < pool.maxSize; i++ {

		if pool.MakeFunc == nil {
			pool.Out <- nil
		} else {
			pool.Out <- pool.MakeFunc()
		}

	}
	pool.ctx.Infof("pool.serve() produced instances")

	for pool.running {
		select {
		case <-pool.closeChan:
			pool.running = false
			break
		// Items being returned
		case item := <-pool.In:
			// perform user function if set
			if pool.ReturnFunc != nil {
				pool.ReturnFunc(item)
			}

			pool.Out <- item
		default:
			time.Sleep(time.Millisecond * 100)
		}

	}
	pool.ctx.Infof("pool.serve() coming down")
	close(pool.In)
	close(pool.Out)
	close(pool.closeChan)
}

func (pool *Pool) BorrowWait() interface{} {
	return <-pool.Out
}

func (pool *Pool) Borrow(wait time.Duration) interface{} {
	timeout := make(chan bool, 1)
	pool.ctx.Infof("pool.Borrow() called")
	go func() {
		time.Sleep(wait)
		timeout <- true
	}()

	select {
	case item := <-pool.Out:
		pool.ctx.Infof("pool.Borrow() success")
		return item
	case <-timeout:
		pool.ctx.Infof("pool.Borrow() timeout")
		return nil
	}
}

func (pool *Pool) Return(item interface{}) {
	pool.ctx.Infof("pool.Return() called")
	pool.In <- item
	pool.ctx.Infof("pool.Return() complete")
}

func (pool *Pool) Close() {
	pool.ctx.Infof("pool.Close() called")
	pool.closeChan <- true
	pool.ctx.Infof("pool.Close() complete")
}
