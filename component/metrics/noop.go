package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var defNoopClient Client = noopClient{}

type noopClient struct{}

func (n noopClient) RegistryCounter(name, help string, constLabels Labels, labels ...string) {}
func (n noopClient) Counter(name string, labels Labels) ICounter                             { return defNoopCounter }
func (n noopClient) CounterWithLabelValue(name string, labelValues ...string) ICounter {
	return defNoopCounter
}

func (n noopClient) RegistryGauge(name, help string, constLabels Labels, labels ...string) {}
func (n noopClient) Gauge(name string, labels Labels) IGauge                               { return defNoopGauge }
func (n noopClient) GaugeWithLabelValue(name string, labelValues ...string) IGauge {
	return defNoopGauge
}

func (n noopClient) RegistryHistogram(name, help string, buckets []float64, constLabels Labels, labels ...string) {
}
func (n noopClient) Histogram(name string, labels Labels) IHistogram { return defNoopHistogram }
func (n noopClient) HistogramWithLabelValue(name string, labelValues ...string) IHistogram {
	return defNoopHistogram
}

func (n noopClient) RegistrySummary(name, help string, constLabels Labels, labels ...string) {}
func (n noopClient) Summary(name string, labels Labels) ISummary                             { return defNoopSummary }
func (n noopClient) SummaryWithLabelValue(name string, labelValues ...string) ISummary {
	return defNoopSummary
}

func (n noopClient) Close() {}

var defNoopCounter ICounter = noopCounter{}

type noopCounter struct{}

func (n noopCounter) Inc()                                                      {}
func (n noopCounter) Add(float64)                                               {}
func (n noopCounter) AddWithExemplar(value float64, exemplar prometheus.Labels) {}

var defNoopGauge IGauge = noopGauge{}

type noopGauge struct{}

func (n noopGauge) Set(float64)                                               {}
func (n noopGauge) Inc()                                                      {}
func (n noopGauge) Dec()                                                      {}
func (n noopGauge) Add(float64)                                               {}
func (n noopGauge) Sub(float64)                                               {}
func (n noopGauge) SetToCurrentTime()                                         {}
func (n noopGauge) AddWithExemplar(value float64, exemplar prometheus.Labels) {}

var defNoopHistogram IHistogram = noopHistogram{}

type noopHistogram struct{}

func (n noopHistogram) Observe(float64)                                    {}
func (n noopHistogram) ObserveWithExemplar(value float64, exemplar Labels) {}

var defNoopSummary ISummary = noopSummary{}

type noopSummary struct{}

func (n noopSummary) Observe(float64)                                    {}
func (n noopSummary) ObserveWithExemplar(value float64, exemplar Labels) {}
