package service_util

import (
	"bufio"
	"github.com/spf13/viper"
	log "github.com/xiaomi-tc/log15"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
)

//type map_Config map[string]map[string]interface{}
var (
	strGlobalConfigPath = "/etc/go-framework/GlobalConfigure.yaml"
	strYamlRootPath     = "/etc/go-framework/modules/"

	vipConf	*viper.Viper
)

// Deprecated:
//func GetFileName(fullPath string) string {
//	_, filenameWithSuffix := path.Split(fullPath)
//
//	var fileSuffix string
//	fileSuffix = path.Ext(filenameWithSuffix)
//
//	var filenameOnly string
//	filenameOnly = strings.TrimSuffix(filenameWithSuffix, fileSuffix)
//
//	port, err := getServiceConfigPort(fullPath)
//	if err != nil {
//		port = 0
//	}
//
//	filenameOnly += "-" + strconv.Itoa(port)
//
//	return filenameOnly
//}

func getServiceConfigPort(yamlpath string) (port int, err error) {
	config := viper.New()
	config.SetConfigFile(yamlpath)  //设置配置文件名
	err = config.ReadInConfig()     // 读配置数据
	if err != nil {
		log.Error("\"Fatal error config file: " + err.Error())
		return 0, err
	}
	port = config.GetInt("port")

	return
}

func setServiceConfigPort(yamlpath string, port int) (err error) {

	file, err:= os.OpenFile(yamlpath, os.O_RDWR, 0666)
	if err != nil {
		log.Error("open config file fail","err",err.Error())
		return err
	}
	//defer file.Close()

	isContainPort := false
	buf := bufio.NewReader(file)
	output := make([]byte, 0)
	for {
		line, _, c := buf.ReadLine()
		if c == io.EOF {
			break
		}
		if strings.Contains(string(line),"port:"){
			newline := "port: " + strconv.Itoa(port)
			line = []byte(newline)
			isContainPort =  true
		}

		output = append(output, line...)
		output = append(output, []byte("\n")...)
	}
	file.Close()

	// 原来没有包含port, 加一个port项
	if !isContainPort {
		newline := "port: " + strconv.Itoa(port)
		line := []byte(newline)
		output = append(output, line...)
		output = append(output, []byte("\n")...)
	}

	if err := writeToFile(yamlpath ,output);err != nil{
		log.Error("write config file err","err",err.Error())
		return err
	}
	return nil
}

//func GetConfig(confItem string) (result map[string]interface{}, err error) {
//	if mapGlobalConfig != nil {
//		if result, err := mapGlobalConfig[confItem]; err {
//			return result, nil
//		}
//	}
//	return nil, errors.New("no result")
//}

func GetGlobalViper() *viper.Viper {
	return vipConf
}

func init() {
	if runtime.GOOS == "windows" {
		strGlobalConfigPath = "c:/etc/GlobalConfigure.yaml"
		strYamlRootPath = "c:/etc/modules/"
	}

	strTempEnv := os.Getenv("GWORLD_ENV")
	log.Info("config-env", "env", strTempEnv)

	vipConf = viper.New()
	vipConf.SetConfigFile(strGlobalConfigPath)  //设置配置文件名
	err := vipConf.ReadInConfig()     // 读配置数据
	if err != nil {
		log.Error("\"Fatal error config file: " + err.Error())
		return
	}
}

func writeToFile(filePath string, outPut []byte) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(f)
	_, err = writer.Write(outPut)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}
