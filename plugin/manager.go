package plugin

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/depender"
)

// Manager 是插件运行时模型：拥有「已启用插件实例集合」这份数据，以及
// 「按依赖顺序启动、逆序关闭」这条规则。
//
// 它不负责初始化时序、不触发生命周期钩子 —— 那是业务层（lifecycle）的事。
// 允许调用：core（契约）、pkg/depender（底座机制）。禁止调用 service/component/lifecycle。
type Manager struct {
	app     core.IApp
	logger  core.ILogger
	types   []core.PluginType
	plugins map[core.PluginType]core.IPlugin
	dep     depender.Depender
}

// NewManager 创建一个空的插件管理器
func NewManager(app core.IApp, logger core.ILogger) *Manager {
	return &Manager{
		app:     app,
		logger:  logger,
		plugins: make(map[core.PluginType]core.IPlugin),
	}
}

// Make 解析启用列表并构建全部插件实例
//
// customize 是使用者提供的「自定义启用」钩子，会在去重前依次执行；返回错误表示启用列表本身非法。
func (m *Manager) Make(types []core.PluginType, customize []func(core.IApp, []core.PluginType) []core.PluginType) error {
	resolved := types
	for _, fn := range customize {
		resolved = fn(m.app, resolved)
	}

	seen := make(map[core.PluginType]struct{}, len(resolved))
	for _, t := range resolved {
		if _, ok := seen[t]; ok {
			return fmt.Errorf("插件启用重复: %v", t)
		}
		seen[t] = struct{}{}
	}

	m.types = resolved
	for _, t := range resolved {
		m.plugins[t] = MakePlugin(m.app, t)
	}
	return nil
}

// Start 按依赖顺序启动全部插件
func (m *Manager) Start() error {
	items := make([]depender.Item, len(m.types))
	for i, pluginType := range m.types {
		p, ok := m.plugins[pluginType]
		if !ok {
			return fmt.Errorf("插件查找失败: %v", pluginType)
		}
		var dps []string
		if dp, ok := p.(core.Depender); ok {
			dps = dp.DependsOn()
		}
		items[i] = depender.NewItem(string(pluginType), dps, func() error {
			return p.Start()
		}, func() {
			if err := p.Close(); err != nil {
				m.logger.Error("插件关闭失败", zap.String("pluginType", string(pluginType)), zap.Error(err))
			}
		})
	}

	dep := depender.NewDepender(items)
	if err := dep.Start(); err != nil {
		return err
	}
	m.dep = dep
	return nil
}

// Close 关闭全部插件
//
// 已启动过（dep 存在）则按依赖逆序关闭；未启动过（可能没有调用 Run）则逐个关闭。
func (m *Manager) Close() {
	if m.dep != nil {
		m.dep.Close()
		return
	}
	for _, pluginType := range m.types {
		p, ok := m.plugins[pluginType]
		if !ok {
			continue
		}
		if err := p.Close(); err != nil {
			m.logger.Error("插件关闭失败", zap.String("pluginType", string(pluginType)), zap.Error(err))
		}
	}
}

// Get 查询已构建的插件
func (m *Manager) Get(pluginType core.PluginType) (core.IPlugin, bool) {
	p, ok := m.plugins[pluginType]
	return p, ok
}

// Inject 向指定插件注入依赖
//
// ignoreDisabled 为真时，插件未启用则静默跳过（由 opt 决定）。
func (m *Manager) Inject(pluginType core.PluginType, ignoreDisabled bool, a ...interface{}) {
	p, ok := m.Get(pluginType)
	if !ok {
		if ignoreDisabled {
			return
		}
		m.logger.Fatal("注入失败, 未启用插件", zap.String("pluginType", string(pluginType)))
	}
	p.Inject(a...)
}

// Types 返回已启用的插件类型列表（只读副本）
func (m *Manager) Types() []core.PluginType {
	out := make([]core.PluginType, len(m.types))
	copy(out, m.types)
	return out
}
