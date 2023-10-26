package metrics

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"github.com/zly-app/zapp/logger"
)

var defaultClient Client

func init() {
	zapp.AddHandler(zapp.AfterInitializeHandler, func(app core.IApp, handlerType handler.HandlerType) {
		defaultClient = newClient(app)
	})
}

func GetClient() Client {
	if defaultClient == nil {
		logger.Log.Fatal("metrics client is nil")
	}
	return defaultClient
}

/*
注册计数器

	name 计数器名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistryCounter(name, help string, constLabels Labels, labels ...string) {
	GetClient().RegistryCounter(name, help, constLabels, labels...)
}

// 获取计数器
func Counter(name string, labels Labels) ICounter { return GetClient().Counter(name, labels) }

// 获取计数器
func CounterWithLabelValue(name string, labelValues ...string) ICounter {
	return GetClient().CounterWithLabelValue(name, labelValues...)
}

/*
注册计量器

	name 计量器名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistryGauge(name, help string, constLabels Labels, labels ...string) {
	GetClient().RegistryGauge(name, help, constLabels, labels...)
}

// 获取计量器
func Gauge(name string, labels Labels) IGauge { return GetClient().Gauge(name, labels) }

// 获取计量器
func GaugeWithLabelValue(name string, labelValues ...string) IGauge {
	return GetClient().GaugeWithLabelValue(name, labelValues...)
}

/*
注册直方图

	name 直方图名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistryHistogram(name, help string, constLabels Labels, labels ...string) {
	GetClient().RegistryHistogram(name, help, constLabels, labels...)
}

// 获取直方图
func Histogram(name string, labels Labels) IHistogram { return GetClient().Histogram(name, labels) }

// 获取直方图
func HistogramWithLabelValue(name string, labelValues ...string) IHistogram {
	return GetClient().HistogramWithLabelValue(name, labelValues...)
}

/*
注册汇总

	name 直方图名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistrySummary(name, help string, constLabels Labels, labels ...string) {
	GetClient().RegistrySummary(name, help, constLabels, labels...)
}

// 获取汇总
func Summary(name string, labels Labels) ISummary { return GetClient().Summary(name, labels) }

// 获取汇总
func SummaryWithLabelValue(name string, labelValues ...string) ISummary {
	return GetClient().SummaryWithLabelValue(name, labelValues...)
}