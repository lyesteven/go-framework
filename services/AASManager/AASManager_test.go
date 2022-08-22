package main

import (
	"testing"
)

//func Test_getYamlConfig(t *testing.T)  {
//	m:=getYamlConfig("./AASManager.yaml");
//	if m == nil{
//		t.Error("Yaml is empty!")
//	}
//
//	_, ok:=m["AASIP"]
//	if !ok{
//		t.Error("AASIP is empty")
//	}
//	_, ok =m["AASPort"]
//	if !ok{
//		t.Error("AASPort is empty")
//	}
//	_, ok =m["TokenAesKey"]
//	if !ok{
//		t.Error("TokenAesKey is empty")
//	}
//
//}

func Benchmark_strcata_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		str := "12"
		str += "23"
		str += "56"
	}
}
