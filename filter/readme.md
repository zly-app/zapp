
# 过滤器

将请求相关的特定逻辑组件化，插件化

# 组件请求、响应时接入过滤器

客户端触发

```go
// inject 模式
err := filter.TriggerClientFilterInject(ctx, clientName, req, rsp, func(ctx context.Context, req, rsp interface{}) error{
	return nil
})
// return 模式
rsp, err := filter.TriggerClientFilter(ctx, clientName, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error){
	return "xxx", nil
})
```

服务端触发

```go
// inject 模式
err := filter.TriggerServiceFilterInject(ctx, serviceName, req, rsp, func(ctx context.Context, req, rsp interface{}) error{
	return nil
})
// return 模式
rsp, err := filter.TriggerServiceFilter(ctx, serviceName, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error){
	return "xxx", nil
})
```

# 过滤器开发

首先实现 [core.Filter](../core/filter.go) 接口.

然后调用 `filter.RegisterFilterCreator` 注册过滤器. 如下

```go
clientFilterCreator = func() Filter{
    return newMyFilter()
}
serviceFilterCreator = func() Filter{
	return newMyFilter()
}
filter.RegisterFilterCreator("myFilter", clientFilterCreator, serviceFilterCreator)
```

如果仅注册服务过滤器或客户端过滤器, 可以将不需要的设为nil, 如下

```go
// 仅注册客户端过滤器
filter.RegisterFilterCreator("myFilter", clientFilterCreator, nil)
// 金注册服务端过滤器
filter.RegisterFilterCreator("myFilter", nil, serviceFilterCreator)
```


## 配置

```yaml
filters:
  service: # 服务过滤器
    default: # 对没有独立配置的服务设置默认的过滤器, 默认包含 base
      - filter1
      - filter2
    myService: # 独立设置服务的过滤器 
      - filter2
  client: # 客户端过滤器
    default: # 对没有独立配置的服务设置默认的过滤器
      default:
        - filter1
        - filter2
    sqlx: # sqlx类型
      default: # sqlx类型中对没有独立配置的客户端设置默认的过滤器
        - filter1
        - filter2
      mySqlx: # 独立设置客户端的过滤器 
        - filter2

  config: # 过滤器配置, 不同过滤器配置不同或不需要配置
    filter1:
      foo: bar
    filter2:
      foo: bar
```
