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
	registerClientFilter(filterType, c)
	registerServiceFilter(filterType, s)
}

// 注册客户端过滤器
func registerClientFilter(filterType string, creator core.FilterCreator) {
	if creator == nil {
		return
	}

	l := len(clientFilterCreator)
	clientFilterCreator[filterType] = creator
	if l == len(clientFilterCreator) {
		log.Log.Fatal("client filter creator repeat register", zap.String("filterType", filterType))
	}
}

// 注册服务过滤器
func registerServiceFilter(filterType string, creator core.FilterCreator) {
	if creator == nil {
		return
	}

	l := len(serviceFilterCreator)
	serviceFilterCreator[filterType] = creator
	if l == len(serviceFilterCreator) {
		log.Log.Fatal("service filter creator repeat register", zap.String("filterType", filterType))
	}
}

// 构建过滤器
func MakeFilter() {
	conf := loadConfig()

	// 建造过滤器实例
	buildFilters()

	// 构建客户端过滤器链
	buildClientFilterChains(conf)

	// 构建服务过滤器链
	buildServiceFilterChains(conf)
}

// 构建过滤器实例
func buildFilters() {
	// 构建客户端过滤器
	clientFilter = make(map[string]core.Filter)
	for filterType, creator := range clientFilterCreator {
		c := creator()
		clientFilter[filterType] = c
	}

	// 构建服务过滤器
	serviceFilter = make(map[string]core.Filter)
	for filterType, creator := range serviceFilterCreator {
		s := creator()
		serviceFilter[filterType] = s
	}
}

// 构建客户端过滤器链 - 使用正确的Config类型
func buildClientFilterChains(conf *Config) {
	clientChain = make(map[string]map[string]FilterChain)

	// 确保默认配置存在
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
			filters := buildFilterChain(filterTypes, clientFilter, "client")
			chain[clientName] = filters
		}
	}
}

// 构建服务过滤器链 - 使用正确的Config类型
func buildServiceFilterChains(conf *Config) {
	// 确保默认配置存在
	if len(conf.Service[defName]) == 0 {
		conf.Service[defName] = []string{"base"} // 写入base
	}

	serviceChain = make(map[string]FilterChain)
	for name, filterTypes := range conf.Service {
		filters := buildFilterChain(filterTypes, serviceFilter, "service")
		serviceChain[name] = filters
	}
}

// 构建过滤器链
func buildFilterChain(filterTypes []string, filterMap map[string]core.Filter, filterType string) FilterChain {
	filters := make(FilterChain, len(filterTypes))
	for i, t := range filterTypes {
		f, ok := filterMap[t]
		if !ok {
			log.Log.Fatal(filterType+" filter is not found", zap.String("filter", t))
		}
		filters[i] = f
	}
	return filters
}

// 初始化过滤器
func InitFilter(app core.IApp) {
	initClientFilters(app)
	initServiceFilters(app)
}

// 初始化客户端过滤器
func initClientFilters(app core.IApp) {
	for t, f := range clientFilter {
		if err := f.Init(app); err != nil {
			log.Log.Fatal("init client filter err", zap.String("filter", t), zap.Error(err))
		}
	}
}

// 初始化服务过滤器
func initServiceFilters(app core.IApp) {
	for t, f := range serviceFilter {
		if err := f.Init(app); err != nil {
			log.Log.Fatal("init service filter err", zap.String("filter", t), zap.Error(err))
		}
	}
}

// 关闭过滤器
func CloseFilter() {
	closeClientFilters()
	closeServiceFilters()
}

// 关闭客户端过滤器
func closeClientFilters() {
	for t, f := range clientFilter {
		if err := f.Close(); err != nil {
			log.Log.Error("close client filter err", zap.String("filter", t), zap.Error(err))
		}
	}
	clientFilter = make(map[string]core.Filter)
}

// 关闭服务过滤器
func closeServiceFilters() {
	for t, f := range serviceFilter {
		if err := f.Close(); err != nil {
			log.Log.Error("close service filter err", zap.String("filter", t), zap.Error(err))
		}
	}
	serviceFilter = make(map[string]core.Filter)
}

// 获取调用方文件行号信息
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

// 获取客户端过滤器链
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

// 获取服务过滤器链
func getServiceFilterChain(serviceName string) FilterChain {
	chain, ok := serviceChain[serviceName]
	if ok {
		return chain
	}
	chain, ok = serviceChain[defName]
	if ok {
		return chain
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
