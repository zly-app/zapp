package msgbus_consume

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/service"
)

// 服务类型
const ServiceType core.ServiceType = "msgbus-consume"

// 启用msgbus-consume服务
func WithService() zapp.Option {
	service.RegisterCreatorFunc(ServiceType, func(app core.IApp, opts ...interface{}) core.IService {
		return NewMsgbusConsumeService(app) // todo opts
	})
	return zapp.WithService(ServiceType)
}

// 注册handler
func RegistryHandler(app core.IApp, topic string, threadCount int, handler core.MsgbusHandler) {
	app.InjectService(ServiceType, &ConsumerConfig{
		IsGlobal:    false,
		Topic:       topic,
		ThreadCount: threadCount,
		Handler:     handler,
	})
}

// 注册全局handler
func RegistryGlobalHandler(app core.IApp, threadCount int, handler core.MsgbusHandler) {
	app.InjectService(ServiceType, &ConsumerConfig{
		IsGlobal:    true,
		Topic:       "",
		ThreadCount: threadCount,
		Handler:     handler,
	})
}
