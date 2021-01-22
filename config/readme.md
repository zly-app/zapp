
<!-- TOC -->

- [服务和组件配置说明](#%E6%9C%8D%E5%8A%A1%E5%92%8C%E7%BB%84%E4%BB%B6%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)
- [从文件加载配置](#%E4%BB%8E%E6%96%87%E4%BB%B6%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
- [从viper加载配置](#%E4%BB%8Eviper%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
- [从配置结构体加载配置](#%E4%BB%8E%E9%85%8D%E7%BD%AE%E7%BB%93%E6%9E%84%E4%BD%93%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
- [从apollo加载配置](#%E4%BB%8Eapollo%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE)
    - [apollo命名空间和配置说明](#apollo%E5%91%BD%E5%90%8D%E7%A9%BA%E9%97%B4%E5%92%8C%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)
    - [配置文件和apollo混用](#%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E5%92%8Capollo%E6%B7%B7%E7%94%A8)

<!-- /TOC -->
---

> 配置来源优先级: 命令行指定文件 > WithViper > WithConfig > WithFiles > WithApollo > 默认配置文件
> 任何来源的配置都会构建为 [viper](https://github.com/spf13/viper) 结构, 然后再反序列化为配置结构体 [core.Config](../core/config.go)

# 服务和组件配置说明

> 服务配置的key为 services.serviceType, serviceType是服务注册的类型值.

> 组件配置的key为 components.componentType.componentName, componentType是初始化组件时指定的类型值, componentName是获取组件时传入的名字.

# 从文件加载配置

> 1.一般使用toml作为配置文件, 可以使用命令行 `-c` 手动指定配置文件, 如果有多个配置文件用英文逗号分隔
> 2.可以使用 `WithFiles` 在代码中指定配置文件
> 3.多个配置文件如果存在同配置分片会智能合并, 从apollo拉取的配置会覆盖相同的文件配置节点

+ 框架配置示例
```toml
[frame]
Debug = true # debug 标志
FreeMemoryInterval = 120000 # 主动清理内存间隔时间(毫秒), <= 0 表示禁用
#...
```

+ 服务配置示例
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

+ 组件配置示例
```toml
[components.cache.default]
CacheDB = "memory"
Codec = "msgpack"
DirectReturnOnCacheFault = true
MemoryCacheDB.CleanupInterval = 300000

#[...]
```

+ 其它配置

> 除了 frame; services; components 这三大类, 还可以添加自定义配置, 然后使用 Parse 方法将配置读取到变量中

```toml
[自定义分片名]
key=value
```

+ 更多配置说明参考 [core.Config](../core/config.go)

# 从viper加载配置

> 使用 `WithViper` 设置 [viper](https://github.com/spf13/viper) 结构

# 从配置结构体加载配置

> 使用 `WithConfig` 设置配置结构体 `core.Config`

# 从apollo加载配置

> 使用 `WithApollo` 设置apollo来源和如何加载

## apollo命名空间和配置说明

```text
apollo命名空间主要为三部分:
    frame: 框架配置
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
    services:
        Api                     {json配置}
        Grpc                    {json配置}
        ...
    components:
        Xorm.default            {json配置}
        Redis.default           {json配置}
        ...
```

## 配置文件和apollo混用

> 从apollo拉取的配置会覆盖文件的配置

1. 文件中添加如下设置, 参考 [config.ApolloConfig](./apollo.go)

    ```toml
    [apollo]
    Address = "http://127.0.0.1:8080"
    AppId = "your-appid"
    AccessKey = "" # 验证key, 优先级高于基础认证
    AuthBasicUser = "" # 基础认证用户名
    AuthBasicPassword = "" # 基础认证密码
    Cluster = "default" # 集群名
    AlwaysLoadFromRemote = false # 总是从远程获取, 在远程加载失败时不会从备份文件加载, 这将导致无法启动app
    BackupFile = "./configs/backup.apollo" # 本地备份文件, 留空表示不使用备份
    ```

2. 对于混用配置, 由于apollo的配置是扁平化的, 多级key用点连接, 所以文件的配置应该改为如下样式以适应:

    ```toml
    [frame]
    Debug = true
    FreeMemoryInterval = 120000
    #...
    Log.Level = true
    Log.WriteToStream = true
    #...

    [services]
    Api.Bind = ":8080"
    #...
    Grpc.Bind = ":3000"
    #...

    [components]
    Xorm.default.Driver = "mysql"
    #...
    Redis.default.Address = "127.0.0.1:6379"
    #...
    ```

3. 由于配置项越来越多, 扁平化的apollo变得不是很好管理, 我们支持将多个配置key合并为一个json值. apollo配置示例:
    ```toml
    [frame]
    Debug = true
    Log = {
         "Level": "info",
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
