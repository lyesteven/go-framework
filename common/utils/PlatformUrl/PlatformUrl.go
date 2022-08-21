package PlatformUrl

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	log "github.com/xiaomi-tc/log15"
	"gopkg.in/yaml.v2"
)

const (
	kPlatformKey = "platform_api_urls"
	HOTMALL         = "hotmall_api"
	FRAMEWORK         = "framework_api"
)

var platformApiUrls map[string]string
var loadConfigOnce sync.Once

func GetPlatformUrl(platform string) (string, error) {
	var result, ok = platformApiUrls[platform]
	if !ok {
		return "", fmt.Errorf("no result")
	}
	return result, nil
}

func init() {
	var yml string
	if runtime.GOOS == "windows" {
		yml = "c:\\etc\\gworld\\GlobalConfigure.yaml"
	} else {
		yml = "/etc/gworld/GlobalConfigure.yaml"
	}

	loadConfigOnce.Do(func() {
		var err error
		platformApiUrls, err = parseYaml(yml)
		if err != nil {
			log.Error("parse GlobalConfigure.yaml err", "err", err)
			platformApiUrls = make(map[string]string)
		}
	})
}

func parseYaml(yml string) (map[string]string, error) {
	data, err := ioutil.ReadFile(yml)
	if err != nil {
		log.Error("parseYaml", "read yaml err", err)
		return nil, err
	}

	var configMap = make(map[string]map[string]map[string]interface{})
	err = yaml.Unmarshal([]byte(data), &configMap)
	if err != nil {
		log.Error("parseYaml", "unmarshal err", err)
		return nil, err
	}

	var env = getEnv()
	if _, ok := configMap[env]; !ok {
		log.Error("parseYaml", "env doesn't exists", env)
		return nil, fmt.Errorf("expect env(%s)", env)
	}

	if _, ok := configMap[env][kPlatformKey]; !ok {
		log.Error("parseYaml", "env doesn't exists", env)
		return nil, fmt.Errorf("expect key(%s)", kPlatformKey)
	}

	var platformUrls = configMap[env][kPlatformKey]
	var result = make(map[string]string)
	for k, v := range platformUrls {
		if _, ok := v.(string); !ok {
			return nil, fmt.Errorf("key(%s) error", k)
		}

		result[k] = v.(string)
	}
	return result, nil
}

func getEnv() string {
	var env = os.Getenv("WODA_ENV")
	if env == "" {
		env = "development"
	}

	return env
}
