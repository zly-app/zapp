package filter

import (
	"context"
	"runtime"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
)

// 过滤器链
type FilterChain []core.Filter

func (c FilterChain) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	meta := GetCallMeta(ctx)
	if v, ok := meta.(*callMeta); ok {
		ctx = v.fill(ctx)
	}

	opts := getFilterOpts(ctx)

	for i := len(c) - 1; i >= 0; i-- {
		invoke, curFilter := next, c[i]

		if opts.InWithoutFilterName(curFilter.Name()) {
			continue
		}

		next = func(ctx context.Context, req, rsp interface{}) error {
			return curFilter.HandleInject(ctx, req, rsp, invoke)
		}
	}
	return next(ctx, req, rsp)
}
func (c FilterChain) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	meta := GetCallMeta(ctx)
	if v, ok := meta.(*callMeta); ok {
		ctx = v.fill(ctx)
	}

	opts := getFilterOpts(ctx)

	for i := len(c) - 1; i >= 0; i-- {
		invoke, curFilter := next, c[i]

		if opts.InWithoutFilterName(curFilter.Name()) {
			continue
		}

		next = func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			return curFilter.Handle(ctx, req, invoke)
		}
	}
	return next(ctx, req)
}

var (
	clientFilterCreator  = make(map[string]core.FilterCreator)
	serviceFilterCreator = make(map[string]core.FilterCreator)

	clientFilter = make(map[string]core.Filter)
	clientChain  map[string]map[string]FilterChain // 指定客户端的链

	serviceFilter = make(map[string]core.Filter)
	serviceChain  map[string]FilterChain // 指定服务的链
)

// 注册服务/客户端过滤器建造者
func RegisterFilterCreator(filterType string, c core.FilterCreator, s core.FilterCreator) {
	if c != nil {
		l := len(clientFilterCreator)
		clientFilterCreator[filterType] = c
		if l == len(clientFilterCreator) {
			log.Log.Fatal("client filter creator repeat register", zap.String("filterType", filterType))
		}
	}
	if s != nil {
		l := len(serviceFilterCreator)
		serviceFilterCreator[filterType] = s
		if l == len(serviceFilterCreator) {
			log.Log.Fatal("service filter creator repeat register", zap.String("filterType", filterType))
		}
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
	clientChain = make(map[string]map[string]FilterChain)
	if len(conf.Client[defName]) == 0 {
		conf.Client[defName] = make(map[string][]string)
	}
	if len(conf.Client[defName][defName]) == 0 {
		conf.Client[defName][defName] = []string{"base"} // 写入base
	}
	for clientType, clientConf := range conf.Client {
		chain, ok := clientChain[clientType]
		if !ok {
			chain = make(map[string]FilterChain)
			clientChain[clientType] = chain
		}

		for clientName, filterTypes := range clientConf {
			filters := make(FilterChain, len(filterTypes))
			for i, t := range filterTypes {
				f, ok := clientFilter[t]
				if !ok {
					log.Log.Fatal("client filter is not found", zap.String("filter", t))
				}
				filters[i] = f
			}
			chain[clientName] = filters
		}
	}

	// 分配
	if len(conf.Service[defName]) == 0 {
		conf.Service[defName] = []string{"base"} // 写入base
	}
	serviceChain = make(map[string]FilterChain)
	for name, filterTypes := range conf.Service {
		filters := make(FilterChain, len(filterTypes))
		for i, t := range filterTypes {
			f, ok := serviceFilter[t]
			if !ok {
				log.Log.Fatal("service filter is not found", zap.String("filter", t))
			}
			filters[i] = f
		}
		serviceChain[name] = filters
	}
}

// 初始化过滤器
func InitFilter(app core.IApp) {
	for t, f := range clientFilter {
		err := f.Init(app)
		if err != nil {
			log.Log.Fatal("init client filter err", zap.String("filter", t), zap.Error(err))
		}
	}
	for t, f := range serviceFilter {
		err := f.Init(app)
		if err != nil {
			log.Log.Fatal("init service filter err", zap.String("filter", t), zap.Error(err))
		}
	}
}

// 关闭过滤器
func CloseFilter() {
	for t, f := range clientFilter {
		err := f.Close()
		if err != nil {
			log.Log.Error("close client filter err", zap.String("filter", t), zap.Error(err))
		}
	}
	clientFilter = make(map[string]core.Filter)
	for t, f := range serviceFilter {
		err := f.Close()
		if err != nil {
			log.Log.Error("close service filter err", zap.String("filter", t), zap.Error(err))
		}
	}
	serviceFilter = make(map[string]core.Filter)
}

func funcFileLine(skip int) (string, string, int) {
	const depth = 16
	const defSkip = 5 // 调试结果
	var pcs [depth]uintptr
	n := runtime.Callers(defSkip+skip, pcs[:])
	ff := runtime.CallersFrames(pcs[:n])

	f, ok := ff.Next()
	if !ok {
		return "", "", 0
	}
	return f.Function, f.File, f.Line
}

func getClientFilterChain(clientType, clientName string) FilterChain {
	chainMap, ok := clientChain[clientType]
	if ok {
		chain, ok := chainMap[clientName]
		if ok {
			return chain
		}
		chain, ok = chainMap[defName]
		if ok {
			return chain
		}
	}

	chainMap, ok = clientChain[defName]
	if ok {
		chain, ok := chainMap[defName]
		if ok {
			return chain
		}
	}
	return nil
}

func getServiceFilterChain(serviceName string) FilterChain {
	l, ok := serviceChain[serviceName]
	if ok {
		return l
	}
	l, ok = serviceChain[defName]
	if ok {
		return l
	}
	return nil
}

// 获取客户端过滤器
func GetClientFilter(ctx context.Context, clientType, clientName, methodName string) (context.Context, FilterChain) {
	chain := getClientFilterChain(clientType, clientName)
	meta := newClientMeta(clientType, clientName, methodName)
	ctx = SaveCallMata(ctx, meta)
	return ctx, chain
}

// 获取服务过滤器
func GetServiceFilter(ctx context.Context, serviceName, methodName string) (context.Context, FilterChain) {
	chain := getServiceFilterChain(serviceName)
	meta := newServiceMeta(serviceName, methodName)
	ctx = SaveCallMata(ctx, meta)
	return ctx, chain
}

func init() {
	baseFilter := WrapFilterCreator("base", newTimeoutFilter, newTraceFilter, newMetricsFilter, newLogFilter, newRecoverFilter, newGPoolFilter)
	RegisterFilterCreator("base", baseFilter, baseFilter)
}
