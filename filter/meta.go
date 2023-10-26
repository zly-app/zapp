package filter

import (
	"context"
	"time"
)

type CallMeta interface {
	// 是否为客户端meta
	IsClientMeta() bool
	ClientType() string
	ClientName() string

	// 是否为服务meta
	IsServiceMeta() bool
	ServiceName() string

	// 获取被调服务
	CalleeService() string
	// 获取被调方法
	CalleeMethod() string

	// 获取开始时间
	StartTime() int64
	// 获取结束时间
	EndTime() int64

	// 增加堆栈信息获取的skip, 必须在执行 Handler/HandlerInject 之前调用
	AddCallersSkip(skip int)
	// 获取触发过滤器的函数名/文件路径/代码行
	FuncFileLine() (fn, file string, line int)
	// 是否存在panic
	HasPanic() bool
	// 设置panic
	SetPanic()
}

type callMeta struct {
	isClientMeta bool   // 是否为客户端的meta
	clientType   string // 客户端类型
	clientName   string // 客户端名

	isServiceMeta bool   // 是否为服务meta
	serviceName   string // 服务类型

	calleeService string // 被调服务
	calleeMethod  string // 被调方法

	callersSkip int

	startTime int64 // 开始时间/纳秒级
	endTime   int64 // 结束时间/纳秒级

	fn, file string
	line     int

	hasPanic bool // 是否存在panic
}

func (m *callMeta) fill() {
	_ = m.StartTime()
	m.fn, m.file, m.line = funcFileLine(m.callersSkip)
}

// 是否为客户端meta
func (m *callMeta) IsClientMeta() bool { return m.isClientMeta }
func (m *callMeta) ClientType() string { return m.clientType }
func (m *callMeta) ClientName() string { return m.clientName }

func (m *callMeta) IsServiceMeta() bool { return m.isServiceMeta }
func (m *callMeta) ServiceName() string { return m.serviceName }

func (m *callMeta) CalleeService() string { return m.calleeService }
func (m *callMeta) CalleeMethod() string  { return m.calleeMethod }

func (m *callMeta) StartTime() int64 {
	if m.startTime == 0 {
		m.startTime = time.Now().UnixNano()
	}
	return m.startTime
}
func (m *callMeta) EndTime() int64 {
	if m.endTime == 0 {
		m.endTime = time.Now().UnixNano()
	}
	return m.endTime
}

func (m *callMeta) AddCallersSkip(skip int)                   { m.callersSkip += skip }
func (m *callMeta) FuncFileLine() (fn, file string, line int) { return m.fn, m.file, m.line }
func (m *callMeta) HasPanic() bool                            { return m.hasPanic }
func (m *callMeta) SetPanic()                                 { m.hasPanic = true }

type metaKey struct{}

func GetCallMeta(ctx context.Context) CallMeta {
	v := ctx.Value(metaKey{})
	if v != nil {
		m, ok := v.(CallMeta)
		if ok {
			return m
		}
	}
	return &callMeta{}
}

func SaveCallMata(ctx context.Context, meta CallMeta) context.Context {
	return context.WithValue(ctx, metaKey{}, meta)
}
