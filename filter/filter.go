package filter

import (
	"context"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
)

// 过滤器链
type FilterChain []core.Filter

func (c FilterChain) FilterInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	nowHandle := func(ctx context.Context, req, rsp interface{}) error {
		ctx, span := utils.Otel.StartSpan(ctx, "CallFunc")
		err := next(ctx, req, rsp)
		span.End()
		return err
	}

	for i := len(c) - 1; i >= 0; i-- {
		nextHandle, curFilter := nowHandle, c[i]
		nowHandle = func(ctx context.Context, req, rsp interface{}) error {
			return curFilter.HandleInject(ctx, req, rsp, nextHandle)
		}
	}
	return nowHandle(ctx, req, rsp)
}
func (c FilterChain) Filter(ctx context.Context, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	nowHandle := func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		ctx, span := utils.Otel.StartSpan(ctx, "CallFunc")
		rsp, err = next(ctx, req)
		span.End()
		return rsp, err
	}

	for i := len(c) - 1; i >= 0; i-- {
		nextHandle, curFilter := nowHandle, c[i]
		nowHandle = func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			return curFilter.Handle(ctx, req, nextHandle)
		}
	}
	return nowHandle(ctx, req)
}

var (
	clientFilterCreator  = make(map[string]core.FilterCreator)
	serviceFilterCreator = make(map[string]core.FilterCreator)

	clientFilter   = make(map[string]core.Filter)
	defClientChain FilterChain            // 默认的链
	clientChain    map[string]FilterChain // 指定客户端的链

	serviceFilter   = make(map[string]core.Filter)
	defServiceChain FilterChain            // 默认的链
	serviceChain    map[string]FilterChain // 指定服务的链
)

// 注册服务/客户端过滤器建造者
func RegisterFilterCreator(filterType string, c core.FilterCreator, s core.FilterCreator) {
	if c != nil {
		clientFilterCreator[filterType] = c
	}
	if s != nil {
		serviceFilterCreator[filterType] = s
	}
}

// 构建过滤器
func MakeFilter() {
	conf := loadConfig()

	// 建造
	clientFilter = make(map[string]core.Filter)
	for filterType, creator := range clientFilterCreator {
		c := creator()
		clientFilter[filterType] = c
	}

	serviceFilter = make(map[string]core.Filter)
	for filterType, creator := range serviceFilterCreator {
		s := creator()
		serviceFilter[filterType] = s
	}

	// 分配
	clientChain = make(map[string]FilterChain)
	for name, filterTypes := range conf.Client {
		filters := make(FilterChain, len(filterTypes))
		for i, t := range filterTypes {
			f, ok := clientFilter[t]
			if !ok {
				logger.Log.Fatal("client filter is not found", zap.String("filter", t))
			}
			filters[i] = f
		}
		if name == "default" {
			defClientChain = filters
		} else {
			clientChain[name] = filters
		}
	}

	// 分配
	serviceChain = make(map[string]FilterChain)
	for name, filterTypes := range conf.Service {
		filters := make(FilterChain, len(filterTypes))
		for i, t := range filterTypes {
			f, ok := serviceFilter[t]
			if !ok {
				logger.Log.Fatal("service filter is not found", zap.String("filter", t))
			}
			filters[i] = f
		}
		if name == "default" {
			defServiceChain = filters
		} else {
			serviceChain[name] = filters
		}
	}
}

// 启动过滤器
func StartFilter() {
	for t, f := range clientFilter {
		err := f.Start()
		if err != nil {
			logger.Log.Fatal("start client filter err", zap.String("filter", t), zap.Error(err))
		}
	}
	for t, f := range serviceFilter {
		err := f.Start()
		if err != nil {
			logger.Log.Fatal("start service filter err", zap.String("filter", t), zap.Error(err))
		}
	}
}

// 关闭过滤器
func CloseFilter() {
	for t, f := range clientFilter {
		err := f.Start()
		if err != nil {
			logger.Log.Error("start client filter err", zap.String("filter", t), zap.Error(err))
		}
	}
	for t, f := range serviceFilter {
		err := f.Start()
		if err != nil {
			logger.Log.Error("start service filter err", zap.String("filter", t), zap.Error(err))
		}
	}
}

// 触发客户端过滤器(inject)
func TriggerClientFilterInject(ctx context.Context, clientName string, req, rsp interface{}, next core.FilterInjectFunc) error {
	chain, ok := clientChain[clientName]
	if !ok {
		chain = defClientChain
	}
	return chain.FilterInject(ctx, req, rsp, next)
}

// 触发客户端过滤器
func TriggerClientFilter(ctx context.Context, clientName string, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	chain, ok := clientChain[clientName]
	if !ok {
		chain = defClientChain
	}
	return chain.Filter(ctx, req, next)
}

// 触发服务过滤器(inject)
func TriggerServiceFilterInject(ctx context.Context, serviceName string, req, rsp interface{}, next core.FilterInjectFunc) error {
	chain, ok := serviceChain[serviceName]
	if !ok {
		chain = defServiceChain
	}
	return chain.FilterInject(ctx, req, rsp, next)
}

// 触发服务过滤器
func TriggerServiceFilter(ctx context.Context, serviceName string, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	chain, ok := serviceChain[serviceName]
	if !ok {
		chain = defServiceChain
	}
	return chain.Filter(ctx, req, next)
}
