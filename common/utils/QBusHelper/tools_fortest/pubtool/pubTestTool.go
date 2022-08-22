package main

import (
	"flag"
	"fmt"
	su "github.com/lyesteven/go-framework/common/service_util"
	qbh "github.com/lyesteven/go-framework/common/utils/QBusHelper"
	"github.com/hashicorp/consul/api"
	log "github.com/xiaomi-tc/log15"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var	c chan os.Signal
var chLoopTimes   = make(chan int)

func waitToExit() {
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGHUP,syscall.SIGUSR1)
	go func() {
		for v := range c {
			log.Info("get signal:" + v.String())
			// just close connection, no need unsubscribe !!
			qbh.CloseConn()
			os.Exit(1)
		}
	}()
}

func udp_server(address string) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer conn.Close()
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		//n, remoteAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Println("failed to read UDP msg because of ", err.Error())
			return
		}

		loopTimes, _  := strconv.Atoi(string(data[:n]))
		//log.Info("will send req:","num:",loopTimes)
		chLoopTimes <- loopTimes
	}
}

func main() {
	var clientID, topic, address string

	flag.StringVar(&clientID, "id", "MyTestPubID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic,"tp","hello","The topic name")
	flag.StringVar(&address,"a","127.0.0.1:12333","UDP server address")
	flag.Parse()

	// Open log file
	h, _ := log.FileHandler("./pubTest.log" ,log.LogfmtFormat())
	log.Root().SetHandler(h)

	//log.SetOutLevel(log.LvlInfo)
	log.SetOutLevel(log.LvlDebug)

	serviceUtil := su.GetInstanceWithOptions(su.WithConsulSrvAddr("127.0.0.1:18500"))

	qbh.SetDelayMsgDir("./data/pub_delayMsg")
	qbh.Init(serviceUtil, clientID)

	waitToExit()

	go udp_server(address)

	kv := qbh.ConsulClient.KV()
	j := 0
	for {
		looptimes := <-chLoopTimes
		log.Info("get a send request,number:"+strconv.Itoa(looptimes))

		for i:= 0; i < looptimes; i++ {
			msg := "msg-" + clientID + "-" + strconv.Itoa(j)
			//j++
			data := []byte(msg)

			log.Info("before pub:"+msg)
			if err := qbh.Publish(topic, data); err != nil {

				log.Error(" \""+string(data[:])+"\"", "error=", err)
			} else {
				log.Info(msg)
			}

			log.Info("after pub:"+msg)

			pair, meta, err := kv.Get(clientID, nil)
			if err != nil {
				// consul should not be error !
				log.Error("consul get kv error! in publishDelayMsg(), will retry 1 second later")
				time.Sleep(1 * time.Second)
			}

			if pair != nil {
				j,_ = strconv.Atoi(string(pair.Value))
				j++
				log.Info("CAS:","i",j)
				p := &api.KVPair{Key: clientID, Value: []byte(strconv.Itoa(j))}
				// change kv in consul. if failed will do nothing, and have another chance one second after
				p.ModifyIndex = meta.LastIndex
				kv.CAS(p, nil)
			} else {
				j = 1  // other process can reclean count number by clear KV
				log.Info("Put:","i",i)
				p := &api.KVPair{Key: clientID, Value: []byte(strconv.Itoa(j))}
				// first time, put new kv in consul
				kv.Put(p, nil)
			}
		}
	}
}
