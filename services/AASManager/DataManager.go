package main

import (
	"io"
	"runtime/debug"
	"sync"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	log "github.com/lyesteven/go-framework/common/new_log15"
)

var (
	//gateway white list
	map_whiteList          = make(map[string]string)
	ch_inquiryWhiteList    = make(chan string)
	ch_inquiryWhiteListRet = make(chan bool)
	ch_newWhiteList        = make(chan *map[string]string)

	//services grpc conn store
	map_servicesList          = make(map[string]ConnKVP)
	rwLock_servicesList       sync.RWMutex
	ch_inquiryServiceConn     = make(chan string)
	ch_inquiryServiceConntRet = make(chan *grpc.ClientConn)
	ch_updateServiceConn      = make(chan ConnKVP)

	//service trancing level
	mapServiceLevel           = make(map[string]TrancLevel)
	ch_inquiryServiceLevel    = make(chan string)
	ch_inquiryServiceLevelRet = make(chan TrancLevel)
	ch_updateServiceLevel     = make(chan *map[string]TrancLevel)

	mapTrancLevel = make(map[TrancLevel]TrancInfo)
)

type TrancInfo struct {
	opentracing.Tracer
	io.Closer
}

type TrancLevel int

const (
	_ TrancLevel = iota
	TrancLevel_Default
	TrancLevel_All
	TrancLevel_NO
)

type ConnKVP struct {
	key         string
	value       *grpc.ClientConn
	servicename string
}

func InitWhiteList() {
	//ch_newWhiteList <- su.GetWhiteList
}

//manager white list
func processWhiteList() {
	defer func() {
		if err := recover(); err != nil {
			log.Crit("processWhiteList", "server crash: ", err)
			log.Crit("processWhiteList", "stack: ", string(debug.Stack()))
		}
	}()

	for {
		select {
		case key := <-ch_inquiryWhiteList:
			ret := isInWhiteList(key)
			ch_inquiryWhiteListRet <- ret
		case list := <-ch_newWhiteList:
			map_whiteList = *list
		}
	}
}

func isInWhiteList(key string) bool {
	if _, ok := map_whiteList[key]; ok {
		return true
	}
	return false
}

func getServiceList(target string) *grpc.ClientConn {
	//ch_inquiryServiceConn <- target
	//return <-ch_inquiryServiceConntRet

	return getServiceListFunc(target)
}

func updateServiceList(kvp ConnKVP) {
	ch_updateServiceConn <- kvp
}

//manager service grpc connection
func processServiceList() {
	defer func() {
		if err := recover(); err != nil {
			log.Crit("processServiceList", "server crash: ", err)
			log.Crit("processServiceList", "stack: ", string(debug.Stack()))
		}
	}()

	for {
		select {
		case key := <-ch_inquiryServiceConn:
			ret := getServiceListFunc(key)
			ch_inquiryServiceConntRet <- ret
		case kvp := <-ch_updateServiceConn:
			updateServiceListFunc(kvp)
		}
	}
}

func updateServiceListFunc(kvp ConnKVP) {
	if len(kvp.key) > 0 {
		rwLock_servicesList.Lock()
		//update with target info
		if value, ok := map_servicesList[kvp.key]; ok {
			if value.value != nil {
				value.value.Close()
			}
		}
		map_servicesList[kvp.key] = kvp

		rwLock_servicesList.Unlock()
	} else {
		//delete with service name
		if len(kvp.servicename) > 0 && kvp.value == nil {
			rwLock_servicesList.Lock()

			for key, value := range map_servicesList {
				if value.servicename == kvp.servicename {

					if value.value != nil {
						value.value.Close()
					}
					delete(map_servicesList, key)
				}
			}
			rwLock_servicesList.Unlock()
		}
	}
}

func getServiceListFunc(key string) *grpc.ClientConn {
	rwLock_servicesList.RLock()
	defer rwLock_servicesList.RUnlock()

	value, ok := map_servicesList[key]
	if ok {
		return value.value
	} else {
		return nil
	}
}

//manager service grpc connection
func processServiceLevel() {

	defer func() {
		if err := recover(); err != nil {
			log.Crit("processServiceLevel", "server crash: ", err)
			log.Crit("processServiceLevel", "stack: ", string(debug.Stack()))
		}
	}()

	for {
		select {
		case servicename := <-ch_inquiryServiceLevel:
			ch_inquiryServiceLevelRet <- getServiceLevelFunc(servicename)
		case mapTemp := <-ch_updateServiceLevel:
			mapServiceLevel = *mapTemp
		}
	}
}

func getServiceLevelFunc(servicename string) TrancLevel {

	tmpLevel := TrancLevel_Default
	if level, ok := mapServiceLevel[servicename]; ok {
		tmpLevel = level
	}
	return tmpLevel
}
