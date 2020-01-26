package common

import (
	. "github.com/yotron/goConfigurableLogger"
	"os"
)

var Cluster string
var ExportType string

func init() {
	Cluster = os.Getenv("YOPRO_ENVIRONMENT")
	if Cluster == "" {
		Error.Println("No variable YOPRO_ENVIRONMENT set.")
		panic("No variable YOPRO_ENVIRONMENT set")
	}
	ExportType = os.Getenv("YOPRO_COLLECTOR")
	if ExportType == "" {
		Error.Println("No variable YOPRO_COLLECTOR set.")
		panic("No variable YOPRO_COLLECTOR set.")
	}
}
