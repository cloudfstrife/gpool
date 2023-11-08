package gpool

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

//testingItem testing Item
type testingItem struct {
	initialed bool
	disabled  bool
	checked   bool
}

//Initial initial tesing item
func (c *testingItem) Initial(params map[string]string) error {
	c.initialed = true
	if msg, ok := params["InitError"]; ok {
		return errors.New(msg)
	}
	return nil
}

//Destory Destory tesing item
func (c *testingItem) Destory(params map[string]string) error {
	c.disabled = true
	if msg, ok := params["DestoryError"]; ok {
		return errors.New(msg)
	}
	return nil
}

//Check check tesing item avaiable
func (c *testingItem) Check(params map[string]string) error {
	c.checked = true
	if msg, ok := params["CheckError"]; ok {
		return errors.New(msg)
	}
	return nil
}

//newTestingItem New item
func newTestingItem() Item {
	return &testingItem{}
}

func TestDefaultPool(t *testing.T) {
	got := DefaultPool(newTestingItem)
	defConfig := DefaultConfig()

	if got.Config.InitialPoolSize != defConfig.InitialPoolSize ||
		got.Config.MinPoolSize != defConfig.MinPoolSize ||
		got.Config.MaxPoolSize != defConfig.MaxPoolSize ||
		got.Config.AcquireRetryAttempts != defConfig.AcquireRetryAttempts ||
		got.Config.AcquireIncrement != defConfig.AcquireIncrement ||
		got.Config.TestDuration != defConfig.TestDuration ||
		got.Config.TestOnGetItem != defConfig.TestOnGetItem ||
		got.newItemFunc == nil || got.items.Len() != 0 || got.shutdown {
		t.Errorf("FIND : %#v", got)
	}
}

func TestGetOne(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Initial()

	i, err := pool.GetOne()
	if err != nil {
		t.Error(err)
	}
	_, ok := i.(*testingItem)
	if !ok {
		t.Error(ErrTypeConvert)
	}
}

func TestGetOneWithExtend(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Initial()
	for i := 0; i < pool.Config.InitialPoolSize-pool.Config.MinPoolSize; i++ {
		_, err := pool.GetOne()
		if err != nil {
			t.Error(err)
		}
	}
	i, err := pool.GetOne()
	if err != nil {
		t.Error(err)
	}
	_, ok := i.(*testingItem)
	if !ok {
		t.Error(ErrTypeConvert)
	}
	time.Sleep(time.Second)
	if pool.items.Len() != pool.Config.MinPoolSize+pool.Config.AcquireIncrement-1 {
		t.Errorf("WANT item count : %d FIND : %d", pool.Config.MinPoolSize-1+pool.Config.AcquireIncrement, pool.items.Len())
	}
}

func TestExtend(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.extend(pool.Config.InitialPoolSize)
	if pool.items.Len() != pool.Config.InitialPoolSize {
		t.Errorf("WANT item count : %d FIND : %d", pool.Config.InitialPoolSize, pool.items.Len())
	}
}

func TestExtendOverMaxPoolSize(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.extend(pool.Config.MaxPoolSize + 1)
	if pool.items.Len() != pool.Config.MaxPoolSize {
		t.Errorf("WANT item count : %d FIND : %d", pool.Config.InitialPoolSize+pool.Config.AcquireIncrement, pool.items.Len())
	}
}

func TestExtendWithShutdown(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Initial()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pool.Shutdown(wg)
	wg.Wait()
	pool.extend(pool.Config.MaxPoolSize)
	if pool.items.Len() != 0 {
		t.Errorf("WANT item count : %d FIND : %d", 0, pool.items.Len())
	}
}

func TestBackOne(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Initial()

	i, err := pool.GetOne()
	if err != nil {
		t.Error(err)
	}
	item, ok := i.(*testingItem)
	if !ok {
		t.Error(ErrTypeConvert)
	}
	pool.BackOne(item)
	if pool.items.Len() != pool.Config.InitialPoolSize {
		t.Errorf("WANT item count : %d FIND : %d", pool.Config.InitialPoolSize, pool.items.Len())
	}
}

