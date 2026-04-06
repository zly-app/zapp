# Zapp 项目 AI 开发参考文档

> 本文档为 AI 快速开发参考，涵盖项目架构、核心接口、关键模块及使用方法。

---

## 1. 项目概述

**zapp** 是一个 Go 语言应用框架，用于快速构建项目。提供组件化管理、插件系统、服务生命周期控制、配置热加载等能力。

**快速启动示例**:
```go
import "github.com/zly-app/zapp"

app := zapp.NewApp("test")
app.Run()
```

---

## 2. 核心架构

### 2.1 架构层次

| 层级 | 说明 |
|------|------|
| **App** | 应用主入口，管理生命周期 |
| **Component** | 基础组件(db/rpc/cache/mq等)，提供 IApp、ILogger、IGPools、IMsgbus |
| **Plugin** | 扩展插件，注入后启动 |
| **Service** | 业务服务，启动后持续运行 |
| **Filter** | 过滤器，在初始化后执行 |

### 2.2 App 生命周期

```
初始化 → 用户操作 → 启动 → 退出
```

**初始化阶段**: 生成 BaseContext → 加载配置 → 初始化日志 → 构建组件 → 构建插件 → 构建 filter → 构建服务

**启动阶段**: 启动插件 → 启动服务 → 启动内存释放任务 → 阻塞等待退出信号

**退出阶段**: 关闭 BaseContext → 停止内存释放 → 关闭服务 → 关闭 filter → 关闭插件 → 释放组件

---

## 3. 核心接口定义

### 3.1 IComponent (组件接口)

```go
// 位置: core/component.go
type IComponent interface {
    App() IApp                // 获取 app
    Config() *Config          // 获取配置
    ILogger                   // 嵌入日志接口
    Close()                   // 关闭所有组件
    IGPools                   // 嵌入协程池接口
    IMsgbus                   // 嵌入消息总线接口
}
```

### 3.2 IPlugin (插件接口)

```go
// 位置: core/plugin.go
type IPlugin interface {
    Inject(a ...interface{})  // 注入，不同插件作用不同
    Start() error             // 启动插件
    Close() error             // 关闭插件
}

type IPluginCreator interface {
    Create(app IApp) IPlugin  // 创建插件
}

type PluginType string       // 插件类型标识
```

### 3.3 IService (服务接口)

```go
// 位置: core/service.go
type IService interface {
    Inject(a ...interface{})  // 注入，不同服务作用不同
    Start() error             // 开始服务
    Close() error             // 关闭服务
}

type IServiceCreator interface {
    Create(app IApp) IService // 创建服务
}

type ServiceType string      // 服务类型标识
```

### 3.4 ILogger (日志接口)

```go
// 位置: core/logger.go
type ILogger interface {
    Log(level string, v ...interface{})
    Debug(v ...interface{})
    Info(v ...interface{})
    Warn(v ...interface{})
    Error(v ...interface{})
    DPanic(v ...interface{})
    Panic(v ...interface{})
    Fatal(v ...interface{})
    NewSessionLogger(fields ...zap.Field) ILogger    // 创建会话日志(带全局日志ID)
    NewTraceLogger(ctx context.Context, fields ...zap.Field) ILogger  // 创建链路日志(带traceID)
}
```

**日志特点**:
- 可传入 `ctx`，若包含 traceID 则日志输出中会带上
- 会话日志产生全局日志ID，不同ID显示不同颜色
- 链路日志根据不同链路ID显示不同颜色

### 3.5 IConfigWatchKeyObject (配置观察接口)

```go
// 位置: core/config.watch.go
type IConfigWatchKeyObject interface {
    GroupName() string                              // 获取组名
    KeyName() string                                // 获取key名
    AddCallback(callback ...ConfigWatchKeyCallback) // 添加回调，启动时立即触发一次
    GetData() []byte                                // 获取原始数据副本
    Expect(v interface{}) bool                      // 检查是否符合预期值
    
    // 类型转换方法
    GetString() string
    GetBool(def ...bool) bool
    GetInt(def ...int) int
    GetInt64(def ...int64) int64
    GetFloat64(def ...float64) float64
    // ... 其他整数类型
    
    ParseJSON(outPtr interface{}) error             // 解析为JSON
    ParseYaml(outPtr interface{}) error             // 解析为YAML
}

// 回调函数签名
type ConfigWatchKeyCallback func(isInit bool, oldData, newData []byte)
```

