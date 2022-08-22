package main

import (
	pbng "github.com/lyesteven/go-framework/pb/NodeAgent"
	log "github.com/xiaomi-tc/log15"
	"golang.org/x/net/context"
)

//定义server，实现pb中接口
type server struct{}

func (s *server) Node_Agent_GetSingleServiceList(ctx context.Context, in *pbng.SingleServiceListReq) (*pbng.ServicesList, error) {

	stResponse := new(pbng.ServicesList)
	stResponse.ServiceName = in.ServiceName
	bFlag := structGetSingleServiceList(in, stResponse)
	if bFlag == false {
		log.Error("structGetSingleServiceList error")
	}

	return stResponse, nil
}

func (s *server) Node_Agent_GetServicesChangeList(ctx context.Context, in *pbng.ServicesChangeListReq) (*pbng.ServicesChangeListRsp, error) {

	stResponse := new(pbng.ServicesChangeListRsp)

	bFlag := structGetServicesChangeList(in, stResponse)
	if bFlag == false {
		log.Error("structGetServicesChangeList error")
	}

	return stResponse, nil
}
