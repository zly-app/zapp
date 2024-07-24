package metrics

var defaultClient Client = defNoopClient

func GetClient() Client { return defaultClient }

func SetClient(client Client) { defaultClient = client }

/*
注册计数器

	name 计数器名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistryCounter(name, help string, constLabels Labels, labels ...string) ICounter {
	return GetClient().RegistryCounter(name, help, constLabels, labels...)
}

// 获取计数器
func Counter(name string) ICounter { return GetClient().Counter(name) }

/*
注册计量器

	name 计量器名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistryGauge(name, help string, constLabels Labels, labels ...string) IGauge {
	return GetClient().RegistryGauge(name, help, constLabels, labels...)
}

// 获取计量器
func Gauge(name string) IGauge { return GetClient().Gauge(name) }

/*
注册直方图

	name 直方图名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	buckets 桶列表
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistryHistogram(name, help string, buckets []float64, constLabels Labels, labels ...string) IHistogram {
	return GetClient().RegistryHistogram(name, help, buckets, constLabels, labels...)
}

// 获取直方图
func Histogram(name string) IHistogram { return GetClient().Histogram(name) }

/*
注册汇总

	name 直方图名, 一般为 需要检测的对象_数值类型_单位
	help 一段描述文字
	constLabels 固定不变的标签值, 如主机名, ip 等
	labels 允许使用的标签, 可为nil
*/
func RegistrySummary(name, help string, constLabels Labels, labels ...string) ISummary {
	return GetClient().RegistrySummary(name, help, constLabels, labels...)
}

// 获取汇总
func Summary(name string) ISummary { return GetClient().Summary(name) }
