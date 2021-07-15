/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/21
   Description :
-------------------------------------------------
*/

package core

// 插件
type IPlugin interface {
	// 注入, 根据插件不同具有不同作用, 具体参考插件实现说明
	Inject(a ...interface{})
	// 启动插件
	Start() error
	// 关闭插件
	Close() error
}

// 插件建造者
type IPluginCreator interface {
	// 创建插件
	Create(app IApp) IPlugin
}

// 插件类型
type PluginType string
