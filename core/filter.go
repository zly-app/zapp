package core

import (
	"context"
)

// 过滤器
type Filter interface {
	// 初始化
	Init() error
	// 注射模式
	HandleInject(ctx context.Context, req, rsp interface{}, next FilterInjectFunc) error
	// return模式
	Handle(ctx context.Context, req interface{}, next FilterFunc) (rsp interface{}, err error)
	// 关闭
	Close() error
}
type FilterInjectFunc func(ctx context.Context, req, rsp interface{}) error
type FilterFunc func(ctx context.Context, req interface{}) (rsp interface{}, err error)

// 过滤器建造者
type FilterCreator func() Filter
