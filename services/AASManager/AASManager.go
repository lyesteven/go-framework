package main

import (
	"errors"
	"github.com/lyesteven/go-framework/common/bjwt"
	"strconv"
	"strings"
	"time"

	log "github.com/lyesteven/go-framework/common/new_log15"
	gtrace "github.com/lyesteven/go-framework/common/utils/grpcTracing"
	pb "github.com/lyesteven/go-framework/pb/AASMessage"

	gm "github.com/grpc-ecosystem/go-grpc-middleware"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type stHttpFormat struct { // modified by jeffrey
	// 客户端必填字段
	AppKey string //`json:"appKey"`
	AppVer string //`json:"appver"`
	Id     string //`json:"id"`
	Token  string //`json:"token"`
	Data   string //`json:"data"`
	//Sign        string

	// 客户端选填字段
	// usertype:　33333,587643 means do not check token. (such as in Hotmall: 0--普通用户,１--企业用户,33333--游客, 587643--内部系统调用)
	UserType int64 // `json:"usertype"`

	// 服务端设置字段
	// 这里修改为不由客户端上传,而在收包时设置 (httpCompact.go)
	TimeStamp   int64	//`json:"timestamp"`
	// 从Token中解出的ID，用于与ID对比判断
	IdFromToken string
	// 说明: 根据程序逻辑servicename 和 methodname 必须保留，否则程序改动过大
	// servicename he methodname 在httpServer.go 内部重置   by jeffrey
	ServiceName string //`json:"servicename"`
	MethodName  string //`json:"methodname"`

	ClientType string // 从 http 的 Request.Header.Peek("Clienttype") 中取出来,用于识别客户端类型iOS or android
	//SrcIP       string	//`json:"srcip"`
}

func (s *stHttpFormat) InvokeServiceByAAS() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//rootSpan, ctx := gtrace.NewRootSpan(ctx, mapTrancLevel[getTrancingWithServiceName(s.ServiceName)], s.ServiceName)
	//defer rootSpan.Finish()

	if innerLogSwitch {
		log.Info("InvokeServiceByAAS", "request", *s)
	} else {
		log.Info("InvokeServiceByAAS", "ServiceName", s.ServiceName, "MethodName", s.MethodName,
			"AppKey", s.AppKey, "Uuid", s.Id, "ClientType", s.ClientType)
	}

	newToken := ""

	// token的校验
	//if s.UserType != NoNeedCheckToken {	// if UserType==33333, no need check token
	if s.UserType != Tourist && s.UserType != InternalSysCaller { // if UserType==33333,587643 no need check token
		// check token first
		aboutToken, hs, cs, _ := bjwt.CheckToken(s.Token)
		switch aboutToken {
		case 0:
		case 1:
			return Forbidden, errors.New("token无效")
			case 2: //在正常token，但过期了, 这里给token续期
			newToken = bjwt.RefreshToken(hs, cs)
			case 3: //正常的临时token
			//在用户是临时匿名token的情况下，应该由登录业务返回新token
			case 100: //临时token且已过期
			return TSTimeOut, errors.New("临时token无效")
		}
		s.IdFromToken = bjwt.GetId(cs) //将解密后的id设置，给后面的业务使用
		// 传入ID 与IdFromToken　不匹配,　先只打日志。　(可能是测试时的情况)
		if s.IdFromToken != s.Id && s.MethodName != "PHLogin" {
			log.Error("[Id!=IdFromToken]", "HttpFormat", s)
		}
	}

	//server invoke
	ret, err := s.InvokeService(ctx)
	if err != nil {
		log.Error("InvokeServiceByAAS", "InvokeService error", err)
	}

	//吴锦锋增加，寻求最快的不对ret做json unmarshe的办法
	//如果已经生成了新的token, 通过attach字段更新token
	if newToken != "" {
		attach := `,"attachtype":"newtoken","attach":{"token":"` + newToken + `"}`
		ret = ret[:len(ret)-1] + attach + ret[len(ret)-1:]
	} else {
		attach := `,"attachtype":"nouse","attach":""`
		ret = ret[:len(ret)-1] + attach + ret[len(ret)-1:]
	}

	return ret, nil
}

