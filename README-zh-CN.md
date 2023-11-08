# gpool

Go语言实现，用于快速构建资源池的库

## 使用方法

### 创建池元素struct并实现Item接口方法

需要实现的方法签名列表

```go
Initial(map[string]string) error
Destory(map[string]string) error
Check(map[string]string) error
```

#### 示例

```go
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
func (c *Connection) Destory(params map[string]string) error {
	return c.TCPConn.Close()
}

//Check 检查元素连接是否可用
func (c *Connection) Check(params map[string]string) error {
	fmt.Println("Check item Avaiable")
	return nil
}
```

### 实现创建池元素工厂方法

```go
//NewConnection 获取新连接
func NewConnection() gpool.Item {
	return &Connection{}
}
```

### 创建单例模式的Pool

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

### 实现获取池元素与归还池元素的方法

```go
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
```

### 实现关闭池方法

```go
//ClosePool 关闭连接池
func ClosePool() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pool.Shutdown(wg)
	wg.Wait()
}
```

### 使用

略

## 配置说明

| 名称                 | 说明                                                     | 类型              | 默认值 |
| -------------------- | -------------------------------------------------------- | ----------------- | ------ |
| InitialPoolSize      | 初始化池中元素数量，取值应在MinPoolSize与MaxPoolSize之间 | int               | 5      |
| MinPoolSize          | 池中保留的最小元素数量                                   | int               | 2      |
| MaxPoolSize          | 池中保留的最大连元素数量                                 | int               | 15     |
| AcquireRetryAttempts | 定义在新连接失败后重复尝试的次数                         | int               | 5      |
| AcquireIncrement     | 当池中的元素耗尽时，一次同时创建的元素数                 | int               | 5      |
| TestDuration         | 连接有效性检查间隔，单位毫秒                             | int               | 1000   |
| TestOnGetItem        | 如果设为true那么在取得元素的同时将校验元素的有效性       | bool              | false  |
| Params               | 元素初始化参数                                           | map[string]string |        |

## 示例

一个示例展示这个库的用法 : [gpool_example](https://github.com/cloudfstrife/gpool_example)
