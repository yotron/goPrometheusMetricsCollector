#  [![yotron](logo-yotron.png)](http://www.yotron.de)

[YOTRON](http://www.yotron.de) is a consultancy company which is focused on DevOps, Cloudmanagement and 
Data Management with NOSQL and SQL-Databases. Visit us on [ www.yotron.de ](http://www.yotron.de)

[![https://github.com/yotron/goPrometheusMetricsCollector/actions?query=workflow:GoBuild](https://github.com/yotron/goPrometheusMetricsCollector/workflows/GoBuild/badge.svg?branch=master&event=push)](https://github.com/yotron/goPrometheusMetricsCollector/actions?query=workflow:GoBuild)


# PrometheusMetricsCollector
This Golang project offers an general API endpoint to grab metrics from [Splunk](https://www.splunk.com/) or any webbased API (e.g., RestBased) to provide 
then for Prometheus in their special format.

The collection of the source API can run on the fly, when Prometheus sends a request or it produced and cache a metric.
This is crucial for long running Splunk requests.

Out of the Box the PrometheusMetricsCollector provides the grabbing of metrics from a [Splunk-API](https://docs.splunk.com/Documentation/Splunk/8.0.1/RESTREF/RESTprolog) and from a simple API like REST
or other APIs which sends a JSON in the response. 

For long running request a caching mechanism and an exact analysis is possible, For Splunk you can define the start and the end of
the time frame you want to generate a metric from.  

PrometheusMetricsCollector caching the last call of a request. For that you can collect the data for the metrics from that point on.

If you need a metric in a low frequency (e.g., daily, weekly, monthly) then PrometheusMetricsCollector caches the data and send
cached values to Prometheus when he requested new metrics.

We have a one project-for-all approach:

- You can setup all Collectors in one project. But you must start each Collector separately in different nodes via Environment variables.
- You can setup a each Collector for different environments. The project supports out of the box the environment
  localTest, staging and production.

So you can provide different Collectors for different environments with different Jobs in one single project. 


# table of content

 - [PrometheusMetricsCollector](#prometheusmetricscollector)

   - [general concept behind the PrometheusMetricsCollector](#general-concept-behind-the-prometheusmetricscollector)

   - [Important files](#important-files)

   - [Important environment variables](#important-environment-variables)

   - [Start PrometheusMetricsCollector](#start-prometheusmetricscollector)

   - [possible environments](#possible-environments)

   - [Configure the Golang API server](#configure-the-golang-api-server)

   - [Setting up passwords in a file](#setting-up-passwords-in-a-file)

   - [Configure the Collectors](#configure-the-collectors)

     - [The setup of a Collector](#the-setup-of-a-collector)

       - [general settings](#general-settings)

         - [Block Type](#block-type)

         - [Block authentication](#block-authentication)

       - [Block: Settings for splunkRequests](#block-settings-for-splunkrequests)

       - [Block: Settings for simpleAPIRequest](#block-settings-for-simpleapirequest)

     - [How to test PrometheusMetricsCollector](#how-to-test-prometheusmetricscollector)

 - [How to deploy PrometheusMetricsCollector](#how-to-deploy-prometheusmetricscollector)

 - [How to develop goCommands](#how-to-develop-gocommands)

   - [Parameter of the configuration file](#parameter-of-the-configuration-file)

   - [Needed parameter](#needed-parameter)

   - [Debug the code](#debug-the-code)

   - [some examples of goCommands](#some-examples-of-gocommands)

     - [decimal points](#decimal-points)

     - [Format human readable bytes](#format-human-readable-bytes)

     - [Analyse a JSON string](#analyse-a-json-string)

     - [Switch from true/false](#switch-from-truefalse)

   - [Testing area](#testing-area)

 - [own credentials](#own-credentials)

(this table of content was created by [Markdown Menu](https://www.markdownmenu.com))

## general concept behind the PrometheusMetricsCollector
PrometheusMetricsCollector contains different objects. These are
- Collectors
- Jobs 

A Collector is a running web server from which Prometheus can grab metrics from. A Collector runs always as a service and is deployed on 
 a physical or virtual server. Every Collector has a unique name.

Each Collector has different Jobs. A Job can be described best as one single metric for Prometheus. When Prometheus wants to grab some metrics 
from the Collector, each Jobs of the Collectors will run and provide metrics for Prometheus. Every Job has a unique `sid` for 
identification of the Job.

Here is the main difference between a 
- `simpleAPIRequest` and a 
- `splunkRequest`. 

In a `simpleAPIRequest` PrometheusMetricsCollector needed metrics are provided by the source API. The PrometheusMetricsCollector-Job is filtering the right metric 
out of the response and transforms the value accordingly to the requirements of Prometheus.

In a `splunkRequest` each Job is a single and separated request to Splunk-API. This Job expects only one single and for Prometheus formatted 
result, for example the amount of requests stored as log entries of the last week of one NGINX-Application.
The SplunkAPI-request must provide that.  

`splunkRequest` can run very long depends on the kind of request (for example, aggregate the amount of log entries for a complete month as a metric).
You can define if the Collector shall wait till the end of a request of a Job or it will deliver it 
when it is available. In between PrometheusMetricsCollector delivers only available metrics.

For `splunkRequest` you can also define the frequency of a request. When you collect for example the amount of log entries for a complete month
it makes no sense to do it on every request of Prometheus. So you can define a frequency to do it every month for example. 
In between Prometheus gets cached values.  

## Important files
There are some important configuration files for this project needed.

API-configuration:
- `conf.Collector.yml` (mandatory): Configure the kind of collection for the different purposes.
   This is the main configuration file for this project. 
- `conf.server.yml` (mandatory): Configure the server ports for the project API.

Authentication:
- `passwordV2.yml` (optional): For the basic authentication against the source API 
   you can provide the password within this file. The alternatives for the password location are a environment variable,
   the manual input or as a direct part of the URL.
   
Caching:   
- `splunk_last_call.json` (mandatory): Information of the last time in Epoche the Splunk request was done. 
- `splunk_last_metric.json` (mandatory): Information of the last value the Splunk request was done.

Logging (for more information please see the github-project [goConfigurableLogger](https://github.com/yotron/goConfigurableLogger)): 
- `conf.logging.yml` (mandatory): Setup the logging. 
- `metrics_exporter.log` (optional): Logfile with the content.

## Important environment variables
When you want to run this project two environment variables must be provided.

- `YOPRO_ENVIRONMENT`: The name of the environment, this API shall run. The value can be `localTest`, `staging` or `production`.
- `YOPRO_COLLECTOR`: The name of the Collector to grab data from and provide that for Prometheus. The name of the Collector 
   is freely definable and is defined by you in the configuration file to "Configure the Collectors" (please see below).

## Start PrometheusMetricsCollector
Define the environment variables and simply:

`go run main.go`

## possible environments
You can deploy the project on different environments. Important parameters in der configuration files
are separated in:

- `localTest: 9005`
- `staging: 9101`
- `production: 9101`

When you want to deploy your project you must setup a environment variable called `YOPRO_ENVIRONMENT` to define
in which environment the Collector shall run.

## Configure the Golang API server
The PrometheusMetricsCollector runs with the hostname the Golang server is running on as a service.
But must define the port the API is listening on. The ports can be defined for the different environments.

The ports are defined in the file `conf.server.yml`:

```
---
port: 
  localTest: 9101
  staging: 8080
  production: 8080
```

## Setting up passwords in a file
For authentication against the source API (e.g., Splunk) name and password is needed. This project provides 
different mechanism to provide the password. One mechanism is to store the password in clear text within the password 
file `passwordV2.yml` for the different Collectors:

```
---
passwords:
  SplunkExample: password
  TomcatAppJMX: password
  JMXExample2: password
```
Remarks:
- **Please do use that mechanism 
  to store password for staging or production environments of the source API.
  The danger of submitting passwords into your versioning system is very high.** 
  This file shall only be used in local environments with dummy passwords for development! 
- A example of the file is located in that project.
- With the creation of the file, the password will not be used immediately. 
  Which mechanism for password providing is defined in the configuration file for the Collectors.

## Configure the Collectors
All Collectors for the different environments are defined in one single configuration file `conf.Collector.yml`.
This file contains the complete setup for all Collectors. 

You find an complete example in this project.

The setup is separated in Collectors. Each Collector has a name as his id. 
It is recommended, that the name shall not contain any special character or white spaces.  

In `conf.collector.yml` the root entries is always the name of the Collector. In the example file it is `SplunkExample`
`TomcatAppJMX` and `JMXExample2`
 
### The setup of a Collector
#### general settings
For each Collector you must define the type of the Collector and the authentication setup.

##### Block Type
The type defines which kind of request the Collector is using. 
```
type: splunkRequest
```
`type`: Could be `splunkRequest` or `simpleAPIRequest`

##### Block authentication
The `authentication` block of that yaml-file is used to authenticate against the source API (Simple-API or Splunk). If you don't need
authentication, please leave the entries empty.

```
authentication:
  username:
    localTest: <user_name>
    staging: sys.splunk
    production: sys.splunk
  password:
    localTest: passwordV2.yml
    staging: manually
    production: YOPRO_AUTH_PASSWORD
```
`username`: The username to use for authentication.

`password`: The source which contains the password. The sources can be:
- *Name of a yml-File*: A password file contains the password (please see above). The name of the password file can be freely 
  selected. The location of the file is in the root folder of that project (in this example `passwordV2.yml`) and must end with `.yml`.
- *YOPRO_AUTH_PASSWORD*: YOPRO_AUTH_PASSWORD is the environment variable to provide the password as a value.
- *MANUALLY*: You can add the keyword *MANUALLY* to add the password during the startup of PrometheusMetricsCollector.     
- *password*: The lazy way is to add the password directly in that parameter.

Remarks:
- We recommend to use environment variables to setup the password.
- PrometheusMetricsCollector only supports basic authentication.

#### Block: Settings for splunkRequests
The general setup for Splunk requests contains the following parameters: 
```
type: splunkRequest
restApiHost: https://splunk.yotron.de
restApiPort: 8089
restSearchParaGeneric: '&search_mode=realtime&output_mode=json'  ## do not change
exporter:
  lastCallFilename: splunk_last_call.json
  lastMetricFilename: splunk_last_metric.json    
  parallelization:
    localTest: 2
    staging: 5        
    production: 10
``` 

`restApiHost`: The host with the protocol (http or https) of the Splunk-API.

`restApiPort`: The port of the Splunk-API.

`restSearchParaGeneric`: The general setting for each request against the Splunk-API. This setting in our example is 
  mandatory as a minimum to run proper requests against Splunk.
   
`exporter`:
- `lastCallFilename`: The name of the file where the timestamp of the last call is cached for each Job.
   The file must be located in the root folder of that project.
- `lastMetricFilename`: The name of the file where the metrics of the last call for each Job is cached.
- `parallelization`: The amount of parallel processes to run with every single request by Prometheus. With that you can run Jobs in parallel 
   and not sequentially, so the runtime of a call by Prometheus can be reduced.
   
The setup of every Job has the following parameter:
```       
sid: splunk_query_example_1
description: Maximum of Splunk delivered since last call aggregated on 1 second.
splunkAnalyzeFrequencySeconds: -1
splunkRestApiSearch: 'search=search index=splunk_pruduction  source="/var/log/nginx/splunk-access.log" query_type="GET" ups_status=200 request_path!=/splunk/api/system/ping'
splunkRestApiTimechart: ' | timechart span=1s count by source | sort "/var/log/nginx/splunk-access.log" desc | head 1'
splunkRestApiAdditional: '&id=<sid>'
splunkRestApiEarliestTime: last_call
splunkRestApiLatestTime: now
splunkResultKeyName: /var/log/nginx/splunk-access.log
timeOut: 5
```
 
`sid`: The unique id of that Job.

`description`: A simple description of that Job which describes the content of the metric.

`splunkAnalyzeFrequencySeconds`: The frequency of the request against the Splunk-API of this Job in seconds.  
  If the frequency is lower than the frequency of requests by prometheus, then this Job will run on every request of Prometheus.
  Please use `-1` if you want to run that Job on every request of Prometheus. 

`splunkRestApiSearch`: The Splunk request for logs to generate the metrics from. Please use the ordinary Splunk syntax for the request 
  against the [REST API of Splunk](https://docs.splunk.com/Documentation/Splunk/8.0.1/RESTTUT/RESTsearches). 

`splunkRestApiTimechart`: PrometheusMetricsAPI expects one single and formatted value to get back from the request. A Splunk-timechart 
  can be used to aggregate data. 

`splunkRestApiAdditional`: Additional parameter to use for the request. It is recommended to keep `&id=<sid>`. With that the request gets the
  id of the Job with which you can identify the request within Splunk.

`splunkRestApiEarliestTime`: The start of the time frame for the content to generate the metric from can be defined with that parameter. 
  Please use the Splunk syntax to define the start point (e.g., -24h@d for midnight yesterday). If you want to start from the Startpoint of the last request
  please type `last_call`.
  
`splunkRestApiLatestTime`: The end of the time frame for the content to generate the metric from can be defined with that parameter. 
   Please use the Splunk syntax to define the end point (e.g., -1s@d for 0:00 today morning). If you want just to end now
   simply type `now`. 
   
`splunkResultKeyName`: The result of the search must be a single value for that metric. 
   Please add a key of the resulting value with the proper value for the metric.  
   
`timeOut`: Please add the allowed duration for the request in seconds. After that the request will be timeout. A Job running in a timout response no value for that metric.
    The setting shall be lower than the timeout for Prometheus for the request against this API.
    
#### Block: Settings for simpleAPIRequest
The general setup for simple API requests contains the following parameters: 

```
type: simpleAPIRequest
apiUrlComplete: 
  localTest: http://192.168.101.101/api/storageinfo
  staging: https://tomcatjmx.stage.yotron.de/api/storageinfo
  production: https://tomcatjmx.prod.yotron.de/api/storageinfo
timeOut: 5
```

`apiUrlComplete`: The URLs of the API to collect the unformatted metrics from.

`timeOut`: Please add the allowed duration for the request in seconds. After that the request will be timeout. A Job running in a timout response no value for that metric.
    The setting shall be lower than the timeout for Prometheus for the request against this API.

The setup of every Job has the following parameter:

```
sid: objectory_repositories_Count
description: ObjectsCount count of the repositories available
simpleJsonResultPath: binariesSummary.Count
goCommand: |
  goCommand: if queryEntry=="true" {queryResult = "1"} else {queryResult = "0"}
goCommandImports:
  - strings    
```

`sid`: The unique id of that Job.

`description`: A simple description of that Job which describes the content of the metric.

`simpleJsonResultPath`: The path within the responding json to get the value from. The branches can be 
separated by a point `.`, e.g., `root.count.today`. This value is the entry for the parameter `queryEntry`
 in the formatting code. 

`goCommand` (optional): The value from the request must be formatted in a way Prometheus can read it.
   The formatting must be done by Golang code. A more detailed explanation of that function you find below.
   
`goCommandImports` (optional): If you need Golang imports to get the code run, you can put them in that parameter.

### How to test PrometheusMetricsCollector
For testing the setup you can simply start PrometheusMetricsCollector with the right environment variables and
call in your browser
*http://localhost:9101/metrics/*
depend on you port setup.

Even if no metrics are available by the source API, PrometheusMetricsCollector returns these values:
```
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_hageneral concept behind the PrometheusMetricsCollectorndler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 2
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
```

### How to deploy PrometheusMetricsCollector
To deploy PrometheusMetricsCollector simply 

1. build the project from your hostsystem with

   `go build -o PrometheusMetricsCollector main.go`

2. Deploy the new package `PrometheusMetricsCollector` to your hostsystem (e.g., microservice)
   * with the configuration 
   * the needed setting files
   * the environment variables  
   
3. You can start the project with

   `./PrometheusMetricsCollector`
   
### How to develop goCommands
#### Parameter of the configuration file
In `simpleAPIRequests` it is possible to add Golang code to run the formatting tasks. The Job parameter are
`goCommand` and `goCommandImports`.
 
`goCommandImports` are used to add some Golang Packages to the code simply by naming it.
 
`goCommand` contains the own golang code you must create.

#### Needed parameter
The code has two important parameters `queryEntry` and `queryResult`.

`queryEntry`: Contains the unformatted metric value you get from the source system. Please be aware that this 
  parameter can also ba a json string (please see examples below).

`queryResult`: This variable must stand at the end of the code und provides the formatted value to give back
  to Prometheus. In between you are free to add as much code as you like. 

#### Debug the code
You are able to run the code without this API. In folder `/tmp` you find the generated code which 
you can debug for errors.

This code will be recreated on every request of Prometheus.  

#### some examples of goCommands
##### decimal points 
Prometheus needs decimals point instead of the comma character. For a simple replace
you can use that code:
```
goCommand: |
  queryResult = strings.Replace(queryEntry, ",", "", -1)
goCommandImports:
  - strings
```

##### Format human readable bytes
Prometheus needs float values for bytes representation to show the values 
in a proper way in Grafana. Human readable syntax like '14 MB' is not acceptable. To transform that:  
``` 
goCommand: |
  words := strings.Fields(queryEntry)
  valueFloat, _ := strconv.ParseFloat(words[0], 64)
  switch words[1] {
  case "bytes":
    valueFloat = valueFloat
  case "KB":
    valueFloat = valueFloat * 1000
  case "MB":
    valueFloat = valueFloat * 1000 * 1000
  case "GB":
    valueFloat = valueFloat * 1000 * 1000 * 1000
  case "TB":
    valueFloat = valueFloat * 1000 * 1000 * 1000 * 1000
  }
  queryResult = strconv.FormatFloat(valueFloat, 'f', 0, 64)
goCommandImports:
  - strings
  - strconv
```
##### Analyse a JSON string
You can also analyse a JSON string, e.g. to get the amount of entries in a array. 
```
goCommand: |
  var repos []interface{}
  _ = json.Unmarshal([]byte(queryEntry), &repos)
  queryResult = strconv.Itoa(len(repos))
goCommandImports:
  - strconv
  - encoding/json
```
##### Switch from true/false
Prometheus do not like "true/false" entries. You must switch them to "1/0". 
```
goCommand: |
  goCommand: if queryEntry=="true" {queryResult = "1"} else {queryResult = "0"}
goCommandImports:
  - strings  
```

## Testing area
We provide a simple metricsSimulator for PrometheusMetricsCollector in the folder *./metricsSimulator*.
The raw metrics are stored in file `metricsSimulator/metrics.yml`. 

You can start the simulator with
```
cd metricsSimulator
go run main.go
```
With the start of the simulator, you get the following logs.
```
INFO: 2020/01/21 16:12:15 main.go:26: {"metric1":200,"metric2":555,"metric3":458}
INFO: 2020/01/21 16:12:17 main.go:58: /rawmetrics/
```
In the browser you can grab the data with *http://localhost:8080/rawmetrics*. 
To provide the metrics for Prometheus you must setup the config file for the Collector.

To provide `metric1` for Prometheus, the Collector must be setup that way:
```
localTestStatusCodes:
  type: simpleAPIRequest
  authentication:
    username:
      localTest:
      staging:
      production:
    password:
      localTest:
      staging:
      production:
  apiUrlComplete:
    localTest: http://localhost:8080/rawmetrics
    staging: http://localhost:8080/rawmetrics
    production: http://localhost:8080/rawmetrics
  timeOut: 5
  jobs:
    -
      sid: simpleAPIRequest_in_testing
      description: This is a description of the simpleAPIRequest_in_testing
      simpleJsonResultPath: metric1
```
When you make a request against PrometheusMetricsCollector via the URL *http://localhost:9101/metrics/*,
you get the following result:
```
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP simpleAPIRequest_in_testing This is a description of the simpleAPIRequest_in_testing
# TYPE simpleAPIRequest_in_testing gauge
simpleAPIRequest_in_testing 200
```

Try to grab `metric2` and `metric3` by changing the Collector configuration. Try to manipulate the metrics 
with the `goCommands`.

### own credentials
There are issues, you have questions? Don't hesitate to get in touch.

created by [Joern Kleinbub](https://github.com/joernkleinbub), [YOTRON](http://www.yotron.de), 26.01.2020

Vist me at [LinkedIn](https://www.linkedin.com/in/j%C3%B6rn-kleinbub/) 

Or via EMail <joern.kleinbub@yotron.de>, www.yotron.de