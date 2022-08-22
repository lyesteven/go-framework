package main

import (
	"github.com/lyesteven/go-framework/pb/ComMessage"
	pbng "github.com/lyesteven/go-framework/pb/NodeAgent"
	log "github.com/xiaomi-tc/log15"
)

func structGetSingleServiceList(in *pbng.SingleServiceListReq, out *pbng.ServicesList) bool {
	log.Info("structGetSingleServiceList in", "in", in)
	if in.CbsType != ComMessage.CBSType_Own && !sumanager.HasSDTask(in.CbsType, in.ServiceName) {
		//获取其他dc上的服务列表
		items, version, err := sumanager.GetServiceListOtherDC(in.CbsType, in.ServiceName, true)
		if err != nil {
			log.Error("structGetSingleServiceList", "GetServiceListOtherDC error", err)
			return false
		}

		out.ServiceVersion = version
		for _, value := range items {
			stInfo := new(pbng.ServiceInfo)
			stInfo.ServiceID = value.ServiceID
			stInfo.ServiceIP = value.IP
			stInfo.ServicePort = int64(value.Port)
			stInfo.ServiceStatus = value.Status
			stInfo.PackageName = value.PackageName
			out.Info = append(out.Info, stInfo)
		}
	} else {
		version, list := sumanager.GetServiceInfoByServiceName(in.CbsType, in.ServiceName)
		if len(list) > 0 {
			out.ServiceVersion = version
			for _, value := range list {
				stInfo := &pbng.ServiceInfo{
					ServiceID:     value.ServiceID,
					ServiceIP:     value.IP,
					ServicePort:   int64(value.Port),
					ServiceStatus: value.Status,
					PackageName:   value.PackageName,
				}
				out.Info = append(out.Info, stInfo)
			}
		} else {
			log.Info("structGetSingleServiceList GetServiceInfoByServiceName result error", "cbsType", in.CbsType, "ServiceName", in.ServiceName)
		}
	}

	log.Info("structGetSingleServiceList out", "out", out)
	return true
}

func structGetServicesChangeList(in *pbng.ServicesChangeListReq, out *pbng.ServicesChangeListRsp) bool {

	log.Info("-----structGetServicesChangeList", "in", in)
	if len(in.CheckList) > 0 {
		for _, value := range in.CheckList {
			version, list := sumanager.GetServiceInfoByServiceName(value.CbsType, value.ServiceName)
			if value.ServiceVersion != version {
				log.Info("###########service name version change", "CbsType", value.CbsType, "service name", value.ServiceName)
				stNode := new(pbng.ServicesList)
				stNode.CbsType = value.CbsType
				stNode.ServiceName = value.ServiceName
				stNode.ServiceVersion = version

				for _, ser := range list {
					stInfo := new(pbng.ServiceInfo)
					stInfo.ServiceID = ser.ServiceID
					stInfo.ServiceIP = ser.IP
					stInfo.ServicePort = int64(ser.Port)
					stInfo.ServiceStatus = ser.Status
					stInfo.PackageName = ser.PackageName
					stNode.Info = append(stNode.Info, stInfo)
				}

				out.ChangeList = append(out.ChangeList, stNode)
			}
		}
	}

	log.Info("structGetServicesChangeList", "out", out)
	return true
}
