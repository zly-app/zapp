package component

import (
	"github.com/zly-app/zapp/core"
)

// Container 是组件容器的持有者：拥有「当前组件实例」这一份数据。
//
// 它取代了原先 component 包的包级全局变量（defaultComponent / GetComponent / ResetComponent）——
// 全局变量的问题是不止一个 app 时会互相覆盖，且谁都能改、改坏了没人负责。
// 现在组件实例只有一个所有者：App 持有 Container，其余人通过 core.IApp.GetComponent() 读。
//
// 允许调用：core（契约）、同包 NewComponent。禁止调用 plugin/service/lifecycle。
type Container struct {
	inner core.IComponent
}

// NewContainer 创建一个空容器
func NewContainer() *Container {
	return &Container{}
}

// Make 构建组件，并依次应用使用者的自定义包装函数
//
// customize 允许把默认组件替换成带额外能力的实现；为空表示使用默认组件。
func (c *Container) Make(app core.IApp, customize []func(core.IApp, core.IComponent) core.IComponent) {
	c.inner = NewComponent(app)
	for _, fn := range customize {
		c.inner = fn(app, c.inner)
	}
}

// Get 返回当前组件实例（未构建时为 nil，由调用方决定如何处理）
func (c *Container) Get() core.IComponent {
	return c.inner
}

// Ready 报告组件是否已构建
func (c *Container) Ready() bool {
	return c.inner != nil
}

// Close 释放组件资源
func (c *Container) Close() {
	if c.inner != nil {
		c.inner.Close()
	}
}