//func (s *stHttpFormat) addParam(serv, method, srcIp, clientType string) {
func (s *stHttpFormat) addParam(serv, method, clientType string) {
	s.ServiceName = serv
	s.MethodName = method
	s.ClientType = clientType
	//s.SrcIP = srcIp
}

func (s *stHttpFormat) checkWhiteListInquiry() bool {
	ch_inquiryWhiteList <- s.ServiceName + s.MethodName
	return <-ch_inquiryWhiteListRet
}

func (s *stHttpFormat) checkTimeout() bool {
	cur := time.Now().Unix()
	if abs(cur-s.TimeStamp) > maxSignTimeout {
		log.Error("isTimeout", "request ts", s.TimeStamp, "host ts", cur, "timeout", maxSignTimeout)
		return true
	}
	return false
}

func (s *stHttpFormat) InvokeService(ctx context.Context) (string, error) {
	var ret string
	var err error

	//inquiry service
	listSrvs := sumanager.GetServiceByNameInternal(s.ServiceName)
	if len(listSrvs) == 0 {
		ret = NotFound
		return ret, NotFoundErr
	}

	for _, value := range listSrvs {
		params := new(InvokeServiceParams)
		params.paramPush(*s)
		params.target = value.IP + ":" + strconv.Itoa(value.Port)
		params.method = value.PackageName + "." + value.ServiceName + "/" + s.MethodName
		params.ctx = ctx
		if isCachedService {
			ret, err = InvokeRemoteServiceWithConnectionCache(params)
		} else {
			ret, err = InvokeRemoteService(params)
		}

		if err == nil {
			break
		}

		//report failed service to su
		sumanager.ReportFailedService(s.ServiceName, value.ServiceID)
		if ret == RetryExceedError || ret == NotFound {
			continue
		}
		break
	}

	if retLogSwitch {
		log.Info("InvokeService", "InvokeRemoteService return", ret)
	}

	return ret, err
}

func getNewestTarget(serviceName string) string {

	tmpInfo, err := sumanager.GetServiceListOwnDC(serviceName, false)
	if err != nil || len(tmpInfo) == 0 {
		return ""
	}
	log.Info("getNewestTarget result", "tmpInfo", tmpInfo)

	for _, value := range tmpInfo {

		if value.IP != "" && value.Port > 0 {
			return value.IP + ":" + strconv.Itoa(value.Port)
		}
	}

	return ""
}

func getTrancingWithServiceName(serviceName string) TrancLevel {
	ch_inquiryServiceLevel <- serviceName
	return <-ch_inquiryServiceLevelRet
}

