# gpool

library for create pool easy , write in google go language 

## USAGE

### create pool item struct and implement Item interface 

Item interface function list

```go
Initial(map[string]string) error
Destory(map[string]string) error
Check(map[string]string) error
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

//Destory Destory Connection
func (c *Connection) Destory(params map[string]string) error {
	return c.TCPConn.Close()
}

//Check check item avaiable
func (c *Connection) Check(params map[string]string) error {
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
		pool = gpool.DefaultPool(NewConnection)
		pool.Config.LoadToml("general.toml")
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

```go
func ClosePool() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pool.Shutdown(wg)
	wg.Wait()
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
| Params               | item initial params                                              | map[string]string |         |

## Complete Example

here is a Complete Example : [gpool_example](https://github.com/cloudfstrife/gpool_example)
