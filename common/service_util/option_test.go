package service_util

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"gworld/git/GoFrameWork/pb/ComMessage"
)

func TestWithConsulSrvAddr(t *testing.T) {
	su := GetInstanceWithOptions(WithConsulSrvAddr("127.0.0.1:8500"))
	assert.Equal(t, "127.0.0.1:8500", su.consulAddr, "unexpected consul agent addr")
	suSingleton = nil
}

func TestWithServiceName(t *testing.T) {
	su := GetInstanceWithOptions(WithServiceName("Microservice1"))
	assert.Equal(t, "Microservice1", su.serviceName, "unexpected service name")
	suSingleton = nil
}

func TestWithMultiDataCenter(t *testing.T) {
	options := []Option {
		WithMultiDataCenter(ComMessage.CBSType_Base, "base"),
		WithMultiDataCenter(ComMessage.CBSType_HotMall, "hotmall"),
		WithMultiDataCenter(ComMessage.CBSType_GCard, "gcard"),
		WithMultiDataCenter(ComMessage.CBSType_GHit, "ghit"),
	}
	su := GetInstanceWithOptions(options...)
	assert.Containsf(t, su.dcItems, ComMessage.CBSType_Base, "not contain base item")
	assert.Containsf(t, su.dcItems, ComMessage.CBSType_HotMall, "not contain hotmall item")
	assert.Containsf(t, su.dcItems, ComMessage.CBSType_GCard, "not contain gcard item")
	assert.Containsf(t, su.dcItems, ComMessage.CBSType_GHit, "not contain ghit item")

	suSingleton = nil
}
