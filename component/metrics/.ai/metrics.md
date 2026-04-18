# metrics 组件 - AI 详细说明

## 概述

metrics 是一个指标收集器组件，提供 Counter（计数器）、Gauge（计量器）、Histogram（直方图）、Summary（汇总）四种指标类型。它采用接口抽象 + 默认实例 + Noop 降级的模式设计，是 zapp 框架的核心组件之一。

## 架构

### 核心模型

metrics 采用 **Client 抽象 + 包级便捷函数** 模式：

```
调用方 --> 包级函数 (RegistryCounter/Counter/...) --> defaultClient --> 实际实现 (如 prometheus client)
                                                   \--> DefNoopClient (默认，空操作)
```

**调用流程：**

1. 应用启动后通过 `SetClient()` 注入实际实现（如 prometheus client）
2. 调用方通过包级函数 `metrics.RegistryCounter(...)` 操作指标
3. 包级函数委托给 `defaultClient`，如果未注入则使用 Noop 实现（所有操作为空操作）

### 文件结构

| 文件 | 职责 |
|------|------|
| `client.go` | 核心接口定义：`Client`、`ICounter`、`IGauge`、`IHistogram`、`ISummary` |
| `default.go` | 包级便捷函数，基于默认 `Client` 实例的代理 |
| `noop.go` | Noop 实现，所有方法为空操作，作为默认降级方案 |

### 关键接口

```go
// Labels - 标签键值对，用于指标维度区分
type Labels = map[string]string

// Client - 指标收集器客户端接口
type Client interface {
    // 计数器
    RegistryCounter(name, help string, constLabels Labels, labels ...string) ICounter
    Counter(name string) ICounter

    // 计量器
    RegistryGauge(name, help string, constLabels Labels, labels ...string) IGauge
    Gauge(name string) IGauge

    // 直方图
    RegistryHistogram(name, help string, buckets []float64, constLabels Labels, labels ...string) IHistogram
    Histogram(name string) IHistogram

    // 汇总
    RegistrySummary(name, help string, constLabels Labels, labels ...string) ISummary
    Summary(name string) ISummary
}

// ICounter - 计数器接口
type ICounter = interface {
    Inc(labels Labels, exemplar Labels)
    Add(v float64, labels Labels, exemplar Labels)
}

// IGauge - 计量器接口
type IGauge = interface {
    Set(v float64, labels Labels)
    Inc(labels Labels)
    Dec(labels Labels)
    Add(v float64, labels Labels)
    Sub(v float64, labels Labels)
    SetToCurrentTime(labels Labels)
}

// IHistogram - 直方图接口
type IHistogram = interface {
    Observe(v float64, labels Labels, exemplar Labels)
}

// ISummary - 汇总接口
type ISummary = interface {
    Observe(v float64, labels Labels, exemplar Labels)
}
```

## 核心组件详解

### Client 接口

- 定义四种指标类型的注册和获取方法
- 注册方法（`Registry*`）同时完成指标创建和注册，返回对应接口实例
- 获取方法（`Counter/Gauge/Histogram/Summary`）通过 name 获取已注册的指标实例
- 注册时需提供 `constLabels`（固定标签）和可变 `labels`（动态标签键名列表）

### default 包级函数

- 持有 `defaultClient` 全局变量，初始值为 `DefNoopClient`
- 提供 `GetClient()` / `SetClient()` 管理默认实例
- 所有包级函数（如 `RegistryCounter`、`Counter`）均委托给 `defaultClient`
- **必须在 `zapp.NewApp()` 之后调用**，因为应用启动时会通过 `SetClient()` 注入实际实现

### Noop 实现

- `noopClient`：所有注册方法返回对应的 Noop 指标实例，获取方法返回默认 Noop 实例
- `noopCounter`：`Inc`、`Add` 为空操作
- `noopGauge`：`Set`、`Inc`、`Dec`、`Add`、`Sub`、`SetToCurrentTime` 为空操作
- `noopHistogram` / `noopSummary`：`Observe` 为空操作
- 作为默认实现，确保未注入实际 Client 时不会 panic

