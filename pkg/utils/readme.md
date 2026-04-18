# utils

zapp 框架的通用工具库，提供并发控制、panic 恢复、反射判断、通配符匹配、网络代理、OpenTelemetry 链路追踪等基础能力。

## 模块概览

| 模块 | 全局变量 | 文件 | 说明 |
|------|---------|------|------|
| 并发控制 | `Go` | `go.go` | 协程并发执行与等待 |
| Context | `Ctx` | `context.go` | context 中存取 Logger，克隆无超时 context |
| 实例标识 | — | `get_instance.go` | 获取本机 IP / 实例名 |
| 网络代理 | — | `proxy.go` | SOCKS5 / HTTP 代理创建 |
| Panic 恢复 | `Recover` | `recover.go` | panic 捕获与调用栈提取 |
| 反射工具 | `Reflect` | `reflect.go` | 零值判断 |
| 三元运算 | `Ternary` | `ternary.go` | 三元表达式与短路取值 |
| 文本匹配 | `Text` | `text.go` | 通配符模糊匹配 |
| 链路追踪 | `Trace` | `trace.go` | OpenTelemetry span 封装 |

---

## Go — 并发控制

`var Go = goCli{}`

### GoAndWait

并行执行多个函数，等待全部完成后返回第一个非 nil 错误。自动 Recover panic。

```go
err := utils.Go.GoAndWait(
    func() error { return nil },
    func() error { return errors.New("fail") },
)
```

### GoRetWait

同 `GoAndWait`，但立即返回一个 wait 函数，由调用方决定何时等待。

```go
wait := utils.Go.GoRetWait(fn1, fn2)
// 做其他事情...
err := wait()
```

### GoQuery

泛型并发查询，按输入顺序返回结果。

```go
results, err := utils.GoQuery[string, User]([]string{"id1", "id2"}, func(id string) (User, error) {
    return getUser(id)
}, false) // ignoreErr=false 时查询出错直接返回错误
```

参数 `ignoreErr`：为 `true` 时忽略单条查询错误，结果切片会跳过该项。

---

## Ctx — Context 工具

`var Ctx = &ctxCli{}`

```go
// 存取 Logger
ctx = utils.Ctx.SaveLogger(ctx, logger)
log := utils.Ctx.GetLogger(ctx) // 不存在返回 nil

// 克隆 context（保留 Values，去除 Deadline/Done/Err）
newCtx := utils.Ctx.CloneContext(ctx)
```

---

## GetInstance / GetLocalIPs

```go
name := utils.GetInstance("default") // 返回本机第一个非回环 IP，无法获取则返回默认值
ips := utils.GetLocalIPs()            // 返回所有符合条件的本机 IP 列表
```

自动排除回环、链路本地、组播等无效地址。

---

## Proxy — 网络代理

### SOCKS5 代理

```go
sp, err := utils.NewSocks5Proxy("socks5://user:pwd@127.0.0.1:1080")
conn, err := sp.Dial("tcp", "example.com:80")
conn, err := sp.DialContext(ctx, "tcp", "example.com:80")
```

支持 `socks5` 和 `socks5h` scheme。

### HTTP 代理

```go
hp, err := utils.NewHttpProxy("https://user:pwd@127.0.0.1:1080")
// 或 socks5 地址也可以
transport := &http.Transport{}
hp.SetProxy(transport)
```

支持 `http`、`https`、`socks5`、`socks5h` scheme。

---

## Recover — Panic 恢复

`var Recover = new(recoverCli)`

```go
// 包装函数调用，捕获 panic 并转为 error
err := utils.Recover.WrapCall(func() error {
    panic("oops")
})

// 判断是否为 Recover 产生的错误
utils.Recover.IsRecoverError(err) // true

// 获取调用栈详情
detail := utils.Recover.GetRecoverErrorDetail(err) // 错误信息 + 调用栈
lines := utils.Recover.GetRecoverErrors(err)       // []string
```

`RecoverError` 接口包含原始错误和调用栈信息：

```go
type RecoverError interface {
    error
    Err() error
    Callers() []*Caller
}
```

---

## Reflect — 反射工具

`var Reflect = new(reflectUtil)`

```go
utils.Reflect.IsZero(nil)           // true
utils.Reflect.IsZero("")            // true
utils.Reflect.IsZero(0)             // true
utils.Reflect.IsZero([]int{})       // true (nil slice)
utils.Reflect.IsZero(myStruct{})    // 递归判断所有字段是否为零值
```

支持基本类型、数组、切片、Map、结构体（递归）等。

---

## Ternary — 三元运算

`var Ternary = &ternaryUtil{}`

```go
// 三元表达式
val := utils.Ternary.Ternary(ok, "yes", "no")

// 短路取值：返回第一个非零值，全为零则返回最后一个
val := utils.Ternary.Or("", 0, "fallback") // "fallback"
val := utils.Ternary.Or("", "hello", "fallback") // "hello"
```

---

## Text — 通配符匹配

`var Text = &textUtil{}`

```go
utils.Text.IsMatchWildcard("foo.txt", "f*.txt")   // true
utils.Text.IsMatchWildcard("foo.txt", "f?.txt")    // false (? 匹配单个字符)
utils.Text.IsMatchWildcard("foo.txt", "*.txt")     // true

// 匹配任意一个模式即可
utils.Text.IsMatchWildcardAny("foo.txt", "bar*", "f*") // true
```

- `?` 匹配单个字符
- `*` 匹配任意字符串（含空字符串）

---

## Trace — OpenTelemetry 链路追踪

`var Trace = &otelCli{}`

```go
// 创建 span
ctx, span := utils.Trace.StartSpan(ctx, "operation", utils.Trace.AttrKey("key").String("val"))
defer utils.Trace.EndSpan(span)

// 快捷方式：CtxStart / CtxEnd
ctx = utils.Trace.CtxStart(ctx, "operation")
defer utils.Trace.CtxEnd(ctx)

// 设置属性 / 添加事件 / 标记错误
utils.Trace.SetSpanAttributes(span, utils.Trace.AttrKey("key").String("val"))
utils.Trace.AddSpanEvent(span, "event-name")
utils.Trace.MarkSpanAnError(span, err)

// 错误事件（含 deadline 和 panic 信息）
utils.Trace.CtxErrEvent(ctx, "failed", err)

// 获取 TraceID / SpanID
traceID, spanID := utils.Trace.GetOTELTraceID(ctx)

// 传播上下文到 HTTP Headers / Map
utils.Trace.SaveToHeaders(ctx, req.Header)
ctx, span = utils.Trace.GetSpanWithHeaders(ctx, req.Header)

utils.Trace.SaveToMap(ctx, map[string]string{...})
ctx, span = utils.Trace.GetSpanWithMap(ctx, map[string]string{...})

// Deadline 信息作为 span 属性
kv := utils.Trace.GetSpanKVWithDeadline(ctx)
```
