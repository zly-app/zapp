package metrics

var defNoopClient Client = noopClient{}

type noopClient struct{}

func (n noopClient) RegistryCounter(name, help string, constLabels Labels, labels ...string) ICounter {
	return defNoopCounter
}
func (n noopClient) Counter(name string) ICounter { return defNoopCounter }

func (n noopClient) RegistryGauge(name, help string, constLabels Labels, labels ...string) IGauge {
	return defNoopGauge
}
func (n noopClient) Gauge(name string) IGauge { return defNoopGauge }

func (n noopClient) RegistryHistogram(name, help string, buckets []float64, constLabels Labels, labels ...string) IHistogram {
	return defNoopHistogram
}
func (n noopClient) Histogram(name string) IHistogram { return defNoopHistogram }

func (n noopClient) RegistrySummary(name, help string, constLabels Labels, labels ...string) ISummary {
	return defNoopSummary
}
func (n noopClient) Summary(name string) ISummary { return defNoopSummary }

var defNoopCounter ICounter = noopCounter{}

type noopCounter struct{}

func (n noopCounter) Inc(labels Labels, exemplar Labels)                {}
func (n noopCounter) Add(value float64, labels Labels, exemplar Labels) {}

var defNoopGauge IGauge = noopGauge{}

type noopGauge struct{}

func (n noopGauge) Set(v float64, labels Labels)   {}
func (n noopGauge) Inc(labels Labels)              {}
func (n noopGauge) Dec(labels Labels)              {}
func (n noopGauge) Add(v float64, labels Labels)   {}
func (n noopGauge) Sub(v float64, labels Labels)   {}
func (n noopGauge) SetToCurrentTime(labels Labels) {}

var defNoopHistogram IHistogram = noopHistogram{}

type noopHistogram struct{}

func (n noopHistogram) Observe(value float64, labels Labels, exemplar Labels) {}

var defNoopSummary ISummary = noopSummary{}

type noopSummary struct{}

func (n noopSummary) Observe(value float64, labels Labels, exemplar Labels) {}
