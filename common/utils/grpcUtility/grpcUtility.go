package grpcutil

import (
	"errors"
	//gtrace "github.com/lyesteven/go-framework/common/utils/grpcTracing"
	su "github.com/lyesteven/go-framework/common/service_util"
	"github.com/lyesteven/go-framework/pb/ComMessage"
	//gm "github.com/grpc-ecosystem/go-grpc-middleware"
	//"github.com/opentracing/opentracing-go"
	log "github.com/xiaomi-tc/log15"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	//export Err
	IllegalCBSTypeErr     = errors.New("illegal call business system type")
	NotFoundServiceErr    = errors.New("can't find services")
	NoAvailableServiceErr = errors.New("no available service to invoke")

	//internal Err
	grpcConnErr = errors.New("grpc conn err")
)

type grpcCall struct {

	mapServicesList         map[string]*grpc.ClientConn // target <-> *grpc.ClientConn
	chInquiryServiceConn    chan string
	chInquiryServiceConnRet chan *grpc.ClientConn
	chCloseServiceConn      chan string

	serviceUtil *su.ServiceUtil
	//tracer opentracing.Tracer
}

//func newGrpcCall(trancers ...opentracing.Tracer) *grpcCall {
func newGrpcCall() *grpcCall {
	//var tracer opentracing.Tracer
	//for _, value := range trancers{
	//	tracer = value
	//}

	return &grpcCall{
		mapServicesList:         make(map[string]*grpc.ClientConn),
		chInquiryServiceConn:    make(chan string),
		chInquiryServiceConnRet: make(chan *grpc.ClientConn),
		chCloseServiceConn:      make(chan string),

		serviceUtil: nil,
		//tracer: tracer,
	}
}

func (c *grpcCall) initGrpcCall() {
	c.serviceUtil = su.GetInstance()
	go c.processServiceConnList()
}

func (c *grpcCall) processServiceConnList() {
	for {
		select {
		case key := <-c.chInquiryServiceConn:
			c.chInquiryServiceConnRet <- c.getServiceConnFunc(key)
		case kvp := <-c.chCloseServiceConn:
			c.closeServiceConnFunc(kvp)
		}
	}
}

func (c *grpcCall) getServiceConnFunc(target string) *grpc.ClientConn {
	if conn, ok := c.mapServicesList[target]; ok && conn != nil {
		return conn
	}

	//conn, err := grpc.Dial(target, grpc.WithInsecure(),
	//	grpc.WithUnaryInterceptor(gm.ChainUnaryClient(gtrace.ClientInterceptor(c.tracer))))
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		log.Error("getServiceConnFunc", "target", target, "grpc.Dial", err)
	} else {
		log.Debug("getServiceConnFunc grpc.Dial ok", "target", target)
		c.mapServicesList[target] = conn
	}
	return conn
}

func (c *grpcCall) closeServiceConnFunc(target string) {
	if conn, ok := c.mapServicesList[target]; ok && conn != nil {
		log.Debug("Close grpc conn,target="+target)
		conn.Close()
		delete(c.mapServicesList, target)
	}
}

func (c *grpcCall) getServiceConn(target string) *grpc.ClientConn {
	c.chInquiryServiceConn <- target
	return <-c.chInquiryServiceConnRet
}

func (c *grpcCall) closeServiceConn(target string) {
	c.chCloseServiceConn <- target
}

//
type GrpcUtility struct {
	*grpcCall
}

var (
	grpcSingleton *GrpcUtility
	grpcInstance sync.Once
)

// Deprecated: Use GetGrpcInstance instead.
//func NewGrpcUtility(su *su.ServiceUtil, trancers ...opentracing.Tracer) *GrpcUtility {
func NewGrpcUtility(su *su.ServiceUtil) *GrpcUtility {
	//return GetGrpcInstance(trancers...)
	return GetGrpcInstance()
}

//func GetGrpcInstance(trancers ...opentracing.Tracer) *GrpcUtility {
func GetGrpcInstance() *GrpcUtility {
	grpcInstance.Do(func(){
		grpcSingleton = &GrpcUtility{
			//grpcCall: newGrpcCall(trancers...),
			grpcCall: newGrpcCall(),
		}
		grpcSingleton.init()
	})

	return grpcSingleton
}

func (u *GrpcUtility) init() {
	u.grpcCall.initGrpcCall()
}

