package main

import (
	"flag"
	su "github.com/lyesteven/go-framework/common/service_util"
	pbng "github.com/lyesteven/go-framework/pb/NodeAgent"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/xiaomi-tc/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"syscall"
)

//
//	TODO: 1 modify to your ServiceName and PackageName First !!!
//

var serviceName, packageName = "NodeAgent", "NodeAgent"

var logPath, logLevel, yamlPath string

//var serviceId string
var srvPort int
var sumanager *su.ServiceUtil
var listen net.Listener

func waitToExit(srvId string) {
	c := make(chan os.Signal)
	signal.Notify(c)
	for {
		s := <-c
		log.Error("waitToExit", "get signal", s)
		if s == syscall.SIGINT || s == syscall.SIGQUIT || s == syscall.SIGTERM {
			sumanager.UnRegistService(srvId)
			listen.Close()
			os.Exit(1)
		}
	}
}

//错误检查
func checkErr(err error, errMessage string, isQuit bool) {
	if err != nil {
		log.Error(errMessage, "error", err)
		if isQuit {
			if listen != nil {
				listen.Close()
			}
			os.Exit(1)
		}
	}
}

//获取配置文件
func loadInitConfig(yamlPath string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	data, err := ioutil.ReadFile(yamlPath)
	checkErr(err, "ReadFile error", true)

	err = yaml.Unmarshal([]byte(data), &m)
	checkErr(err, "yamlConfig Unmarshal error", true)

	if runtime.GOOS == "windows" {
		logPath = "c:/log/lyesteven/go-framework/" + serviceName + "/"
	} else {
		logPath = "/data/logs/lyesteven/go-framework/" + serviceName + "/"
	}

	logPath = logPath + serviceName + ".log"

	return m, err
}

//向consul注册服务本身
func registerToConsul(svName, pgName string) string {
	svId, err := sumanager.RegistService(svName, pgName)

	if err != nil {
		checkErr(err, "registerToConsul", true)
	} else {
		checkErr(err, "registerToConsul", false)
	}

	return svId
}

func init() {
	var projectConf string
	if runtime.GOOS == "windows" {
		projectConf = "c:\\etc\\lyesteven\\go-framework\\modules\\" + serviceName + "\\" + serviceName + ".yaml"
	} else {
		projectConf = "/etc/lyesteven/go-framework/modules/" + serviceName + "/" + serviceName + ".yaml"
	}
	flag.StringVar(&yamlPath, "c", projectConf, "config_file path,example /etc/lyesteven/go-framework/NodeAgent.yaml")
	flag.StringVar(&logLevel, "l", "debug", "log level: debug, info, error")
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Crit("main", "server crash: ", err)
			log.Crit("main", "stack: ", string(debug.Stack()))
		}
	}()

	flag.Parse()

	_, err := loadInitConfig(yamlPath)
	if err != nil {
		checkErr(err, "loadInitConfig error", true)
	}

	sumanager, srvPort = su.NewServiceUtil(serviceName, yamlPath)
	serviceId := registerToConsul(serviceName, packageName)

	//tracer, closer, err := gtrace.NewJaegerTracer(serviceName, "127.0.0.1:6831", gtrace.WithSamplePro(1))
	//if err != nil {
	//	checkErr(err, "NewJaegerTracer error", true)
	//}
	//defer closer.Close()

	//监听信号，退出时，反向注销服务
	go waitToExit(serviceId)

	// Open log file
	//h, _ := log.FileHandler(logPath, log.LogfmtFormat())
	h, err := log.NetFileHandler(logPath, serviceName, log.LogfmtFormat(), log.WithDstAddr("127.0.0.1:9999"))
	if err != nil {
		log.Error("log.NetFileHandler", "error", err)
		return
	}
	log.Root().SetHandler(h)

	switch logLevel {
	case "debug":
		log.SetOutLevel(log.LvlDebug)
	case "info":
		log.SetOutLevel(log.LvlInfo)
	case "error":
		log.SetOutLevel(log.LvlError)
	default:
		log.SetOutLevel(log.LvlDebug)
	}
	log.Debug("start...", "serviceid", serviceId, "port", srvPort)

	listen, err = net.Listen("tcp", ":"+strconv.Itoa(srvPort))
	if err != nil {
		log.Error("tcp port error", "port", srvPort)
	}

	s := grpc.NewServer(
		//数据上报
		//grpc.UnaryInterceptor(gm.ChainUnaryServer(gtrace.ServerInterceptor(tracer), grpc_prometheus.UnaryServerInterceptor)),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	//  注册服务
	pbng.RegisterNodeAgentServer(s, &server{})

	//数据上报
	grpc_prometheus.Register(s)
	grpc_prometheus.EnableHandlingTimeHistogram()

	reflection.Register(s)
	if err := s.Serve(listen); err != nil {
		log.Error("ServiceStartFailed", "error", err)
	}
}
