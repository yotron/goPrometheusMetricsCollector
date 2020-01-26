package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	. "github.com/yotron/goConfigurableLogger"
	"github.com/yotron/goPrometheusMetricsCollector/common"
	"github.com/yotron/goPrometheusMetricsCollector/common/jsontools"
	"github.com/yotron/goPrometheusMetricsCollector/grabber"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type collectors struct {
	desc map[string]*prometheus.Desc
}

var wg sync.WaitGroup

func NewCollector(confPar common.Conf) *collectors {
	grabber.Conf = confPar
	grabber.PW = common.GetPassword(grabber.Conf.Authentication)
	if grabber.Conf.Exporter.LastCallFilename != "" && grabber.Conf.Exporter.LastMetricFilename != "" {
		grabber.PersistorLastCall.InitConfig(grabber.Conf.Exporter.LastCallFilename)
		grabber.PersistorMetric.InitConfig(grabber.Conf.Exporter.LastMetricFilename)
	}
	Debug.Println("conf", grabber.Conf.Authentication)
	var test1 = grabber.Conf.Exporter.Parallelization
	Debug.Println(test1)
	if confPar.Type == "splunkRequest" {
		if threads, err := common.GetInt(common.GetDataByCluster(grabber.Conf.Exporter.Parallelization)); err == nil {
			Info.Println("Parallelization set to:", threads)
			runtime.GOMAXPROCS(int(threads))
		} else {
			panic("Could not read Parallelization value. Skipped.")
		}
	}
	var coll collectors
	coll.desc = make(map[string]*prometheus.Desc)
	for _, job := range grabber.Conf.Jobs {
		Debug.Println(job.Sid, "Set collector type", grabber.Conf.Type)
		coll.desc[job.Sid] = prometheus.NewDesc(
			job.Sid,
			job.Description,
			nil, nil,
		)
	}
	Debug.Println("Collection set:", &coll)
	return &coll
}

func (collector *collectors) Describe(ch chan<- *prometheus.Desc) {
	Info.Println("Describe SplunkCollector")
	for _, job := range grabber.Conf.Jobs {
		Debug.Println(job.Sid, "SplunkCollector described")
		ch <- collector.desc[job.Sid]
	}
}

func (collector *collectors) Collect(ch chan<- prometheus.Metric) {
	Info.Println("Start Collect")
	Debug.Println(grabber.Conf)
	if grabber.Conf.Type == "splunkRequest" {
		for _, job := range grabber.Conf.Jobs {
			wg.Add(1)
			Info.Println(job.Sid, "Job recognized", wg)
			go func(job common.ConfJobs) {
				secsStart := time.Now().Unix()
				defer func() {
					Info.Println(job.Sid, "Process removed", wg)
					wg.Done()
					secsDuration := time.Now().Unix() - secsStart
					Info.Println(job.Sid, "Process duration", secsDuration)
				}()
				Debug.Println("Job handled in process:", job)
				var metricResult grabber.SplunkMetric
				metricResult.SetMetric(job)
				Debug.Println("Metric value:", &metricResult)
				if metricResult.Metricvalue != "" {
					mv, err := strconv.ParseFloat(metricResult.Metricvalue, 64)
					Info.Println(job.Sid, "Floated Metric value:", mv)
					if err != nil {
						Error.Println(job.Sid, "Error during transformation of:", metricResult.Metricvalue, " Error:", err, "will proceed")
					} else {
						Debug.Println(job.Sid, "Metric to send:", metricResult.Collectorname, "Value:", mv)
						ch <- prometheus.MustNewConstMetric(collector.desc[metricResult.Collectorname], prometheus.GaugeValue, mv)
					}
				}
			}(job)
		}
		Info.Println("Wait To Finish")
		wg.Wait()
	} else if grabber.Conf.Type == "simpleAPIRequest" {
		defer func() {
			if r := recover(); r != nil {
				Error.Println("Was panic. Send nothing.", r)
			}
		}()
		err := json.Unmarshal(grabber.PollAPI(), &grabber.Data)
		if err != nil {
			Error.Println("Could not parse json from response:", err)
			panic("Could not parse json from response")
		}
		var metricJSON string
		for _, job := range grabber.Conf.Jobs {
			metricJSON = jsontools.GetDataOfSlice(grabber.Data, job.SimpleJsonResultPath, 0)
			Debug.Println(job.Sid, "JobId:", job.SimpleJsonResultPath, "Value:", metricJSON)
			if job.GoCommand != "" {
				metricJSON = common.GetGoCommandValue(metricJSON, job)
			}
			valueToReturn, err := strconv.ParseFloat(metricJSON, 64)
			if err != nil {
				Error.Println(job.Sid, "Cloud not parse value for", job.SimpleJsonResultPath, ", Value:", metricJSON)
				panic("Cloud not parse value for" + job.SimpleJsonResultPath + ", Value:" + metricJSON)
			}
			ch <- prometheus.MustNewConstMetric(collector.desc[job.Sid], prometheus.GaugeValue, valueToReturn)
		}
	}
}
