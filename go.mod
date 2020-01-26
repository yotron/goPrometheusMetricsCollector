module github.com/yotron/goPrometheusMetricsCollector

go 1.13

require (
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/prometheus/client_golang v1.3.0
	github.com/yotron/goConfigurableLogger v0.0.0-20200125155224-b4950fd3e34c
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/yotron/goPrometheusMetricsCollector => ./
