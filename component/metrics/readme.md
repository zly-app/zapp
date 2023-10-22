
# 指标收集器组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/prometheus/client_golang/prometheus](https://github.com/prometheus/client_golang)

# 配置

> 默认组件类型为 `metrics`

```yaml
components:
  metrics:
    default:
      ProcessCollector: true     # 启用进程收集器
      GoCollector: true          # 启用go收集器

      PullBind: ""          # pull模式bind地址, 如: ':8080', 如果为空则不启用pull模式
      PullPath: "/metrics"       # pull模式拉取路径, 如: '/metrics'

      PushAddress: "" # push模式 pushGateway地址, 如果为空则不启用push模式, 如: 'http://127.0.0.1:9091'
      PushInstance: "" # 实例名, 一般为ip或主机名
      PushTimeInterval: 10000 # push模式推送时间间隔, 单位毫秒
      PushRetry: 2 # push模式推送重试次数
      PushRetryInterval: 1000 # push模式推送重试时间间隔, 单位毫秒
```

# 示例

```go
app := zapp.NewApp("myapp") // 必须在 NewApp 之后才能调用 metrics 相关方法

// 注册计数器
metrics.RegistryCounter("logger_level_num", "测试app", map[string]string{"app": "myapp"}, "level")

metrics.Counter("logger_level_num", metrics.Labels{"level": "info"}).Inc()
metrics.Counter("logger_level_num", metrics.Labels{"level": "debug"}).Inc()
```