func (u *GrpcUtility) doInvoke(request, response interface{}, target, method string, timeout time.Duration) error {
	conn := u.getServiceConn(target)
	if conn == nil {
		log.Info("doInvoke getServiceConn nil", "target", target)
		return grpcConnErr
	}

	ctx := context.Background()
	if Context, ok := log.GetReqContextForGoroutine(); ok{
		ctx = Context
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // terminal routine tree
	return grpc.Invoke(ctx, method, request, response, conn)
}

// Within a data center grpc invoke.
func (u *GrpcUtility) InvokeGrpc(request, response interface{}, packageName, serviceName, methodName string) error {
	listSrvs := u.serviceUtil.GetServiceByNameInternal(serviceName)
	if len(listSrvs) == 0 {
		log.Error("InvokeGrpc can't find services", "serviceName", serviceName)
		return NotFoundServiceErr
	}

	var method string
	if packageName == "" {
		method = serviceName + "/" + methodName
	} else {
		method = packageName + "." + serviceName + "/" + methodName
	}

	for _, value := range listSrvs {
		target := value.IP + ":" + strconv.Itoa(value.Port)
		err := u.doInvoke(request, response, target, method, 10 * time.Second)
		if err != nil {
			log.Error("InvokeGrpc", "serviceName", serviceName, "target", target, "error", err)

			switch {
			case strings.Contains(err.Error(), "context deadline exceeded"):
				//rpc调用超时，不再重试其他节点，由调用方决定是否需要重试
				// TODO 告警上报
			case strings.Contains(err.Error(), "unknown method"):
				//rpc接口未实现，不再重试其他节点
				//TODO 告警上报
			case err == grpcConnErr:
				continue
			case strings.Contains(err.Error(), "unknown service"):
				//通过consul发现的服务与rpc调用发现的结果不一致，即consul中存在过期或者伪造的服务节点
				//从consul服务列表中注销，并向其他节点发起重试
				go u.serviceUtil.UnRegistServiceID(ComMessage.CBSType_Own, value.ServiceID)
				fallthrough
			case strings.Contains(err.Error(), "connection is unavailable"):
				fallthrough
			case strings.Contains(err.Error(), "client connection is closing"):
				fallthrough
			case strings.Contains(err.Error(), "transport is closing"):
				// TODO --这里其实有风险，因为请求可能已经到达服务端进行处理了，对于非幂等接口可能产生重放攻击
				u.closeServiceConn(target)

				u.serviceUtil.ReportFailedService(serviceName, value.ServiceID)

				//链路问题引发的调用失败，向其他节点发起重试
				continue
			}
		}
		return err
	}

	return NoAvailableServiceErr
}

// Cross data center grpc invoke.
func (u *GrpcUtility) InvokeInterSystemGrpc(request, response interface{}, packageName, serviceName, methodName string, cbsType ComMessage.CBSType) error {

	//cbsType is own data center, turn to InvokeGrpc
	if cbsType == ComMessage.CBSType_Own || u.serviceUtil.IsOwnDC(cbsType) {
		return u.InvokeGrpc(request, response, packageName, serviceName, methodName)
	}

	listSrvs := u.serviceUtil.GetServiceByNameCrossSystem(cbsType, serviceName)
	if len(listSrvs) == 0 {
		log.Error("InvokeInterSystemGrpc can't find services", "cbsType", cbsType, "serviceName", serviceName) //ComMessage.CBSType implement String()
		return NotFoundServiceErr
	}

	var method string
	if packageName == "" {
		method = serviceName + "/" + methodName
	} else {
		method = packageName + "." + serviceName + "/" + methodName
	}

	for _, value := range listSrvs {
		target := value.IP + ":" + strconv.Itoa(value.Port)
		err := u.doInvoke(request, response, target, method, 10 * time.Second)
		if err != nil {
			log.Error("InvokeInterSystemGrpc", "cbsType", cbsType, "serviceName", serviceName, "target", target, "error", err)

			switch {
			case strings.Contains(err.Error(), "context deadline exceeded"):
				//rpc调用超时，不再重试其他节点，由调用方决定是否需要重试
				// TODO 告警上报
			case strings.Contains(err.Error(), "unknown method"):
				//rpc接口未实现，不再重试其他节点
				//TODO 告警上报
			case err == grpcConnErr:
				continue
			case strings.Contains(err.Error(), "unknown service"):
				//通过consul发现的服务与rpc调用发现的结果不一致，即consul中存在过期或者伪造的服务节点
				//从consul服务列表中注销，并向其他节点发起重试
				go u.serviceUtil.UnRegistServiceID(cbsType, value.ServiceID)
				fallthrough
			case strings.Contains(err.Error(), "connection is unavailable"):
				fallthrough
			case strings.Contains(err.Error(), "client connection is closing"):
				fallthrough
			case strings.Contains(err.Error(), "transport is closing"):
				//TODO 这里其实有风险，因为请求可能已经到达服务端进行处理了，对于非幂等接口可能产生重放攻击
				u.closeServiceConn(target)

				u.serviceUtil.ReportFailedServiceWithCBSType(cbsType, serviceName, value.ServiceID)

				//链路问题引发的调用失败，向其他节点发起重试
				continue
			}
		}

		return err
	}

	return NoAvailableServiceErr
}

func (u *GrpcUtility) CloseService(serviceName string) {
	listSrvs := u.serviceUtil.GetServiceByNameInternal(serviceName)
	if len(listSrvs) == 0 {
		log.Error("CloseService can't find services", "serviceName", serviceName)
		return
	}
	for _, value := range listSrvs {
		target := value.IP + ":" + strconv.Itoa(value.Port)
		u.closeServiceConnFunc(target)
	}
}
