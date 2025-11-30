package filter

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/cast"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/zly-app/zapp/component/metrics"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

const (
	metricsRpcServerStartedTotal = "rpc_server_started_total" // 服务rpc开始计数器
	metricsRpcServerHandledTotal = "rpc_server_handled_total" // 服务rpc调用计数器
	metricsRpcServerPanicTotal   = "rpc_server_panic_total"   // 服务rpc调用panic计数器
	metricsRpcServerHandledMsec  = "rpc_server_handled_msec"  // 服务耗时桶

	metricsRpcClientStartedTotal = "rpc_client_started_total" // 客户端rpc开始计数器
	metricsRpcClientHandledTotal = "rpc_client_handled_total" // 客户端rpc调用计数器
	metricsRpcClientPanicTotal   = "rpc_client_panic_total"   // 客户端rpc调用panic计数器
	metricsRpcClientHandledMsec  = "rpc_client_handled_msec"  // 客户端耗时桶

	metricsProcessCpuCores    = "process_cpu_cores"    // cpu数量
	metricsProcessMemoryQuota = "process_memory_quota" // 内存总量
)

const (
	LabelKind          = "kind"
	LabelCallerService = "caller_service"
	LabelCallerMethod  = "caller_method"
	LabelCalleeService = "callee_service"
	LabelCalleeMethod  = "callee_method"
	LabelCodeType      = "code_type"
	LabelCode          = "code"
)

func init() {
	RegisterFilterCreator("base.metrics", newMetricsFilter, newMetricsFilter)
}

var metricsOnce sync.Once
var defaultMetrics = &metricsFilter{}

func newMetricsFilter() core.Filter {
	return defaultMetrics
}

type metricsFilter struct {
	RpcServerStartedTotal metrics.ICounter
	RpcServerHandledTotal metrics.ICounter
	RpcServerPanicTotal   metrics.ICounter
	RpcServerHandledMsec  metrics.IHistogram

	RpcClientStartedTotal metrics.ICounter
	RpcClientHandledTotal metrics.ICounter
	RpcClientPanicTotal   metrics.ICounter
	RpcClientHandledMsec  metrics.IHistogram

	ProcessCpuCores    metrics.IGauge
	ProcessMemoryQuota metrics.IGauge
}

func (*metricsFilter) Name() string { return "base.metrics" }

func (m *metricsFilter) Init(app core.IApp) error {
	metricsOnce.Do(func() {
		startLabels := []string{LabelKind, LabelCallerService, LabelCallerMethod, LabelCalleeService, LabelCalleeMethod}
		labels := append(startLabels, LabelCodeType, LabelCode)
		buckets := []float64{10, 20, 30, 50, 100, 200, 300, 500, 1000, 2000, 3000, 5000}

		m.RpcServerStartedTotal = metrics.RegistryCounter(metricsRpcServerStartedTotal, "服务rpc开始计数器", nil, startLabels...)
		m.RpcServerHandledTotal = metrics.RegistryCounter(metricsRpcServerHandledTotal, "服务rpc调用计数器", nil, labels...)
		m.RpcServerPanicTotal = metrics.RegistryCounter(metricsRpcServerPanicTotal, "服务rpc调用panic计数器", nil, labels...)
		m.RpcServerHandledMsec = metrics.RegistryHistogram(metricsRpcServerHandledMsec, "耗时桶", buckets, nil, labels...)

		m.RpcClientStartedTotal = metrics.RegistryCounter(metricsRpcClientStartedTotal, "客户端rpc开始计数器", nil, startLabels...)
		m.RpcClientHandledTotal = metrics.RegistryCounter(metricsRpcClientHandledTotal, "客户端rpc调用计数器", nil, labels...)
		m.RpcClientPanicTotal = metrics.RegistryCounter(metricsRpcClientPanicTotal, "客户端rpc调用panic计数器", nil, labels...)
		m.RpcClientHandledMsec = metrics.RegistryHistogram(metricsRpcClientHandledMsec, "客户端耗时桶", buckets, nil, labels...)

		m.ProcessCpuCores = metrics.RegistryGauge(metricsProcessCpuCores, "cpu数量", nil)
		m.ProcessMemoryQuota = metrics.RegistryGauge(metricsProcessMemoryQuota, "内存总量", nil)
		go func() {
			m.reportSysInfo() // 立即报告

			t := time.NewTicker(time.Second * 15)
			for {
				select {
				case <-t.C:
					m.reportSysInfo()
				case <-app.BaseContext().Done():
					t.Stop()
				}
			}
		}()

	})
	return nil
}

