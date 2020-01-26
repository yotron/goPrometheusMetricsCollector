package grabber

import (
	. "github.com/yotron/goConfigurableLogger"
	"github.com/yotron/goPrometheusMetricsCollector/common"
	"io/ioutil"
	"net/http"
	"time"
)

func PollAPI() []byte {
	client := &http.Client{
		CheckRedirect: redirectPolicyFunc,
		Timeout:       time.Duration(Conf.TimeOut) * time.Second,
	}
	req, err := http.NewRequest("GET", common.GetDataByCluster(Conf.ApiUrlComplete), nil)
	if err != nil {
		Error.Println("Request creation failed:", err)
		panic("Request creation failed")
	}
	Info.Println("Connecting to: ", common.GetDataByCluster(Conf.ApiUrlComplete))
	if common.GetDataByCluster(Conf.Authentication.Username) != "" {
		req.SetBasicAuth(common.GetDataByCluster(Conf.Authentication.Username), PW)
	}
	resp, err := client.Do(req)
	if err != nil {
		Error.Println("Request failed:", err)
		panic("Request failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Error.Println("Could not read response:", err)
		panic("Could not read response")
	}
	return body
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if common.GetDataByCluster(Conf.Authentication.Username) != "" {
		req.SetBasicAuth(common.GetDataByCluster(Conf.Authentication.Username), PW)
	}
	return nil
}
