# gpool 组件 - AI 详细说明

## 概述

gpool 是一个基于 worker 池模型的协程池组件，用于控制并发 goroutine 数量，避免无限制创建协程导致的资源耗尽。它是 zapp 框架的核心组件之一，实现了 `core.IGPool` 和 `core.IGPools` 接口。

## 架构

### 核心模型

gpool 采用经典的 **worker-pool + job-queue** 模式：

```
调用方 --> jobQueue(任务队列) --> dispatcher(调度协程) --> workerQueue(工人队列) --> worker(工人协程)
```

**工作流程：**

1. 调用方通过 `Go`/`TryGo` 等方法提交任务到 `jobQueue`
2. `dispatcher` 协程从 `jobQueue` 取任务，从 `workerQueue` 取空闲 worker，将任务分配给 worker
3. worker 执行任务完毕后自动回到 `workerQueue` 等待下一个任务

### 文件结构

| 文件 | 职责 |
|------|------|
| `gpool.go` | 协程池核心实现，包含任务提交、调度、关闭逻辑 |
| `gpools.go` | 协程池管理器，支持按名称创建和管理多个池实例，集成 zapp 配置和生命周期 |
| `config.go` | 配置定义（`GPoolConfig`），含默认值校验 |
| `default_creator.go` | 包级默认管理器，提供 `GetGPool`/`GetDefGPool`/`GetCreator` 便捷函数 |
| `worker.go` | Worker 实现，每个 worker 是一个长期运行的 goroutine |
| `job.go` | Job 封装，包含执行函数和回调，带自动 Recover |
| `nopool.go` | 无池模式实现，当 `ThreadCount < 0` 时使用，直接在提交协程中同步执行 |

### 关键接口

```go
// core.IGPool - 协程池接口
type IGPool interface {
    Go(fn func() error, callback func(err error))                    // 异步执行，队列满时阻塞
    GoSync(fn func() error) (result error)                           // 同步执行
    TryGo(fn func() error, callback func(err error)) (ok bool)      // 尝试异步执行，队列满返回 false
    TryGoSync(fn func() error) (result error, ok bool)              // 尝试同步执行，队列满返回 false
    GoAndWait(fn ...func() error) error                              // 批量异步执行并等待，返回第一个错误
    GoRetWait(fn ...func() error) func() error                       // 批量异步执行，返回等待函数
    Wait()                                                           // 等待所有已提交任务完成
    Close()                                                          // 关闭池
}

// core.IGPools - 协程池管理器接口
type IGPools interface {
    GetGPool(name ...string) IGPool                                  // 获取命名池，未指定则获取默认池
}
```

## 配置

### GPoolConfig

```go
type GPoolConfig struct {
    JobQueueSize int  // 任务队列大小，最小值 100000
    ThreadCount  int  // 并发 worker 数
}
```

**ThreadCount 特殊值：**
- `0`：使用逻辑 CPU 数 × 2（由 check 逻辑保证最小为 100）
- `< 0`（负数）：使用 `NoPool` 模式，无并发限制，每次提交直接在当前 goroutine 执行
- 正数：指定固定 worker 数量，最小值 100

### 配置读取

`gpools.makeGPoolGroup` 通过 `config.Conf.ParseComponentConfig` 从 zapp 配置中读取，组件类型为 `"gpool"`。

## 关键行为

### 任务提交

| 方法 | 阻塞行为 | 队列满时 | 返回值 |
|------|---------|---------|--------|
| `Go` | 入队时阻塞 | 阻塞直到有空位 | 无 |
| `GoSync` | 阻塞等待结果 | 同 Go | error |
| `TryGo` | 非阻塞 | 返回 false | ok bool |
| `TryGoSync` | 非阻塞 | 返回 false | error, ok bool |
| `GoAndWait` | 阻塞等待全部 | 同 Go | error（第一个） |
| `GoRetWait` | 非阻塞，返回 wait 函数 | 同 Go | func() error |

### 关闭行为

`Close()` 的语义：
1. 向 dispatcher 发送停止信号
2. 空闲 worker 立即停止
3. 正在执行任务的 worker 完成当前任务后停止
4. 任务队列中未分配的任务不会被丢弃，而是被消费掉（仅 wg.Done，不执行）
5. 等待所有 worker 停止后返回
6. 通过 `sync.Once` 保证只关闭一次

### 自动 Recover

所有通过 gpool 执行的函数都经过 `utils.Recover.WrapCall` 包装，panic 会被捕获并转为 error 返回。

### NoPool 模式

当 `ThreadCount < 0` 时使用 `NoPool`：
- `Go`/`TryGo`：同步执行（在提交方 goroutine 中直接调用 fn），`TryGo` 始终返回 true
- `GoAndWait`：使用 `utils.Go.GoAndWait` 并发执行
- `GoRetWait`：使用 `utils.Go.GoRetWait` 并发执行
- `Wait`：使用独立 WaitGroup，仅追踪 Go/TryGo 提交的任务

## 生命周期集成

`gpools` 通过 `handler.AddHandler(handler.AfterCloseComponent, ...)` 注册关闭钩子，在 zapp 组件关闭后自动关闭所有池实例。

## 包级便捷函数

```go
gpool.GetGPool(name)    // 获取命名池
gpool.GetDefGPool()     // 获取默认池
gpool.GetCreator()      // 获取管理器实例
```

## 注意事项

- `Close()` 后再调用 `Go` 等方法，callback 会收到 `ErrGPoolClosed` 错误
- `GoAndWait` 和 `GoRetWait` 中，如果多个函数返回错误，只返回第一个错误
- dispatcher 是单协程调度，不会出现任务分配的竞态条件
- worker 数量在创建后不可动态调整
