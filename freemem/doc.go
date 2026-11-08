// Package freemem 是内存释放模型 —— 拥有「释放间隔」。
//
// # 契约
//
// 拥有：释放间隔（来自配置 Frame.FreeMemoryInterval）。
// 提供：Releaser.Start —— 在后台按间隔调用 debug.FreeOSMemory，ctx 取消即停。
//
// 允许调用：context / runtime/debug（底座机制）。
// 禁止调用：任何其他模型。
//
// # 终止保障
//
// 后台协程只有一个出口：ctx.Done()。未启用时根本不启动协程——
// 没有终止条件的常驻协程是不允许的。
package freemem
