package service

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
)

// Manager 是服务运行时模型：拥有「已启用服务实例集合」这份数据，以及
// 「逐个等待就绪、逐个关闭」这条规则。
//
// 它不负责初始化时序、不触发生命周期钩子 —— 那是业务层（lifecycle）的事。
// 允许调用：core（契约）、同包 WaitRun（就绪观察）。禁止调用 plugin/component/lifecycle。
type Manager struct {
	app      core.IApp
	logger   core.ILogger
	types    []core.ServiceType
	services map[core.ServiceType]core.IService

	// exitOnErrOfObserve 在服务不稳定观察阶段出现错误时是否退出
	exitOnErrOfObserve bool
}

// NewManager 创建一个空的服务管理器
func NewManager(app core.IApp, logger core.ILogger, exitOnErrOfObserve bool) *Manager {
	return &Manager{
		app:                app,
		logger:             logger,
		services:           make(map[core.ServiceType]core.IService),
		exitOnErrOfObserve: exitOnErrOfObserve,
	}
}

// Make 解析启用列表并构建全部服务实例
func (m *Manager) Make(types []core.ServiceType, customize []func(core.IApp, []core.ServiceType) []core.ServiceType) error {
	resolved := types
	for _, fn := range customize {
		resolved = fn(m.app, resolved)
	}

	seen := make(map[core.ServiceType]struct{}, len(resolved))
	for _, t := range resolved {
		if _, ok := seen[t]; ok {
			return fmt.Errorf("服务启用重复: %v", t)
		}
		seen[t] = struct{}{}
	}

	m.types = resolved
	for _, t := range resolved {
		m.services[t] = MakeService(m.app, t)
	}
	return nil
}

// Start 逐个启动服务，并等待其通过就绪观察
func (m *Manager) Start() error {
	for _, serviceType := range m.types {
		s, ok := m.services[serviceType]
		if !ok {
			return fmt.Errorf("服务查找失败: %v", serviceType)
		}

		err := WaitRun(m.app, &WaitRunOption{
			ServiceType:        serviceType,
			ExitOnErrOfObserve: m.exitOnErrOfObserve,
			RunServiceFn:       s.Start,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Close 逐个关闭服务
func (m *Manager) Close() {
	for _, serviceType := range m.types {
		s, ok := m.services[serviceType]
		if !ok {
			continue
		}
		if err := s.Close(); err != nil {
			m.logger.Error("服务关闭失败", zap.String("serviceType", string(serviceType)), zap.Error(err))
		}
	}
}

// Get 查询已构建的服务
func (m *Manager) Get(serviceType core.ServiceType) (core.IService, bool) {
	s, ok := m.services[serviceType]
	return s, ok
}

// Inject 向指定服务注入依赖
func (m *Manager) Inject(serviceType core.ServiceType, ignoreDisabled bool, a ...interface{}) {
	s, ok := m.Get(serviceType)
	if !ok {
		if ignoreDisabled {
			return
		}
		m.logger.Fatal("注入失败, 未启用服务", zap.String("serviceType", string(serviceType)))
	}
	s.Inject(a...)
}

// Types 返回已启用的服务类型列表（只读副本）
func (m *Manager) Types() []core.ServiceType {
	out := make([]core.ServiceType, len(m.types))
	copy(out, m.types)
	return out
}
