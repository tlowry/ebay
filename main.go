package main

import (
	"github.com/tlowry/ebay/util"
	"log"
	"net/http"
	"sync"
	"time"
)

/*
import (
	"encoding/json"
	"flag"
	"github.com/tlowry/ebay/app"
	"github.com/tlowry/ebay/util"
	"io/ioutil"
	"log"

)


func main() {

	log.Println("About to parse args")
	var confFile = flag.String("confFile", "hierarchy.json", "json configuration file")
	flag.Parse()

	fileBytes, err := ioutil.ReadFile(*confFile)
	util.FailOnErr("failed to read json conf", err)

	conf := util.TierConf{}
	err = json.Unmarshal(fileBytes, &conf)
	util.FailOnErr("Failed to parse json", err)

	log.Println("About to run tiers ", conf.MaxPrice)

	if len(conf.Tiers) < 1 {
		util.Fail("ERROR: no tiers in json")
	}

	searcher := app.NewTierSearch(&conf)
	searcher.RunSearch()

}
*/

func getPool(max int) *util.Pool {

	pool := util.NewPool(max)
	pool.MakeFunc = func() interface{} {
		return &http.Client{}
	}

	return pool
}

func doBorrow(num int, pool *util.Pool, wg *sync.WaitGroup) {
	defer func() {
		log.Println("wg done")
		wg.Done()
	}()

	count := 0
	for i := 0; i < num; i++ {
		client := pool.BorrowWait().(*http.Client)
		time.Sleep(time.Second * 10)
		pool.Return(client)
		count++
	}

	log.Println("all returned ", count, " clients")

}

func stress() {
	max := 5
	pool := getPool(max)

	wg := sync.WaitGroup{}

	for count := 0; count < max*2; count++ {
		wg.Add(1)
		go doBorrow(max, pool, &wg)
	}

	wg.Wait()
	log.Println("WG complete")
	pool.Close()
	log.Println("All done")

}

func main() {
	stress()
}

func valid() {
	max := 10
	pool := getPool(max)

	clients := make([]*http.Client, 0, max)

	for i := 0; i < max; i++ {
		client := pool.BorrowWait().(*http.Client)
		clients = append(clients, client)
	}

	borrowed := len(clients)
	log.Println("Borrowed ", borrowed)

	limit := pool.Borrow(time.Second * 1)

	if limit != nil {
		log.Println("Got a client over limit")
	} else {
		log.Println("Pool is enforcing limits")
	}

	for i, client := range clients {
		log.Println("Returning ", i, " to pool")
		pool.Return(client)
	}

	returned := pool.Borrow(time.Second * 1)

	if returned == nil {
		log.Println("failed to checkout item")
	} else {
		log.Println("All good")
	}

	pool.Close()

}