// 报告系统信息
func (m *metricsFilter) reportSysInfo() {
	m.ProcessCpuCores.Set(float64(runtime.NumCPU()), nil)

	total := float64(0)
	if memory, _ := mem.VirtualMemory(); memory != nil {
		total = float64(memory.Total)
	}
	m.ProcessMemoryQuota.Set(total, nil)
}

func (m *metricsFilter) start(ctx context.Context) CallMeta {
	meta := GetCallMeta(ctx)
	_ = meta.StartTime()

	switch meta.Kind() {
	case MetaKindService:
		label := metrics.Labels{
			LabelKind:          "server",
			LabelCallerService: meta.CallerService(),
			LabelCallerMethod:  meta.CallerMethod(),
			LabelCalleeService: meta.CalleeService(),
			LabelCalleeMethod:  meta.CalleeMethod(),
		}
		m.RpcServerStartedTotal.Inc(label, nil)
	case MetaKindClient:
		label := metrics.Labels{
			LabelKind:          "client",
			LabelCallerService: meta.CallerService(),
			LabelCallerMethod:  meta.CallerMethod(),
			LabelCalleeService: meta.CalleeService(),
			LabelCalleeMethod:  meta.CalleeMethod(),
		}
		m.RpcClientStartedTotal.Inc(label, nil)
	}

	return meta
}
func (m *metricsFilter) end(ctx context.Context, meta CallMeta, rsp interface{}, err error) {
	exemplar := (metrics.Labels)(nil)

	duration := meta.EndTime() - meta.StartTime()
	code, codeType, _ := DefaultGetErrCodeFunc(ctx, rsp, err)

	// 不成功的则上报标本
	if codeType != CodeTypeSuccess {
		traceID, _ := utils.Otel.GetOTELTraceID(ctx)
		exemplar = metrics.Labels{"traceID": traceID}
		utils.Otel.SaveToMap(ctx, exemplar)
	}

	switch meta.Kind() {
	case MetaKindService:
		label := metrics.Labels{
			LabelKind:          "server",
			LabelCallerService: meta.CallerService(),
			LabelCallerMethod:  meta.CallerMethod(),
			LabelCalleeService: meta.CalleeService(),
			LabelCalleeMethod:  meta.CalleeMethod(),
			LabelCodeType:      cast.ToString(codeType),
			LabelCode:          cast.ToString(code),
		}
		m.RpcServerHandledTotal.Inc(label, exemplar)
		if meta.HasPanic() {
			m.RpcServerPanicTotal.Inc(label, exemplar)
		}
		m.RpcServerHandledMsec.Observe(float64(time.Duration(duration)/time.Millisecond), label, exemplar)
	case MetaKindClient:
		label := metrics.Labels{
			LabelKind:          "client",
			LabelCallerService: meta.CallerService(),
			LabelCallerMethod:  meta.CallerMethod(),
			LabelCalleeService: meta.CalleeService(),
			LabelCalleeMethod:  meta.CalleeMethod(),
			LabelCodeType:      cast.ToString(codeType),
			LabelCode:          cast.ToString(code),
		}
		m.RpcClientHandledTotal.Inc(label, exemplar)
		if meta.HasPanic() {
			m.RpcClientPanicTotal.Inc(label, exemplar)
		}
		m.RpcClientHandledMsec.Observe(float64(time.Duration(duration)/time.Millisecond), label, exemplar)
	}
}

func (m *metricsFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	meta := m.start(ctx)
	err := next(ctx, req, rsp)
	m.end(ctx, meta, rsp, err)
	return err
}

func (m *metricsFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	meta := m.start(ctx)
	rsp, err := next(ctx, req)
	m.end(ctx, meta, rsp, err)
	return rsp, err
}

func (m *metricsFilter) Close() error { return nil }

var Metrics = metricsCli{}

type metricsCli struct{}

func (metricsCli) StartClient(ctx context.Context, clientType, clientName, methodName string) (context.Context, CallMeta) {
	meta := newClientMeta(clientType, clientName, methodName)
	ctx = SaveCallMata(ctx, meta)
	ctx = meta.fill(ctx)
	return ctx, defaultMetrics.start(ctx)
}
func (metricsCli) StartService(ctx context.Context, serviceName, methodName string) (context.Context, CallMeta) {
	meta := newServiceMeta(serviceName, methodName)
	ctx = SaveCallMata(ctx, meta)
	ctx = meta.fill(ctx)
	return ctx, defaultMetrics.start(ctx)
}

func (metricsCli) End(ctx context.Context, meta CallMeta, rsp interface{}, err error) {
	defaultMetrics.end(ctx, meta, rsp, err)
}
