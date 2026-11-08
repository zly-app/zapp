// Package filter 是过滤器链模型 —— 拥有「过滤器实例与调用链」。
//
// # 契约
//
// 拥有：客户端/服务端过滤器实例表、按客户端或服务名构建的链。
// 提供：MakeFilter（按配置构建）、InitFilter（初始化并注入 app）、CloseFilter（关闭）、
// GetClientFilter / GetServiceFilter（取链）。
// 允许调用：core（契约）、log（默认日志）、本包配置读取。
//
// # 归属说明（待定项）
//
// 过滤器实例与链当前存放在**包级变量**里，装配入口是 MakeFilter / CloseFilter，
// 读取入口 GetClientFilter 不接受 app 句柄——因为消费方（HTTP/Kafka 等客户端组件）
// 在外部仓库中，改签名会波及整个生态。
//
// 现状的定性：这是「共享模型 + 单一激活实例」——链数据由本包独占，装配只允许
// 组合根（lifecycle.Initialize）调用一次，模型内部不得自行重建。它不是纯粹的数据独占：
// 两个 app 会互相覆盖链。这一点已登记为待定债务（见 docs/TODO.md）。
//
// 彻底方案（未做）：引入链集合实例 + 组合根显式注入，GetClientFilter 改为从注入句柄读取。
// 需要同步改动外部消费方，单独立项。
package filter