---

## 4. 配置系统

> 详细说明见 `config/readme.md`

### 4.1 配置加载优先级

```
命令行 -c 指定文件 > WithViper > WithConfig > WithFiles > WithApollo > 默认配置文件
```

- 使用命令 `-t` 测试配置是否正确
- 多个配置文件存在同配置分片会智能合并
- 默认配置文件(按优先级自动加载一个): `./configs/default.yaml` > `./configs/default.yml` > `./configs/default.toml` > `./configs/default.json`

### 4.2 配置选项

```go
// 位置: config/opts.go

config.WithViper(vi *viper.Viper) Option      // 设置 viper 实例
config.WithConfig(conf *core.Config) Option   // 设置配置结构体
config.WithFiles(files ...string) Option      // 设置配置文件
config.WithApollo(conf *ApolloConfig) Option  // 从 Apollo 配置中心加载
config.WithoutFlag() Option                   // 禁用 flag 解析
```

### 4.3 配置结构 Key 规则

| 类型 | 配置 Key 格式 | 示例 |
|------|---------------|------|
| 框架配置 | `frame` | `frame.debug: true` |
| 插件配置 | `plugins.{pluginType}` | `plugins.my_plugin.foo: bar` |
| 服务配置 | `services.{serviceType}` | `services.api.bind: :8080` |
| 组件配置 | `components.{componentType}.{componentName}` | `components.cache.cache1.cacheDB: memory` |
| 自定义配置 | 自定义分片名 | `myconfig.foo: bar` |

### 4.4 框架配置示例 (所有字段可选)

```yaml
frame:
    debug: true                                    # debug标志
    Env: ''                                        # 环境名
    FreeMemoryInterval: 120000                     # 清理内存间隔(ms), <=0 禁用
    WaitServiceRunTime: 1000                       # 等待服务启动时间(ms)
    ServiceUnstableObserveTime: 10000              # 服务不稳定观察时间(ms)
    Flags: []                                      # flag, 忽略大小写, 如 ['a', 'B']
    Labels:                                        # 标签, 忽略大小写
        Foo: Bar
    Log:
        Level: 'debug'                             # 日志等级: debug/info/warn/error/dpanic/panic/fatal
        TraceLevel: 'debug'                        # 附加到trace的等级
        Json: false                                # 启用json编码器
        WriteToStream: true                        # 输出到屏幕
        WriteToFile: false                         # 输出到文件
        Name: ''                                   # 日志文件名, 自动附加 .log
        AppendPid: false                           # 文件名附加进程号
        Path: './log'                              # 日志路径
        FileMaxSize: 32                            # 日志最大尺寸(M)
        FileMaxBackupsNum: 3                       # 备份数量, 0=永久
        FileMaxDurableTime: 7                      # 保存天数, 0=永久
        Compress: false                            # 压缩历史日志
        TimeFormat: '2006-01-02 15:04:05'          # 时间格式
        Color: true                                # 彩色日志等级
        CapitalLevel: false                        # 大写日志等级
        DevelopmentMode: true                      # 开发者模式
        ShowFileAndLinenumMinLevel: 'debug'        # 显示文件行号的最小等级
        ShowStacktraceLevel: 'error'               # 显示调用链的等级
        MillisDuration: true                       # Duration转为毫秒
    PrintConfig: true                              # 初始化时打印配置
```

### 4.5 插件/服务/组件配置示例

