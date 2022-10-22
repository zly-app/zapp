
<!-- TOC -->

- [我们不需要任何配置就能跑起来](#%E6%88%91%E4%BB%AC%E4%B8%8D%E9%9C%80%E8%A6%81%E4%BB%BB%E4%BD%95%E9%85%8D%E7%BD%AE%E5%B0%B1%E8%83%BD%E8%B7%91%E8%B5%B7%E6%9D%A5)
- [服务,插件和组件配置说明](#%E6%9C%8D%E5%8A%A1%E6%8F%92%E4%BB%B6%E5%92%8C%E7%BB%84%E4%BB%B6%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)
- [从文件加载配置](#%E4%BB%8E%E6%96%87%E4%BB%B6%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
    - [框架配置示例](#%E6%A1%86%E6%9E%B6%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
    - [插件配置示例](#%E6%8F%92%E4%BB%B6%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
    - [服务配置示例](#%E6%9C%8D%E5%8A%A1%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
    - [组件配置示例](#%E7%BB%84%E4%BB%B6%E9%85%8D%E7%BD%AE%E7%A4%BA%E4%BE%8B)
    - [其它配置](#%E5%85%B6%E5%AE%83%E9%85%8D%E7%BD%AE)
- [从viper加载配置](#%E4%BB%8Eviper%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
- [从配置结构体加载配置](#%E4%BB%8E%E9%85%8D%E7%BD%AE%E7%BB%93%E6%9E%84%E4%BD%93%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
- [从apollo加载配置](#%E4%BB%8Eapollo%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
    - [apollo命名空间和配置说明](#apollo%E5%91%BD%E5%90%8D%E7%A9%BA%E9%97%B4%E5%92%8C%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)
    - [在配置文件中设置从apollo加载](#%E5%9C%A8%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E4%B8%AD%E8%AE%BE%E7%BD%AE%E4%BB%8Eapollo%E5%8A%A0%E8%BD%BD)
- [引用配置文件](#%E5%BC%95%E7%94%A8%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)
- [远程配置变更观察](#%E8%BF%9C%E7%A8%8B%E9%85%8D%E7%BD%AE%E5%8F%98%E6%9B%B4%E8%A7%82%E5%AF%9F)
    - [内置apollo提供者示例](#%E5%86%85%E7%BD%AEapollo%E6%8F%90%E4%BE%9B%E8%80%85%E7%A4%BA%E4%BE%8B)

<!-- /TOC -->
---

# 我们不需要任何配置就能跑起来

+ 配置来源优先级: 命令行`-c`指定文件 > WithViper > WithConfig > WithFiles > WithApollo > 默认配置文件
+ 使用命令 `-t` 来测试你的任何来源的配置是否正确.
+ 任何来源的配置都会构建为 [viper](https://github.com/spf13/viper) 结构, 然后再反序列化为配置结构体 [core.Config](../core/config.go)

# 服务,插件和组件配置说明

+ 插件配置的key为 `plugins.{pluginType}`, pluginType是插件注册的类型值.
+ 服务配置的key为 `services.{serviceType}`, serviceType是服务注册的类型值.
+ 组件配置的key为 `components.{componentType}.{componentName}`, componentType是初始化组件时指定的类型值, componentName是获取组件时传入的名字.

# 从文件加载配置

+ 一般使用yaml作为配置文件, 可以使用命令行 `-c` 手动指定配置文件, 如果有多个配置文件用英文逗号分隔
+ 可以使用 `WithFiles` 在代码中指定配置文件
+ 多个配置文件如果存在同配置分片会智能合并
+ 如果下面的某个配置文件存在, 会按从上到下的优先级自动加载<b><font color='red'>一个</font></b>配置文件.

   ```
   ./configs/default.yaml
   ./configs/default.yml
   ./configs/default.toml
   ./configs/default.json
   ```

## 框架配置示例

以下所有配置字段都是可选的

```yaml
frame: # 框架配置
    debug: true # debug标志
    Name: '' # app名
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
        ShowFileAndLinenumMinLevel: 'warn' # 最小显示文件路径和行号的等级
        MillisDuration: true # 对zap.Duration转为毫秒
    PrintConfig: true # app初始时是否打印配置
```

## 插件配置示例
```yaml
plugins: # 插件配置
    my_plugin: # 插件类型
        Foo: Bar # 示例, 不代表真实插件配置情况
```

## 服务配置示例
```yaml
services: # 服务配置
    api: # 服务类型
        Bind: :8080 # 示例, 不代表真实插件配置情况
    grpc: # 服务类型
        Bind: :3000 # 示例, 不代表真实插件配置情况
```

## 组件配置示例

```yaml
components: # 组件配置
    cache: # 组件类型
        cacheName1: # 组件名, 比如提供多个redis客户端连接不同的redis集群
            CacheDB: memory # 示例, 不代表真实插件配置情况
        cacheName2: # 组件名, 比如提供多个redis客户端连接不同的redis集群
            CacheDB: memory # 示例, 不代表真实插件配置情况
```

## 其它配置

> 除了 frame / plugins / services / components 这几类, 还可以添加自定义配置, 然后使用 `Parse` 方法将配置读取到变量中

```yaml
myconfig: #自定义分片名
    Foo: Bar # 示例, 不代表真实情况
```

+ 更多配置说明阅读源码 [core.Config](../core/config.go)

# 从viper加载配置

> 使用 `WithViper` 设置 [viper](https://github.com/spf13/viper) 结构

# 从配置结构体加载配置

> 使用 `WithConfig` 设置配置结构体 [core.Config](../core/config.go)

# 从apollo加载配置

> 使用 `WithApollo` 设置apollo来源和如何加载

## apollo命名空间和配置说明

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

## 在配置文件中设置从apollo加载

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
    ApplicationParseKeys: []       # application命名空间下哪些key数据会被解析, 无论如何默认的key(frame/components/plugins/services)会被解析
    Namespaces: []                 # 其他自定义命名空间
    IgnoreNamespaceNotFound: false # 是否忽略命名空间不存在, 无论如何设置application命名空间必须存在
```

# 引用配置文件

可以在配置中引用另一个配置文件, 可以使用相对路径, 相对于程序运行时当前目录.

被引用的配置文件中不能再添加引用了, 它不会被识别.

引用的配置文件必须存在.

```yaml
include:
    files: './1.toml,./2.toml'
```

# 远程配置变更观察

## 内置apollo提供者示例

自动解析泛型结构示例

通过 `config.WatchKeyStruct` 观察一个key数值变更, 每次变更时会自动解析一次数据. 当解析失败会忽略该数据并打印一个错误日志, 返回的配置数据不会产生变化.

```go
package main

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/plugin/apollo_provider"
)

func main() {
	app := zapp.NewApp("test",
		apollo_provider.WithPlugin(true), // 启用apollo配置提供者并设置为默认提供者
	)
	defer app.Exit()

	type AA struct {
		A int `json:"a"`
	}

	// 获取key对象
	keyObj := config.WatchKeyStruct[*AA]("group_name", "generic_key", config.WithWatchStructJson())
	a := keyObj.Get()
	app.Info("数据", a)

	keyObj.AddCallback(func(first bool, oldData, newData *AA) {
		app.Info("回调",
			zap.Bool("first", first),
			zap.Any("oldData", oldData),
			zap.Any("newData", newData),
		)
	})
	app.Run()
}
```

手动解析配置数据示例

```go
package main

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin/apollo_provider"
)

func main() {
	app := zapp.NewApp("test",
		apollo_provider.WithPlugin(true), // 启用apollo配置提供者并设置为默认提供者
	)
	defer app.Exit()

	callback := func(w core.IConfigWatchKeyObject, first bool, oldData, newData []byte) {
		app.Info("回调",
			zap.String("groupName", w.GroupName()),
			zap.String("keyName", w.KeyName()),
			zap.Bool("first", first),
			zap.String("oldData", string(oldData)),
			zap.String("newData", string(newData)),
		)
	}
	config.WatchKey("watch", "a").AddCallback(callback)
	config.WatchKey("watch", "b").AddCallback(callback)
	config.WatchKey("watch2", "a").AddCallback(callback)

	app.Run()
}
```

[其它示例代码](./watch_example)

