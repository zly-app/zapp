package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/zlyuancn/zretry"
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"github.com/zly-app/zapp/logger"
)

type (
	Labels   = map[string]string
	ICounter = interface {
		Inc()
		Add(float64)
	}
	IGauge = interface {
		Set(float64)
		Inc()
		Dec()
		Add(float64)
		Sub(float64)
		SetToCurrentTime()
	}
	IHistogram = interface {
		Observe(float64)
	}
	ISummary = interface {
		Observe(float64)
	}
)

type Client interface {
	/*注册计数器
	  name 计数器名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryCounter(name, help string, constLabels Labels, labels ...string)
	// 获取计数器
	Counter(name string, labels Labels) ICounter
	// 获取计数器
	CounterWithLabelValue(name string, labelValues ...string) ICounter

	/*注册计量器
	  name 计量器名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryGauge(name, help string, constLabels Labels, labels ...string)
	// 获取计量器
	Gauge(name string, labels Labels) IGauge
	// 获取计量器
	GaugeWithLabelValue(name string, labelValues ...string) IGauge

	/*注册直方图
	  name 直方图名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  buckets 桶列表
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryHistogram(name, help string, buckets []float64, constLabels Labels, labels ...string)
	// 获取直方图
	Histogram(name string, labels Labels) IHistogram
	// 获取直方图
	HistogramWithLabelValue(name string, labelValues ...string) IHistogram

	/*注册汇总
	  name 直方图名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistrySummary(name, help string, constLabels Labels, labels ...string)
	// 获取汇总
	Summary(name string, labels Labels) ISummary
	// 获取汇总
	SummaryWithLabelValue(name string, labelValues ...string) ISummary

	// 关闭
	Close()
}

type clientCli struct {
	app core.IApp

	counterCollector       map[string]*prometheus.CounterVec // 计数器
	counterCollectorLocker sync.RWMutex

	gaugeCollector       map[string]*prometheus.GaugeVec // 计量器
	gaugeCollectorLocker sync.RWMutex

	histogramCollector       map[string]*prometheus.HistogramVec // 直方图
	histogramCollectorLocker sync.RWMutex

	summaryCollector       map[string]*prometheus.SummaryVec // 汇总
	summaryCollectorLocker sync.RWMutex

	pullRegistry prometheus.Registerer // pull模式注册器
	pusher       *push.Pusher          // push模式推送器
}

func newClient(app core.IApp) Client {
	p := &clientCli{
		app:                app,
		counterCollector:   make(map[string]*prometheus.CounterVec),
		gaugeCollector:     make(map[string]*prometheus.GaugeVec),
		histogramCollector: make(map[string]*prometheus.HistogramVec),
		summaryCollector:   make(map[string]*prometheus.SummaryVec),
	}

	key := fmt.Sprintf("components.%s.default", DefaultComponentType)
	conf := newConfig()
	if app.GetConfig().GetViper().IsSet(key) {
		if err := app.GetConfig().GetViper().UnmarshalKey(key, conf); err != nil {
			app.Fatal("解析 metrics 配置失败", zap.Error(err))
		}
	}
	conf.Check()

	p.startPullMode(conf)
	p.startPushMode(conf)

	return p
}

// 启动pull模式
func (p *clientCli) startPullMode(conf *Config) {
	if conf.PullBind == "" {
		return
	}

	// 创建注册器
	r := prometheus.NewRegistry()
	p.pullRegistry = r
	if conf.ProcessCollector {
		r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	if conf.GoCollector {
		r.MustRegister(collectors.NewGoCollector())
	}

	p.app.Info("启用 metrics pull模式", zap.String("PullBind", conf.PullBind), zap.String("PullPath", conf.PullPath))

	// 构建server
	handle := promhttp.InstrumentMetricHandler(r, promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	mux := http.NewServeMux()
	mux.Handle(conf.PullPath, handle)
	server := &http.Server{Addr: conf.PullBind, Handler: mux}

	zapp.AddHandler(zapp.AfterExitHandler, func(app core.IApp, handlerType handler.HandlerType) {
		_ = server.Close()
	})
	// 开始监听
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			logger.Log.Fatal("启动pull模式失败", zap.Error(err))
		}
	}(server)
}

// 启动push模式
func (p *clientCli) startPushMode(conf *Config) {
	if conf.PushAddress == "" {
		return
	}

	// 创建推送器
	pusher := push.New(conf.PushAddress, p.app.Name())
	p.pusher = pusher
	if conf.PushInstance == "" {
		conf.PushInstance = p.app.Name()
	}
	pusher.Grouping("instance", conf.PushInstance)

	if conf.ProcessCollector {
		pusher.Collector(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	if conf.GoCollector {
		pusher.Collector(collectors.NewGoCollector())
	}

	p.app.Info("启用 metrics push 模式", zap.String("PushAddress", conf.PushAddress), zap.String("PushInstance", conf.PushInstance))

	// 开始推送
	done, cancel := context.WithCancel(context.Background())
	zapp.AddHandler(zapp.AfterExitHandler, func(app core.IApp, handlerType handler.HandlerType) {
		cancel()
	})
	go func(ctx context.Context, conf *Config, pusher *push.Pusher) {
		for {
			t := time.NewTimer(time.Duration(conf.PushTimeInterval) * time.Millisecond)
			select {
			case <-ctx.Done():
				t.Stop()
				p.push(conf, pusher) // 最后一次推送
				return
			case <-t.C:
				p.push(conf, pusher)
			}
		}
	}(done, conf, pusher)
}

// 推送
func (p *clientCli) push(conf *Config, pusher *push.Pusher) {
	err := zretry.DoRetry(int(conf.PushRetry+1), time.Duration(conf.PushRetryInterval)*time.Millisecond, pusher.Push,
		func(nowAttemptCount, remainCount int, err error) {
			p.app.Error("metrics 状态推送失败", zap.Error(err))
		},
	)
	if err == nil {
		p.app.Debug("metrics 状态推送成功")
	}
}

// 注册收集器
func (p *clientCli) registryCollector(collector prometheus.Collector) error {
	if p.pullRegistry != nil {
		if err := p.pullRegistry.Register(collector); err != nil {
			return err
		}
	}
	if p.pusher != nil {
		p.pusher.Collector(collector)
	}
	return nil
}

func (p *clientCli) RegistryCounter(name, help string, constLabels Labels, labels ...string) {
	p.counterCollectorLocker.Lock()
	defer p.counterCollectorLocker.Unlock()

	if _, ok := p.counterCollector[name]; ok {
		p.app.Fatal("重复注册 metrics 计数器")
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(counter)
	if err != nil {
		p.app.Fatal("注册 metrics 计数器失败", zap.Error(err))
	}
	p.counterCollector[name] = counter
}
func (p *clientCli) Counter(name string, labels Labels) ICounter {
	p.counterCollectorLocker.RLock()
	defer p.counterCollectorLocker.RUnlock()

	coll, ok := p.counterCollector[name]
	if !ok {
		p.app.Fatal("metrics 计数器不存在", zap.String("name", name))
	}
	counter, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取 metrics 计数器失败", zap.Error(err))
	}
	return counter
}
func (p *clientCli) CounterWithLabelValue(name string, labelValues ...string) ICounter {
	p.counterCollectorLocker.RLock()
	defer p.counterCollectorLocker.RUnlock()

	coll, ok := p.counterCollector[name]
	if !ok {
		p.app.Fatal("metrics 计数器不存在", zap.String("name", name))
	}
	counter, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取 metrics 计数器失败", zap.Error(err))
	}
	return counter
}

func (p *clientCli) RegistryGauge(name, help string, constLabels Labels, labels ...string) {
	p.gaugeCollectorLocker.Lock()
	defer p.gaugeCollectorLocker.Unlock()

	if _, ok := p.gaugeCollector[name]; ok {
		p.app.Fatal("重复注册 metrics 计量器")
	}

	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(gauge)
	if err != nil {
		p.app.Fatal("注册 metrics 计量器失败", zap.Error(err))
	}

	p.gaugeCollector[name] = gauge
}
func (p *clientCli) Gauge(name string, labels Labels) IGauge {
	p.gaugeCollectorLocker.RLock()
	defer p.gaugeCollectorLocker.RUnlock()

	coll, ok := p.gaugeCollector[name]
	if !ok {
		p.app.Fatal("metrics 计量器不存在", zap.String("name", name))
	}
	gauge, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取 metrics 计量器失败", zap.Error(err))
	}
	return gauge
}
func (p *clientCli) GaugeWithLabelValue(name string, labelValues ...string) IGauge {
	p.gaugeCollectorLocker.RLock()
	defer p.gaugeCollectorLocker.RUnlock()

	coll, ok := p.gaugeCollector[name]
	if !ok {
		p.app.Fatal("metrics 计量器不存在", zap.String("name", name))
	}
	gauge, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取 metrics 计量器失败", zap.Error(err))
	}
	return gauge
}

func (p *clientCli) RegistryHistogram(name, help string, buckets []float64, constLabels Labels, labels ...string) {
	p.histogramCollectorLocker.Lock()
	defer p.histogramCollectorLocker.Unlock()

	if _, ok := p.histogramCollector[name]; ok {
		p.app.Fatal("重复注册 metrics 直方图")
	}

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
		Buckets:     buckets,
	}, labels)
	err := p.registryCollector(histogram)
	if err != nil {
		p.app.Fatal("注册 metrics 直方图失败", zap.Error(err))
	}

	p.histogramCollector[name] = histogram
}
func (p *clientCli) Histogram(name string, labels Labels) IHistogram {
	p.histogramCollectorLocker.RLock()
	defer p.histogramCollectorLocker.RUnlock()

	coll, ok := p.histogramCollector[name]
	if !ok {
		p.app.Fatal("metrics 直方图不存在", zap.String("name", name))
	}
	histogram, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取 metrics 直方图失败", zap.Error(err))
	}
	return histogram
}
func (p *clientCli) HistogramWithLabelValue(name string, labelValues ...string) IHistogram {
	p.histogramCollectorLocker.RLock()
	defer p.histogramCollectorLocker.RUnlock()

	coll, ok := p.histogramCollector[name]
	if !ok {
		p.app.Fatal("metrics 直方图不存在", zap.String("name", name))
	}
	histogram, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取 metrics 直方图失败", zap.Error(err))
	}
	return histogram
}

func (p *clientCli) RegistrySummary(name, help string, constLabels Labels, labels ...string) {
	p.summaryCollectorLocker.Lock()
	defer p.summaryCollectorLocker.Unlock()

	if _, ok := p.summaryCollector[name]; ok {
		p.app.Fatal("重复注册 metrics 汇总")
	}

	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(summary)
	if err != nil {
		p.app.Fatal("注册 metrics 汇总失败", zap.Error(err))
	}

	p.summaryCollector[name] = summary
}
func (p *clientCli) Summary(name string, labels Labels) ISummary {
	p.summaryCollectorLocker.RLock()
	defer p.summaryCollectorLocker.RUnlock()

	coll, ok := p.summaryCollector[name]
	if !ok {
		p.app.Fatal("metrics 汇总不存在", zap.String("name", name))
	}
	summary, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取 metrics 汇总失败", zap.Error(err))
	}
	return summary
}
func (p *clientCli) SummaryWithLabelValue(name string, labelValues ...string) ISummary {
	p.summaryCollectorLocker.RLock()
	defer p.summaryCollectorLocker.RUnlock()

	coll, ok := p.summaryCollector[name]
	if !ok {
		p.app.Fatal("metrics 汇总不存在", zap.String("name", name))
	}
	summary, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取 metrics 汇总失败", zap.Error(err))
	}
	return summary
}

func (p *clientCli) Close() {}
