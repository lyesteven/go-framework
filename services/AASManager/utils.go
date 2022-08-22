package main

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	log "github.com/lyesteven/go-framework/common/new_log15"
	pb "github.com/lyesteven/go-framework/pb/AASMessage"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func printRequest(in *pb.InvokeServiceRequest) {
	fmt.Printf("inRequest:  %+v\n", in.String())
}

//func waitToExit(server *http.Server) {
func waitToExit(server net.Listener) {
	c := make(chan os.Signal)
	signal.Notify(c)
	for {
		s := <-c
		//log.Debug("waitToExit", "get signal", s)

		if s == syscall.SIGINT || s == syscall.SIGQUIT || s == syscall.SIGTERM {
			log.Error("waitToExit", "get signal", s)
			if server != nil {
				// 关闭监听端口
				if err := server.Close(); err != nil {
					log.Error("Server shutdown error:" + err.Error())
				}
				// 等待所有处理结束，最长等待２秒
				flagChan := make(chan struct{}, 1)
				go func() {
					rWG.Wait()
					flagChan <- struct{}{}
				}()
				select {
				case <-flagChan:
					log.Error("Close after All request closeed")
				case <-time.After(2 * time.Second):
					log.Error(("Close after 2 seconds timeout"))
				}

				Trancclose()
			}
			sumanager.UnRegistService(servId)
			cleanupDone <- true
		}
	}
}

//错误检查
func checkErr(err error, errMessage string, isQuit bool) {
	if err != nil {
		log.Error(errMessage, "error", err)
		if isQuit {
			os.Exit(1)
		}
	}
}

func logInit(logPath, logLevel string) error {
	// log rotate setting
	// maxSize:500M, 10day, backup:100 files, compress
	log.SetRotatePara(500, 10, 100, true)

	// Open Log file
	h, err := log.NetFileHandler(logPath, serviceName, log.LogfmtFormat(), log.WithDstAddr("127.0.0.1:9999"))
	if err != nil {
		log.Error("log.NetFileHandler", "error", err)
		return err
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
		log.SetOutLevel(log.LvlInfo)
	}
	return nil
}

//获取本地配置文件，设置日志路径
func loadLocalConfig(yamlPath string) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigFile(yamlPath) //设置配置文件名
	err := config.ReadInConfig()   // 读配置数据
	if err != nil {
		log.Error("Fatal error config file: " + err.Error())
		return nil, err
	}

	if config.IsSet("LogPath") {
		logPath = config.GetString("LogPath")
		if logPath[len(logPath)-1] != '/' {
			logPath += "/"
		}
		logPath = logPath + "lyesteven/" + serviceName + "/"
		logPath = logPath + serviceName + ".log"
	}

	if config.IsSet("IsCachedService") {
		isCachedService = config.GetBool("IsCachedService")
	}

	if config.IsSet("port") && config.GetInt("port") <= 0 {
		return config, errors.New("yaml get port error")
	}

	return config, err
}

//func loadInitConfig(yamlPath string) (map[string]interface{}, error) {
//	m := make(map[string]interface{})
//
//	data, err := ioutil.ReadFile(yamlPath)
//	checkErr(err, "ReadFile error", false)
//
//	err = yaml.Unmarshal([]byte(data), &m)
//	checkErr(err, "yamlConfig Unmarshal error", true)
//
//	isCachedService = m["IsCachedService"].(bool)
//
//	if m["port"].(int) <= 0 {
//		return m, errors.New("yaml get port error")
//	}
//
//	return m, err
//}
