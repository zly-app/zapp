// Package component 是组件模型 —— 拥有「当前组件实例」。
//
// # 契约
//
// 拥有：容器中的唯一组件实例（Container.inner）。
// 提供：
//   - Container.Make  —— 构建默认组件并应用自定义包装
//   - Container.Get   —— 读取组件（唯一读入口）
//   - Container.Close —— 释放组件资源
//
// 允许调用：core（契约）、component/gpool、component/msgbus（同聚合内的机制）。
// 禁止调用：plugin / service / daemon / lifecycle。
//
// # 相对旧版的变化
//
// 旧版用包级变量 defaultComponent 持有组件，并提供 GetComponent / ResetComponent
// 两个全局改写入口。这违反「数据独占」——任何包都能改，且第二个 app 会覆盖第一个。
// 现在组件只由 Container 持有，App 经 core.IApp.GetComponent() 暴露；全局写入路径已删除。
package component
