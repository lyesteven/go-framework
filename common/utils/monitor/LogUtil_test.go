package monitor

import (
	"github.com/lyesteven/go-framework/pb/ComMessage"
	"testing"
)

func TestMonitorBriefly(t *testing.T) {
	InitCfg(ComMessage.CBSType_Woda, "TestMonitorBriefly")
	err := MonitorBriefly("test_metric", "test_method", Counter, 1)
	if err != nil {
		t.Failed()
	}
}

func TestMonitor(t *testing.T) {
	InitCfg(ComMessage.CBSType_Base, "TestMonitor")
	err := Monitor("test_metric", "test_method", Counter, 1, "This is a test", "!db", "100000", 0, "0", "")
	if err != nil {
		t.Failed()
	}
}

func TestMonitorS(t *testing.T) {
	InitCfg(ComMessage.CBSType_Woda, "TestMonitorS")
	var m = Metric{
		Metric:     "test_metric",
		Method:     "test_method",
		MetricType: Counter.String(),
		Value:      1,
	}
	err := MonitorS(&m)
	if err != nil {
		t.Failed()
	}
}