```yaml
plugins:                    # 插件配置
    my_plugin:              # 插件类型
        Foo: Bar

services:                   # 服务配置
    api:                    # 服务类型
        Bind: :8080
    grpc:
        Bind: :3000

components:                 # 组件配置
    cache:                  # 组件类型
        cacheName1:         # 组件名(支持多实例)
            CacheDB: memory
        cacheName2:
            CacheDB: redis

myconfig:                   # 自定义配置分片
    Foo: Bar
```

### 4.6 引用配置文件

```yaml
include:
    files: './1.toml,./2.toml'   # 相对路径, 引用文件不能再引用
```

### 4.7 Apollo 配置中心

**最小配置**:
```yaml
apollo:
    Address: "http://127.0.0.1:8080"
    AppId: "your-appid"
```

**完整配置**:
```yaml
apollo:
    Address: "http://127.0.0.1:8080"
    AppId: "your-appid"
    AccessKey: ""                           # 验证key
    AuthBasicUser: ""                       # 基础认证用户名
    AuthBasicPassword: ""                   # 基础认证密码
    Cluster: "default"                      # 集群名
    AlwaysLoadFromRemote: false             # 总是从远程获取
    BackupFile: "./configs/backup.apollo"   # 本地备份文件
    ApplicationDataType: "yaml"             # application命名空间数据类型: yaml/yml/toml/json
    ApplicationParseKeys: []                # 额外解析的key
    Namespaces: []                          # 其他命名空间
    IgnoreNamespaceNotFound: false          # 忽略命名空间不存在
```

**Apollo 命名空间规则**:
- 命名空间名作为配置顶级key, 其内容作为二级key
- `application` 命名空间下的 `frame/components/plugins/filters/services` 作为顶级key并解析其值
- 支持扁平化: 命名空间 `frame.json` 直接配置框架内容

**示例 - Apollo 配置解析**:
| 命名空间 | field | value |
|----------|-------|-------|
| namespace1 | key1 | value |
| namespace1 | key2 | {"foo": "bar"} |

解析为:
```yaml
namespace1:
    key1: 'value'
    key2: '{"foo": "bar"}'
```

### 4.8 配置观察 (推荐用法)

**基础用法 - 观察原始数据**:
```go
// 可在变量定义时初始化
var MyConfigWatch = zapp.WatchConfigKey("group_name", "key_name")

func main() {
    app := zapp.NewApp("test", apollo_provider.WithPlugin(true))
    defer app.Exit()
    
    // 获取数据
    y1 := MyConfigWatch.GetString()
    y2 := MyConfigWatch.GetInt()
    y3 := MyConfigWatch.GetBool()
    
    // 检查预期值
    if MyConfigWatch.Expect("1") { ... }
    
    // 添加回调 (启动时立即触发一次)
    MyConfigWatch.AddCallback(func(first bool, oldData, newData []byte) {
        app.Info("回调", zap.Bool("first", first))
    })
    
    app.Run()
}
```

**推荐用法 - 泛型结构自动解析**:
```go
type MyConfig struct {
    A int `json:"a"`
}

var MyConfigWatch = zapp.WatchConfigJson[*MyConfig]("group_name", "generic_key")

func main() {
    app := zapp.NewApp("test", apollo_provider.WithPlugin(true))
    defer app.Exit()
    
    for {
        a := MyConfigWatch.Get()  // 直接获取结构体
        app.Info("数据", a)
        time.Sleep(time.Second)
    }
}
```

**特性**:
- 一行代码接入配置中心
- 自动监听配置变更
- 开始watch失败会 Fatal 退出
- 泛型方式解析失败会 Error 并忽略变更

**Apollo 非properties命名空间**:
```go
// 命名空间为 watch 的 json 类型, key固定为 "content"
var MyConfigWatch = zapp.WatchConfigJson[*MyConfig]("watch.json", "content")
```

### 4.9 用户匹配器 (灰度/白名单)

```go
type UserMatcher struct {
    Uids    []string  // 指定用户ID
    Percent uint8     // 灰度比例 0~100
    Tails   []string  // 用户尾号(后两位)
}

// 使用示例
type MyConfig struct {
    A string
    zapp.UserMatcher            // 直接继承
    Matcher2 zapp.UserMatcher   // 嵌入字段
}

func (u *UserMatcher) IsHit(userID string) bool  // 判断是否命中
```

