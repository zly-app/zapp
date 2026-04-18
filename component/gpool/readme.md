# 协程池组件

> 基于 worker-pool 模型的协程池，控制并发 goroutine 数量

## 特性

- 可配置的 worker 数量和任务队列大小
- 支持同步/异步、阻塞/非阻塞多种提交方式
- 自动 Recover，panic 不会导致协程崩溃
- 支持无池模式（ThreadCount < 0），不做并发限制
- 多池管理，支持按名称创建和获取池实例
- 集成 zapp 配置和生命周期

## 示例

```go
package main

import (
	"fmt"
	"sync"

	"github.com/zly-app/zapp/component/gpool"
)

func main() {
	// 获取默认协程池
	pool := gpool.GetDefGPool()
	defer pool.Close()

	var wg sync.WaitGroup
	wg.Add(3)

	// 异步执行
	pool.Go(func() error {
		fmt.Println("task 1")
		return nil
	}, func(err error) {
		wg.Done()
	})

	// 同步执行
	err := pool.GoSync(func() error {
		fmt.Println("task 2")
		return nil
	})
	fmt.Println("task 2 result:", err)

	// 尝试执行（非阻塞）
	ok := pool.TryGo(func() error {
		fmt.Println("task 3")
		return nil
	}, func(err error) {
		wg.Done()
	})
	if !ok {
		fmt.Println("队列已满，提交失败")
	}

	wg.Wait()
}
```

## 配置

```yaml
components:
  gpool:
    default:
      JobQueueSize: 100000  # 任务队列大小，最小 100000
      ThreadCount: 100      # 并发 worker 数，最小 100，负数表示无限制
```

**ThreadCount 特殊值：**

| 值 | 行为 |
|----|------|
| 正数 | 指定固定 worker 数量 |
| 0 | 使用逻辑 CPU 数 × 2 |
| 负数 | 无池模式，不做并发限制 |

## API

### 异步提交

| 方法 | 说明 |
|------|------|
| `Go(fn, callback)` | 异步执行，队列满时阻塞 |
| `TryGo(fn, callback)` | 尝试异步执行，队列满返回 false |

### 同步提交

| 方法 | 说明 |
|------|------|
| `GoSync(fn)` | 同步执行，阻塞等待结果 |
| `TryGoSync(fn)` | 尝试同步执行，队列满返回 false |

### 批量提交

| 方法 | 说明 |
|------|------|
| `GoAndWait(fn...)` | 批量异步执行并等待，返回第一个错误 |
| `GoRetWait(fn...)` | 批量异步执行，返回等待函数 |

### 管理

| 方法 | 说明 |
|------|------|
| `Wait()` | 等待所有已提交任务完成 |
| `Close()` | 关闭池，正在执行的任务会等待完成 |

### 包级函数

| 函数 | 说明 |
|------|------|
| `gpool.GetDefGPool()` | 获取默认池 |
| `gpool.GetGPool(name)` | 获取命名池 |
| `gpool.GetCreator()` | 获取池管理器 |

## 说明

+ 所有执行函数自动 Recover，panic 不会导致程序崩溃
+ `Close()` 后再提交任务，callback 会收到 `ErrGPoolClosed`
+ worker 数量创建后不可动态调整

## AI 使用说明

如需了解本组件的详细架构和 AI 辅助开发指南，请参阅 [AI 说明文件](./.ai/gpool.md)。
