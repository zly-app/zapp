
# 指标收集器

```go
app := zapp.NewApp("myapp") // 必须在 NewApp 之后才能调用 metrics 相关方法

// 注册计数器
metrics.RegistryCounter("logger_level_num", "测试app", map[string]string{"app": "myapp"}, "level")

metrics.Counter("logger_level_num", metrics.Labels{"level": "info"}).Inc()
metrics.Counter("logger_level_num", metrics.Labels{"level": "debug"}).Inc()
```
