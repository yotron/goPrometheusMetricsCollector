package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	. "github.com/yotron/goConfigurableLogger"
	"github.com/yotron/goPrometheusMetricsCollector/collector"
	"github.com/yotron/goPrometheusMetricsCollector/common"
	"net/http"
	"time"
)

var confColls map[string]common.Conf
var confServer common.ConfServer

type appHandler func(http.ResponseWriter, *http.Request) (common.Response, error)

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if response, err := fn(w, r); err != nil {
		Error.Println("Error in Errorhandling of server response: ", response, "Error:", err)
		switch response.Status {
		case http.StatusInternalServerError:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		w.Write(response.ContentBody)
		w.Header().Set("Content-Type", response.ContentType)
	}
}

func main() {
	confColls = common.ReadCollectorConfig()
	confServer.ReadServerConfig()
	var PrometheusCollector prometheus.Collector
	conf := confColls[common.ExportType]
	if conf.Type != "" {
		b, _ := json.Marshal(conf)
		Debug.Println("Conf as json:", string(b))
		PrometheusCollector = collector.NewCollector(conf)
		prometheus.MustRegister(PrometheusCollector)
		prometheus.Unregister(prometheus.NewGoCollector())
		prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		http.Handle("/metrics/", promhttp.Handler())
		Info.Println("Metrics handler set")
		s := &http.Server{
			Addr:           ":" + common.GetDataByCluster(confServer.Port),
			Handler:        nil,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		Info.Println(s.ListenAndServe())
	} else {
		Error.Println("Could not find Collector with name: " + common.ExportType)
	}
}
