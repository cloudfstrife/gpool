package gpool

import (
	"container/list"
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrCanNotGetItem return when get item failed
	ErrCanNotGetItem = errors.New("Unable GET Item")
	// ErrTypeConvert return when Convert item to require type
	ErrTypeConvert = errors.New("Class Cast Failed")
	// ErrItemInitialFailed return when Initial item failed
	ErrItemInitialFailed = errors.New("Iem Initial Failed")
	// ErrPoolExtendFailed return when Initial pool  failed
	ErrPoolExtendFailed = errors.New("Pool Extend Failed")
	// ErrHasBeenShotdown return when do something after pool has been shutdown
	ErrHasBeenShotdown = errors.New("Pool has been shutdown")
)

//Pool pool class
type Pool struct {
	Config      Config
	items       *list.List
	newItemFunc Creator
	cond        *sync.Cond
	shutdown    bool
}

//DefaultPool create a pool with default config
func DefaultPool(creater Creator) *Pool {
	return &Pool{
		Config:      DefaultConfig(),
		items:       list.New(),
		newItemFunc: creater,
		cond:        sync.NewCond(&sync.Mutex{}),
		shutdown:    false,
	}
}

//Initial initial pool
func (pool *Pool) Initial() {
	log.WithField(
		"config", pool.Config,
	).Debug("START Initial Pool")

	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	// Initial pool item
	go pool.extend(pool.Config.InitialPoolSize)
	pool.cond.Wait()
	// start cyclicity check goroutine
	go pool.startCheck()
	log.Debug("Done Initial Pool")
}

//Extend push item into pool
func (pool *Pool) extend(count int) {
	log.WithField(
		"count", count,
	).Debug("START | Extend Pool")

	pool.cond.L.Lock()
	defer pool.cond.Signal()
	defer pool.cond.L.Unlock()
	if pool.shutdown {
		log.Debug(ErrHasBeenShotdown)
		return
	}
	for i := 0; i < count; i++ {
		if pool.items.Len() >= pool.Config.MaxPoolSize {
			break
		}
		var item = pool.newItemFunc()
		err := item.Initial(pool.Config.Params)
		if err != nil {
			log.Error(ErrItemInitialFailed)
			continue
		}
		pool.items.PushBack(item)
	}
	log.Debug("DONE | Extend Pool")
}

//StartCheck start check avaiable goroutine
func (pool *Pool) startCheck() {
	t := time.NewTicker(time.Duration(pool.Config.TestDuration) * time.Millisecond)
	for {
		<-t.C
		pool.checkAvaiable()
	}
}

//CheckAvaiable check item avaiable
func (pool *Pool) checkAvaiable() {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	log.WithField(
		"pool Size", pool.items.Len(),
	).Debug("START | CheckAvaiable ")
	if pool.shutdown {
		return
	}
	for i := pool.items.Front(); i != nil; {
		item, Ok := i.Value.(Item)
		if !Ok {
			log.Error(ErrTypeConvert)
		}
		err := item.Check(pool.Config.Params)
		if err == nil {
			i = i.Next()
			continue
		} else {
			log.WithError(err).Error("CheckAvaiable ERROR ")
			n := i.Next()
			pool.items.Remove(i)
			i = n
		}
	}
	log.Debug("DONE | CheckAvaiable")
}

//GetOne get a pool item
func (pool *Pool) GetOne() (Item, error) {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	if pool.shutdown {
		return nil, ErrHasBeenShotdown
	}
	retry := 0
	var i *list.Element
	for {
		i = pool.items.Front()
		if i != nil {
			break
		}
		if retry <= pool.Config.AcquireRetryAttempts {
			retry++
			go pool.extend(pool.Config.AcquireIncrement)
			pool.cond.Wait()
			continue
		} else {
			return nil, ErrCanNotGetItem
		}
	}
	pool.items.Remove(i)
	if pool.items.Len() < pool.Config.MinPoolSize {
		go pool.extend(pool.Config.AcquireIncrement)
	}
	item, ok := i.Value.(Item)
	if !ok {
		return nil, ErrTypeConvert
	}
	if !pool.Config.TestOnGetItem {
		return item, nil
	}
	if err := item.Check(pool.Config.Params); err != nil {
		return nil, err
	}
	return item, nil
}

//BackOne  give back a pool item
func (pool *Pool) BackOne(item Item) {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	if pool.items.Len() >= pool.Config.MaxPoolSize {
		err := item.Destory(pool.Config.Params)
		if err != nil {
			log.WithError(err).Error("Item Destory ERROR While BackOne")
		}
		return
	}
	pool.items.PushBack(item)
}

//Shutdown shutdown pool
func (pool *Pool) Shutdown(wg *sync.WaitGroup) {
	log.Debug("START | Shutdown Pool")
	defer wg.Done()
	pool.cond.L.Lock()
	pool.shutdown = true
	defer pool.cond.L.Unlock()
	for i := pool.items.Front(); i != nil; i = pool.items.Front() {
		pool.items.Remove(i)
		item, ok := i.Value.(Item)
		if !ok {
			log.Error(ErrTypeConvert)
			continue
		}
		err := item.Destory(pool.Config.Params)
		if err != nil {
			log.WithError(err).Error("Item Destory ERROR While Shutdown ")
		}
	}
	log.Debug("DONE | Shutdown Pool")
}
