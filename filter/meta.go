package filter

import (
	"context"
	"time"

	"github.com/zly-app/zapp/config"
)

type CallMeta interface {
	Kind() MetaKind
	// 是否为客户端meta
	IsClientMeta() bool
	ClientType() string
	ClientName() string

	// 是否为服务meta
	IsServiceMeta() bool
	// 获取服务名
	ServiceName() string

	// 获取主调服务
	CallerService() string
	// 获取主调方法
	CallerMethod() string
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

type MetaKind int

const (
	MetaKindClient  MetaKind = 1
	MetaKindService MetaKind = 2
)

type callMeta struct {
	kind MetaKind

	clientType string // 客户端类型
	clientName string // 客户端名

	serviceName string // 服务名

	callerService string // 主调服务
	callerMethod  string // 主调方法
	calleeService string // 被调服务
	calleeMethod  string // 被调方法

	callersSkip int

	startTime int64 // 开始时间/纳秒级
	endTime   int64 // 结束时间/纳秒级

	fn, file string
	line     int

	hasPanic bool // 是否存在panic
}

func newClientMeta(clientType, clientName, methodName string) *callMeta {
	return &callMeta{
		kind: MetaKindClient,

		clientType: clientType,
		clientName: clientName,

		calleeService: clientType + "/" + clientName,
		calleeMethod:  methodName,
	}
}

func newServiceMeta(serviceName, methodName string) *callMeta {
	return &callMeta{
		kind: MetaKindService,

		serviceName: serviceName,

		calleeService: serviceName,
		calleeMethod:  methodName,
	}
}

func (m *callMeta) fillByCallerMeta(callerMeta CallerMeta) {
	if callerMeta.CallerService != "" {
		m.callerService = callerMeta.CallerService
	}
	if callerMeta.CallerMethod != "" {
		m.callerMethod = callerMeta.CallerMethod
	}
	if callerMeta.CalleeService != "" {
		m.calleeService = callerMeta.CalleeService
	}
	if callerMeta.CalleeMethod != "" {
		m.calleeMethod = callerMeta.CalleeMethod
	}
}
func (m *callMeta) fill(ctx context.Context) context.Context {
	_ = m.StartTime()
	m.fn, m.file, m.line = funcFileLine(m.callersSkip)

	// 填充主调被调信息, 应用可以调用 SaveCallerMeta 来修改主调被调信息
	callerMeta, ok := GetCallerMeta(ctx)
	if ok {
		m.fillByCallerMeta(callerMeta)
	} else { // 没有主调数据, 通过 app 获取
		m.callerService = config.Conf.Config().Frame.Name
		m.callerMethod = "rpc"
	}

	if m.IsServiceMeta() {
		// 将当前服务信息存入ctx, 那么client就会从ctx中获取到当前服务信息作为主调, 这里仅设置主调信息, 因为被调只有执行时才能确认
		return SaveCallerMeta(ctx, CallerMeta{
			CallerService: m.calleeService,
			CallerMethod:  m.calleeMethod,
		})
	}

	return ctx
}

func (m *callMeta) Kind() MetaKind { return m.kind }

// 是否为客户端meta
func (m *callMeta) IsClientMeta() bool { return m.kind == MetaKindClient }
func (m *callMeta) ClientType() string { return m.clientType }
func (m *callMeta) ClientName() string { return m.clientName }

func (m *callMeta) IsServiceMeta() bool { return m.kind == MetaKindService }
func (m *callMeta) ServiceName() string { return m.serviceName }

func (m *callMeta) CallerService() string { return m.callerService }
func (m *callMeta) CallerMethod() string  { return m.callerMethod }
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

type callerMetaKey struct{}

// 主调信息
type CallerMeta struct {
	CallerService string // 主调服务
	CallerMethod  string // 主调方法
	CalleeService string // 被调服务
	CalleeMethod  string // 被调方法
}

func GetCallerMeta(ctx context.Context) (CallerMeta, bool) {
	v := ctx.Value(callerMetaKey{})
	if v != nil {
		m, ok := v.(CallerMeta)
		if ok {
			return m, true
		}
	}
	return CallerMeta{}, false
}

func SaveCallerMeta(ctx context.Context, callerMeta CallerMeta) context.Context {
	return context.WithValue(ctx, callerMetaKey{}, callerMeta)
}
