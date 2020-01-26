package grabber

import (
	"crypto/tls"
	"encoding/json"
	. "github.com/yotron/goConfigurableLogger"
	"github.com/yotron/goPrometheusMetricsCollector/common"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

type response struct {
	status_code int
	body        []byte
}

type SplunkMetric struct {
	Metricname    string
	Metricvalue   string
	Collectorname string
}

type JobSid struct {
	Sid string `json:"sid"`
}

type entry struct {
	Name    string `json:"name"`
	Author  string `json:"author"`
	Content struct {
		SearchEarliestTime int64  `json:"searchEarliestTime"`
		SearchLatestTime   int64  `json:"searchLatestTime"`
		Sid                string `json:"sid"`
		DispatchState      string `json:"dispatchState"`
	} `json:"content"`
}

type JobStatus struct {
	Updated string  `json:"updated"`
	Entry   []entry `json:"entry"`
}

type message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type field struct {
	Name          string `json:"name"`
	Groupby_rank  string `json:"groupby_rank"`
	Data_source   string `json:"data_source"`
	Splitby_field string `json:"splitby_field"`
	Splitby_value string `json:"splitby_value"`
}

type SearchResults struct {
	Init_offset int       `json:"init_offset"`
	Messages    []message `json:"messages"`
	Fields      []field
	Results     []interface{}
}

var timeout time.Duration
var confServer common.ConfServer

func (sm *SplunkMetric) SetMetric(job common.ConfJobs) {
	Debug.Println(uniqueSplunkSid(job.Sid), "Job handled")
	defer func() {
		if r := recover(); r != nil {
			Error.Println(uniqueSplunkSid(job.Sid), "Was panic. Send nothing.", r, "will proceed")
		}
	}()
	timeout = time.Duration(job.TimeOut) * time.Second
	diff := int64(time.Now().Unix()) - getLastCall(job)
	Debug.Println(uniqueSplunkSid(job.Sid), "difference:", diff, "frequency setting:", common.GetIntOnly(job.SplunkAnalyzeFrequencySeconds))
	if common.GetIntOnly(job.SplunkAnalyzeFrequencySeconds) < 0 {
		sm.realtimeResultRun(job)
	} else if getLastMetricRun(job) == "-1" {
		sm.lastMetricResultRun(job)
	} else if getLastMetricRun(job) == "-2" || getLastCall(job) == 0 || diff >= common.GetIntOnly(job.SplunkAnalyzeFrequencySeconds) {
		cacheCreateRun(job)
	} else {
		sm.getCachedValue(job)
	}
	Debug.Println("Metric at the end:", sm)
}

func (sm *SplunkMetric) realtimeResultRun(job common.ConfJobs) {
	Info.Println(job.Sid, "realtimeResultRun started")
	var searchResult SearchResults
	searchResult.createJobGetResult(job)
	Debug.Println(searchResult)
	*sm = searchResult.getMetric(job)
	Info.Println(*sm)
}

func cacheCreateRun(job common.ConfJobs) {
	Info.Println(uniqueSplunkSid(job.Sid), "cacheCreateRun started")
	persistLastMetric(job, "-1")
	var jobSid JobSid
	jobSid.createJob(job)
}

func (sm *SplunkMetric) getCachedValue(job common.ConfJobs) {
	Info.Println(job.Sid, "getCachedValue started.")
	sm.new("dummy", getLastMetricRun(job), job.Sid)
}

func (sm *SplunkMetric) lastMetricResultRun(job common.ConfJobs) {
	Info.Println(uniqueSplunkSid(job.Sid), "lastMetricResultRun started.")
	var jobStatus JobStatus
	var searchResult SearchResults
	jobStatus.analyzeJobStatus(job)
	if jobStatus.Entry[0].Content.DispatchState == "DONE" {
		Debug.Println(uniqueSplunkSid(job.Sid), "Dispatch State:", jobStatus.Entry[0].Content.DispatchState)
		searchResult.analyzeJob(job)
		*sm = searchResult.getMetric(job)
		if sm != nil {
			persistLastMetric(job, sm.Metricvalue)
			Info.Println(job.Sid, "Latest time to persist:", jobStatus.Entry[0].Content.SearchLatestTime, "Sid:", jobStatus.Entry[0].Content.Sid)
			persistLastCall(job, jobStatus.Entry[0].Content.SearchLatestTime)
		}
		deleteJob(uniqueSplunkSid(job.Sid))
	}
}

func persistLastCall(job common.ConfJobs, lastest_time int64) {
	Debug.Println(job.Sid, "Write latest time", lastest_time)
	PersistorLastCall.Write(job.Sid, lastest_time)
}

func persistLastMetric(job common.ConfJobs, last_metric string) {
	Debug.Println(job.Sid, "Write metric", last_metric)
	PersistorMetric.Write(job.Sid, last_metric)
}

func (js *JobSid) createJob(job common.ConfJobs) {
	Debug.Println(uniqueSplunkSid(job.Sid), ": Beginn of job creation")
	Debug.Println("Job content: ", uniqueSplunkSid(job.Sid))
	ea, la := getTimeRange(job)
	Debug.Println("Earliest: ", ea, "Latest:", la)
	postBodyData := job.SplunkRestApiSearch + job.SplunkRestApiTimechart + job.SplunkRestApiAdditional + "&earliest_time=" + ea + "&latest_time=" + la
	replacer := getReplacerForConfigSetting(uniqueSplunkSid(job.Sid))
	urlStringParsed := replacer.Replace("<splunk.rest_api_host>:<splunk.rest_api_port>/services/search/jobs?" + Conf.RestSearchParaGeneric)
	Debug.Println(urlStringParsed)
	bodydataStringParsed := replacer.Replace(replacer.Replace(postBodyData))
	Debug.Println(bodydataStringParsed)
	var resp response
	resp.postResponse(urlStringParsed, bodydataStringParsed)
	js.handleJobCreationPOSTResponse(resp)
}

func (sr *SearchResults) createJobGetResult(job common.ConfJobs) {
	Debug.Println(uniqueSplunkSid(job.Sid), ": Begin of job creation")
	Debug.Println("Job content: ", uniqueSplunkSid(job.Sid))
	ea, la := getTimeRange(job)
	Debug.Println("Earliest: ", ea, "Latest:", la)
	postBodyData := job.SplunkRestApiSearch + job.SplunkRestApiTimechart + job.SplunkRestApiAdditional + "&earliest_time=" + ea + "&latest_time=" + la + "&exec_mode=oneshot"
	replacer := getReplacerForConfigSetting(uniqueSplunkSid(job.Sid))
	urlStringParsed := replacer.Replace("<splunk.rest_api_host>:<splunk.rest_api_port>/services/search/jobs?" + Conf.RestSearchParaGeneric)
	Debug.Println(urlStringParsed)
	bodydataStringParsed := replacer.Replace(replacer.Replace(postBodyData))
	Debug.Println(bodydataStringParsed)
	var resp response
	resp.postResponse(urlStringParsed, bodydataStringParsed)
	sr.handleJobAnalyzeResultGETResponse(resp)
}

func getTimeRange(job common.ConfJobs) (string, string) {
	var earliest_time string
	var latest_time string
	time_now := time.Now()
	if job.SplunkRestApiEarliestTime == "last_call" {
		_, err := os.Stat(Conf.Exporter.LastMetricFilename)
		if err == nil {
			earliest_time = epocheToSplunkString(getLastCallSet15SecBeforeIfNull(job))
			Debug.Println("File splunk_last_call exists. Earliest set to", earliest_time)
		} else if os.IsNotExist(err) {
			Debug.Println("File splunk_last_call not exists. Earliest set to dummy -15s")
			earliest_time = "-15s"
		} else {
			Error.Println(uniqueSplunkSid(job.Sid), "file splunk_last_call stat error: %v", err)
		}
	} else {
		Debug.Println(uniqueSplunkSid(job.Sid), "Earliest time set to:", job.SplunkRestApiEarliestTime)
		earliest_time = job.SplunkRestApiEarliestTime
	}
	if job.SplunkRestApiLatestTime == "now" {
		latest_time = time_now.Format(time.RFC3339)
		Debug.Println("File splunk_last_call set to 'now'. Latest set to ", latest_time)
	} else {
		Debug.Println(uniqueSplunkSid(job.Sid), "Latest time set to:", job.SplunkRestApiLatestTime)
		latest_time = job.SplunkRestApiLatestTime
	}
	return earliest_time, latest_time
}

func getLastCallSet15SecBeforeIfNull(job common.ConfJobs) int64 {
	last_call := getLastCall(job)
	if last_call == 0 {
		last_call = int64(time.Now().Unix()) - 15
		Debug.Println("Created new LastCall typeOf:", last_call)
	}
	return last_call
}

func getLastCall(job common.ConfJobs) int64 {
	last_call_all := make(map[string]int64)
	json.Unmarshal(common.ReadFile(Conf.Exporter.LastCallFilename), &last_call_all)
	last_call := last_call_all[job.Sid]
	Debug.Println("LastCall TypeOf:", reflect.TypeOf(last_call), "LastCall Value:", last_call)
	return last_call
}

func getLastMetricRun(job common.ConfJobs) string {
	last_metric_all := make(map[string]string)
	json.Unmarshal(common.ReadFile(Conf.Exporter.LastMetricFilename), &last_metric_all)
	last_metric := last_metric_all[job.Sid]
	Debug.Println("Status TypeOf:", reflect.TypeOf(last_metric), "last metric Value:", last_metric)
	if last_metric == "" {
		Debug.Println("Last metric not found:", last_metric)
	}
	return last_metric
}

func deleteJob(sid string) {
	Debug.Println("SID to delete: ", sid)
	replacer := getReplacerForConfigSetting(sid)
	urlString := replacer.Replace("<splunk.rest_api_host>:<splunk.rest_api_port>/services/search/jobs/<sid>")
	var resp response
	resp.deleteResponse(urlString)
	handleJobDeleteResultDELETEResponse(resp)
}

func (sr *SearchResults) analyzeJob(job common.ConfJobs) {
	Debug.Println("SID to be analyzed: " + uniqueSplunkSid(job.Sid))
	replacer := getReplacerForConfigSetting(uniqueSplunkSid(job.Sid))
	urlString := replacer.Replace("<splunk.rest_api_host>:<splunk.rest_api_port>/services/search/jobs/<sid>/results?output_mode=json")
	var resp response
	resp.getResponse(urlString)
	sr.handleJobAnalyzeResultGETResponse(resp)
}

func (js *JobStatus) analyzeJobStatusWaitForDone(job common.ConfJobs) {
	js.analyzeJobStatus(job)
	Debug.Println(uniqueSplunkSid(job.Sid), "JobStatus:", js)
	startSec := time.Now().Unix()
	for js.Entry[0].Content.DispatchState != "DONE" {
		Debug.Println(job.Sid, "Current Duration", time.Now().Unix()-startSec, "Current Timeout:", int64(timeout/time.Second))
		if time.Now().Unix()-startSec > int64(timeout/time.Second)-3 {
			Error.Println(job.Sid, "Response time near to the timeout")
			break
		}
		time.Sleep(time.Second * 1)
		js.analyzeJobStatus(job)
	}
}

func (js *JobStatus) analyzeJobStatus(job common.ConfJobs) {
	Debug.Println("SID to analyze: " + uniqueSplunkSid(job.Sid))
	replacer := getReplacerForConfigSetting(uniqueSplunkSid(job.Sid))
	urlString := replacer.Replace("<splunk.rest_api_host>:<splunk.rest_api_port>/services/search/jobs/<sid>?output_mode=json")
	var resp response
	resp.getResponse(urlString)
	js.handleJobStatusResultGETResponse(resp)
	Debug.Println(uniqueSplunkSid(job.Sid), "after job status anylyzing. Result:", js)
	if js.Entry[0].Content.DispatchState == "FAILED" {
		Error.Println(job.Sid, "SID analyzeJobStaus request failed")
		panic("request failed")
	} else if js.Entry[0].Content.DispatchState == "FATAL" {
		Error.Println(uniqueSplunkSid(job.Sid), "SID analyzeJobStaus request failed")
		persistLastMetric(job, "-2")
		panic("request failed")
	}
}

func (js *JobStatus) handleJobStatusResultGETResponse(response response) {
	Debug.Println("Reponse to handle:", response.status_code, "body:", response.body)
	if response.status_code == 200 {
		Debug.Println("Request success: HTTP-Status ", response.status_code, " Body: ", string(response.body))
		err := json.Unmarshal(response.body, &js)
		if err != nil {
			Error.Println("Could not create a result:", err)
		} else {
			Debug.Println("Result created: ", js)
		}
	} else if response.status_code == 404 {
		Error.Println("Request not succesfull: HTTP-Status ", response.status_code, " Body: ", string(response.body))
		var sr *SearchResults
		err := json.Unmarshal(response.body, &sr)
		if err != nil {
			Error.Println("Could not handle a 404 error response as a search result:", err, "response body:", response.body)
		} else {
			Debug.Println("Analyzing Job Status failed. Prepare for setting back job for recreation.")
			var ent entry
			ent.Content.DispatchState = "FATAL"
			js.Entry = append(js.Entry, ent)
		}
	}
}

func (jcr *JobSid) handleJobCreationPOSTResponse(response response) {
	if response.status_code == 201 {
		err := json.Unmarshal(response.body, &jcr)
		if err != nil {
			Error.Println("Error during parsing:", err)
		} else {
			Debug.Println("Sid created: " + jcr.Sid)
		}
	}
}

func (sr *SearchResults) handleJobAnalyzeResultGETResponse(response response) {
	Debug.Println("Body:", string(response.body))
	if response.status_code == 200 {
		err := json.Unmarshal(response.body, &sr)
		if err != nil {
			Error.Println("The result cannot be analyzed. Error: %s", err)
		} else {
			for _, message := range sr.Messages {
				if message.Type == "ERROR" {
					Error.Println("The result has errors: " + message.Text)
				}
			}
			Debug.Println("Result created: ", sr)
		}
	}
}

func handleJobDeleteResultDELETEResponse(response response) {
	if response.status_code != 201 {
		Debug.Println("Delete Request success: HTTP-Status ", response.status_code, " Body: ", string(response.body))
	}
}

func (thisresp *response) deleteResponse(getURLString string) {
	Debug.Println("DELETE Request to fire: " + getURLString)
	req, err := http.NewRequest("DELETE", getURLString, nil)
	if err != nil {
		Error.Println("Error setting request: %s", err)
	} else {
		Debug.Println("DELETE Request set: %s", req)
	}
	req.SetBasicAuth(common.GetDataByCluster(Conf.Authentication.Username), PW)
	thisresp.doRequest(req)
}

func (thisresp *response) postResponse(getURLString string, body_data string) {
	postContent := strings.NewReader(body_data)
	Debug.Println("Post-URL:", getURLString, "BodyData:", body_data)
	req, err := http.NewRequest("POST", getURLString, postContent)
	if err != nil {
		Error.Println("Error during request creation: %s", err)
	} else {
		Debug.Println(req)
	}
	req.SetBasicAuth(common.GetDataByCluster(Conf.Authentication.Username), PW)
	thisresp.doRequest(req)
}

func (thisresp *response) getResponse(getURLString string) {
	Debug.Println("GET-URL:", getURLString)
	req, err := http.NewRequest("GET", getURLString, nil)
	if err != nil {
		Error.Println("Error: %s", err)
	} else {
		Debug.Println(req)
	}
	req.SetBasicAuth(common.GetDataByCluster(Conf.Authentication.Username), PW)
	thisresp.doRequest(req)
}

func (thisresp *response) doRequest(req *http.Request) {
	tr := getTransportConf()
	Debug.Println("TR: ", tr, "Request: ", timeout)
	client := http.Client{
		Transport: tr,
	}
	var resp *http.Response
	var err error
	Debug.Println("Client: ", client, "Request: ", req, "Transport:", tr)
	Debug.Println("Start Epoche ", time.Now().Unix())
	Debug.Println("Request:", req.URL.Path, "Host:", req.Host)
	resp, err = client.Do(req)
	if err != nil {
		Error.Println("Error during request:", err)
		panic("Request failed: HTTP-Status ")
	} else {
		thisresp.status_code = resp.StatusCode
		thisresp.body, err = ioutil.ReadAll(resp.Body)
		Debug.Println("Stop Epoche ", time.Now().Unix())
		if thisresp.status_code < 300 {
			Debug.Println("Request success: HTTP-Status ", thisresp.status_code, "Response:", string(thisresp.body))
		} else {
			Error.Println("Request failed: HTTP-Status ", thisresp.status_code, "Error:", string(thisresp.body))
			panic("Request failed: HTTP-Status ")
		}
	}
}

func (sr *SearchResults) getMetric(job common.ConfJobs) SplunkMetric {
	Debug.Println("Result for Metricextraction: ", sr)
	var sp SplunkMetric
	splunkResultKey := job.SplunkResultKeyName
	Info.Println(job.Sid, "resultKey to analyze:", splunkResultKey, "Amount of results:", len(sr.Results))
	if len(sr.Results) > 0 {
		for _, result := range sr.Results {
			value := reflect.ValueOf(result)
			Debug.Println("Value to reflect:", value)
			if value.Kind() == reflect.Map {
				Debug.Println("Value is array:", value.Kind())
				idx, struc := common.GetDataOfSlice(value, splunkResultKey)
				Debug.Println("Index in array:", idx, "Structure:", struc)
				if idx != -1 {
					sp.new(splunkResultKey, struc.Interface().(string), job.Sid)
				} else {
					Error.Println(job.Sid, "Could not find", splunkResultKey, "in result")
					panic("Parameter not found.")
				}
			} else {
				Error.Println(job.Sid, "Could not find", splunkResultKey, "in result")
				panic("Parameter not found.")
			}
		}
	} else {
		sp.new(splunkResultKey, "0", job.Sid)
	}
	Debug.Println("Metric to return: ", sp)
	return sp
}

func (sp *SplunkMetric) new(name string, value string, collector string) {
	sp.Metricname = name
	sp.Metricvalue = value
	sp.Collectorname = collector
	Debug.Println("Metric set:", sp)
}

func getTransportConf() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout: time.Duration(timeout) * time.Second,
	}
}

func getReplacerForConfigSetting(sid string) *strings.Replacer {
	Debug.Println("SID to replace: " + sid)
	return strings.NewReplacer(
		"<splunk.rest_api_host>", Conf.RestApiHost,
		"<splunk.rest_api_port>", Conf.RestApiPort,
		"<sid>", sid)
}

func epocheToSplunkString(epoche int64) string {
	Debug.Println("Epoche: ", epoche)
	t := (time.Unix(epoche, 0)).Format(time.RFC3339)
	Debug.Println("Time: ", t)
	return t
}

func uniqueSplunkSid(sid string) string {
	return common.Cluster + "_" + sid
}
