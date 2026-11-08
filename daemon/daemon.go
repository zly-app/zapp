package daemon

import (
	"flag"
	"fmt"
	"os"

	kservice "github.com/kardianos/service"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/log"
)

// Manager 是守护进程模型：拥有「进程级服务包装器」这份数据，以及
// 「install/start/stop/restart/status/uninstall 参数处理」这条规则。
//
// 它不参与应用初始化时序，只被壳在 Run() 时驱动。
// 允许调用：kardianos/service（底座机制）、log（默认日志 —— 此处的参数处理发生在
// 配置与 Logger 就绪之前，只能用默认日志）。禁止调用 plugin/service/component/lifecycle。
type Manager struct {
	name    string
	enabled bool
	runFn   func()
	stopFn  func()
	svc     kservice.Service
}

// NewManager 创建守护进程管理器
//
// runFn 在服务被拉起时执行（应为应用的启动用例），stopFn 在服务被停止时执行（应为应用的退出用例）。
func NewManager(name string, enabled bool, runFn, stopFn func()) *Manager {
	return &Manager{name: name, enabled: enabled, runFn: runFn, stopFn: stopFn}
}

// Start 实现 kardianos/service 的 Interface
func (m *Manager) Start(_ kservice.Service) error {
	go m.runFn()
	return nil
}

// Stop 实现 kardianos/service 的 Interface
func (m *Manager) Stop(_ kservice.Service) error {
	m.stopFn()
	return nil
}

// setService 构建进程级服务包装器
func (m *Manager) setService() {
	svcConfig := &kservice.Config{
		Name:      m.name,
		Arguments: os.Args[1:],
	}
	s, err := kservice.New(m, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	m.svc = s
}

// Prepare 处理守护进程相关命令行参数
//
// 启用守护且首个参数是 install/start/stop/restart/status/uninstall 之一时，执行对应动作并退出；
// 否则构建服务包装器，等待 Run() 接管。
func (m *Manager) Prepare() {
	if !m.enabled || len(os.Args) < 2 {
		m.setService()
		return
	}

	flag.String("install", "", "安装服务, string 是运行时传递给 app 的参数, 请不要使用相对路径")
	flag.Bool("start", false, "启动服务")
	flag.Bool("stop", false, "停止服务")
	flag.Bool("restart", false, "重启服务")
	flag.Bool("status", false, "获取状态")
	flag.Bool("uninstall", false, "移除服务")

	svcConfig := &kservice.Config{Name: m.name}

	switch os.Args[1] {
	case "install", "-install":
		svcConfig.Arguments = os.Args[2:]
	case "remove", "start", "stop", "restart", "status", "uninstall":
	case "-remove", "-start", "-stop", "-restart", "-status", "-uninstall":
	default:
		m.setService()
		return
	}

	s, err := kservice.New(m, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "install", "-install":
		err = s.Install()
	case "start", "-start":
		err = s.Start()
	case "stop", "-stop":
		err = s.Stop()
	case "restart", "-restart":
		err = s.Restart()
	case "uninstall", "-uninstall":
		err = s.Uninstall()
	case "status", "-status":
		status, err := s.Status()
		if err != nil {
			log.Log.Error(zap.Error(err))
			os.Exit(1)
		}

		switch status {
		case kservice.StatusRunning:
			log.Log.Info(fmt.Sprintf("service %s is Running", m.name))
		case kservice.StatusStopped:
			log.Log.Info(fmt.Sprintf("service %s is Stopped", m.name))
		default:
			log.Log.Info(fmt.Sprintf("service %s status is Unknown", m.name))
		}
		os.Exit(0)
	}

	if err != nil {
		log.Log.Error(zap.Error(err))
		os.Exit(1)
	}

	log.Log.Info("ok")
	os.Exit(0)
}

// Run 以前台/服务方式运行，阻塞直到进程结束
func (m *Manager) Run() error {
	return m.svc.Run()
}

// Enabled 报告守护进程是否启用
func (m *Manager) Enabled() bool {
	return m.enabled
}

var _ kservice.Interface = (*Manager)(nil)
