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

func init() {
	RegisterFilterCreator("base.metrics", newMetricsFilter, newMetricsFilter)
}

var metricsOnce sync.Once

func newMetricsFilter() core.Filter {
	return metricsFilter{}
}

type metricsFilter struct{}

func (m metricsFilter) Init(app core.IApp) error {
	metricsOnce.Do(func() {
		startLables := []string{"kind", "caller_service", "caller_method", "callee_service", "callee_method"}
		lables := append(startLables, "code_type", "code")
		buckets := []float64{10, 20, 30, 50, 100, 200, 300, 500, 1000, 2000, 3000, 5000}

		metrics.RegistryCounter(metricsRpcServerStartedTotal, "服务rpc开始计数器", nil, startLables...)
		metrics.RegistryCounter(metricsRpcServerHandledTotal, "服务rpc调用计数器", nil, lables...)
		metrics.RegistryCounter(metricsRpcServerPanicTotal, "服务rpc调用panic计数器", nil, lables...)
		metrics.RegistryHistogram(metricsRpcServerHandledMsec, "耗时桶", buckets, nil, lables...)

		metrics.RegistryCounter(metricsRpcClientStartedTotal, "客户端rpc开始计数器", nil, startLables...)
		metrics.RegistryCounter(metricsRpcClientHandledTotal, "客户端rpc调用计数器", nil, lables...)
		metrics.RegistryCounter(metricsRpcClientPanicTotal, "客户端rpc调用panic计数器", nil, lables...)
		metrics.RegistryHistogram(metricsRpcClientHandledMsec, "客户端耗时桶", buckets, nil, lables...)

		metrics.RegistryGauge(metricsProcessCpuCores, "cpu数量", nil)
		metrics.RegistryGauge(metricsProcessMemoryQuota, "内存总量", nil)
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
func (m metricsFilter) reportSysInfo() {
	metrics.GaugeWithLabelValue(metricsProcessCpuCores).Set(float64(runtime.NumCPU()))

	total := float64(0)
	if memory, _ := mem.VirtualMemory(); memory != nil {
		total = float64(memory.Total)
	}
	metrics.GaugeWithLabelValue(metricsProcessMemoryQuota).Set(total)
}

func (m metricsFilter) start(ctx context.Context) CallMeta {
	meta := GetCallMeta(ctx)
	_ = meta.StartTime()

	traceID, _ := utils.Otel.GetOTELTraceID(ctx)
	exemplar := metrics.Labels{"traceID": traceID}

	switch meta.Kind() {
	case MetaKindService:
		values := []string{"server", meta.CallerService(), meta.CallerMethod(), meta.CalleeService(), meta.CalleeMethod()}
		c := metrics.CounterWithLabelValue(metricsRpcServerStartedTotal, values...)
		if e, ok := c.(metrics.ExemplarAdder); ok {
			e.AddWithExemplar(1, exemplar)
		} else {
			c.Inc()
		}
	case MetaKindClient:
		values := []string{"client", meta.CallerService(), meta.CallerMethod(), meta.CalleeService(), meta.CalleeMethod()}
		c := metrics.CounterWithLabelValue(metricsRpcClientStartedTotal, values...)
		if e, ok := c.(metrics.ExemplarAdder); ok {
			e.AddWithExemplar(1, exemplar)
		} else {
			c.Inc()
		}
	}

	return meta
}
func (m metricsFilter) end(ctx context.Context, meta CallMeta, rsp interface{}, err error) {
	duration := meta.EndTime() - meta.StartTime()

	code, codeType, _ := DefaultGetErrCodeFunc(ctx, rsp, err)
	traceID, _ := utils.Otel.GetOTELTraceID(ctx)
	exemplar := metrics.Labels{"traceID": traceID}

	var values []string
	var k1, k2, k3 string

	switch meta.Kind() {
	case MetaKindService:
		values = []string{"server", meta.CallerService(), meta.CallerMethod(), meta.CalleeService(), meta.CalleeMethod(), cast.ToString(codeType), cast.ToString(code)}
		k1 = metricsRpcServerHandledTotal
		k2 = metricsRpcServerPanicTotal
		k3 = metricsRpcServerHandledMsec
	case MetaKindClient:
		values = []string{"client", meta.CallerService(), meta.CallerMethod(), meta.CalleeService(), meta.CalleeMethod(), cast.ToString(codeType), cast.ToString(code)}
		k1 = metricsRpcClientHandledTotal
		k2 = metricsRpcClientPanicTotal
		k3 = metricsRpcClientHandledMsec
	}

	c := metrics.CounterWithLabelValue(k1, values...)
	if e, ok := c.(metrics.ExemplarAdder); ok {
		e.AddWithExemplar(1, exemplar)
	} else {
		c.Inc()
	}

	if meta.HasPanic() {
		c = metrics.CounterWithLabelValue(k2, values...)
		if e, ok := c.(metrics.ExemplarAdder); ok {
			e.AddWithExemplar(1, exemplar)
		} else {
			c.Inc()
		}
	}

	h := metrics.HistogramWithLabelValue(k3, values...)
	if e, ok := h.(metrics.ExemplarObserver); ok {
		e.ObserveWithExemplar(float64(time.Duration(duration)/time.Millisecond), exemplar)
	} else {
		h.Observe(float64(time.Duration(duration) / time.Millisecond))
	}
}

func (m metricsFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	meta := m.start(ctx)
	err := next(ctx, req, rsp)
	m.end(ctx, meta, rsp, err)
	return err
}

func (m metricsFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	meta := m.start(ctx)
	rsp, err := next(ctx, req)
	m.end(ctx, meta, rsp, err)
	return rsp, err
}

func (m metricsFilter) Close() error { return nil }
