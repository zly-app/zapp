/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/21
   Description :
-------------------------------------------------
*/

package core

// 服务
type IService interface {
	// 注入, 根据服务不同具有不同作用, 具体参考服务实现说明
	Inject(a ...interface{})
	// 开始服务
	Start() error
	// 关闭服务
	Close() error
}

// 服务建造者
type IServiceCreator interface {
	// 创建服务
	Create(app IApp) IService
}

// 服务类型
type ServiceType string
