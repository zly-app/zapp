
# 用于快速构建项目的基础库

---
<!-- TOC -->

- [开始](#%E5%BC%80%E5%A7%8B)
- [插件使用](#%E6%8F%92%E4%BB%B6%E4%BD%BF%E7%94%A8)
    - [组件插件](#%E7%BB%84%E4%BB%B6%E6%8F%92%E4%BB%B6)
    - [服务插件](#%E6%9C%8D%E5%8A%A1%E6%8F%92%E4%BB%B6)
- [配置](#%E9%85%8D%E7%BD%AE)
- [依赖包说明](#%E4%BE%9D%E8%B5%96%E5%8C%85%E8%AF%B4%E6%98%8E)
- [插件开发规范](#%E6%8F%92%E4%BB%B6%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
    - [组件插件开发规范](#%E7%BB%84%E4%BB%B6%E6%8F%92%E4%BB%B6%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
    - [服务插件开发规范](#%E6%9C%8D%E5%8A%A1%E6%8F%92%E4%BB%B6%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
- [使用说明](#%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E)
    - [守护进程](#%E5%AE%88%E6%8A%A4%E8%BF%9B%E7%A8%8B)
    - [Handler](#handler)
    - [根据配置决定启动哪些服务](#%E6%A0%B9%E6%8D%AE%E9%85%8D%E7%BD%AE%E5%86%B3%E5%AE%9A%E5%90%AF%E5%8A%A8%E5%93%AA%E4%BA%9B%E6%9C%8D%E5%8A%A1)
    - [会话独立日志](#%E4%BC%9A%E8%AF%9D%E7%8B%AC%E7%AB%8B%E6%97%A5%E5%BF%97)

<!-- /TOC -->
---

# 开始

```go
app := zapp.NewApp("test")
app.Run()
```

# 插件使用

## 组件插件

> 我们实现了一些组件插件, 可以在 [这里](https://github.com/zly-app/component) 找到

## 服务插件

> 我们实现了一些服务插件, 可以在 [这里](https://github.com/zly-app/service) 找到

# 配置

> 请转到 [这里](./config)

# 依赖包说明

> 作为一个基础库, 我们尽量减少了依赖包, 只需要以下包依赖

+ [github.com/spf13/viper](https://github.com/spf13/viper) v1.7.1
+ [github.com/takama/daemon](https://github.com/takama/daemon) v1.0.0
+ [go.uber.org/zap](https://github.com/uber-go/zap) v1.16.0

# 插件开发规范

## 组件插件开发规范

> 我们提供了自定义组件创建选项 `zapp.WithCustomComponent`, 这个核心功能能让我们支持任何组件

> ... 待完善

## 服务插件开发规范

> ... 待完善

# 使用说明

## 守护进程

> 初始化时添加 `zapp.WithEnableDaemon()` 选项, 构建后使用以下命令

```text
./app文件
    install [args]          安装服务, args 是运行时传递给 app 的参数
    remove                  移除服务
    start                   启动app
    stop                    停止app
    status                  查看运行状态
```

## Handler

> 初始化时添加 `zapp.WithHandler(...)` 选项 

```text
BeforeInitializeHandler         在app初始化前
AfterInitializeHandler          在app初始化后
BeforeStartHandler              在app启动前
AfterStartHandler               在app启动后
BeforeExitHandler               在app退出前
AfterExitHandler                在app退出后
```

## 根据配置决定启动哪些服务

> 初始化时添加 `zapp.WithCustomEnableService(...)` 选项, zapp 会根据返回值来决定开启和关闭哪些服务

```go
zapp.WithCustomEnableService(func(app core.IApp) (servers map[core.ServiceType]bool) {
	servers = make(map[core.ServiceType]bool)
	if app.GetConfig().GetLabel("api") == "true" {
		servers["api"] = true
	}
	return servers
})
```

## 会话独立日志

```text
core.ILogger 提供了 NewMirrorLogger(tag ...string) ILogger 方法用于创建一个镜像 logger.
在会话开始时可以通过这个方法创建会话内使用的logger, 会话结束后无需考虑销毁它, 它会自动回收.
使用镜像logger打印日志会产生一个全局日志id, 并且我们会根据不同的全局日志id输出不同的颜色.
```
