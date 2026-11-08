// Package service 是服务模型 —— 拥有「已启用服务实例集合」。
//
// # 契约
//
// 拥有：启用列表、服务实例表、就绪观察选项。
// 提供：
//   - Manager.Make  —— 解析启用清单（含自定义启用钩子）并构建实例
//   - Manager.Start —— 逐个启动并等待就绪
//   - Manager.Close —— 逐个关闭
//   - Manager.Get / Inject —— 窄查询与注入
//
// 允许调用：core（契约）、本包 WaitRun（就绪观察）、本包建造者注册表。
// 禁止调用：plugin / component / daemon / lifecycle。
package service
