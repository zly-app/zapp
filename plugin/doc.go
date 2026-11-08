// Package plugin 是插件模型 —— 拥有「已启用插件实例集合」。
//
// # 契约
//
// 拥有：启用列表、插件实例表、依赖启动器。
// 提供：
//   - Manager.Make  —— 解析启用清单（含自定义启用钩子）并构建实例
//   - Manager.Start —— 按依赖顺序启动
//   - Manager.Close —— 按依赖逆序关闭
//   - Manager.Get / Inject —— 窄查询与注入
//
// 允许调用：core（契约）、pkg/depender（底座机制）、本包建造者注册表。
// 禁止调用：service / component / daemon / lifecycle —— 插件不认识服务，也不编排时序。
//
// # 归属说明
//
// 本包同时承担两个职责：建造者注册表（RegisterCreator，进程级、底座机制）
// 与运行时实例管理（Manager，模型数据）。二者分离在文件层面：
// plugin.go 是注册表，manager.go 是模型。
package plugin
