package metrics

type (
	Labels = map[string]string

	ICounter = interface {
		Inc(labels Labels, exemplar Labels)
		Add(v float64, labels Labels, exemplar Labels)
	}

	IGauge = interface {
		Set(v float64, labels Labels)
		Inc(labels Labels)
		Dec(labels Labels)
		Add(v float64, labels Labels)
		Sub(v float64, labels Labels)
		SetToCurrentTime(labels Labels)
	}

	IHistogram = interface {
		Observe(v float64, labels Labels, exemplar Labels)
	}

	ISummary = interface {
		Observe(v float64, labels Labels, exemplar Labels)
	}
)

type Client interface {
	/*注册计数器
	  name 计数器名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryCounter(name, help string, constLabels Labels, labels ...string) ICounter
	// 获取计数器
	Counter(name string) ICounter

	/*注册计量器
	  name 计量器名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryGauge(name, help string, constLabels Labels, labels ...string) IGauge
	// 获取计量器
	Gauge(name string) IGauge

	/*注册直方图
	  name 直方图名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  buckets 桶列表
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryHistogram(name, help string, buckets []float64, constLabels Labels, labels ...string) IHistogram
	// 获取直方图
	Histogram(name string) IHistogram

	/*注册汇总
	  name 直方图名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistrySummary(name, help string, constLabels Labels, labels ...string) ISummary
	// 获取汇总
	Summary(name string) ISummary
}
