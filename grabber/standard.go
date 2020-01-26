/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package grabber

import (
	"github.com/yotron/goPrometheusMetricsCollector/common"
	"github.com/yotron/goPrometheusMetricsCollector/common/securewrite"
)

var Conf common.Conf
var PW string
var PersistorLastCall securewrite.Persistor
var PersistorMetric securewrite.Persistor
var Data map[string]interface{}
