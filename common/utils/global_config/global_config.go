package global_config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	//"os"
	"runtime"
	"sync"

	log "github.com/xiaomi-tc/log15"
)

var loadOnce sync.Once
var globalConfig interface{}

func Init() {
	loadOnce.Do(func() {
		var yml string
		if runtime.GOOS == "windows" {
			yml = "c:\\etc\\GlobalConfigure.yaml"
		} else {
			yml = "/etc/gworld/GlobalConfigure.yaml"
		}

		var err error
		globalConfig, err = parseYaml(yml)
		if err != nil {
			log.Error("parse GlobalConfigure.yaml err", "err", err)
		}
	})
}

func GetString(keys ...string) (string, error) {
	var r, err = get(globalConfig, keys...)
	if err != nil {
		log.Error("GlobalConfig GetString", "keys", keys, "err", err)
		return "", err
	}

	var str, ok = r.(string)
	if !ok {
		return "", fmt.Errorf("key fmt error")
	}

	return str, nil
}

func GetInt(keys ...string) (int, error) {
	var r, err = get(globalConfig, keys...)
	if err != nil {
		log.Error("GlobalConfig GetString", "keys", keys, "err", err)
		return -1, err
	}

	var i, ok = r.(int)
	if !ok {
		return -1, fmt.Errorf("key fmt error")
	}

	return i, nil
}

func parseYaml(yml string) (interface{}, error) {
	data, err := ioutil.ReadFile(yml)
	if err != nil {
		log.Error("parseYaml", "read yaml err", err)
		return nil, err
	}

	var configMap = make(map[string]interface{})
	err = yaml.Unmarshal([]byte(data), &configMap)
	if err != nil {
		log.Error("parseYaml", "unmarshal err", err)
		return nil, err
	}

	//var env = getEnv()
	//result, ok := configMap[env]
	//if !ok {
	//	log.Error("parseYaml", "env doesn't exists", env)
	//	return nil, fmt.Errorf("expect project(%s)", projectName)
	//}
	result := configMap

	return result, nil
}

func get(config interface{}, keys ...string) (interface{}, error) {
	for _, k := range keys {
		r, err := getkey(config, k)
		if err != nil {
			return nil, err
		}
		config = r
	}
	return config, nil
}

func getkey(config interface{}, key string) (interface{}, error) {
	var m, ok = config.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("get key(%s) error", key)
	}

	result, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("expect key %s", key)
	}
	return result, nil
}

//func getEnv() string {
//	var env = os.Getenv("WODA_ENV")
//	if env == "" {
//		env = "production"
//	}
//
//	return env
//}
