
# 用于快速构建项目的基础库

---
<!-- TOC -->

- [开始](#%E5%BC%80%E5%A7%8B)
- [扩展性](#%E6%89%A9%E5%B1%95%E6%80%A7)
    - [组件](#%E7%BB%84%E4%BB%B6)
    - [服务](#%E6%9C%8D%E5%8A%A1)
- [配置](#%E9%85%8D%E7%BD%AE)
- [依赖包说明](#%E4%BE%9D%E8%B5%96%E5%8C%85%E8%AF%B4%E6%98%8E)
- [扩展开发规范](#%E6%89%A9%E5%B1%95%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
    - [组件开发规范](#%E7%BB%84%E4%BB%B6%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
    - [服务开发规范](#%E6%9C%8D%E5%8A%A1%E5%BC%80%E5%8F%91%E8%A7%84%E8%8C%83)
- [使用说明](#%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E)
    - [守护进程](#%E5%AE%88%E6%8A%A4%E8%BF%9B%E7%A8%8B)
    - [Handler](#handler)
    - [运行时自定义服务](#%E8%BF%90%E8%A1%8C%E6%97%B6%E8%87%AA%E5%AE%9A%E4%B9%89%E6%9C%8D%E5%8A%A1)
    - [会话独立日志](#%E4%BC%9A%E8%AF%9D%E7%8B%AC%E7%AB%8B%E6%97%A5%E5%BF%97)
- [app生命周期](#app%E7%94%9F%E5%91%BD%E5%91%A8%E6%9C%9F)
    - [初始化](#%E5%88%9D%E5%A7%8B%E5%8C%96)
    - [用户操作](#%E7%94%A8%E6%88%B7%E6%93%8D%E4%BD%9C)
    - [启动](#%E5%90%AF%E5%8A%A8)
    - [退出](#%E9%80%80%E5%87%BA)

<!-- /TOC -->
---

# 开始

```go
app := zapp.NewApp("test")
app.Run()
```

# 扩展性

## 组件

> 我们实现了一些组件, 可以在 [这里](https://github.com/zly-app/component) 找到

## 服务

> 我们实现了一些服务, 可以在 [这里](https://github.com/zly-app/service) 找到

# 配置

> 请转到 [这里](./config)

# 依赖包说明

> 作为一个基础库, 我们尽量减少了依赖包, 只需要以下包依赖

+ [github.com/spf13/viper](https://github.com/spf13/viper) v1.7.1
+ [github.com/takama/daemon](https://github.com/takama/daemon) v1.0.0
+ [go.uber.org/zap](https://github.com/uber-go/zap) v1.16.0

# 扩展开发规范

## 组件开发规范

> 我们提供了自定义组件创建选项 `zapp.WithCustomComponent`, 这个核心功能能让我们支持任何组件

> ... 待完善

## 服务开发规范

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

## 运行时自定义服务

> 初始化时添加 `zapp.WithCustomService(...)` 选项, zapp 会根据返回值来决定开启和关闭哪些服务

```go
zapp.WithCustomService(func(app core.IApp, services map[core.ServiceType]bool, servicesOpts map[core.ServiceType][]interface{}) {
    if app.GetConfig().HasFlag("api_service") {
        services["api"] = true
    }
})
```

## 会话独立日志

`core.ILogger` 提供了 `NewSessionLogger(fields ...zap.Field) ILogger` 方法用于创建一个会话 logger.<br>
使用会话logger打印日志会产生一个全局日志id, 并且我们会根据不同的全局日志id输出不同的颜色.

# app生命周期

初始化 > 用户操作 > 启动 > 退出

## 初始化

`app := zapp.NewApp(...)` > 生成`BaseContext` > 加载配置 > 初始化日志记录器 > 初始化组件 > 创建服务   

## 用户操作

用户在这里对服务进行注入

## 启动

`app.Run()` > 启动服务 > 启动内存释放任务 > 阻塞等待退出信号

## 退出

`app.Exit() 或收到退出信号` > 关闭`BaseContext` > 停止内存释放任务 > 关闭服务 > 释放组件资源 > `结束之前调用app.Run()的阻塞` 

