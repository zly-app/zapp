
<!-- TOC -->

- [配置说明](#%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)
- [配置加载方式](#%E9%85%8D%E7%BD%AE%E5%8A%A0%E8%BD%BD%E6%96%B9%E5%BC%8F)
    - [从文件加载配置](#%E4%BB%8E%E6%96%87%E4%BB%B6%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
        - [框架配置示例](#%E6%A1%86%E6%9E%B6%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
        - [插件配置示例](#%E6%8F%92%E4%BB%B6%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
        - [服务配置示例](#%E6%9C%8D%E5%8A%A1%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
        - [组件配置示例](#%E7%BB%84%E4%BB%B6%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
        - [其它配置](#%E5%85%B6%E5%AE%83%E9%85%8D%E7%BD%AE)
    - [从viper加载配置](#%E4%BB%8Eviper%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
    - [从配置结构体加载配置](#%E4%BB%8E%E9%85%8D%E7%BD%AE%E7%BB%93%E6%9E%84%E4%BD%93%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
    - [引用配置文件](#%E5%BC%95%E7%94%A8%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)
    - [从apollo配置中心加载配置](#%E4%BB%8Eapollo%E9%85%8D%E7%BD%AE%E4%B8%AD%E5%BF%83%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
        - [apollo配置中心命名空间和配置说明](#apollo%E9%85%8D%E7%BD%AE%E4%B8%AD%E5%BF%83%E5%91%BD%E5%90%8D%E7%A9%BA%E9%97%B4%E5%92%8C%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)
        - [扁平化配置说明](#%E6%89%81%E5%B9%B3%E5%8C%96%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)
        - [在配置文件中设置从apollo配置中心加载](#%E5%9C%A8%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E4%B8%AD%E8%AE%BE%E7%BD%AE%E4%BB%8Eapollo%E9%85%8D%E7%BD%AE%E4%B8%AD%E5%BF%83%E5%8A%A0%E8%BD%BD)
- [配置观察](#%E9%85%8D%E7%BD%AE%E8%A7%82%E5%AF%9F)
    - [使用示例](#%E4%BD%BF%E7%94%A8%E7%A4%BA%E4%BE%8B)
    - [自动解析泛型结构示例-强烈推荐用法](#%E8%87%AA%E5%8A%A8%E8%A7%A3%E6%9E%90%E6%B3%9B%E5%9E%8B%E7%BB%93%E6%9E%84%E7%A4%BA%E4%BE%8B-%E5%BC%BA%E7%83%88%E6%8E%A8%E8%8D%90%E7%94%A8%E6%B3%95)
- [配置工具](#%E9%85%8D%E7%BD%AE%E5%B7%A5%E5%85%B7)
    - [用户白名单](#%E7%94%A8%E6%88%B7%E7%99%BD%E5%90%8D%E5%8D%95)

<!-- /TOC -->
---

# 配置说明

我们不需要任何配置就能直接跑起来, 当然你也可以使用配置

+ 配置来源优先级: 命令行`-c`指定文件 > WithViper > WithConfig > WithFiles > WithApollo > 默认配置文件
+ 使用命令 `-t` 来测试你的任何来源的配置是否正确.
+ 任何来源的配置都会构建为 [viper](https://github.com/spf13/viper) 结构, 然后再反序列化为配置结构体 [core.Config](../core/config.go)

# 配置加载方式

## 从文件加载配置

+ 一般使用yaml作为配置文件, 可以使用命令行 `-c` 手动指定配置文件, 如果有多个配置文件用英文逗号分隔
+ 也可以使用 `WithFiles` 在代码中指定配置文件
+ 多个配置文件如果存在同配置分片会智能合并
+ 如果下面的某个配置文件存在, 会按从上到下的优先级自动加载<b><font color='red'>一个</font></b>配置文件.

   ```
   ./configs/default.yaml
   ./configs/default.yml
   ./configs/default.toml
   ./configs/default.json
   ```

通用配置写法如下, 某些特定写法根据`它`的文档为准

+ 插件配置的key为 `plugins.{pluginType}`, pluginType是插件注册的类型值.
+ 服务配置的key为 `services.{serviceType}`, serviceType是服务注册的类型值.
+ 组件配置的key为 `components.{componentType}.{componentName}`, componentType是初始化组件时指定的类型值, componentName是获取组件时传入的名字.

### 框架配置示例

以下所有配置字段都是可选的

```yaml
frame: # 框架配置
    debug: true # debug标志
    Name: '' # app名
    Env: '' # 环境名
    FreeMemoryInterval: 120000 # 主动清理内存间隔时间(毫秒), <= 0 表示禁用
    WaitServiceRunTime: 1000 # 默认等待服务启动阶段, 等待时间(毫秒), 如果时间到未收到服务启动成功信号则将服务标记为不稳定状态然后继续开始工作(我们总不能一直等着吧)
    ServiceUnstableObserveTime: 10000 # 默认服务不稳定观察时间, 等待时间(毫秒), 如果时间到仍未收到服务启动成功信号也将服务标记为启动成功
    Flags: [] # flag, 注意: flag是忽略大小写的, 示例 ['a', 'B', 'c']
    Labels: # 标签, 注意: 标签名是忽略大小写的
        #Foo: Bar
    Log: # 日志配置
        Level: 'debug' # 日志等级, debug, info, warn, error, dpanic, panic, fatal
        Json: false # 启用json编码器, 输出的每一行日志转为json格式
        WriteToStream: true # 输出到屏幕
        WriteToFile: false # 日志是否输出到文件
        Name: '' # 日志文件名, 末尾会自动附加 .log 后缀
        AppendPid: false # 是否在日志文件名后附加进程号
        Path: './log' # 默认日志存放路径
        FileMaxSize: 32 # 每个日志最大尺寸,单位M
        FileMaxBackupsNum: 3 # 日志文件最多保存多少个备份, 0表示永久
        FileMaxDurableTime: 7 # 文件最多保存多长时间,单位天, 0表示永久
        Compress: false # 是否压缩历史日志
        TimeFormat: '2006-01-02 15:04:05' # 时间显示格式
        Color: true # 是否打印彩色日志等级, 只有关闭json编码器才生效
        CapitalLevel: false # 是否大写日志等级
        DevelopmentMode: true # 开发者模式, 在开发者模式下日志记录器在写完DPanic消息后程序会感到恐慌
        ShowFileAndLinenum: true # 显示文件路径和行号
        ShowFileAndLinenumMinLevel: 'debug' # 最小显示文件路径和行号的等级. 推荐所有等级都打印代码行, 相对于能快速定位问题来说, 这点性能损耗无关紧要
        MillisDuration: true # 对zap.Duration转为毫秒
    PrintConfig: true # app初始时是否打印配置
```

### 插件配置示例
```yaml
plugins: # 插件配置
    my_plugin: # 插件类型
        Foo: Bar # 示例, 不代表真实插件配置情况
```

### 服务配置示例
```yaml
services: # 服务配置
    api: # 服务类型
        Bind: :8080 # 示例, 不代表真实插件配置情况
    grpc: # 服务类型
        Bind: :3000 # 示例, 不代表真实插件配置情况
```

### 组件配置示例

```yaml
components: # 组件配置
    cache: # 组件类型
        cacheName1: # 组件名, 比如提供多个redis客户端连接不同的redis集群
            CacheDB: memory # 示例, 不代表真实插件配置情况
        cacheName2: # 组件名, 比如提供多个redis客户端连接不同的redis集群
            CacheDB: memory # 示例, 不代表真实插件配置情况
```

### 其它配置

> 除了 frame / plugins / services / components 这几类, 还可以添加自定义配置, 然后使用 `Parse` 方法将配置读取到变量中

```yaml
myconfig: #自定义分片名
    Foo: Bar # 示例, 不代表真实情况
```

+ 更多配置说明阅读源码 [core.Config](../core/config.go)

## 从viper加载配置

> 使用 `WithViper` 设置 [viper](https://github.com/spf13/viper) 结构

## 从配置结构体加载配置

> 使用 `WithConfig` 设置配置结构体 [core.Config](../core/config.go)

## 引用配置文件

+ 可以在配置中引用另一个配置文件, 允许使用相对路径, 它相对于程序运行时当前目录.
+ 被引用的配置文件中不能再添加引用了, 它不会被识别.
+ 引用的配置文件必须存在

示例:

```yaml
include:
    files: './1.toml,./2.toml'
```

## 从`apollo`配置中心加载配置

> 使用 `WithApollo` 设置apollo来源和如何加载

### `apollo`配置中心命名空间和配置说明

zapp会将命名空间的名称作为配置顶级key, 该命名空间的配置会作为二级key和其值.

示例 apollo 配置数据:

 | 命名空间   | field | value          |
 | ---------- | ----- | -------------- |
 | namespace1 | key1  | value          |
 | namespace1 | key2  | {"foo": "bar"} |

以上apollo配置数据会被解析为以下配置

```yaml
namespace1:
  key1: 'value'
  key2: '{"foo": "bar"}'
```

zapp会将 `applicaiont` 命名空间下的 `frame`,`components`,`plugins`,`services`配置作为配置顶级key, 并将其值按照指定格式解析后赋予其子集(默认是`yaml`格式解析, 需要在apollo配置中设为你需要的格式).

示例 apollo 配置数据:

 | 命名空间    | field    | value                                     | 备注                   |
 | ----------- | -------- | ----------------------------------------- | ---------------------- |
 | application | frame    | {"debug": true, "log": {"level": "info"}} | 会按照指定格式解析其值 |
 | application | plugins  | {"myplugin": {"foo": "bar"}}              | 会按照指定格式解析其值 |
 | application | otherKey | {"debug": true, "log": {"level": "info"}} | 不会解析其值           |

以上apollo配置数据会被解析为以下配置

```yaml
frame:
  debug: true
  log:
    level: 'info'
plugins:
  foo: 'bar'
application:
  otherKey: '{"debug": true, "log": {"level": "info"}}'
```

### 扁平化配置说明

上一步所有配置都放在命名空间`application`中的, 这里支持将 `frame`/`components`/`plugins`/`filters`/`services`直接作为命名空间来配置数据. 如下

示例 apollo 配置数据:

| 命名空间     | content                                   | 备注                   |
| ------------ | ----------------------------------------- | ---------------------- |
| frame.json   | {"debug": true, "log": {"level": "info"}} | 会按照指定格式解析其值 |
| plugins.json | {"myplugin": {"foo": "bar"}}              | 会按照指定格式解析其值 |

当前其它在配置`ApplicationParseKeys`中指定的命名空间也会被解析.

注意. 此处不支持 `properties` 格式, 在创建Namespace时需要选择匹配的格式

### 在配置文件中设置从`apollo`配置中心加载

> 文件中添加如下设置, 参考 [config.ApolloConfig](./apollo.go). 从apollo拉取的配置会和文件的配置智能合并, 以apollo配置优先

最小配置

```yaml
apollo:
    Address: "http://127.0.0.1:8080"
    AppId: "your-appid"
```

完整配置

```yaml
apollo:
    Address: "http://127.0.0.1:8080"
    AppId: "your-appid"
    AccessKey: ""                  # 验证key, 优先级高于基础认证
    AuthBasicUser: ""              # 基础认证用户名, 可用于nginx的基础认证扩展
    AuthBasicPassword: ""          # 基础认证密码
    Cluster: "default"             # 集群名, 默认default
    AlwaysLoadFromRemote: false    # 总是从远程获取, 在远程加载失败时不会从备份文件加载
    BackupFile: "./configs/backup.apollo" # 本地备份文件, 留空表示不使用备份
    ApplicationDataType: "yaml"    # application命名空间下key的值的数据类型, 支持yaml,yml,toml,json
    ApplicationParseKeys: []       # application命名空间下哪些key数据会被解析, 无论如何默认的key(frame/components/plugins/filters/services)会被解析
    Namespaces: []                 # 其他自定义命名空间
    IgnoreNamespaceNotFound: false # 是否忽略命名空间不存在, 无论如何设置application命名空间必须存在
```

---

# 配置观察

拥有以下特性

+ <b>一行代码接入配置中心, 解放心智负担</b>
+ 自动监听配置变更, 每次获取时拿到的是最新的配置数据
+ 开始watch时会立即从远程获取一次数据, 如果失败会立即打印`Fatal`日志并退出程序(配置有问题程序不应该启动)

## 使用示例

```go
package main

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/plugin/apollo_provider"
)

// 可以在定义变量时初始化
var MyConfigWatch = zapp.WatchConfigKey("group_name", "key_name")

func main() {
	app := zapp.NewApp("test",
		// 使用apollo配置中心作为配置提供者并设置为默认的配置提供者
		apollo_provider.WithPlugin(true),
	)
	defer app.Exit()

	// 也可以在这里初始化
	//MyConfigWatch = zapp.WatchConfigKey("group_name", "key_name")

	// 获取原始数据
	y1 := MyConfigWatch.GetString()
	app.Info(y1) // 1

	// 转为 int 值
	y2 := MyConfigWatch.GetInt()
	app.Info(y2) // 1

	// 转为 boolean 值
	y3 := MyConfigWatch.GetBool()
	app.Info(y3) // true

	// 检查复合预期
	b1 := MyConfigWatch.Expect("1")
	b2 := MyConfigWatch.Expect(1)
	b3 := MyConfigWatch.Expect(true)
	app.Info(b1, b2, b3) // true, true, true

	// 添加回调函数
	MyConfigWatch.AddCallback(func(first bool, oldData, newData []byte) {
		app.Info("回调",
			zap.Bool("first", first),
			zap.String("oldData", string(oldData)),
			zap.String("newData", string(newData)),
		)
	})

	app.Run()
}
```

## 自动解析泛型结构示例-`(强烈推荐用法)`

通过 `config.WatchKeyStruct` 观察一个key数值变更, 拥有以下额外特性

+ 自动将配置数据作为指定格式解析到一个类型中, 通过`Get`能直接拿到想要的配置数据
+ 开始watch时会立即对数据结构进行解析, 如果失败会立即打印`Fatal`日志并退出程序(配置有问题程序不应该启动)
+ 当配置变更时如果解析失败会打印一个`Error`日志并忽略该配置变更(获取到的配置是上一次正确解析的配置数据)
+ 只有配置变更时才会解析数据, 并不是每次获取数据都解析一次

```go
package main

import (
	"time"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/plugin/apollo_provider"
)

type MyConfig struct {
	A int `json:"a"`
}

// 可以在定义变量时初始化
var MyConfigWatch = zapp.WatchConfigJson[*MyConfig]("group_name", "generic_key")

func main() {
	app := zapp.NewApp("test",
		// 使用apollo配置中心作为配置提供者并设置为默认的配置提供者
		apollo_provider.WithPlugin(true),
	)
	defer app.Exit()

	// 也可以在这里初始化
	//MyConfigWatch = zapp.WatchConfigJson[*MyConfig]("group_name", "generic_key")

	// 获取数据
	for {
		a := MyConfigWatch.Get()
		app.Info("数据", a)
		time.Sleep(time.Second)
	}
}
```

对于 `apollo` 中的非 `properties` 类型命名空间, 其 `group_name={{namespace}}.{{type}}` 其 `key=content`. 如:

```go
// 先创建一个命名空间为 watch 的 json 类型. 其key固定为"content"
var MyConfigWatch = zapp.WatchConfigJson[*MyConfig]("watch.json", "content")
```

[其它示例代码](./watch_example)

---

# 配置工具

## 用户白名单

功能支持

- [x] 指定用户id
- [x] 指定尾号(后两位)
- [x] 灰度比例(1%细粒度)


示例

```go
package main

import (
	"fmt"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/plugin/apollo_provider"
)

type MyConfig struct {
	A                  string              // 其它业务需要的字段
	B                  string              // 其它业务需要的字段
	zapp.UserWhiteList                     // 直接继承这个结构
	UserWhitelist2     zapp.UserWhiteList  // 另一个白名单
	UserWhitelist3     *zapp.UserWhiteList // 又一个白名单
}

// 定义一个获取配置的函数
var GetMyConfig = zapp.WatchConfigJson[*MyConfig]("group_name", "generic_key")

func main() {
	app := zapp.NewApp("test",
		// 使用apollo配置中心作为配置提供者并设置为默认的配置提供者
		apollo_provider.WithPlugin(true),
	)
	defer app.Exit()

	v := GetMyConfig.Get() // 获取数据

	// 检查用户是否在白名单中
	fmt.Println(v.IsWhiteList("111"))   // true
	fmt.Println(v.IsWhiteList("222"))   // true
	fmt.Println(v.IsWhiteList("12301")) // true
	fmt.Println(v.IsWhiteList("12302")) // true

	// 检查用户是否在白名单中
	fmt.Println(v.UserWhitelist2.IsWhiteList("111"))   // true
	fmt.Println(v.UserWhitelist2.IsWhiteList("222"))   // true
	fmt.Println(v.UserWhitelist2.IsWhiteList("12301")) // true
	fmt.Println(v.UserWhitelist2.IsWhiteList("12302")) // true
}
```

配置如下

```json
{
    "Uids": [
        "111",
        "222"
    ],
    "Percent": 0,
    "Tails": [
        "01",
        "02"
    ],

    "UserWhitelist2": {
      "Uids": [
        "111",
        "222"
      ],
      "Percent": 0,
      "Tails": [
        "01",
        "02"
      ]
    }
}
```

底层结构说明

```go
// 用户白名单, 多个数据同时存在时只要匹配任意一个条件就行
type UserWhiteList struct {
	Uids    []string // 指定的用户
	Percent uint8    // 灰度比例, 百分比, 0~100
	Tails   []string // 用户后两位尾号
}
```
