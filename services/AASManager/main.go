package main

import (
	"flag"
	"fmt"
	log "github.com/lyesteven/go-framework/common/new_log15"
	su "github.com/lyesteven/go-framework/common/service_util"
	gtrace "github.com/lyesteven/go-framework/common/utils/grpcTracing"
	"github.com/spf13/viper"
	"runtime"
	"runtime/debug"
	"sync"
)

var serviceName, packageName = "AASManager", "AASManager"
var (
	logPath, logLevel, YamlPath string
	sumanager *su.ServiceUtil
	servId string = ""
	rWG sync.WaitGroup
	cleanupDone chan bool
	LocalConfig *viper.Viper
	NoCheckTokenSvcsMap = make(map[string]int)
)

func init() {
	// 一个服务多个配置的情况有待处理
	var project_conf string
	if runtime.GOOS == "windows" {
		project_conf = "c:\\etc\\lyesteven\\go-framework\\modules\\" + serviceName + "\\" + serviceName + ".yaml"
		logPath = "c:\\log\\lyesteven\\go-framework\\" + serviceName + "\\"
	} else {
		project_conf = "/etc/lyesteven/go-framework/modules/" + serviceName + "/" + serviceName + ".yaml"
		logPath = "/data/logs/lyesteven/go-framework/" + serviceName + "/"
	}

	flag.StringVar(&YamlPath, "c", project_conf, "config_file path,example /etc/lyesteven/go-framework/AASManager.yaml")
	flag.StringVar(&logLevel, "l", "info", "log level: debug, info, error")

	logPath = logPath + serviceName + ".log"
}

func Trancclose() {
	for _, value := range mapTrancLevel {
		_ = value.Closer.Close()
	}
}

// JaegerTracerInit func JaegerTracerInit(mapYaml map[string]interface{}) error {
func JaegerTracerInit() error {

	//trancingAll, closerAll, err := gtrace.NewJaegerTracerWithPlatformPrefix(serviceName, mapYaml["JagentHost"].(string), gtrace.WithSamplePro(1))
	trancingAll, closerAll, err := gtrace.NewJaegerTracerWithPlatformPrefix(serviceName, LocalConfig.GetString("JagentHost"), gtrace.WithSamplePro(1))
	if err != nil {
		return err
	}
	mapTrancLevel[TrancLevel_All] = TrancInfo{trancingAll, closerAll}

	var traceSampleDefault = 0.1
	//if sample, ok := mapYaml["TraceSampleDefault"]; ok {
	//	traceSampleDefault = sample.(float64)
	//}
	if LocalConfig.IsSet("TraceSampleDefault") {
		traceSampleDefault = LocalConfig.GetFloat64("TraceSampleDefault")
	}

	//trancingDefault, closerDefault, err := gtrace.NewJaegerTracerWithPlatformPrefix(serviceName, mapYaml["JagentHost"].(string), gtrace.WithSamplePro(traceSampleDefault))
	trancingDefault, closerDefault, err := gtrace.NewJaegerTracerWithPlatformPrefix(serviceName, LocalConfig.GetString("JagentHost"), gtrace.WithSamplePro(traceSampleDefault))
	if err != nil {
		return err
	}
	mapTrancLevel[TrancLevel_Default] = TrancInfo{trancingDefault, closerDefault}

	//trancingNo, closerNo, err := gtrace.NewJaegerTracerWithPlatformPrefix(serviceName, mapYaml["JagentHost"].(string), gtrace.WithSamplePro(0))
	trancingNo, closerNo, err := gtrace.NewJaegerTracerWithPlatformPrefix(serviceName, LocalConfig.GetString("JagentHost"), gtrace.WithSamplePro(0))
	if err != nil {
		return err
	}
	mapTrancLevel[TrancLevel_NO] = TrancInfo{trancingNo, closerNo}

	//if mapYaml["TrackList_All"] != nil {
	//	for _, value := range mapYaml["TrackList_All"].([]interface{}) {
	//		mapServiceLevel[value.(string)] = TrancLevel_All
	//	}
	//}
	tracklist_all := LocalConfig.GetStringSlice("TrackList_All")
	for _, value := range tracklist_all {
		mapServiceLevel[value] = TrancLevel_All
	}

	//if mapYaml["TrackList_NO"] != nil {
	//	for _, value := range mapYaml["TrackList_NO"].([]interface{}) {
	//		mapServiceLevel[value.(string)] = TrancLevel_NO
	//	}
	//}
	tracklist_no := LocalConfig.GetStringSlice("TrackList_NO")
	for _, value := range tracklist_no {
		mapServiceLevel[value] = TrancLevel_NO
	}

	return nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Crit("main", "server crash: ", err)
			log.Crit("main", "stack: ", string(debug.Stack()))
		}
	}()

	flag.Parse()

	// 读本地配置文件
	var err error
	LocalConfig, err = loadLocalConfig(YamlPath)
	if err != nil {
		checkErr(err, "loadLocalConfig error", true)
	}

	// 打开日志文件，设置日志级别 (可以开始在日志文件中输出日志)
	if LocalConfig.IsSet("LogLevel") {
		logLevel = LocalConfig.GetString("LogLevel")
	}
	err = logInit(logPath, logLevel)
	fmt.Println("logPaht=", logPath)
	if err != nil {
		checkErr(err, "logInit error", true)
	}

	// 加载不检查 token 的服务白名单
	initNoCheckTokenSvcs()

	//get consul instance
	sumanager, _ = su.NewServiceUtil(serviceName, YamlPath)
	servId, err = sumanager.RegistService(serviceName, packageName)
	checkErr(err, "registerToConsul", true)

	log.Info("", "version", su.GetVersion())

	//trancing init
	//err = JaegerTracerInit(m)
	err = JaegerTracerInit()
	if err != nil {
		checkErr(err, "JaegerTracerInit error", true)
	}

	//white manager
	go processWhiteList()
	InitWhiteList()
	//grpc connect manager
	go processServiceList()
	//service level manager
	go processServiceLevel()

	//http server
	go httpServer(LocalConfig.GetInt("port"))
	cleanupDone = make(chan bool)
	<-cleanupDone
}