func InvokeRemoteServiceWithConnectionCache(params *InvokeServiceParams) (string, error) {
	var err error

	//get grpc connection from cache
	conn := getServiceList(params.target)
	if conn == nil {
		//conn, err = grpc.Dial(params.target, grpc.WithInsecure(), gtrace.ClientDialOption(mapTrancLevel[getTrancingWithServiceName(params.req.ServiceName)].Tracer))
		conn, err = grpc.Dial(params.target,
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(gm.ChainUnaryClient(gtrace.ClientInterceptor(mapTrancLevel[getTrancingWithServiceName(params.req.ServiceName)].Tracer))))
		if err != nil {
			log.Error("InvokeRemoteServiceWithConnectionCache", "did not connect", err)
			return NotFound, NotFoundErr
		}
		kvp := ConnKVP{params.target, conn, params.req.ServiceName}
		updateServiceList(kvp)
	}

	out := new(pb.InvokeServiceResponse)

	//ctx := getContextWithReqID()
	for retry := 0; ; retry++ {
		//err = grpc.Invoke(params.ctx, params.method, params.req, out, conn)
		err = conn.Invoke(params.ctx, params.method, params.req, out)
		if err == nil {
			break
		} else {
			log.Error("InvokeRemoteServiceWithConnectionCache grpc invoke error", "target", params.target, "method", params.method, "error", err)
			conn.Close()
			kvp := ConnKVP{params.target, nil, params.req.ServiceName}
			updateServiceList(kvp)

			if strings.Contains(err.Error(), "context deadline exceeded") {
				return TimeOutError, TimeOutErr //超时
			}

			if strings.Contains(err.Error(), "connection is unavailable") || strings.Contains(err.Error(), "the client connection is closing") || strings.Contains(err.Error(), "transport is closing") {
				if retry < 3 {

					conn, err = grpc.Dial(params.target, grpc.WithInsecure(), grpc.WithUnaryInterceptor(gm.ChainUnaryClient(gtrace.ClientInterceptor(mapTrancLevel[getTrancingWithServiceName(params.req.ServiceName)].Tracer))))
					if err != nil {
						log.Error("InvokeRemoteServiceWithConnectionCache", "did not connect", err)
						return NotFound, NotFoundErr
					}
					kvp := ConnKVP{params.target, conn, params.req.ServiceName}
					updateServiceList(kvp)
					continue
				}
				return RetryExceedError, RetryExceedErr
			}

			if strings.Contains(err.Error(), "all SubConns are in TransientFailure") {
				if retry < 3 {
					time.Sleep(2 * time.Second)
					tmpTarget := getNewestTarget(params.req.ServiceName)
					if "" == tmpTarget {
						log.Info("tmpTarget is null")
						tmpTarget = params.target
					}

					conn, err = grpc.Dial(tmpTarget, grpc.WithInsecure(), grpc.WithUnaryInterceptor(gm.ChainUnaryClient(gtrace.ClientInterceptor(mapTrancLevel[getTrancingWithServiceName(params.req.ServiceName)].Tracer))))
					if err != nil {
						log.Error("InvokeRemoteServiceWithConnectionCache", "did not connect", err)
						return NotFound, NotFoundErr
					}
					kvp := ConnKVP{tmpTarget, conn, params.req.ServiceName}
					updateServiceList(kvp)
					continue
				}
				return RetryExceedError, RetryExceedErr
			}

			if strings.Contains(err.Error(), "unknown service") {
				return NotFound, NotFoundErr
			}

			return InternalServerError, InternalServerErr
		}
	}

	return out.ServiceResponseData, nil
}

func InvokeRemoteService(params *InvokeServiceParams) (string, error) {
	var ret string

	conn, err := grpc.Dial(params.target, grpc.WithInsecure(), grpc.WithUnaryInterceptor(gm.ChainUnaryClient(gtrace.ClientInterceptor(mapTrancLevel[getTrancingWithServiceName(params.req.ServiceName)].Tracer))))
	if err != nil {
		log.Error("InvokeRemoteService grpc dial error", "error", err)
		ret = NotFound
		return ret, NotFoundErr
	}
	defer conn.Close()

	out := new(pb.InvokeServiceResponse)

	//err = grpc.Invoke( /*getContextWithReqID()*/ params.ctx, params.method, params.req, out, conn)
	err = conn.Invoke(params.ctx, params.method, params.req, out)
	if err != nil {
		log.Error("InvokeRemoteService grpc Invoke error", "target", params.target, "method", params.method, "error", err)
		ret = InternalServerError
		return ret, InternalServerErr
	}

	return out.ServiceResponseData, nil
}

func getContextWithReqID() context.Context {

	var reqidString string
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	reqidString, ctx = gtrace.NewAndSetReqIDToCtx(ctx)
	log.Info(" --- begin aas call", "reqidString", reqidString)
	return ctx
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0 // return correctly abs(-0)
	}
	return x
}

// InvokeServiceParams
type InvokeServiceParams struct {
	req    *pb.InvokeServiceRequest
	target string
	method string
	ctx    context.Context
}

func (s *InvokeServiceParams) paramPush(in stHttpFormat) {
	// annotated and modified by jeffrey
	temp := new(pb.InvokeServiceRequest)
	temp.ServiceName = in.ServiceName
	temp.MethodName = in.MethodName
	temp.AppKey = in.AppKey
	temp.AppVer = in.AppVer // added
	temp.Id = in.Id
	temp.Token = in.Token
	temp.UserType = in.UserType
	temp.Data = in.Data
	temp.IdFromToken = in.IdFromToken

	//temp.Sign = in.Sign
	//temp.TimeStamp = in.TimeStamp
	//temp.SrcIP = in.SrcIP

	s.req = temp
}
