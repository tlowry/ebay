package util

import (
	"log"
	"strings"
	"time"
)

type TierConf struct {
	Tiers     [][]string
	MaxPrice  float64
	OutputDir string
	LogDir    string
}

func SanitizeNum(str string) string {
	ret := strings.Replace(str, ",", "", -1)
	ret = strings.Replace(ret, " ", "", -1)
	return ret
}

// Quit the app with the provided error string message if iface is null
func FailOnNil(msg string, iface interface{}) {
	if iface == nil {
		log.Fatal(msg)
	}
}

// Quit the app with the provided error string message if iface is null
func FailOnErr(msg string, err error) {
	if err != nil {
		log.Fatal(msg, " ", err.Error())
	}
}

func Fail(msg string) {
	log.Fatal(msg)
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
}

func NewPool(maxSize int) *Pool {
	pool := Pool{}
	pool.In = make(chan interface{}, maxSize)
	pool.Out = make(chan interface{}, maxSize)
	pool.closeChan = make(chan bool, 1)
	pool.maxSize = maxSize
	pool.running = true
	pool.MakeFunc = nil
	pool.ReturnFunc = nil

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

	for pool.running {
		select {
		case <-pool.closeChan:
			pool.running = false
			break
		// Items being returned
		case item := <-pool.In:
			log.Println("Pool return")
			if pool.ReturnFunc != nil {
				pool.ReturnFunc(item)
			}

			pool.Out <- item
			log.Println("Pool return complete")
		}

	}
	log.Println("Shutting down pool")
	close(pool.In)
	close(pool.Out)
	close(pool.closeChan)
}

func (pool *Pool) BorrowWait() interface{} {
	return <-pool.Out
}

func (pool *Pool) Borrow(wait time.Duration) interface{} {
	timeout := make(chan bool, 1)

	go func() {
		time.Sleep(wait)
		timeout <- true
	}()

	select {
	case item := <-pool.Out:
		return item
	case <-timeout:
		return nil
	}
}

func (pool *Pool) Return(item interface{}) {
	log.Println("Pool Return func")
	pool.In <- item
}

func (pool *Pool) Close() {
	log.Println("Pool Closing")
	pool.closeChan <- true
}
