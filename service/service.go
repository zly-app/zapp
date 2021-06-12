/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package service

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 服务建造者
type serviceCreator func(app core.IApp, opts ...interface{}) core.IService

func (h serviceCreator) Create(app core.IApp, opts ...interface{}) core.IService {
	return h(app, opts...)
}

// 建造者列表
var creators = make(map[core.ServiceType]core.IServiceCreator)

// 注册服务建造者
func RegisterCreator(serviceType core.ServiceType, creator core.IServiceCreator) {
	if _, ok := creators[serviceType]; ok {
		logger.Log.Fatal("重复注册建造者", zap.String("serviceType", string(serviceType)))
	}
	creators[serviceType] = creator
}

// 注册服务建造者函数
func RegisterCreatorFunc(serviceType core.ServiceType, creatorFunc func(app core.IApp, opts ...interface{}) core.IService) {
	RegisterCreator(serviceType, serviceCreator(creatorFunc))
}

// 构建服务
func MakeService(app core.IApp, serviceType core.ServiceType, opts ...interface{}) core.IService {
	if creator, ok := creators[serviceType]; ok {
		return creator.Create(app, opts...)
	}
	app.Fatal("使用了未注册建造者的服务", zap.String("serviceType", string(serviceType)))
	return nil
}
