package gpool

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

//Item pool item
type Item interface {
	Initial(map[string]string) error
	Destory() error
	Check() error
}

//Creater create item function
type Creater func() Item

//Pool pool class
type Pool struct {
	Config   *Config
	items    *list.List
	lock     sync.Mutex
	shutdown context.CancelFunc
	Creater  Creater
}

//DefaultPool create a pool with default config
func DefaultPool(creater Creater) *Pool {
	var result = &Pool{
		Config:  DefaultConfig(),
		Creater: creater,
	}
	return result
}

//Initial initial pool
func (pool *Pool) Initial() {
	if pool.Config == nil {
		log.Fatal("pool config is nil")
	}

	pool.Log("START", "Pool Initial")

	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.items = list.New()

	//push item into pool
	go pool.Extend(pool.Config.InitialPoolSize)

	//start check avaiable goroutine
	ctx, cf := context.WithCancel(context.Background())
	pool.shutdown = cf
	go pool.StartCheck(ctx)

	pool.Log("DONE", "Pool Initial")
}

//Extend push item into pool
func (pool *Pool) Extend(count int) {
	pool.Log("START", fmt.Sprintf("Extend Count : %d", count))

	if pool.items.Len() >= pool.Config.MaxPoolSize {
		return
	}
	for i := 0; i < count; i++ {
		var item = pool.Creater()
		err := item.Initial(pool.Config.Params)
		if err != nil {
			log.Println("ERROR : Iem Initial ERROR \n", err)
			continue
		}
		pool.items.PushBack(item)
	}

	pool.Log("DONE", fmt.Sprintf("Extend Count : %d ,Pool size : %d", count, pool.items.Len()))
}

//StartCheck start check avaiable goroutine
func (pool *Pool) StartCheck(ctx context.Context) {
	t := time.NewTicker(time.Duration(pool.Config.TestDuration) * time.Millisecond)
a:
	for {
		select {
		case <-ctx.Done():
			break a
		case <-t.C:
			pool.Log("START", "CheckAvaiable")
			pool.CheckAvaiable()
			pool.Log("DONE", "CheckAvaiable")
		}
	}
}

//CheckAvaiable check item avaiable
func (pool *Pool) CheckAvaiable() {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	for i := pool.items.Front(); i != nil; i = i.Next() {
		item, Ok := i.Value.(Item)
		if !Ok {
			log.Println("ERROR : Class Cast ERROR while CheckAvaiable")
		}
		err := item.Check()
		if err != nil {
			log.Println("ERROR : CheckAvaiable ERROR \n", err)
			pool.items.Remove(i)
		}
	}
}

//GetOne get a pool item
func (pool *Pool) GetOne() (Item, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	retry := 0
	for {
		//检查链表头元素是否为空，防止链表遍历结束依然未获取到连接时报错
		i := pool.items.Front()
		if i == nil {
			if retry <= pool.Config.AcquireRetryAttempts {
				retry++
				go pool.Extend(pool.Config.AcquireIncrement)
				continue
			}
			return nil, errors.New("Unable GET Item")
		}
		pool.items.Remove(i)
		item, ok := i.Value.(Item)
		if !ok {
			return nil, errors.New("Class Cast ERROR while Get Item")
		}
		if pool.items.Len() < pool.Config.MinPoolSize {
			go pool.Extend(pool.Config.AcquireIncrement)
		}
		if pool.Config.TestOnGetItem {
			err := item.Check()
			return item, err
		}
		return item, nil
	}
}

//BackOne  give back a pool item
func (pool *Pool) BackOne(item Item) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if pool.items.Len() >= pool.Config.MaxPoolSize {
		err := item.Destory()
		if err != nil {
			log.Println("ERROR : Item Destory ERROR While BackOne \n", err)
		}
		return
	}
	pool.items.PushBack(item)
	return
}

//Shutdown shutdown pool
func (pool *Pool) Shutdown() {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.Log("START", "Shutdown Pool")
	for i := pool.items.Front(); i != nil; i = pool.items.Front() {
		item, ok := i.Value.(Item)
		pool.items.Remove(i)
		if !ok {
			log.Println("ERROR : Class Cast ERROR while shutdown pool")
			continue
		}
		err := item.Destory()
		if err != nil {
			log.Println("ERROR : Item Destory ERROR While Shutdown \n", err)
		}
	}
	pool.shutdown()
	pool.Log("DONE", "Shutdown Pool")
}

//Log record log
func (pool *Pool) Log(status, msg string) {
	if pool.Config.Debug {
		log.Printf("INFO : [ %5s] %s\n", status, msg)
	}
}
