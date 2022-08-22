package main

import (
	"fmt"
	"time"
	"github.com/buger/jsonparser"
)

// the proto of data from http post:
//  '{ "appkey": str,
//     "appver": str,
//     "data":  string(json) / json,
//     "timestamp": int,
//     "token":  str,
//     "id": str,
//     "srcip": str,
//    }',

func parseHttpFormatCompat(body []byte) (*stHttpFormat, []string, error) {
	var http_format = new(stHttpFormat)
	var result []string

	var err = jsonparser.ObjectEach(body, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if dataType != jsonparser.String && dataType != jsonparser.Number {
			// "data" 兼容 "string" and "json" obj　--steven 2020-06-20
			if string(key) != "data" || dataType != jsonparser.Object {
				return fmt.Errorf("key(%s) error type", string(key))
			}
		}

		var err error
		switch string(key) {
		case "appkey":
			http_format.AppKey, err = jsonparser.ParseString(value)
		case "appver":
			http_format.AppVer, err = jsonparser.ParseString(value)
		case "data":
			http_format.Data, err = jsonparser.ParseString(value)
		case "token":
			http_format.Token, err = jsonparser.ParseString(value)
		case "id":
			http_format.Id, err = jsonparser.ParseString(value)
		case "usertype":
			http_format.UserType, err = jsonparser.ParseInt(value)
			if dataType == jsonparser.String {
				result = append(result, string(key))
			}
		//case "timestamp":
		//	http_format.TimeStamp, err = jsonparser.ParseInt(value)
		//	if dataType == jsonparser.String {
		//		result = append(result, string(key))
		//	}
		//case "srcip":
		//	http_format.SrcIP, err = jsonparser.ParseString(value)
		}
		return err
	})

	// 由服务端设置时间戳
	http_format.TimeStamp = time.Now().Unix()

	return http_format, result, err
}