**配置示例**:
```json
{
    "Uids": ["111", "222"],
    "Percent": 0,
    "Tails": ["01", "02"]
}
```

---

## 5. 内置组件

### 5.1 协程池 (gpool)

```go
// 位置: component/gpool/
// 提供并发任务执行能力

// 任务定义
type job struct {
    fn       func() error      // 执行函数
    callback func(err error)   // 回调函数
}

// 工作协程
type worker struct {
    pool       chan<- *worker  // 工人队列池
    jobChannel chan *job       // 工作任务通道
    stop       chan struct{}   // 停止信号
}
```

### 5.2 消息总线 (IMsgbus)

默认队列大小: `DefaultMsgbusQueueSize = 1000`

---

## 6. pkg 工具包

### 6.1 serializer (序列化器)

```go
// 位置: pkg/serializer/
// 支持多种序列化格式

ISerializer interface {
    Marshal(a interface{}, w io.Writer) error
    MarshalBytes(a interface{}) ([]byte, error)
    Unmarshal(r io.Reader, a interface{}) error
    UnmarshalBytes(data []byte, a interface{}) error
}

// 可用序列化器
- BaseSerializer    // 基础序列化器，支持基本类型转换
- JsonSerializer    // JSON 序列化
- YamlSerializer    // YAML 序列化  
- MsgpackSerializer // Msgpack 序列化
- SonicSerializer   // Sonic 高性能 JSON
```

**BaseSerializer 类型转换**:
- 支持 `string`, `[]byte`, `bool`, 所有整数类型, `float32/64`
- bool 转换支持多种字符串格式: `true/false`, `yes/no`, `on/off`, `1/0` 等

### 6.2 compactor (压缩器)

```go
// 位置: pkg/compactor/
// 支持多种压缩格式

- RawCompactor    // 无压缩
- GzipCompactor   // Gzip 压缩
- ZstdCompactor   // Zstd 压缩
```

### 6.3 utils (工具集)

```go
// 位置: pkg/utils/

// 反射工具
utils.Reflect.IsZero(a interface{}) bool  // 判断是否为零值

// 恢复工具(panic捕获)
utils.Recover.WrapCall(fn func() error) error           // 包装函数捕获panic
utils.Recover.IsRecoverError(err error) bool            // 是否为recover错误
utils.Recover.GetRecoverError(err error) (RecoverError, bool)
utils.Recover.GetRecoverErrorDetail(err error) string   // 获取错误详情

// 三元运算
utils.Ternary(condition bool, a, b T) T

// 文本工具
utils.Text...

// 代理工具
utils.NewSocks5Proxy(address string) (ISocks5Proxy, error)  // 创建 socks5 代理
utils.NewHttpProxy(address string) (IHttpProxy, error)      // 创建 http 代理

// 单例获取
utils.GetInstance...
```

### 6.4 lumberjack (日志滚动)

```go
// 位置: pkg/lumberjack/lumberjack.go
// 日志文件滚动管理，支持:
// - 文件大小限制滚动
// - 按时间保留旧文件
// - 按数量保留备份
// - 自动压缩

type Logger struct {
    Filename   string  // 日志文件路径
    MaxSize    int     // 最大大小(MB), 默认100
    MaxAge     int     // 最大保留天数
    MaxBackups int     // 最大备份数量
    LocalTime  bool    // 使用本地时间
    Compress   bool    // 压缩备份文件
}
```

### 6.5 zlog (日志包装)

```go
// 位置: pkg/zlog/
// 日志核心实现包装

// 获取日志输出合成器
zlog.GetLogWriteSyncer(l interface{}) (zapcore.WriteSyncer, bool)

// 添加日志字段
zlog.AddFields(l interface{}, fields ...zap.Field) bool

// 移除日志字段
zlog.RemoveFields(l interface{}, count int, fieldKeys ...string) (int, bool)

// 获取日志核心
zlog.GetLogCore(l core.ILogger) core.ILogger
```

