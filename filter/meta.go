package filter

import (
	"context"
)

type Meta struct {
	isClientMeta bool // 是否为客户端的meta
	ClientType   string
	ClientName   string
	ServiceName  string

	MethodName string

	CallersSkip int
	fn, file    string
	line        int
}

func (m *Meta) fill(isClientMeta bool) {
	m.isClientMeta = isClientMeta
	m.fn, m.file, m.line = funcFileLine(m.CallersSkip)
}

// 是否为客户端meta
func (m *Meta) IsClientMeta() bool {
	return m.isClientMeta
}

// 获取函数名/文件路径/代码行
func (m *Meta) FuncFileLine() (fn, file string, line int) {
	return m.fn, m.file, m.line
}

type metaKey struct{}

func GetMetaFromCtx(ctx context.Context) *Meta {
	v := ctx.Value(metaKey{})
	if v != nil {
		m, ok := v.(*Meta)
		if ok {
			return m
		}
	}
	return &Meta{}
}

func SaveMataToCtx(ctx context.Context, meta *Meta) context.Context {
	return context.WithValue(ctx, metaKey{}, meta)
}
