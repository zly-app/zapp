package filter

import (
	"context"
)

type filterOpts struct {
	WithoutFilterName []string // 排除的过滤器
}

func (f *filterOpts) InWithoutFilterName(filterName string) bool {
	for i := range f.WithoutFilterName {
		if filterName == f.WithoutFilterName[i] {
			return true
		}
	}
	return false
}

type filterOptsKey struct{}

// 不使用一些过滤器
func WithoutFilterName(ctx context.Context, filterName ...string) context.Context {
	v := ctx.Value(filterOptsKey{})
	if v == nil {
		// 初始化
		v = &filterOpts{}
		ctx = context.WithValue(ctx, filterOptsKey{}, v)
	}
	opts := v.(*filterOpts)

	if len(opts.WithoutFilterName) == 0 {
		opts.WithoutFilterName = make([]string, len(filterName))
		copy(opts.WithoutFilterName, filterName)
	} else {
		opts.WithoutFilterName = append(opts.WithoutFilterName, filterName...)
	}
	return ctx
}

// 获取过滤器选项
func getFilterOpts(ctx context.Context) *filterOpts {
	v := ctx.Value(filterOptsKey{})
	if v == nil {
		return &filterOpts{}
	}
	return v.(*filterOpts)
}