func TestBackOneWithDrop(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Config.Params["DestoryError"] = "Testing"
	pool.Initial()

	for i := 0; i < pool.Config.MaxPoolSize-pool.Config.InitialPoolSize; i++ {
		item := newTestingItem()
		pool.BackOne(item)
	}
	item := newTestingItem()
	pool.BackOne(item)
	if pool.items.Len() != pool.Config.MaxPoolSize {
		t.Errorf("WANT item count : %d FIND : %d", pool.Config.InitialPoolSize, pool.items.Len())
	}
}

func TestInitial(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Initial()

	if pool.items.Len() != pool.Config.InitialPoolSize {
		t.Errorf("WANT item count : %d FIND : %d", pool.Config.InitialPoolSize, pool.items.Len())
	}
}

func TestInitialWithError(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Config.Params["InitError"] = "Testing"
	pool.Initial()

	if pool.items.Len() != 0 {
		t.Errorf("WANT item count : %d \n FIND : %d", 0, pool.items.Len())
	}
	if item, err := pool.GetOne(); item != nil || err != ErrCanNotGetItem {
		t.Errorf("WANT : %#v , ErrCanNotGetItem FIND : %#v , %#v", nil, item, err)
	}
}

func TestShutdown(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Initial()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pool.Shutdown(wg)
	wg.Wait()
	item, err := pool.GetOne()
	if item != nil || err == nil || err != ErrHasBeenShotdown {
		t.Errorf("WANT : nil , ErrHasBeenShotdown FIND : %#v , %#v", item, err)
	}
	if pool.items.Len() != 0 || !pool.shutdown {
		t.Errorf("WANT : %#v , %#v FIND : %#v , %#v", 0, true, pool.items.Len(), pool.shutdown)
	}
}

func TestShutdownWithError(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Config.Params["DestoryError"] = "Testing"
	pool.Initial()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pool.Shutdown(wg)
	wg.Wait()
	if pool.items.Len() != 0 || !pool.shutdown {
		t.Errorf("WANT : %#v , %#v FIND : %#v , %#v", 0, true, pool.items.Len(), pool.shutdown)
	}
}

func TestCheckAvaiable(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.extend(pool.Config.InitialPoolSize)
	pool.checkAvaiable()
	i, err := pool.GetOne()
	if err != nil {
		t.Error(err)
	}
	item, ok := i.(*testingItem)
	if !ok {
		t.Error(ErrTypeConvert)
	}

	if !item.checked {
		t.Errorf("WANT : true Got  : %t ", item.checked)
	}
}

func TestCheckAvaiableWithError(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Config.Params["CheckError"] = "Testing"
	pool.Config.TestOnGetItem = true
	pool.extend(pool.Config.InitialPoolSize)
	pool.checkAvaiable()
	if pool.items.Len() != 0 {
		t.Errorf("WANT : 0 FIND : %#v", pool.items.Len())
	}

	wantErr := errors.New("Testing")
	item, err := pool.GetOne()
	if err == nil {
		t.Errorf("WANT : %#v FIND : %#v", err, nil)
	}
	if item != nil || !reflect.DeepEqual(wantErr, err) {
		t.Errorf("WANT : %#v , %#v FIND : %#v , %#v", nil, wantErr, item, err)
	}
}
func TestCheckAvaiableWithShutdown(t *testing.T) {
	pool := DefaultPool(newTestingItem)
	pool.Config.Params["CheckError"] = "Testing"
	pool.Config.TestOnGetItem = true
	pool.extend(pool.Config.InitialPoolSize)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pool.Shutdown(wg)
	wg.Wait()
	pool.checkAvaiable()
	if pool.items.Len() != 0 {
		t.Errorf("WANT : %d FIND : %d", 0, pool.items.Len())
	}
}
