// Package daemon 是守护进程模型 —— 拥有「进程级服务包装器」。
//
// # 契约
//
// 拥有：服务名、是否启用、kardianos 服务实例。
// 提供：
//   - Manager.Prepare —— 处理 install/start/stop/restart/status/uninstall 参数
//   - Manager.Run     —— 以服务或前台方式运行
//
// 允许调用：kardianos/service（底座机制）、log（默认日志）。
// 禁止调用：plugin / service / component / lifecycle。
//
// # 为什么用全局日志
//
// Prepare 发生在配置加载与 Logger 构建之前（install/stop 这类命令即使配置缺失也必须能执行），
// 此时只有默认日志可用。这是本模型唯一允许触碰 log.Log 的场景，已在代码处注明。
package daemon
