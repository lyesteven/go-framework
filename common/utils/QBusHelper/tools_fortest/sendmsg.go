package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	log "github.com/xiaomi-tc/log15"
	su "github.com/lyesteven/go-framework/common/service_util"
	"time"
)

func main() {
	var (
		udpsvr_addr string
		count int
	)

	flag.StringVar(&udpsvr_addr, "a", "127.0.0.1:12333", "UDP server address")
	flag.IntVar(&count, "ct", 3, "number to send")
	flag.Parse()

	log.SetOutLevel(log.LvlInfo)

	//serviceUtil := su.GetInstanceWithOptions(su.WithConsulSrvAddr("127.0.0.1:18500"))
	su.GetInstanceWithOptions(su.WithConsulSrvAddr("127.0.0.1:18500"))
	kv := su.GetConsulclient().KV()

	addr, err := net.ResolveUDPAddr("udp", udpsvr_addr)
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Can't dial: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(strconv.Itoa(count)))
	if err != nil {
		fmt.Println("failed:", err)
		os.Exit(1)
	}

	time.Sleep(2*time.Second)

	pair, _, err := kv.Get("pubTestID", nil)
	if err ==nil {
		log.Info("pubTestID kv value:" + string(pair.Value))
	}
	pair, _, _ = kv.Get("subTestID", nil)
	if err ==nil {
		log.Info("subTestID kv value:" + string(pair.Value))
	}
}
