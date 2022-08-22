package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
	log "github.com/lyesteven/go-framework/common/new_log15"
	"io/ioutil"
	"net"
	"runtime/debug"
	"strings"
)

const (
	SwitchServiceName = "config"
	SwitchMethodName  = "switch"
)

type SwitchType int
type OperType int

var (
	innerLog     SwitchType = 1
	retLog       SwitchType = 2
	serviceLevel SwitchType = 3
)

var (
	openOper  OperType = 1
	closeOper OperType = 2
)

type SwitchInfo struct {
	SType SwitchType `json:"switch_type"`
	OType OperType   `json:"oper_type"`
}

//func doSwitchOperation(c *gin.Context) {
func doSwitchOperation(c *fasthttp.RequestCtx) {
	var s SwitchInfo
	if !strings.Contains(string(c.Request.Header.ContentType()), "application/json") {
		c.SuccessString("application/json", `{"code":90001,"message":"header want json format","data":"","attachtype":"nouse","attach":""}`)
		return
	}

	err := json.Unmarshal(c.PostBody(), &s)
	//err := c.BindJSON(&s)
	if err != nil {
		log.Error("doSwitchOperation", "BindJSON", err)
		c.SuccessString("application/json", `{"code":90002,"message":"Post Data is not a json","data":"","attachtype":"nouse","attach":""}`)
		return
	}

	log.Info("doSwitchOperation", "SwitchType", s.SType, "OperType", s.OType)

	errCode, errMsg := 0, "ok"
	switch s.SType {
	case innerLog:
		if s.OType == closeOper {
			innerLogSwitch = false
		} else if s.OType == openOper {
			innerLogSwitch = true
		} else {
			errCode, errMsg = 90003, "oper_type illegal"
		}
	case retLog:
		if s.OType == closeOper {
			retLogSwitch = false
		} else if s.OType == openOper {
			retLogSwitch = true
		} else {
			errCode, errMsg = 90003, "oper_type illegal"
		}
	case serviceLevel:
		updateServiceLevel()

	default:
		errCode, errMsg = 90003, "switch_type illegal"
	}

	retJson := fmt.Sprintf("{\"code\":%d,\"message\":\"%s\",\"data\":\"\",\"attachtype\":\"nouse\",\"attach\":\"\"}", errCode, errMsg)
	c.SuccessString("application/json", retJson)
}

func updateServiceLevel() {

	data, err := ioutil.ReadFile(YamlPath)
	if err == nil {

		m := make(map[string]interface{})
		err = yaml.Unmarshal([]byte(data), &m)

		if err == nil {

			tempMap := make(map[string]TrancLevel)
			if m["TrackList_All"] != nil {
				for _, value := range m["TrackList_All"].([]interface{}) {
					tempMap[value.(string)] = TrancLevel_All
				}
			}

			if m["TrackList_NO"] != nil {
				for _, value := range m["TrackList_NO"].([]interface{}) {
					tempMap[value.(string)] = TrancLevel_NO
				}
			}

			sendList := new([]ConnKVP)
			for key, value := range mapServiceLevel {

				if new, ok := tempMap[key]; ok {
					if new == value {
						continue
					}
				}
				*sendList = append(*sendList, ConnKVP{"", nil, key})
			}

			for key, _ := range tempMap {

				if _, ok := mapServiceLevel[key]; ok {
					continue
				}
				*sendList = append(*sendList, ConnKVP{"", nil, key})
			}

			ch_updateServiceLevel <- &tempMap
			for _, value := range *sendList {
				updateServiceList(value)
			}
		}
	}
}

func httpServer(port int) {

	// 创建路由
	r := fasthttprouter.New()

	//r := gin.Default()
	//r.RedirectFixedPath = true
	// use prometheus metrics exporter middleware.
	//r.Use(ginprom.PromMiddleware(nil))

	r.POST("/:service/:method", doGrpcCallOperation)

	//获取临时token
	r.GET("/token/get_temp_token", getTempToken)

	// Test
	r.GET("/Test/test", testTest)

	r.NotFound = func(c *fasthttp.RequestCtx) {
		printHttpRequest(&c.Request)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		checkErr(err, "httpServer", true)
	}
	go waitToExit(ln)

	log.Info("-----------	start listen ", "port", port)
	if err := fasthttp.Serve(ln, r.Handler); err != nil {
		checkErr(err, "httpServer", true)
	}
}

//func doGrpcCallOperation(c *gin.Context) {
func doGrpcCallOperation(c *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			log.Crit("InvokeServiceByAAS", "server crash: ", err)
			log.Crit("InvokeServiceByAAS", "stack: ", string(debug.Stack()))
		}
	}()

	servicename, methodname := c.UserValue("service").(string), c.UserValue("method").(string)

	if servicename == SwitchServiceName && methodname == SwitchMethodName {
		//conf change
		doSwitchOperation(c)
		return

	} else {
		bodyBytes := c.PostBody()

		// log
		log.Debug("AAS request", "header", c.Request.Header.String())
		log.Debug("AAS request", "body", string(bodyBytes))

		// parse body
		info, warning, err := parseHttpFormatCompat(bodyBytes)
		if err != nil {
			//c.SetContentType("application/json")
			//c.SetStatusCode(http.StatusBadRequest)
			retJson := fmt.Sprintf("{\"code\":90003,\"message\":\"%s\",\"data\":\"\",\"attachtype\":\"nouse\",\"attach\":\"\"}", err.Error())
			//c.SetBody([]byte(retJson))
			c.SuccessString("application/json", retJson)
			//c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Warn: string to int
		if len(warning) != 0 {
			if logLevel != "debug" {
				log.Error("AAS request", "header", c.Request.Header.String())
				log.Error("AAS request", "body", string(bodyBytes))
			}
			log.Error("Json string to int", "keys", strings.Join(warning, ","))
		}

		//info.addParam(servicename, methodname, c.RemoteIP().String())
		//info.addParam(servicename, methodname, c.RemoteIP().String(), string(c.Request.Header.Peek("Clienttype")))
		info.addParam(servicename, methodname, string(c.Request.Header.Peek("Clienttype")))

		rWG.Add(1)
		defer rWG.Done()

		rspStr, err := info.InvokeServiceByAAS()
		if err != nil {
			log.Error("AAS invoke service error:" + err.Error())

			c.SuccessString("application/json", rspStr)
			//c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debug("AAS response", "body", rspStr)
		c.SuccessString("application/json", rspStr)
		//c.String(http.StatusOK, rspStr)
	}
}

//func printHttpRequest(r *http.Request) {
func printHttpRequest(r *fasthttp.Request) {
	if logLevel == "debug" || logLevel == "info" {
		b := bytes.NewBuffer(make([]byte, 0))
		bw := bufio.NewWriter(b)

		if err := r.Write(bw); err == nil {
			bw.Flush()
			log.Info("AAS request", "content", b.String())
		} else {
			log.Error("AAS write to writer error:" + err.Error())
		}
	}

}
