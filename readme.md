用于快速构建项目的基础库.

---

## AI 开发指引

> 本节专为 AI 助手（如 Claude、GPT 等）设计，帮助 AI 快速理解项目并提供开发协助。

### 何时阅读 AI_REFERENCE.md

**AI 应优先阅读 `AI_REFERENCE.md` 而非扫描整个项目代码**，在以下场景中：

| 场景 | 说明 |
|------|------|
| 快速开发任务 | 需要基于 zapp 创建新组件/插件/服务 |
| 接口查询 | 需要了解核心接口定义（IComponent、IPlugin、IService 等） |
| 架构理解 | 需要快速理解项目整体架构和生命周期 |
| 配置使用 | 需要了解配置加载方式、Apollo 配置或配置热加载 |
| 工具包使用 | 需要使用 pkg 下的序列化器、压缩器、工具函数等 |
| 代码生成 | 需要生成符合 zapp 规范的代码模板 |

### 何时仍需扫描项目代码

以下场景 AI **仍应扫描具体代码文件**：

> **仅在以下情况需要深入阅读源码**:
> - 调试具体问题时
> - 实现高级定制功能时
> - 文档信息不足以解决当前问题时

### AI_REFERENCE.md 定位

- **位置**: `./AI_REFERENCE.md`
- **内容**: 核心接口定义、架构概览、配置系统、工具包索引、开发指引、关键文件索引
- **目的**: 让 AI 无需扫描整个代码仓库即可快速开发

### 构建新项目

如需构建完整业务项目，推荐使用 [uapp](https://github.com/zly-app/uapp)，它封装了 zapp 并预置常用组件/插件和 Apollo 配置支持，且提供脚手架文档可按需求关键词快速定位工具。

---
<!-- TOC -->

- [开始](#%E5%BC%80%E5%A7%8B)
- [扩展性](#%E6%89%A9%E5%B1%95%E6%80%A7)
    - [组件](#%E7%BB%84%E4%BB%B6)
    - [插件](#%E6%8F%92%E4%BB%B6)
    - [filter](#filter)
    - [服务](#%E6%9C%8D%E5%8A%A1)
- [配置](#%E9%85%8D%E7%BD%AE)
- [使用说明](#%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E)
    - [守护进程](#%E5%AE%88%E6%8A%A4%E8%BF%9B%E7%A8%8B)
    - [Handler](#handler)
    - [运行时自定义启用插件](#%E8%BF%90%E8%A1%8C%E6%97%B6%E8%87%AA%E5%AE%9A%E4%B9%89%E5%90%AF%E7%94%A8%E6%8F%92%E4%BB%B6)
    - [运行时自定义启用服务](#%E8%BF%90%E8%A1%8C%E6%97%B6%E8%87%AA%E5%AE%9A%E4%B9%89%E5%90%AF%E7%94%A8%E6%9C%8D%E5%8A%A1)
    - [独特的日志](#%E7%8B%AC%E7%89%B9%E7%9A%84%E6%97%A5%E5%BF%97)
- [app生命周期](#app%E7%94%9F%E5%91%BD%E5%91%A8%E6%9C%9F)
    - [初始化](#%E5%88%9D%E5%A7%8B%E5%8C%96)
    - [用户操作](#%E7%94%A8%E6%88%B7%E6%93%8D%E4%BD%9C)
    - [启动](#%E5%90%AF%E5%8A%A8)
    - [退出](#%E9%80%80%E5%87%BA)

<!-- /TOC -->
---

# 开始

推荐使用 [uapp](https://github.com/zly-app/uapp)，它封装了 zapp 并预置常用组件、插件（pprof/otlp/apollo_provider 等）和 Apollo 配置支持，且提供脚手架文档方便 AI 快速生成项目。

直接使用 zapp：

```go
app := zapp.NewApp("test")
app.Run()
```

# 扩展性

## 组件

+ 我们实现了一些组件, 可以在 [这里](https://github.com/zly-app/component) 找到
+ [这里](./component) 内置了一些组件

## 插件

+ 我们实现了一些插件, 可以在 [这里](https://github.com/zly-app/plugin) 找到

## filter

+ [这里](./filter) 内置了一些 filter

## 服务

+ 我们实现了一些服务, 可以在 [这里](https://github.com/zly-app/service) 找到

# 配置

+ 请转到 [这里](./config)

# 使用说明

## 守护进程

初始化时添加 `zapp.WithEnableDaemon()` 选项, 构建后使用以下命令

```text
./app文件
    install [args]          安装服务, args 是运行时传递给 app 的参数
    remove                  移除服务
    start                   启动app
    stop                    停止app
    status                  查看运行状态
```

## Handler

初始化时添加 `zapp.WithHandler(...)` 选项 

```text
BeforeInitializeHandler         在app初始化前
AfterInitializeHandler          在app初始化后
BeforeStartHandler              在app启动前
AfterStartHandler               在app启动后
BeforeExitHandler               在app退出前
AfterExitHandler                在app退出后
```

## 运行时自定义启用插件

初始化时添加 `zapp.WithCustomEnablePlugin(...)` 选项, zapp 会根据返回值来决定开启和关闭哪些插件

```go
zapp.WithCustomEnablePlugin(func(app core.IApp, plugins []core.PluginType) []core.PluginType {
    if !app.GetConfig().HasFlag("my_plugin") {
        plugins = append(plugins, "my_plugin")
    }
    return plugins
})
```

## 运行时自定义启用服务

初始化时添加 `zapp.WithCustomEnableService(...)` 选项, zapp 会根据返回值来决定开启和关闭哪些服务

```go
zapp.WithCustomEnableService(func(app core.IApp, services []core.ServiceType) []core.ServiceType) Option {
    if !app.GetConfig().HasFlag("api_service") {
        services = append(services, "api")
    }
    return services
})
```

## 独特的日志

`core.ILogger` 提供了 `NewTraceLogger(ctx context.Context, fields ...zap.Field) ILogger` 方法用于创建一个带链路id的 logger(前提是ctx中包含有效的span).<br>
使用生成的log打印日志会带上链路id, 并且我们会根据不同的链路id输出不同的颜色.

`core.ILogger` 提供了 `NewSessionLogger(fields ...zap.Field) ILogger` 方法用于创建一个会话 logger.<br>
使用会话logger打印日志会产生一个全局日志id, 并且我们会根据不同的全局日志id输出不同的颜色.

日志打印时可以将 `ctx` 传入, 如果 `ctx` 中包含 `traceID` 那么在日志输出中会带上 `traceID`. 示例 `app.Info(ctx, "test")`

# app生命周期

初始化 > 用户操作 > 启动 > 退出

## 初始化

`app := zapp.NewApp(...)` > 生成`BaseContext` > 加载配置 > 初始化日志记录器 > 构建组件 > 构建插件 > 构建并初始化 filter > 构建服务   

## 用户操作

用户在这里对服务进行注入, 如注入插件, 注入服务等.

## 启动

`app.Run()` > 启动插件 > 启动服务 > 启动内存释放任务 > 阻塞等待退出信号

## 退出

`app.Exit() 或收到退出信号` > 关闭`BaseContext` > 停止内存释放任务 > 关闭服务 > 关闭 filter > 关闭插件 > 释放组件资源 > `结束之前调用app.Run()的阻塞`