---

## 7. 使用选项

### 7.1 Handler 钩子

```go
zapp.WithHandler(
    BeforeInitializeHandler,  // app 初始化前
    AfterInitializeHandler,   // app 初始化后
    BeforeStartHandler,       // app 启动前
    AfterStartHandler,        // app 启动后
    BeforeExitHandler,        // app 退出前
    AfterExitHandler,         // app 退出后
)
```

### 7.2 自定义插件/服务

```go
// 运行时自定义启用插件
zapp.WithCustomEnablePlugin(func(app core.IApp, plugins []core.PluginType) []core.PluginType {
    if !app.GetConfig().HasFlag("my_plugin") {
        plugins = append(plugins, "my_plugin")
    }
    return plugins
})

// 运行时自定义启用服务
zapp.WithCustomEnableService(func(app core.IApp, services []core.ServiceType) []core.ServiceType {
    if !app.GetConfig().HasFlag("api_service") {
        services = append(services, "api")
    }
    return services
})
```

### 7.3 守护进程

```go
// 启用守护进程模式
zapp.WithEnableDaemon()

// 命令行操作:
// install [args] - 安装服务
// remove        - 移除服务
// start         - 启动 app
// stop          - 停止 app
// status        - 查看状态
```

---

## 8. 常量定义

```go
// 位置: consts/def.go

// 框架常量
DefaultFreeMemoryInterval         = 120000  // 默认清理内存间隔(ms)
DefaultWaitServiceRunTime         = 1000    // 等待服务启动时间(ms)
DefaultServiceUnstableObserveTime = 10000   // 服务不稳定观察时间(ms)

// 配置常量
DefaultConfigFiles = "./configs/default.yaml,./configs/default.yml,..."
ApolloConfigKey    = "apollo"

// 组件常量
DefaultComponentName     = "default"
DefaultMsgbusQueueSize   = 1000
```

---

## 9. 扩展资源

| 资源 | 地址 |
|------|------|
| 组件库 | https://github.com/zly-app/component |
| 插件库 | https://github.com/zly-app/plugin |
| 服务库 | https://github.com/zly-app/service |
| 推荐启动器 | https://github.com/zly-app/uapp |

---

## 10. 开发指引

### 10.1 创建新组件

1. 定义组件接口，嵌入 `IComponent`
2. 实现 `ComponentType` 类型标识
3. 在 `Create(app IApp)` 方法中初始化组件

### 10.2 创建新插件

1. 实现 `IPlugin` 接口 (`Inject`, `Start`, `Close`)
2. 实现 `IPluginCreator` 接口 (`Create`)
3. 定义 `PluginType` 常量

### 10.3 创建新服务

1. 实现 `IService` 接口 (`Inject`, `Start`, `Close`)
2. 实现 `IServiceCreator` 接口 (`Create`)
3. 定义 `ServiceType` 常量

### 10.4 配置观察实现

实现 `IConfigWatchProvider` 接口:
```go
type IConfigWatchProvider interface {
    Get(groupName, keyName string) ([]byte, error)
    Watch(groupName, keyName string, callback ConfigWatchProviderCallback) error
}
```

---

## 11. 关键文件索引

| 文件路径 | 说明 |
|----------|------|
| `core/component.go` | 组件接口定义 |
| `core/plugin.go` | 插件接口定义 |
| `core/service.go` | 服务接口定义 |
| `core/logger.go` | 日志接口定义 |
| `core/config.watch.go` | 配置观察接口定义 |
| `config/opts.go` | 配置选项 |
| `config/config.watch.go` | 配置观察实现 |
| `consts/def.go` | 常量定义 |
| `component/gpool/` | 协程池组件 |
| `pkg/serializer/` | 序列化器 |
| `pkg/compactor/` | 压缩器 |
| `pkg/utils/` | 工具集 |
| `pkg/lumberjack/` | 日志滚动 |
| `pkg/zlog/` | 日志包装 |

---

*文档生成时间: 2026-04-06*