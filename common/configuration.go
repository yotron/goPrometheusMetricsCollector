package common

import (
	. "github.com/yotron/goConfigurableLogger"
	"gopkg.in/yaml.v2"
)

type ConfJobs struct {
	Sid                           string   `yaml:"sid"`
	Description                   string   `yaml:"description"`
	SplunkAnalyzeFrequencySeconds string   `yaml:"splunkAnalyzeFrequencySeconds"`
	SplunkRestApiSearch           string   `yaml:"splunkRestApiSearch"`
	SplunkRestApiTimechart        string   `yaml:"splunkRestApiTimechart"`
	SplunkRestApiAdditional       string   `yaml:"splunkRestApiAdditional"`
	SplunkRestApiEarliestTime     string   `yaml:"splunkRestApiEarliestTime"`
	SplunkRestApiLatestTime       string   `yaml:"splunkRestApiLatestTime"`
	SplunkResultKeyName           string   `yaml:"splunkResultKeyName"`
	SimpleJsonResultPath          string   `yaml:"simpleJsonResultPath"`
	TimeOut                       int64    `yaml:"timeOut"`
	GoCommand                     string   `yaml:"goCommand"`
	GoCommandImports              []string `yaml:"goCommandImports"`
}

type ConfClusterType struct {
	LocalTest  string `yaml:"localTest"`
	Staging    string `yaml:"staging"`
	Production string `yaml:"production"`
}

type Conf struct {
	Type                  string             `yaml:"type"`
	RestApiHost           string             `yaml:"restApiHost"`
	RestApiPort           string             `yaml:"restApiPort"`
	RestSearchParaGeneric string             `yaml:"restSearchParaGeneric"`
	TimeOut               int64              `yaml:"timeOut"`
	ApiUrlComplete        ConfClusterType    `yaml:"apiUrlComplete"`
	Authentication        ConfAuthentication `yaml:"authentication"`
	Exporter              struct {
		LastCallFilename   string          `yaml:"lastCallFilename"`
		LastMetricFilename string          `yaml:"lastMetricFilename"`
		Parallelization    ConfClusterType `yaml:"parallelization"`
	} `yaml:"exporter"`
	Jobs []ConfJobs `yaml:"jobs"`
}

type ConfAuthentication struct {
	Username ConfClusterType `yaml:"username"`
	Password ConfClusterType `yaml:"password"`
}

type ConfServer struct {
	Port ConfClusterType `yaml:"port"`
}

func ReadCollectorConfig() map[string]Conf {
	var conf map[string]Conf
	err := yaml.Unmarshal(ReadYAMLFile("conf.collector.yml", conf), &conf)
	if err != nil {
		Error.Println("Error:", err)
	}
	Debug.Println("Conf set:\n", conf)
	return conf
}

func (conf *ConfServer) ReadServerConfig() {
	err := yaml.Unmarshal(ReadYAMLFile("conf.server.yml", conf), &conf)
	if err != nil {
		Error.Println("Error:", err)
	}
	Debug.Println("Conf set:\n", conf)
}
