
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
        - [apollo配置json支持](#apollo%E9%85%8D%E7%BD%AEjson%E6%94%AF%E6%8C%81)
    - [在配置文件中设置从apollo加载](#%E5%9C%A8%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E4%B8%AD%E8%AE%BE%E7%BD%AE%E4%BB%8Eapollo%E5%8A%A0%E8%BD%BD)
- [引用配置文件](#%E5%BC%95%E7%94%A8%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)

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

+ 一般使用toml作为配置文件, 可以使用命令行 `-c` 手动指定配置文件, 如果有多个配置文件用英文逗号分隔
+ 可以使用 `WithFiles` 在代码中指定配置文件
+ 多个配置文件如果存在同配置分片会智能合并

## 框架配置示例

```toml
[frame]
Debug = true # debug 标志
FreeMemoryInterval = 120000 # 主动清理内存间隔时间(毫秒), <= 0 表示禁用
#...
```

## 插件配置示例
```toml
[plugins.zipkin]
A = 1
B = "v"

#[...]
```

## 服务配置示例
```toml
[services.api]
Bind = ":8080"
IPWithNginxForwarded = false
IPWithNginxReal = false

[services.grpc]
Bind = ":3000"
HeartbeatTime = 20000

#[...]
```

## 组件配置示例
```toml
[components.cache.default]
CacheDB = "memory"
Codec = "msgpack"
DirectReturnOnCacheFault = true
MemoryCacheDB.CleanupInterval = 300000

#[...]
```

## 其它配置

> 除了 frame; plugins services; components 这几类, 还可以添加自定义配置, 然后使用 `Parse` 方法将配置读取到变量中

```toml
[自定义分片名]
key=value
```

+ 更多配置说明阅读源码 [core.Config](../core/config.go)

# 从viper加载配置

> 使用 `WithViper` 设置 [viper](https://github.com/spf13/viper) 结构

# 从配置结构体加载配置

> 使用 `WithConfig` 设置配置结构体 [core.Config](../core/config.go)

# 从apollo加载配置

> 使用 `WithApollo` 设置apollo来源和如何加载

## apollo命名空间和配置说明

```text
apollo命名空间主要为以下部分:
    frame: 框架配置
    plugins: 插件配置
    services: 服务配置
    components: 组件配置
    当然你也通过设置 ApolloConfig.Namespaces 以加载自定义命名空间
apollo的配置是扁平化的, 多级的key应该用点连接起来, 所以配置应该类似于这样:
    frame:
        Debug                   true            debug标志
        FreeMemoryInterval      120000          清理内存间隔时间(毫秒)
        ...
        Log.Level               debug           日志等级, debug, info, warn, error, dpanic, panic, fatal
        Log.WriteToStream       true            输出到屏幕
        ...
    plugins:
        zipkin.A                1               ...
        zipkin.B                v               ...
        ...
    services:
        Api.Bind                :8080           ...
        ...
        Grpc.Bind               :3000           ...
        ...
    components:
        Xorm.default.Driver     mysql           ...
        ...
        Redis.default.Address   127.0.0.1:6379  ...
        ...
apollo的配置也可以使用json, 如下:
    frame:
        Debug                   true            debug标志
        FreeMemoryInterval      120000          清理内存间隔时间(毫秒)
        ...
        Log                     {json配置}
    plugins:
        zipkin                  {json配置}
        ...
    services:
        Api                     {json配置}
        Grpc                    {json配置}
        ...
    components:
        Xorm.default            {json配置}
        Redis.default           {json配置}
        ...
```

### apollo配置json支持

> 由于配置项越来越多, 扁平化的apollo变得不是很好管理, 我们支持将多个配置key合并为一个json值. apollo配置示例:

    ```toml
    [frame]
    Debug = true
    Log = {
         "Level": "info",
         ...
      }
    
    [plugins]
    zipkin = {
        ...
      }
    
    [services]
    Api = {
        ...
      }
    
    [components]
    Xorm.default = {
        ...
      }
    ```

## 在配置文件中设置从apollo加载

> 文件中添加如下设置, 参考 [config.ApolloConfig](./apollo_sdk/sdk.go). 从apollo拉取的配置会和文件的配置智能合并, 以apollo配置优先

    ```toml
    [apollo]
    Address = "http://127.0.0.1:8080"
    AppId = "your-appid"
    AccessKey = ""                  # 验证key, 优先级高于基础认证
    AuthBasicUser = ""              # 基础认证用户名, 可用于nginx的基础认证扩展
    AuthBasicPassword = ""          # 基础认证密码
    Cluster = "default"             # 集群名, 默认default
    AlwaysLoadFromRemote = false    # 总是从远程获取, 在远程加载失败时不会从备份文件加载
    BackupFile = "./configs/backup.apollo" # 本地备份文件, 留空表示不使用备份
    NamespacePrefix = ""            # 命名空间前缀, apollo支持的部门前缀
    Namespaces = ""                 # 其他自定义命名空间, 多个命名空间用英文逗号隔开
    ```

# 引用配置文件

可以在配置中引用另一个配置文件, 可以使用相对路径, 相对于程序运行时当前目录

被引用的配置文件中不能再添加引用了, 它不会被识别

引用的配置文件必须存在

```toml
[include]
files = './1.toml,./2.toml'
```
