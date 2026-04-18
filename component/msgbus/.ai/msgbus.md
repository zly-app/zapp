# msgbus 组件 - AI 详细说明

## 概述

msgbus 是一个进程内的发布/订阅（Pub/Sub）消息总线组件，支持 topic 级别的订阅和全局订阅。它是 zapp 框架的核心组件之一，实现了 `core.IMsgbus` 接口，集成了 OpenTelemetry 链路追踪。

## 架构

### 核心模型

msgbus 采用 **Topic-Subscriber** 模型：

```
Publisher --> MsgBus --> Topic1 --> Subscriber1 (handler goroutine pool)
                      |        \-> Subscriber2 (handler goroutine pool)
                      \-> GlobalTopic --> GlobalSubscriber1
                                        \-> GlobalSubscriber2
```

**消息流：**

1. 发布者调用 `Publish(ctx, topic, msg)` 发布消息
2. 消息先序列化为 JSON 用于日志/追踪，然后存入 context
3. 消息同时发送到 **全局 Topic** 和 **指定 Topic**
4. 每个 Topic 遍历所有订阅者，将消息推入订阅者的消息队列（channel）
5. 订阅者的 worker goroutine 从队列取出消息执行 handler

### 文件结构

| 文件 | 职责 |
|------|------|
| `msgbus.go` | 消息总线核心实现，管理 topics 和 global topic |
| `simple.go` | 包级便捷函数，基于默认 `msgBus` 实例 |
| `msg.go` | 消息封装（`channelMsg`），context 传播，消息序列化数据存取 |
| `topic.go` | Topic 实现，管理订阅者列表，处理发布/订阅/取消订阅 |
| `subscriber.go` | 订阅者实现，包含消息队列和 worker goroutine 池 |
| `subscriber_test.go` | 订阅者测试 |
| `topic_test.go` | Topic 测试 |

### 关键接口

```go
// core.IMsgbus - 消息总线接口
type IMsgbus interface {
    Publish(ctx context.Context, topic string, msg interface{})
    Subscribe(topic string, threadCount int, handler MsgbusHandler) (subscribeId uint32)
    SubscribeGlobal(threadCount int, handler MsgbusHandler) (subscribeId uint32)
    Unsubscribe(topic string, subscribeId uint32)
    UnsubscribeGlobal(subscribeId uint32)
    CloseTopic(topic string)
    Close()
}

// core.IMsgbusMessage - 消息接口
type IMsgbusMessage interface {
    Ctx() context.Context
    Topic() string
    Msg() interface{}
}

// core.MsgbusHandler - 消息处理函数
type MsgbusHandler func(ctx context.Context, msg IMsgbusMessage)
```

## 核心组件详解

### MsgBus

- 持有一个全局 `Topic`（`global`）和一个 `topics map[string]Topic`
- `Publish` 时消息同时发送到 global topic 和指定 topic
- `Subscribe` 使用 double-check locking 确保 topic 的懒创建是并发安全的
- 默认消息队列大小为 1000（`DefaultMsgQueueSize`）

### Topic

- 维护 `subs map[uint32]Subscriber`
- `Publish`：遍历所有订阅者，调用 `sub.Handler(msg)` 将消息推入订阅者队列
- `Subscribe`：创建订阅者，启动 worker goroutine，注册到 subs map
- `Unsubscribe`：从 map 删除并关闭订阅者（后置关闭，先解锁再关闭）
- `Close`：收集所有订阅者后清空 map，逐一关闭（后置关闭模式避免死锁）

### Subscriber

- 每个 Subscriber 有自己的消息队列 `chan IMsgbusMessage` 和 `threadCount` 个 worker goroutine
- `Handler(msg)` 将消息推入队列
- worker goroutine 从队列消费消息，执行 handler
- 自动 Recover：handler 执行经过 `utils.Recover.WrapCall` 包装
- 全局订阅者和普通订阅者在 OpenTelemetry span 名称上有区分

### 消息传播

- `Publish` 时通过 `utils.Ctx.CloneContext(ctx)` 克隆 context，避免 context 取消影响后续处理
- 消息序列化数据通过 `context.WithValue` 传播，subscriber 处理时可通过 `getMsgData(ctx)` 获取
- OpenTelemetry 链路：publish 创建 span，subscriber 处理时创建子 span

## 包级便捷函数

```go
msgbus.Publish(ctx, topic, msg)           // 发布
msgbus.Subscribe(topic, threadCount, handler)  // 订阅
msgbus.SubscribeGlobal(threadCount, handler)   // 全局订阅
msgbus.Unsubscribe(topic, subscribeId)    // 取消订阅
msgbus.UnsubscribeGlobal(subscribeId)     // 取消全局订阅
msgbus.CloseTopic(topic)                  // 关闭主题
msgbus.Close()                            // 关闭消息总线
```

## 关键行为

### 发布

- 消息先通过 `sonic.MarshalString` 序列化为 JSON，用于日志和追踪
- 如果序列化失败，记录错误日志但不会中断发布流程（消息本身仍然发送）
- 即使没有订阅者，消息也不会报错，只是被丢弃

### 订阅

- `threadCount` 指定并发处理消息的 goroutine 数量，最小为 1
- `msgQueueSize` 由 MsgBus 的 `msgQueueSize` 决定（默认 1000）
- 订阅者创建后自动启动 worker goroutine
- 返回 `subscribeId`（全局自增 uint32），用于后续取消订阅

### 取消订阅

- 调用 `Unsubscribe` 后订阅者被关闭，消息队列 channel 被 close
- worker goroutine 通过 `range s.queue` 自然退出
- 取消后该订阅者不再收到消息

### 关闭

- `CloseTopic`：删除并关闭指定 topic 的所有订阅者
- `Close`：关闭所有 topic（包括 global），然后重置为空状态

### 消息队列阻塞

- 如果订阅者处理缓慢导致消息队列满，`sub.Handler(msg)`（即 `s.queue <- msg`）会阻塞
- 这会阻塞 Topic 的 Publish 方法，进而阻塞发布者
- 因此订阅者应尽快处理消息或设置足够的 `threadCount`

## 集成

### OpenTelemetry

- Publish span：`msgbus/publish/{topic}`
- Subscriber span：`msgbus/subsriber/global_{subId}` 或 `msgbus/subsriber/{subId}`
- 消息序列化数据作为 span event 的属性传播

### zapp 生命周期

msgbus 本身不注册 zapp 生命周期钩子，需要使用者自行在合适的时机调用 `Close()`。

## 注意事项

- 发布消息时如果没有对应 topic 的订阅者，消息会被丢弃（全局订阅者仍能收到）
- 新增的订阅者不能收到历史消息
- 创建订阅时会自动创建对应的 topic
- subscriber 的 `Start()` 和 `Close()` 都有 once 保护，多次调用安全
- 全局订阅者会收到所有 topic 的消息，包括后续新创建的 topic
