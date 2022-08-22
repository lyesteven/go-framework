package main

import (
	"errors"
)

var innerLogSwitch = false
var retLogSwitch = false
var isCachedService = false
var InternalServerErr = errors.New(InternalServerError)
var NotFoundErr = errors.New(NotFound)
var TimeOutErr = errors.New(TimeOutError)
var RetryExceedErr = errors.New(RetryExceedError)

const (
	OK                  = "0"
	//TSTimeOut           = "{\"code\":900011,\"Desc\":\"TS过期\"}"
	//Unauthorized        = "{\"Code\":990002,\"Desc\":\"未授权\"}"
	//Forbidden           = "{\"Code\":990003,\"Desc\":\"签名验证无效\"}"
	//NotFound            = "{\"Code\":990004,\"Desc\":\"未找到服务\"}"
	//InternalServerError = "{\"Code\":990005,\"Desc\":\"内部错误\"}"
	//TimeOutError        = "{\"Code\":990006,\"Desc\":\"调用超时\"}"
	//RetryExceedError    = "{\"Code\":990007,\"Desc\":\"超过重试次数\"}"
	//TokenError          = "{\"Code\":990008,\"Desc\":\"token过期\"}"

	TSTimeOut           = `{"code":900011,"message":"临时token无效","data":"","attachtype":"nouse","attach":""}`
	Unauthorized        = `{"code":900012,"message":"未授权","data":"","attachtype":"nouse","attach":""}"`
	//Forbidden           = `{"code":900013,"message":"签名验证无效","data":"","attachtype":"nouse","attach":""}`
	Forbidden           = `{"code":1001005,"message":"token无效","data":"","attachtype":"nouse","attach":""}`
	NotFound            = `{"code":900014,"message":"未找到服务","data":"","attachtype":"nouse","attach":""}`
	InternalServerError = `{"code":900015,"message":"内部错误","data":"","attachtype":"nouse","attach":""}`
	TimeOutError        = `{"code":900016,"message":"调用超时","data":"","attachtype":"nouse","attach":""}`
	RetryExceedError    = `{"code":900017,"message":"超过重试次数","data":"","attachtype":"nouse","attach":""}`
	TokenError          = `{"code":900018,"message":"token过期","data":"","attachtype":"nouse","attach":""}`

	Tourist = 33333
	InternalSysCaller = 587643
)

const (
	appSecret            = "a323f9b6-1f04-420e-adb9-b06d142c5e63"
	tokenAesKey          = "chuanfa@sina.com"
	split                = "`|`"
	splitCount     int   = 3
	maxSignTimeout int64 = 10 * 60
)

type AppReturnJSON struct {
	Code int 			`json:"code"`
	Desc string 		`json:"desc"`
	Data interface{} 	`json:"data"`
	//Attach interface{} 	`json:"attach"`
}