## 四种指标类型

### Counter（计数器）

- 只增不减的累计值
- 典型用途：请求总数、错误总数、完成任务数
- 方法：
  - `Inc(labels, exemplar)`：+1
  - `Add(v, labels, exemplar)`：增加指定值

### Gauge（计量器）

- 可增可减的瞬时值
- 典型用途：当前温度、内存使用量、并发连接数
- 方法：
  - `Set(v, labels)`：设为指定值
  - `Inc(labels)`：+1
  - `Dec(labels)`：-1
  - `Add(v, labels)`：增加指定值
  - `Sub(v, labels)`：减少指定值
  - `SetToCurrentTime(labels)`：设为当前 Unix 时间戳

### Histogram（直方图）

- 对观测值进行采样并统计分布
- 典型用途：请求延迟、响应大小
- 注册时需指定 `buckets`（桶列表），如 `[]float64{0.1, 0.5, 1, 2.5, 5, 10}`
- 方法：
  - `Observe(v, labels, exemplar)`：记录一个观测值

### Summary（汇总）

- 对观测值进行采样并计算分位数
- 典型用途：请求延迟的 P50/P90/P99
- 方法：
  - `Observe(v, labels, exemplar)`：记录一个观测值

## 参数说明

### name（指标名称）

- 格式建议：`检测对象_数值类型_单位`，如 `http_request_duration_seconds`
- 必须唯一，重复注册同名指标会由底层实现决定行为（通常 panic）

### help（描述）

- 对指标的说明文字，用于 Prometheus UI 展示

### constLabels（固定标签）

- `Labels` 类型（`map[string]string`）
- 注册时确定，所有数据点都携带，如 `{"app": "myapp", "host": "10.0.0.1"}`
- 适合不变维度

### labels（动态标签键名）

- `...string` 可变参数，声明允许使用的标签键名
- 使用时通过 `Labels` 传入具体值，如注册时 `labels="level"`，使用时 `Labels{"level": "info"}`

### exemplar（样例）

- `Labels` 类型，用于关联 trace 信息
- Prometheus Exemplar 机制，可将指标与具体 trace 关联
- 不使用时传 `nil`

## 包级便捷函数

```go
// 注册指标
metrics.RegistryCounter(name, help, constLabels, labels...)     // 注册计数器
metrics.RegistryGauge(name, help, constLabels, labels...)       // 注册计量器
metrics.RegistryHistogram(name, help, buckets, constLabels, labels...)  // 注册直方图
metrics.RegistrySummary(name, help, constLabels, labels...)     // 注册汇总

// 获取已注册的指标
metrics.Counter(name)     // 获取计数器
metrics.Gauge(name)       // 获取计量器
metrics.Histogram(name)   // 获取直方图
metrics.Summary(name)     // 获取汇总

// 管理默认实例
metrics.GetClient()       // 获取当前 Client
metrics.SetClient(client) // 设置 Client（通常由框架调用）
```

## 集成

### zapp 生命周期

- metrics 不注册 zapp 生命周期钩子
- `SetClient()` 通常在 `zapp.NewApp()` 内部由框架自动调用，注入实际的 metrics 实现（如 prometheus）
- 在 `NewApp()` 之前调用包级函数会使用 Noop 实现，指标数据将被丢弃

### Prometheus

- 框架中通常提供 prometheus 实现来满足 `Client` 接口
- 通过 `/metrics` 端点暴露指标数据

## 注意事项

- **必须在 `zapp.NewApp()` 之后使用**，否则指标操作为 Noop 空操作
- 注册指标的 `name` 必须全局唯一，重复注册通常会导致 panic
- `constLabels` 和 `labels` 中应避免高基数标签（如 user_id），会导致指标爆炸
- Histogram 的 `buckets` 需根据实际数据分布合理设置
- Noop 实现保证在未初始化时安全运行，不会 panic
