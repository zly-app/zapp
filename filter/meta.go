package filter

import (
	"context"
	"time"
)

type CallMeta struct {
	isClientMeta bool // 是否为客户端的meta

	CalleeService string // 被调服务
	CalleeMethod  string // 被调方法

	CallersSkip int

	startTime int64 // 开始时间/纳秒级
	endTime   int64 // 结束时间/纳秒级

	fn, file string
	line     int

	hasPanic bool // 是否存在panic
}

func (m *CallMeta) fill() {
	_ = m.GetStartTime()
	m.fn, m.file, m.line = funcFileLine(m.CallersSkip)
}

// 是否为客户端meta
func (m *CallMeta) IsClientMeta() bool {
	return m.isClientMeta
}

// 是否为服务meta
func (m *CallMeta) IsServiceMeta() bool {
	return !m.isClientMeta
}

// 获取函数名/文件路径/代码行
func (m *CallMeta) FuncFileLine() (fn, file string, line int) {
	return m.fn, m.file, m.line
}

// 获取开始时间
func (m *CallMeta) GetStartTime() int64 {
	if m.startTime == 0 {
		m.startTime = time.Now().UnixNano()
	}
	return m.startTime
}

// 获取结束时间
func (m *CallMeta) GetEndTime() int64 {
	if m.endTime == 0 {
		m.endTime = time.Now().UnixNano()
	}
	return m.endTime
}

func (m *CallMeta) HasPanic() bool {
	return m.hasPanic
}
func (m *CallMeta) SetPanic() {
	m.hasPanic = true
}

type metaKey struct{}

func GetCallMeta(ctx context.Context) *CallMeta {
	v := ctx.Value(metaKey{})
	if v != nil {
		m, ok := v.(*CallMeta)
		if ok {
			return m
		}
	}
	return &CallMeta{}
}

func SaveCallMata(ctx context.Context, meta *CallMeta) context.Context {
	return context.WithValue(ctx, metaKey{}, meta)
}
