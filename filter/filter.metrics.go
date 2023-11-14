package filter

import (
	"context"
	"sync"
	"time"

	"github.com/spf13/cast"

	"github.com/zly-app/zapp/component/metrics"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

const (
	metricsRpcServerHandledTotal = "rpc_server_handled_total" // rpc调用计数器
	metricsRpcServerPanicTotal   = "rpc_server_panic_total"   // rpc调用panic计数器
	metricsRpcServerHandledMsec  = "rpc_server_handled_msec"  // 耗时桶
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
		constLabels := metrics.Labels{
			"app":            app.Name(),
			"env":            app.GetConfig().Config().Frame.Env,
			"container_name": app.GetConfig().Config().Frame.Instance,
		}
		lables := []string{"kind", "code_type", "code", "callee_service", "callee_method"}

		metrics.RegistryCounter(metricsRpcServerHandledTotal, "rpc调用计数器", constLabels, lables...)
		metrics.RegistryCounter(metricsRpcServerPanicTotal, "rpc调用panic计数器", constLabels, lables...)

		buckets := []float64{10, 20, 30, 50, 100, 200, 300, 500, 1000, 2000, 3000, 5000}
		metrics.RegistryHistogram(metricsRpcServerHandledMsec, "耗时桶", buckets, constLabels, lables...)
	})
	return nil
}

func (m metricsFilter) start(ctx context.Context) CallMeta {
	meta := GetCallMeta(ctx)
	_ = meta.StartTime()
	return meta
}
func (m metricsFilter) end(ctx context.Context, meta CallMeta, rsp interface{}, err error) {
	duration := meta.EndTime() - meta.StartTime()
	kind := "client"
	if meta.IsServiceMeta() {
		kind = "server"
	}

	code, codeType, _ := DefaultGetErrCodeFunc(ctx, rsp, err)
	traceID, _ := utils.Otel.GetOTELTraceID(ctx)
	_ = traceID
	values := []string{kind, cast.ToString(codeType), cast.ToString(code), meta.CalleeService(), meta.CalleeMethod()}

	metrics.CounterWithLabelValue(metricsRpcServerHandledTotal, values...).Inc()
	if meta.HasPanic() {
		metrics.CounterWithLabelValue(metricsRpcServerPanicTotal, values...).Inc()
	}
	metrics.HistogramWithLabelValue(metricsRpcServerHandledMsec, values...).Observe(float64(time.Duration(duration) / time.Millisecond))
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
