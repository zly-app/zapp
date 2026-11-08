package zapp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"github.com/zly-app/zapp/plugin"
	"github.com/zly-app/zapp/service"
)

// 冒烟测试：钉住「壳 + 业务层 + 模型」这条链的行为。
//
// 它不测具体模型内部逻辑，只验证拆分后各步是否仍然按契约发生：
// 组件被构建、插件/服务实例可查、生命周期钩子按顺序触发、退出时按序关闭。

type smokePlugin struct {
	started bool
	closed  bool
}

func (p *smokePlugin) Inject(a ...interface{}) {}
func (p *smokePlugin) Start() error            { p.started = true; return nil }
func (p *smokePlugin) Close() error            { p.closed = true; return nil }

type smokeService struct {
	started bool
	closed  bool
}

func (s *smokeService) Inject(a ...interface{}) {}
func (s *smokeService) Start() error            { s.started = true; return nil }
func (s *smokeService) Close() error            { s.closed = true; return nil }

func TestAppLifecycle(t *testing.T) {
	// 分支一：未初始化时 Component() 必须 panic 且说清原因
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("未初始化时 Component() 应 panic")
			}
			if !strings.Contains(fmt.Sprint(r), "未初始化") {
				t.Fatalf("panic 信息未说明原因: %v", r)
			}
		}()
		Component()
	}()

	p := &smokePlugin{}
	s := &smokeService{}
	plugin.RegisterCreatorFunc("smoke-plugin", func(core.IApp) core.IPlugin { return p })
	service.RegisterCreatorFunc("smoke-service", func(core.IApp) core.IService { return s })

	var order []string
	record := func(name string) Handler {
		return func(core.IApp, HandlerType) { order = append(order, name) }
	}

	app := NewApp("smoke",
		WithPlugin("smoke-plugin"),
		WithService("smoke-service"),
		WithIgnoreInjectOfDisablePlugin(true),
		WithIgnoreInjectOfDisableService(true),
		WithHandler(handler.BeforeInitializeHandler, record("before-init")),
		WithHandler(handler.AfterMakeComponent, record("after-component")),
		WithHandler(handler.AfterInitializeHandler, record("after-init")),
	)
	a, ok := app.(*shell)
	if !ok {
		t.Fatalf("NewApp 返回的类型不是壳: %T", app)
	}

	if a.GetComponent() == nil {
		t.Fatal("组件未构建")
	}
	// 分支二：初始化后 Component() 读到的是同一个实例（不额外持有状态）
	if Component() == nil {
		t.Fatal("初始化后 Component() 返回 nil")
	}
	if Component() != a.GetComponent() {
		t.Fatal("Component() 与 app.GetComponent() 不是同一实例")
	}
	if got, ok := a.GetPlugin("smoke-plugin"); !ok || got != p {
		t.Fatalf("插件实例不可查: %v %v", got, ok)
	}
	if got, ok := a.GetService("smoke-service"); !ok || got != s {
		t.Fatalf("服务实例不可查: %v %v", got, ok)
	}
	if got := a.models.Plugin.Types(); len(got) != 1 || got[0] != "smoke-plugin" {
		t.Fatalf("启用列表不符: %v", got)
	}

	want := []string{"before-init", "after-component", "after-init"}
	if len(order) != len(want) {
		t.Fatalf("初始化钩子数量不符: %v", order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("钩子顺序不符: got %v want %v", order, want)
		}
	}

	// 启动用例：插件与服务都应被 Start
	a.run()
	if !p.started || !s.started {
		t.Fatalf("启动未覆盖全部模型: plugin=%v service=%v", p.started, s.started)
	}

	// 退出用例：逆序关闭，且不触发 os.Exit
	a.exit()
	if !p.closed || !s.closed {
		t.Fatalf("退出未关闭全部模型: plugin=%v service=%v", p.closed, s.closed)
	}
	if a.BaseContext().Err() == nil {
		t.Fatal("退出时应取消基础上下文")
	}

	// 未启用时的注入：按选项静默跳过，不崩
	a.InjectPlugin("never-enabled", 1)
	a.InjectService("never-enabled", 1)
}
