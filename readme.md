
# 用于快速构建项目的基础库

---
<!-- TOC -->

- [示例](#%E7%A4%BA%E4%BE%8B)
- [插件使用](#%E6%8F%92%E4%BB%B6%E4%BD%BF%E7%94%A8)
    - [组件插件](#%E7%BB%84%E4%BB%B6%E6%8F%92%E4%BB%B6)
        - [以 xorm组件 为例](#%E4%BB%A5-xorm%E7%BB%84%E4%BB%B6-%E4%B8%BA%E4%BE%8B)
    - [服务插件](#%E6%9C%8D%E5%8A%A1%E6%8F%92%E4%BB%B6)
        - [以 api服务 为例](#%E4%BB%A5-api%E6%9C%8D%E5%8A%A1-%E4%B8%BA%E4%BE%8B)
- [配置](#%E9%85%8D%E7%BD%AE)
- [依赖包说明](#%E4%BE%9D%E8%B5%96%E5%8C%85%E8%AF%B4%E6%98%8E)
- [插件开发规范](#%E6%8F%92%E4%BB%B6%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
    - [组件插件开发规范](#%E7%BB%84%E4%BB%B6%E6%8F%92%E4%BB%B6%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
    - [服务插件开发规范](#%E6%9C%8D%E5%8A%A1%E6%8F%92%E4%BB%B6%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)

<!-- /TOC -->
---

# 示例

```go
app := zapp.NewApp("test")
app.Run()
```

# 插件使用

## 组件插件

> 我们实现了一些组件插件, 可以在 [这里](https://github.com/zly-app?tab=repositories&q=-plugin&type=&language=) 找到

### 以 [xorm组件](https://github.com/zly-app/xorm-plugin) 为例

```go
// 定义自己的组件
type Component struct {
	core.IComponent
	xorm_plugin.IXormComponent
	// ... 其他组件
}

// 重写Close()
func (c *Component) Close() {
	c.IXormComponent.Close()
	// ... 关闭其他组件
}

app := zapp.NewApp("test",
    zapp.WithCustomComponent(func(app core.IApp) core.IComponent { // 自定义返回自己的组件
        return &Component{
            IComponent:     app.GetComponent(),       // 设置原始组件
            IXormComponent: xorm_plugin.NewXorm(app), //  设置xorm组件
            // ... 设置其他组件
        }
    }),
)

c := app.GetComponent().(*Component) // 直接转换为自己的组件
c.GetXorm()                          // 获取 xorm 组件
// c.GetXXX() 获取其它组件
```

## 服务插件

> 我们实现了一些服务插件, 可以在 [这里](https://github.com/zly-app?tab=repositories&q=-service&type=&language=) 找到

### 以 [api服务](https://github.com/zly-app/api-service) 为例

```go
// 注册api服务
api_service.RegistryService()
// ... 注册其他服务

app := zapp.NewApp("test",
    api_service.WithApiService(), // 启用api服务
    // ... 启用其它服务
)

// 注册路由
api_service.RegistryApiRouter(app, func(c core.IComponent, router iris.Party) {
    router.Get("/", api_service.Wrap(func(ctx *api_service.Context) interface{} {
        return "hello"
    }))
})

// ... 其它服务的注入

// 运行
app.Run()
```

# 配置

> 请转到 [这里](./config)

# 依赖包说明

> 作为一个基础库, 我们尽量减少了依赖包, 但是组件插件和服务插件并没有, 因为它们需要.

# 插件开发规范

## 组件插件开发规范

> 我们提供了自定义组件创建选项, 这个核心功能能让我们支持任何组件

> 待完善

## 服务插件开发规范

> 待完善
