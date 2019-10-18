# gpool

library for create pool easy , write in google go language 

## USAGE

### create pool item struct and implement Item interface 

Item interface function list

```go
Initial(map[string]string) error
Destory() error
Check() error
```

#### example

```go
//Connection pool item struct
type Connection struct {
	TCPConn net.Conn
}

//Initial Initial operation
func (c *Connection) Initial(params map[string]string) error {
	con, err := net.Dial("tcp", params["host"]+":"+params["port"])
	if err != nil {
		return err
	}
	c.TCPConn = con
	return nil
}

//Destory Destory operation
func (c *Connection) Destory() error {
	return c.TCPConn.Close()
}

//Check check item avaiable
func (c *Connection) Check() error {
	fmt.Println("Check item Avaiable")
	return nil
}
```

### create item factory 

```go
//NewConnection New item 
func NewConnection() gpool.Item {
	return &Connection{}
}
```

### create Singleton pool 

```go
var (
	pool *gpool.Pool
	once sync.Once
)

func init() {
	once.Do(func() {
		pool = gpool.DefaultPool()
		pool.Config.LoadToml("general.toml")

		fmt.Println(pool.Config)
		pool.NewFunc = NewConnection
		pool.Initial()

	})
}
```

### implement get Item and give back item 

```go
//GetConnection Get item Connection
func GetConnection() (net.Conn, error) {
	item, err := pool.GetOne()
	if err != nil {
		return nil, err
	}
	con, ok := item.(*Connection)
	if ok {
		return con.TCPConn, nil
	}
	return nil, errors.New("Class cast ERROR")
}

//CloseConnection back item Connection
func CloseConnection(conn net.Conn) {
	pool.BackOne(&Connection{
		TCPConn: conn,
	})
}
```

### implement close pool

```
//ClosePool shutdown the pool
func ClosePool() {
	pool.Shutdown()
}
```

### use pool

omit

## Config

| Name                 | Description                                                      | Type              | Default |
| -------------------- | ---------------------------------------------------------------- | ----------------- | ------- |
| InitialPoolSize      | initial pool size.										          | int               | 5       |
| MinPoolSize          | min item in pool.                                                | int               | 2       |
| MaxPoolSize          | max item in pool.                                                | int               | 15      |
| AcquireRetryAttempts | retry times when get item Failed.                                | int               | 5       |
| AcquireIncrement     | create item count when pool is empty.                            | int               | 5       |
| TestDuration         | interval time between check item avaiable.Unit:Millisecond       | int               | 1000    |
| TestOnGetItem        | test avaiable when get item.                                     | bool              | false   |
| Debug                | show debug information.                                          | bool              | false   |
| Params               | item initial params                                              | map[string]string |         |

## Complete Example

here is a Complete Example

### File Tree 

```
.
|-- cmd
|   |-- client
|   |   `-- client.go
|   `-- server
|       `-- server.go
|-- dial
|   |-- connection.go
|   `-- dial.go
|-- general.toml
|-- go.mod
`-- go.sum

```

**cmd/client/client.go**

```
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"app/testing/dial"
)

func main() {
	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				send(i, j)
			}

		}(i)
	}
	wg.Wait()
	dial.ClosePool()
	end := time.Now()
	fmt.Println(end.Sub(start))
}

func send(i, j int) {
	conn, err := dial.GetConnection()
	if err != nil {
		log.Fatalf("第%d个线程获取连接失败%v", i, err)
	}
	defer dial.CloseConnection(conn)
	_, err = conn.Write([]byte(fmt.Sprintf("%d %d\n", i, j)))
	if err != nil {
		log.Fatal(err)
	}
}

```
**cmd/server/server.go**

```
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:11211")
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for {
		conn, err := listener.Accept()
		i++
		if err != nil {
			log.Fatal(err)
		}
		go Proc(conn, i)
	}

}

//Proc Proc
func Proc(conn net.Conn, i int) {
	defer conn.Close()
	buf := bufio.NewReader(conn)
	for {
		bs, _, err := buf.ReadLine()
		if err != nil {
			return
		}
		v := string(bs)
		slist := strings.Split(v, " ")

		if len(slist) == 2 {
			if slist[1] == "999" {
				fmt.Printf("第%d个连接收到消息 : 线程 %s 第 %s 次发送\n", i, slist[0], slist[1])
			}
		} else {
			fmt.Println(v)
		}
	}
}

```

**dial/dial.go**

```
package dial

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/cloudfstrife/gpool"
)

var (
	pool *gpool.Pool
	once sync.Once
)

func init() {
	once.Do(func() {
		pool = gpool.DefaultPool(NewConnection)
		pool.Config.LoadToml("general.toml")
		fmt.Println(pool.Config)
		pool.Initial()
	})
}

//NewConnection 获取新连接
func NewConnection() gpool.Item {
	return &Connection{}
}

//GetConnection 获取连接
func GetConnection() (net.Conn, error) {
	item, err := pool.GetOne()
	if err != nil {
		return nil, err
	}
	con, ok := item.(*Connection)
	if ok {
		return con.TCPConn, nil
	}
	return nil, errors.New("类型转换错误")
}

//CloseConnection 关闭连接
func CloseConnection(conn net.Conn) {
	pool.BackOne(&Connection{
		TCPConn: conn,
	})
}

//ClosePool 关闭连接池
func ClosePool() {
	pool.Shutdown()
}

```

**dial/connection.go**

```
package dial

import (
	"fmt"
	"net"
)

//Connection 连接池对象
type Connection struct {
	TCPConn net.Conn
}

//Initial 初始化
func (c *Connection) Initial(params map[string]string) error {
	con, err := net.Dial("tcp", params["host"]+":"+params["port"])
	if err != nil {
		return err
	}
	c.TCPConn = con
	return nil
}

//Destory 销毁连接
func (c *Connection) Destory() error {
	return c.TCPConn.Close()
}

//Check 检查元素连接是否可用
func (c *Connection) Check() error {
	fmt.Println("检查连接可用")
	return nil
}

```

**general.toml**

```
InitialPoolSize = 5
MinPoolSize = 2
MaxPoolSize = 15
AcquireRetryAttempts = 5
AcquireIncrement = 5
TestDuration = 60000
TestOnGetItem = false
Debug = false

[Params]
  host = "127.0.0.1"
  port = "11211"
```

**go.mod**

```
module app/testing

go 1.13

require (
	github.com/cloudfstrife/gpool v0.0.2
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/sys v0.0.0-20191010194322-b09406accb47 // indirect
)

```

### RUN & TEST

```
go build app/testing/cmd/server
./server

```
---
```
go build app/testing/cmd/client
./client
```
