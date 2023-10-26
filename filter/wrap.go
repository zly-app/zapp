package filter

import (
	"context"

	"github.com/zly-app/zapp/core"
)

type filterWrap []core.Filter

func (f filterWrap) Init() error {
	for _, filter := range f {
		err := filter.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f filterWrap) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	for i := len(f) - 1; i >= 0; i-- {
		invoke, curFilter := next, f[i]
		next = func(ctx context.Context, req, rsp interface{}) error {
			return curFilter.HandleInject(ctx, req, rsp, invoke)
		}
	}
	return next(ctx, req, rsp)
}

func (f filterWrap) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	for i := len(f) - 1; i >= 0; i-- {
		invoke, curFilter := next, f[i]
		next = func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			return curFilter.Handle(ctx, req, invoke)
		}
	}
	return next(ctx, req)
}

func (f filterWrap) Close() error {
	for _, filter := range f {
		err := filter.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// 将多个过滤器建造者包装为一个过滤器建造者
func WrapFilterCreator(cc ...core.FilterCreator) core.FilterCreator {
	return func() core.Filter {
		filters := make([]core.Filter, len(cc))
		for i, creator := range cc {
			filter := creator()
			filters[i] = filter
		}
		return WrapFilter(filters...)
	}
}

// 将多个过滤器包装为一个过滤器
func WrapFilter(ff ...core.Filter) core.Filter {
	return filterWrap(ff)
}
