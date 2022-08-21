package QBusUtility

import (
	"container/list"
	"errors"
	"fmt"
	su "gworld/git/GoFrameWork/common/service_util"
	"gworld/git/GoFrameWork/pb/ComMessage"
	. "github.com/agiledragon/gomonkey"
	"github.com/hashicorp/consul/api"
	"github.com/latermoon/boltdb"
	"github.com/nats-io/go-nats-streaming"
	. "github.com/smartystreets/goconvey/convey"
	log "github.com/xiaomi-tc/log15"
	"net"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

//　Hepler


type heplerNode struct {
	configFile string
	cmdHandler *exec.Cmd
}

var (
	receiveNo int=0
	listNSCmd = list.New()
	listCSCmd = list.New()
	listPubCmd = list.New()
	listSubCmd = list.New()
)



func contains(l *list.List, value string) (bool, *list.Element) {
	for e := l.Front(); e != nil; e = e.Next() {
		v := e.Value.(*heplerNode)
		if v.configFile == value {
			return true, e
		}
	}
	return false, nil
}

func startConsul(configFile string) {
	go func() {
		cmdString := "./tools_fortest/consul agent" + " -bind "+su.GetHostIP() + " -config-file=" + configFile
		//lsCmd = exec.Command("bash", "-c", "./tools/consul agent -config-file=./tools/cs-cfg.json")

		if contain, e := contains(listCSCmd, cmdString); contain {
			v:= e.Value.(*heplerNode)
			if v.cmdHandler != nil {
				log.Info("startConsul() Do Nothing!: "+cmdString)
				return
			}
		}
		newNode := &heplerNode{configFile: configFile, cmdHandler: exec.Command("bash", "-c", cmdString)}

		listCSCmd.PushBack(newNode)
		err :=newNode.cmdHandler.Run()
		if err != nil {
			//log.Info("startConsul() : "+cmdString +", error:" + err.Error())
		}
	}()
}

func stopConsul(configFile string) {

	if contain, e := contains(listCSCmd, configFile); contain {
		v:= e.Value.(*heplerNode)
		v.cmdHandler.Process.Kill()
		listCSCmd.Remove(e)
		return
	}
}


func startNS(configFile string) {

	go func() {
		cmdString := "./tools_fortest/nats-streaming-server -c " + configFile
		//nsCmd = exec.Command("bash", "-c", "./tools/nats-streaming-server -c ./tools/ns-cfg-1.conf")
		if contain, e := contains(listNSCmd, cmdString); contain {
			v:= e.Value.(*heplerNode)
			if v.cmdHandler != nil {
				log.Info("startNS() Do Nothing!: "+cmdString)
				return
			}
		}
		newNode := &heplerNode{configFile: configFile, cmdHandler: exec.Command("bash", "-c", cmdString)}
		//var stderr bytes.Buffer
		//newNode.cmdHandler.Stderr = &stderr

		listNSCmd.PushBack(newNode)
		err := newNode.cmdHandler.Run()
		if err != nil {
			//log.Info("startNS() : "+cmdString +", error:" + err.Error())
		}

		//err := newNode.cmdHandler.Run()()
		//if err !=nil {
		//	fmt.Println("Start MQ:" + cmdString)
		//	fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		//}
	}()
}

func stopNS(configFile string) {

	if contain, e := contains(listNSCmd, configFile); contain {
		v:= e.Value.(*heplerNode)
		v.cmdHandler.Process.Kill()
		listNSCmd.Remove(e)
		return
	}
}

func setupQBHelplerTestingEnv() *su.ServiceUtil {

	// start "consul" and "nats-streaming" server for testing
	startConsul("./tools_fortest/cs-cfg.json")
	time.Sleep(2 * time.Second)
	startNS("./tools_fortest/ns-cfg-1.conf")

	sUtil := su.GetInstanceWithOptions(su.WithConsulSrvAddr("127.0.0.1:18500"))
	time.Sleep(10*time.Second)
	return sUtil
}

func resetQBHelplerTTestingEnv() {
	stopConsul("./tools_fortest/cs-cfg.json")
	stopNS("./tools_fortest/ns-cfg-1.conf")
	stopNS("./tools_fortest/ns-cfg-2.conf")
	time.Sleep(2 *time.Second)

}

func setupTransTestingEnv(topic, udp_addr string) *su.ServiceUtil {

	sUtil:=setupQBHelplerTestingEnv()

	// prepare pub/sub tools
	log.Info("begin start pub/sub process")
	startPubTestTool("pubTestID",topic,udp_addr)
	startSubTestTool("subTestID",topic)
	time.Sleep(6 *time.Second)

	return sUtil
}

func resetTransTestEnv() {
	resetQBHelplerTTestingEnv()
	log.Info("resetTransTestEnv() stop Pub/Sub!")
	stopPubTestTool()
	stopSubTestTool()
}

func startPubTestTool(clientID,topic,udpsvr_addr string) {
	go func() {
		cmdString := "./tools_fortest/pubtool/pubTestTool -id " + clientID +" -tp " + topic + " -a " + udpsvr_addr

		if contain, e := contains(listPubCmd, cmdString); contain {
			v:= e.Value.(*heplerNode)
			if v.cmdHandler != nil {
				log.Info("startPubTestTool() Do Nothing!: "+cmdString)
				return
			}
		}
		newNode := &heplerNode{configFile: cmdString, cmdHandler: exec.Command("bash", "-c", cmdString)}

		listPubCmd.PushBack(newNode)
		log.Info("startPubTestTool() :"+newNode.configFile)
		err :=newNode.cmdHandler.Run()
		if err != nil {
			//log.Info("startPubTestTool() : "+cmdString +", error:" + err.Error())
		}
	}()
}

func stopPubTestTool() {
	for e := listPubCmd.Front(); e != nil; e = e.Next() {
		v := e.Value.(*heplerNode)
		if err:= v.cmdHandler.Process.Kill(); err != nil {
			log.Info("stopPubTestTool() :"+v.configFile +", error:" + err.Error())
		}
		log.Info("stopPubTestTool() :"+v.configFile)
		listPubCmd.Remove(e)
	}

	return
}

func req_pubmsgs(udpsvr_addr string, count int) {
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
}

func startSubTestTool(clientID, topic string) {
	go func() {
		cmdString := "./tools_fortest/subtool/subTestTool -id " + clientID +" -tp " + topic
		//lsCmd = exec.Command("bash", "-c", "./tools/consul agent -config-file=./tools/cs-cfg.json")

		if contain, e := contains(listSubCmd, cmdString); contain {
			v:= e.Value.(*heplerNode)
			if v.cmdHandler != nil {
				log.Info("startSubTestTool() Do Nothing!: "+cmdString)
				return
			}
		}
		newNode := &heplerNode{configFile: cmdString, cmdHandler: exec.Command("bash", "-c", cmdString)}

		listSubCmd.PushBack(newNode)
		log.Info("startSubTestTool() :"+newNode.configFile)
		err :=newNode.cmdHandler.Run()
		if err != nil {
			//log.Info("startSubTestTool() : "+cmdString +", error:" + err.Error())
		}

	}()
}

func stopSubTestTool() {
	for e := listSubCmd.Front(); e != nil; e = e.Next() {
		v := e.Value.(*heplerNode)
		if err:= v.cmdHandler.Process.Kill(); err != nil {
			log.Error("stopSubTestTool() :"+v.configFile +", error:" + err.Error())
		}
		log.Info("stopSubTestTool() :"+v.configFile)
		listSubCmd.Remove(e)
	}

	return
}


var Msgcb stan.MsgHandler = func(msg *stan.Msg) {
	receiveNo++
	//log.Info("recv["+ strconv.Itoa(receiveNo)+"]:", "msg",string(msg.Data))
}

// Testing

func Test_openDlyMsgLists(t *testing.T) {

	var testpath = "./testpath"
	var testmsg1,testmsg2 = []byte("first msg"),[]byte("second msg")

	//var diskQueue gdq.Interface
	Convey("Test_openDlyMsgLists()", t, func() {
		var inputMsg = func(msg []byte) {
			SetDelayMsgDir(testpath)
			_ = openDlyMsgLists()
			So(dlyMsgDB, ShouldNotBeNil)

			listsCount := len(dlyMsgLists)
			dcCount := len(ComMessage.CBSType_name)
			So(listsCount,ShouldEqual,dcCount)

			err := dlyMsgLists[ComMessage.CBSType_Own].RPush([]byte(msg))
			So(err,ShouldBeNil)

			dlyMsgDB.Close()
			return
		}

		var outMsg = func() []byte {
			SetDelayMsgDir(testpath)
			_ = openDlyMsgLists()
			So(dlyMsgDB, ShouldNotBeNil)

			msgOut,err := dlyMsgLists[ComMessage.CBSType_Own].LPop()
			So(err,ShouldBeNil)

			//after dlyMsgDB.Close(), the space will be release.
			// so need use b to save message
			maxbytes := len(msgOut) // length of b
			b := make([]byte, maxbytes)
			copy(b,msgOut)

			dlyMsgDB.Close()
			return b
		}

		Convey("目录不存在", func(){
			if err:= os.RemoveAll(testpath);err != nil {
				t.Error("Can't remove directory:" + testpath)
			}
			log.SetOutLevel(log.LvlInfo)
			inputMsg(testmsg1)
			msgOut:= outMsg()
			So(string(msgOut),ShouldEqual,string(testmsg1))
		})

		Convey("目录存在,文件不存在", func(){
			if err:= os.RemoveAll(testpath);err != nil {
				t.Error("Can't remove directory:" + testpath)
			}
			err := os.Mkdir(testpath, os.ModePerm)
			if err != nil {
				t.Error("Can't remove directory:" + testpath)
			}
			log.SetOutLevel(log.LvlError)
			inputMsg(testmsg1)
			msgOut:= outMsg()
			So(string(msgOut),ShouldEqual,string(testmsg1))
		})

		Convey("目录文件都存在", func(){

			inputMsg(testmsg2)
			msgOut := outMsg()
			So(string(msgOut),ShouldEqual,string(testmsg2))
		})

		Convey("失败测试", func() {
			Convey("创建目录失败", func() {
				patches := ApplyFunc(os.MkdirAll, func(_ string, _ os.FileMode) (error) {
					return errors.New("MkdirAll Mock failed error")
				})
				defer patches.Reset()

				if err := os.RemoveAll(testpath); err != nil {
					t.Error("Can't remove directory:" + testpath)
				}

				SetDelayMsgDir(testpath)
				err := openDlyMsgLists()
				So(err, ShouldNotBeNil)
			})

			Convey("建DlyMsgLists失败", func(){
				patches := ApplyFunc(boltdb.Open, func(_ string, _ os.FileMode, _ *boltdb.Options) (*boltdb.DB, error){
					return nil, errors.New("create db error")
				})
				defer patches.Reset()

				SetDelayMsgDir(testpath)
				err := openDlyMsgLists()

				So(dlyMsgDB, ShouldBeNil)
				So(err, ShouldNotBeNil)

				if err := os.RemoveAll(testpath); err != nil {
					t.Error("Can't remove directory:" + testpath)
				}
			})
		})
	})

}

func Test_SetMaxCurrHandleMsgNumber(t *testing.T) {
	type args struct {
		maxNumber int
		want	int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"=2",args{2,2}},
		{"=5",args{5,5}},
		{">5",args{6,5}},
		{"=0",args{0,0}},
		{"<0",args{-1,0}},
	}
	Convey("TestSetMaxCurrHandleMsgNumber()",t,func() {
		for _, tt := range tests {
				SetMaxCurrHandleMsgNumber(tt.args.maxNumber)
				So(maxInFlight, ShouldEqual, tt.args.want)
		}
	})
}

func Test_QBusHelper(t *testing.T) {

	datadir := "./data"
	if err := os.RemoveAll(datadir); err != nil {
		t.Fatal("Can't remove directory:" + datadir)
	}

	// Start consul and MQ for this process
	sUtil := setupQBHelplerTestingEnv()
	defer func() {
		resetQBHelplerTTestingEnv()
		if err := os.RemoveAll(datadir); err != nil {
			t.Fatal("Can't remove directory:" + datadir)
		}
	}()

	Convey("Init()前检测", t, func() {
		Convey("变量及参数", func() {
			So(bInit, ShouldBeFalse)
			So(ServiceName, ShouldEqual, "")
			//So(maxInFlight, ShouldEqual,1)
			So(ConsulClient, ShouldBeNil)
		})

		Convey("lists", func() {
			So(mapDCs, ShouldBeNil)
		})

		Convey("map", func() {
			So(len(mapTopicConn), ShouldEqual, 0)
		})

		Convey("DlyMsgLists", func() {
			So(dlyMsgDB,ShouldBeNil)
			listsCount := len(dlyMsgLists)
			dcCount := len(ComMessage.CBSType_name)
			So(listsCount,ShouldEqual,dcCount)
			len,_ :=dlyMsgLists[ComMessage.CBSType_Own].Len()
			So(len, ShouldEqual,int64(0))
		})
	})

	// Init()
	Init(sUtil, "tstQBusID")
	testTopic := "testTopic"

	time.Sleep(5*time.Second)

	Convey("Init()后检测", t,func() {

		Convey("变量及参数", func() {
			So(bInit, ShouldBeTrue)
			So(ServiceName, ShouldEqual, "QBus")
			So(maxInFlight, ShouldEqual, 1)
			So(ConsulClient, ShouldNotBeNil)
		})

	})

	// 模拟一次收发操作
	// publish one message to topic "testTopic"
	for i := 0; i < 1; i++ {
		msg := "msgTest-" + strconv.Itoa(i)
		Publish(testTopic, []byte(msg))
		//log.Info("send["+strconv.Itoa(i)+"]:", "msg", msg)
	}
	// consume one message from topic "testTopic"
	receiveNo = 0
	SubTopic(testTopic, "group-1", "abc", 1, Msgcb)
	time.Sleep(5*time.Second)

	Convey("完成一次发布及订阅后的检测", t, func() {

		Convey("map", func() {
			for k,v := range mapTopicConn {
				fmt.Println(k,":",v)
			}

			So(len(mapTopicConn), ShouldEqual, 1)
			key := ComMessage.CBSType_Own.String()+"_"+testTopic
			pCon, _ := mapTopicConn[key]
			// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
			So(pCon.pubConn.nsClusterID, ShouldEqual, "testMQnode01")
			So(pCon.bCanUse, ShouldBeTrue)
		})

		Convey("diskQueue", func() {
			So(dlyMsgDB,ShouldNotBeNil)

			listsCount := len(dlyMsgLists)
			dcCount := len(ComMessage.CBSType_name)
			So(listsCount,ShouldEqual,dcCount)
			So(dlyMsgLists[ComMessage.CBSType_Own], ShouldNotBeNil)
		})

		Convey("lists", func() {
			Convey("listNSClusterAddr", func() {
				So(mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Len(), ShouldEqual, 1)
				qa := mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Front().Value.(*nsClusterAddr)
				// ns-cfg-1.conf 中设置  cluster_id: testMQnode01
				So(qa.nsClusterID, ShouldEqual, "testMQnode01")
				So(qa.Port, ShouldEqual, 15222)
			})

			Convey("listPublishConn", func() {
				So(mapDCs[ComMessage.CBSType_Own].listPublishConn.Len(), ShouldEqual, 1)
				qa := mapDCs[ComMessage.CBSType_Own].listPublishConn.Front().Value.(*pubConn)
				// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
				So(qa.nsClusterID, ShouldEqual, "testMQnode01")
			})

			Convey("listSubTopic", func() {
				So(mapDCs[ComMessage.CBSType_Own].listSubTopic.Len(), ShouldEqual, 1)
				qa := mapDCs[ComMessage.CBSType_Own].listSubTopic.Front().Value.(*subTopic)
				So(qa.topic, ShouldEqual, testTopic)
				So(qa.qgroup, ShouldEqual, "group-1")
				So(qa.durable, ShouldEqual, "abc")
				So(qa.manualAck,ShouldEqual,false)
				So(qa.hdnum, ShouldEqual, 1)
			})

			Convey("listSubConn", func() {
				So(mapDCs[ComMessage.CBSType_Own].listSubConn.Len(), ShouldEqual, 1)
				qa := mapDCs[ComMessage.CBSType_Own].listSubConn.Front().Value.(*subConn)
				// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
				So(qa.nsClusterID, ShouldEqual, "testMQnode01")
				So(qa.topic, ShouldEqual, testTopic)
				So(qa.qgroup, ShouldEqual, "group-1")
				So(qa.durable, ShouldEqual, "abc")
			})
		})
	})

	// 取消订阅
	UnsubTopic(testTopic, "group-1", "abc")
	time.Sleep(1*time.Second)

	Convey("取消订阅后检测", t, func() {
		Convey("map", func() {
			key := ComMessage.CBSType_Own.String()+"_"+testTopic
			So(len(mapTopicConn), ShouldEqual, 1)
			pCon, _ := mapTopicConn[key]
			// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
			So(pCon.pubConn.nsClusterID, ShouldEqual, "testMQnode01")
			So(pCon.bCanUse, ShouldBeTrue)

		})

		Convey("lists", func() {
			Convey("listNSClusterAddr", func() {
				So(mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Len(), ShouldEqual, 1)
				qa := mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Front().Value.(*nsClusterAddr)
				// ns-cfg-1.conf 中设置  cluster_id: testMQnode01
				So(qa.nsClusterID, ShouldEqual, "testMQnode01")
				So(qa.Port, ShouldEqual, 15222)
			})

			Convey("listPublishConn", func() {
				So(mapDCs[ComMessage.CBSType_Own].listPublishConn.Len(), ShouldEqual, 1)
				qa := mapDCs[ComMessage.CBSType_Own].listPublishConn.Front().Value.(*pubConn)
				// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
				So(qa.nsClusterID, ShouldEqual, "testMQnode01")
			})

			Convey("listSubTopic", func() {
				So(mapDCs[ComMessage.CBSType_Own].listSubTopic.Len(), ShouldEqual, 0)
			})

			Convey("listSubConn", func() {
				So(mapDCs[ComMessage.CBSType_Own].listSubConn.Len(), ShouldEqual, 0)
			})
		})
	})

	// Init() again
	Init(sUtil, "tstQBusID")
	Convey("重复调用Init()状态检测", t, func() {

		Convey("变量及参数", func() {
			So(bInit, ShouldBeTrue)
			So(ServiceName, ShouldEqual, "QBus")
			So(maxInFlight, ShouldEqual, 1)
			So(ConsulClient, ShouldNotBeNil)
		})

		Convey("map 及 List", func() {
			Convey("map", func() {
				So(len(mapTopicConn), ShouldEqual, 1)
				key := ComMessage.CBSType_Own.String()+"_"+testTopic
				pCon, _ := mapTopicConn[key]
				// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
				So(pCon.pubConn.nsClusterID, ShouldEqual, "testMQnode01")
				So(pCon.bCanUse, ShouldBeTrue)
			})

			Convey("lists", func() {
				Convey("listNSClusterAddr", func() {
					So(mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Len(), ShouldEqual, 1)
					qa := mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Front().Value.(*nsClusterAddr)
					// ns-cfg-1.conf 中设置  cluster_id: testMQnode01
					So(qa.nsClusterID, ShouldEqual, "testMQnode01")
					So(qa.Port, ShouldEqual, 15222)
				})

				Convey("listPublishConn", func() {
					So(mapDCs[ComMessage.CBSType_Own].listPublishConn.Len(), ShouldEqual, 1)
					qa := mapDCs[ComMessage.CBSType_Own].listPublishConn.Front().Value.(*pubConn)
					// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
					So(qa.nsClusterID, ShouldEqual, "testMQnode01")
				})

				Convey("listSubTopic", func() {
					So(mapDCs[ComMessage.CBSType_Own].listSubTopic.Len(), ShouldEqual, 0)
				})

				Convey("listSubConn", func() {
					So(mapDCs[ComMessage.CBSType_Own].listSubConn.Len(), ShouldEqual, 0)
				})
			})
		})

	})

	// 再次订阅 topic "testTopic"
	receiveNo = 0
	SubTopic(testTopic, "group-1", "abc", 1, Msgcb)

	Convey("MQ服务发现-当前结点检测", t,func() {
		Convey("当前结点信息", func() {
			So(mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Len(), ShouldEqual, 1)
			qa := mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Front().Value.(*nsClusterAddr)
			// ns-cfg-1.conf 中设置  cluster_id: testMQnode01
			So(qa.nsClusterID, ShouldEqual, "testMQnode01")
			So(qa.Port, ShouldEqual, 15222)
		})

		Convey("结点已经连接", func() {
			So(mapDCs[ComMessage.CBSType_Own].listPublishConn.Len(), ShouldEqual, 1)
			qa := mapDCs[ComMessage.CBSType_Own].listPublishConn.Front().Value.(*pubConn)
			// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
			So(qa.nsClusterID, ShouldEqual, "testMQnode01")
		})

		Convey("结点上已订阅主题", func() {
			So(mapDCs[ComMessage.CBSType_Own].listSubConn.Len(), ShouldEqual, 1)
			qa := mapDCs[ComMessage.CBSType_Own].listSubConn.Front().Value.(*subConn)
			// ./tools/ns-cfg-1.conf 中设置  cluster_id: testMQnode01
			So(qa.nsClusterID, ShouldEqual, "testMQnode01")
			So(qa.topic, ShouldEqual, testTopic)
			So(qa.qgroup, ShouldEqual, "group-1")
			So(qa.durable, ShouldEqual, "abc")
		})

	})

	// 新连一个MQ结点
	// 结点2配置文件: ./tools/ns-cfg-2.conf
	startNS("./tools_fortest/ns-cfg-2.conf")
	//log.Info("start second NATS-Streaming")
	time.Sleep(10*time.Second)

	Convey("MQ服务发现-增加MQ结点", t,func() {

		Convey("新加结点信息", func() {
			So(mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Len(), ShouldEqual, 2)
			qa := mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Back().Value.(*nsClusterAddr)
			// ns-cfg-2.conf 中设置  cluster_id: testMQnode02
			So(qa.nsClusterID, ShouldEqual, "testMQnode02")
			So(qa.Port, ShouldEqual, 16222)
		})

		Convey("新加结点已连接", func() {
			So(mapDCs[ComMessage.CBSType_Own].listPublishConn.Len(), ShouldEqual, 2)
			qa := mapDCs[ComMessage.CBSType_Own].listPublishConn.Back().Value.(*pubConn)
			// ./tools/ns-cfg-2.conf 中设置  cluster_id: testMQnode02
			So(qa.nsClusterID, ShouldEqual, "testMQnode02")
		})

		Convey("新结点订阅了主题", func() {
			So(mapDCs[ComMessage.CBSType_Own].listSubConn.Len(), ShouldEqual, 2)
			qa := mapDCs[ComMessage.CBSType_Own].listSubConn.Back().Value.(*subConn)
			// ./tools/ns-cfg-2.conf 中设置  cluster_id: testMQnode02
			So(qa.nsClusterID, ShouldEqual, "testMQnode02")
			So(qa.topic, ShouldEqual, testTopic)
			So(qa.qgroup, ShouldEqual, "group-1")
			So(qa.durable, ShouldEqual, "abc")
		})
	})

	// 删除结点2. 配置文件: ./tools/ns-cfg-2.conf
	stopNS("./tools_fortest/ns-cfg-2.conf")

	time.Sleep(10*time.Second)

	Convey("删除新增MQ结点", t,func() {

		Convey("结点信息", func() {
			So(mapDCs[ComMessage.CBSType_Own].listNSClusterAddr.Len(), ShouldEqual, 1)
		})

		Convey("新加结点已断连", func() {
			So(mapDCs[ComMessage.CBSType_Own].listPublishConn.Len(), ShouldEqual, 1)
		})

		Convey("新结点已经从订阅表中删除", func() {
			So(mapDCs[ComMessage.CBSType_Own].listSubConn.Len(), ShouldEqual, 1)
		})
	})

	// 取消订阅
	UnsubTopic(testTopic, "group-1", "abc")


}

func Test_Trans(t *testing.T) {
	datadir := "./data"
	if err := os.RemoveAll(datadir); err != nil {
		t.Fatal("Can't remove directory:" + datadir)
	}

	// prepare pub/sub tools
	testTopic := "testPubSubTopic"
	udp_address := "127.0.0.1:12333"
	setupTransTestingEnv(testTopic,udp_address)

	defer func() {
		resetTransTestEnv()
		if err := os.RemoveAll(datadir); err != nil {
			t.Fatal("Can't remove directory:" + datadir)
		}
		if err := os.RemoveAll(("./delayMessages")); err != nil {
			t.Fatal("Can't remove directory:./delayMessages")
		}
	}()

	//su.GetInstanceWithOptions(su.WithConsulSrvAddr("127.0.0.1:18500"))
	kv := su.GetConsulclient().KV()

	time.Sleep(5*time.Second)

	Convey("传输测试-单MQ结点", t, func() {

		Convey("通讯正常性测试", func() {
			// 发布 3条消息
			req_pubmsgs(udp_address, 3)
			//log.Info("send 3 message")
			time.Sleep(1*time.Second)

			p := &api.KVPair{Key: "testKV", Value: []byte("testvalue")}
			kv.Put(p, nil)
			time.Sleep(1*time.Second)
			pa, _, _ := kv.Get("testKV", nil)
			So(string(pa.Value),ShouldEqual,"testvalue")

			//time.Sleep(10*time.Second)

			pair, _, _ := kv.Get("pubTestID", nil)
			for pair == nil {
				log.Info("pair is nil" )
				time.Sleep(3*time.Second)
				pair, _, _ = kv.Get("pubTestID", nil)
			}

			value, _ := strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)
			// 检查是否接受了 3条消息
			pair, _, _ = kv.Get("subTestID", nil)
			value, _ = strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)

			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)
			time.Sleep(1*time.Second)
		})

		Convey("通讯中订阅方中断再恢复", func() {
			// 发布 3条消息
			req_pubmsgs(udp_address, 3)
			// 中断订阅方
			stopSubTestTool()
			time.Sleep(1*time.Second)
			// 再发布 3条消息
			req_pubmsgs(udp_address, 3)
			// 恢复订阅
			startSubTestTool("subTestID", testTopic)

			time.Sleep(2*time.Second)
			// 检查结果
			pair, _, _ := kv.Get("pubTestID", nil)
			value, _ := strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 6)
			// 检查是否接受了 3条消息
			pair, _, _ = kv.Get("subTestID", nil)
			value, _ = strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 6)

			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)
			time.Sleep(1*time.Second)
		})

		Convey("通讯中发布方中断再恢复", func() {
			// 发布 3条消息
			req_pubmsgs(udp_address, 3)
			time.Sleep(1*time.Second)
			// 中断发阅方
			stopPubTestTool()
			// 检查结果
			pair, _, _ := kv.Get("pubTestID", nil)
			value, _ := strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)
			// 检查是否接受了 3条消息
			pair, _, _ = kv.Get("subTestID", nil)
			value, _ = strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)

			// 恢复发布
			startPubTestTool("pubTestID", testTopic, udp_address)
			time.Sleep(2*time.Second)
			// 再发布 3条消息
			req_pubmsgs(udp_address, 3)
			time.Sleep(1*time.Second)

			// 检查结果
			pair, _, _ = kv.Get("pubTestID", nil)
			value, _ = strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 6)
			//
			pair, _, _ = kv.Get("subTestID", nil)
			value, _ = strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 6)

			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)
			time.Sleep(1*time.Second)
		})

		Convey("新主题先发布再订阅", func() {
			stopPubTestTool()
			stopSubTestTool()
			time.Sleep(1*time.Second)
			// 发布新主题
			newTestTopic := "NowtestPubSubTopic001"
			startPubTestTool("pubTestID", newTestTopic, udp_address)
			time.Sleep(1*time.Second)
			// 发布 3条消息
			req_pubmsgs(udp_address, 3)
			// 开始订阅
			startSubTestTool("subTestID", newTestTopic)
			time.Sleep(2*time.Second)

			// 检查结果
			pair, _, _ := kv.Get("pubTestID", nil)
			value, _ := strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)
			//
			pair, _, _ = kv.Get("subTestID", nil)
			value, _ = strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)

			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)
			time.Sleep(1*time.Second)

		})

		Convey("新主题先订阅再发布", func() {
			stopPubTestTool()
			stopSubTestTool()
			time.Sleep(1*time.Second)
			// 订阅新主题
			newTestTopic := "NowtestPubSubTopic002"
			startSubTestTool("subTestID", newTestTopic)
			time.Sleep(2 * time.Second)
			// 发布新主题
			startPubTestTool("pubTestID", newTestTopic, udp_address)
			time.Sleep(2*time.Second)
			req_pubmsgs(udp_address, 3)
			time.Sleep(1*time.Second)

			// 检查结果
			pair, _, _ := kv.Get("pubTestID", nil)
			value, _ := strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)
			//
			pair, _, _ = kv.Get("subTestID", nil)
			value, _ = strconv.Atoi(string(pair.Value))
			So(value, ShouldEqual, 3)

			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)
			time.Sleep(1*time.Second)
		})

		Convey("MQ异常及恢复测试", func() {

			Convey("通讯中恢复", func() {
				stopPubTestTool()
				stopSubTestTool()
				time.Sleep(1 * time.Second)
				startPubTestTool("pubTestID", testTopic, udp_address)
				startSubTestTool("subTestID", testTopic)
				time.Sleep(5 * time.Second)

				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)

				// 中断 MQ
				stopNS("./tools_fortest/ns-cfg-1.conf")
				time.Sleep(1 * time.Second)
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)

				// 恢复MQ
				startNS("./tools_fortest/ns-cfg-1.conf")

				//***** MQ 快速restart, serviceUtil可能未感知,订阅方也不知道变化,
				//*****  短时间会造成'死'等, go nats库有ping机制(5sx3),并在30s后reconnect
				//***** 这里设置超过30s 以覆盖这种情况
				//time.Sleep(32 * time.Second)
				time.Sleep(32 * time.Second)

				// 检查结果
				var a1,a2 int
				if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil {
					a1, _ = strconv.Atoi(string(pair.Value))
				}
				// 检查是否接受了 6条消息
				if pair, _, err := kv.Get("subTestID", nil); err ==nil && pair !=nil {
					a2, _ = strconv.Atoi(string(pair.Value))
				}

				//clean kv
				kv.Delete("pubTestID", nil)
				kv.Delete("subTestID", nil)
				time.Sleep(1 * time.Second)

				So(a1, ShouldEqual, 6)
				So(a2, ShouldEqual, 6)
			})

			Convey("异常时候增加(变更)订阅者", func() {

				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)
				// MQ 中断
				stopNS("./tools_fortest/ns-cfg-1.conf")
				time.Sleep(2 * time.Second)

				//增加订阅方
				//stopSubTestTool()
				// start the second sub process
				startSubTestTool("subTestID-02", testTopic)
				time.Sleep(3 * time.Second)
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)

				// 恢复MQ
				startNS("./tools_fortest/ns-cfg-1.conf")
				time.Sleep(15 * time.Second)

				//log.Info("Sleep 100s")
				//time.Sleep(100 * time.Second)
				// 检查结果
				pair, _, err := kv.Get("pubTestID", nil)
				value, _ := strconv.Atoi(string(pair.Value))
				So(value, ShouldEqual, 6)
				//
				a1, a2 := 0,0
				if pair, _, err = kv.Get("subTestID-02", nil); err ==nil && pair !=nil {
					a1, _ = strconv.Atoi(string(pair.Value))
				}

				if pair, _, err = kv.Get("subTestID", nil); err ==nil && pair !=nil {
					a2, _ = strconv.Atoi(string(pair.Value))
				}

				//clean kv
				kv.Delete("pubTestID", nil)
				kv.Delete("subTestID", nil)
				kv.Delete("subTestID-02", nil)

				log.Info("total recv msg: a1["+strconv.Itoa(a1)+"], a2["+strconv.Itoa(a2) +"]")
				time.Sleep(1 * time.Second)
				a1+=a2
				So(a1, ShouldEqual, 6)
			})

			Convey("异常时发布新主题并订阅", func() {
				stopPubTestTool()
				stopSubTestTool()
				time.Sleep(1 * time.Second)

				// 中断 MQ
				stopNS("./tools_fortest/ns-cfg-1.conf")
				time.Sleep(1 * time.Second)
				//
				newTestTopic := "NowtestPubSubTopic003"
				//发布订阅新主题
				startPubTestTool("pubTestID", newTestTopic, udp_address)
				startSubTestTool("subTestID", newTestTopic)
				time.Sleep(2 * time.Second)
				req_pubmsgs(udp_address, 3)
				// 恢复MQ
				startNS("./tools_fortest/ns-cfg-1.conf")
				time.Sleep(15* time.Second)

				// 检查结果
				var a1,a2 int
				if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil{
					a1, _ = strconv.Atoi(string(pair.Value))
				}
				//
				if pair, _, err := kv.Get("subTestID", nil); err ==nil && pair !=nil {
					a2, _ = strconv.Atoi(string(pair.Value))
				}
				//clean kv
				kv.Delete("pubTestID", nil)
				kv.Delete("subTestID", nil)
				time.Sleep(1 * time.Second)

				So(a1, ShouldEqual, 3)
				So(a2, ShouldEqual, 3)
			})

			Convey("异常时发布新主题,恢复后订阅", func() {
				stopPubTestTool()
				stopSubTestTool()
				time.Sleep(1 * time.Second)
				// 中断 MQ
				stopNS("./tools_fortest/ns-cfg-1.conf")
				time.Sleep(1 * time.Second)
				//
				newTestTopic := "NowtestPubSubTopic004"
				//发布订阅新主题
				startPubTestTool("pubTestID", newTestTopic, udp_address)
				time.Sleep(2 * time.Second)
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)
				// 恢复MQ
				startNS("./tools_fortest/ns-cfg-1.conf")
				time.Sleep(10* time.Second)
				// 订阅新主题
				startSubTestTool("subTestID002", newTestTopic)
				time.Sleep(5 * time.Second)

				// 检查结果
				var a1,a2 int
				if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil {
					a1, _ = strconv.Atoi(string(pair.Value))
				}

				//
				if pair, _, err := kv.Get("subTestID002", nil); err ==nil && pair !=nil {
					a2, _ = strconv.Atoi(string(pair.Value))
				}
				//clean kv
				kv.Delete("pubTestID", nil)
				kv.Delete("subTestID", nil)
				time.Sleep(1 * time.Second)

				So(a1, ShouldEqual, 3)
				So(a2, ShouldEqual, 3)



			})
		})
	})

	stopPubTestTool()
	stopSubTestTool()
	time.Sleep(2 * time.Second)

	Convey("双MQ结点测试", t, func() {

		Convey("单结点通讯时新增一个结点", func() {
			startPubTestTool("pubTestID", testTopic, udp_address)
			startSubTestTool("subTestID", testTopic)
			time.Sleep(2 * time.Second)

			// 发布 3条消息
			req_pubmsgs(udp_address, 3)
			time.Sleep(1 * time.Second)
			//新增一个MQ结点
			startNS("./tools_fortest/ns-cfg-2.conf")
			time.Sleep(16 * time.Second)

			req_pubmsgs(udp_address, 3)
			time.Sleep(2 * time.Second)


			// 检查结果
			var a1,a2 int
			if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil {
				a1, _ = strconv.Atoi(string(pair.Value))
			}

			// 检查是否接受了 6条消息
			if pair, _, err := kv.Get("subTestID", nil); err ==nil && pair !=nil {
				a2, _ = strconv.Atoi(string(pair.Value))
			}
			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)
			// 恢复单MQ结点
			stopNS("./tools_fortest/ns-cfg-2.conf")
			time.Sleep(3 * time.Second)

			So(a1, ShouldEqual, 6)
			So(a2, ShouldEqual, 6)
		})

		Convey("新主题先发布再订阅", func() {

			//新增一个MQ结点(两个MQ结点)
			startNS("./tools_fortest/ns-cfg-2.conf")
			time.Sleep(15 * time.Second)

			stopPubTestTool()
			stopSubTestTool()
			time.Sleep(1 * time.Second)

			// 新主题
			newTestTopic := "NowtestPubSubTopic005"
			time.Sleep(1 * time.Second)
			startPubTestTool("pubTestID", newTestTopic, udp_address)
			time.Sleep(3 * time.Second)
			// 发布 3条消息
			req_pubmsgs(udp_address, 3)
			time.Sleep(1 * time.Second)
			// 开始订阅
			startSubTestTool("subTestID", newTestTopic)
			time.Sleep(3 * time.Second)

			// 检查结果
			var a1,a2 int
			if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil {
				a1, _ = strconv.Atoi(string(pair.Value))
				log.Info("pubTest kv:"+string(pair.Value))
			}

			// 检查是否接受了 6条消息
			if pair, _, err := kv.Get("subTestID", nil); err ==nil && pair !=nil {
				a2, _ = strconv.Atoi(string(pair.Value))
				log.Info("subTest kv:"+string(pair.Value))
			}
			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)

			// 恢复单MQ结点
			stopNS("./tools_fortest/ns-cfg-2.conf")
			time.Sleep(2 * time.Second)

			So(a1, ShouldEqual, 3)
			So(a2, ShouldEqual, 3)

		})

		Convey("新主题先订阅再发布", func() {
			//新增一个MQ结点
			startNS("./tools_fortest/ns-cfg-2.conf")
			time.Sleep(15 * time.Second)

			stopPubTestTool()
			stopSubTestTool()
			time.Sleep(1 * time.Second)
			// 新主题
			newTestTopic := "NowtestPubSubTopic006"

			// 先订阅后发布
			startSubTestTool("subTestID", newTestTopic)
			time.Sleep(3 * time.Second)
			startPubTestTool("pubTestID", newTestTopic, udp_address)
			time.Sleep(3 * time.Second)

			// 发布3条消息
			req_pubmsgs(udp_address, 3)
			time.Sleep(1 * time.Second)

			// 检查结果
			var a1,a2 int
			if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil {
				a1, _ = strconv.Atoi(string(pair.Value))
			}

			// 检查是否接受了 6条消息
			if pair, _, err := kv.Get("subTestID", nil); err ==nil && pair !=nil {
				a2, _ = strconv.Atoi(string(pair.Value))
			}

			//clean kv
			kv.Delete("pubTestID", nil)
			kv.Delete("subTestID", nil)

			// 恢复单MQ结点
			stopNS("./tools_fortest/ns-cfg-2.conf")
			time.Sleep(2 * time.Second)

			So(a1, ShouldEqual, 3)
			So(a2, ShouldEqual, 3)
		})

		Convey("单MQ异常及恢复测试", func() {

			//新增一个MQ结点
			startNS("./tools_fortest/ns-cfg-2.conf")
			time.Sleep(15 * time.Second)

			Convey("通讯中删除一个MQ结点及恢复", func() {
				stopPubTestTool()
				stopSubTestTool()
				time.Sleep(2 * time.Second)
				startPubTestTool("pubTestID", testTopic, udp_address)
				startSubTestTool("subTestID", testTopic)
				time.Sleep(3 * time.Second)
				// 发布 3条消息
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)
				// 删除一个MQ结点
				stopNS("./tools_fortest/ns-cfg-2.conf")
				time.Sleep(2 * time.Second)
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)

				//恢复到2个MQ结点
				startNS("./tools_fortest/ns-cfg-2.conf")
				time.Sleep(15 * time.Second)
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)

				// 检查结果
				var a1,a2 int
				if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil {
					a1, _ = strconv.Atoi(string(pair.Value))
				}

				// 检查是否接受了 6条消息
				if pair, _, err := kv.Get("subTestID", nil); err ==nil && pair !=nil {
					a2, _ = strconv.Atoi(string(pair.Value))
				}

				//clean kv
				kv.Delete("pubTestID", nil)
				kv.Delete("subTestID", nil)
				time.Sleep(1 * time.Second)

				So(a1, ShouldEqual, 9)
				So(a2, ShouldEqual, 9)

			})

			Convey("删除一个MQ结点,新发布主题再订阅", func() {
				stopPubTestTool()
				stopSubTestTool()
				time.Sleep(1 * time.Second)

				// 删除一个MQ结点
				stopNS("./tools_fortest/ns-cfg-2.conf")
				time.Sleep(1 * time.Second)

				// 发布新主题
				newTestTopic := "NowtestPubSubTopic007"
				startPubTestTool("pubTestID", newTestTopic, udp_address)
				time.Sleep(3 * time.Second)
				// 发布 3条消息
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)
				// 开始订阅
				startSubTestTool("subTestID", newTestTopic)
				time.Sleep(3 * time.Second)

				// 检查结果
				var a1,a2 int
				if pair, _, err := kv.Get("pubTestID", nil); err ==nil && pair !=nil {
					a1, _ = strconv.Atoi(string(pair.Value))
				}

				// 检查是否接受了 6条消息
				if pair, _, err := kv.Get("subTestID", nil); err ==nil && pair !=nil {
					a2, _ = strconv.Atoi(string(pair.Value))
				}

				//clean kv
				kv.Delete("pubTestID", nil)
				kv.Delete("subTestID", nil)

				// 恢复双MQ结点
				startNS("./tools_fortest/ns-cfg-2.conf")
				time.Sleep(5 * time.Second)

				So(a1, ShouldEqual, 3)
				So(a2, ShouldEqual, 3)
			})

			Convey("删除一个MQ结点,先订阅新主题再发布", func() {
				stopPubTestTool()
				stopSubTestTool()
				time.Sleep(1 * time.Second)

				// 删除一个MQ结点
				stopNS("./tools_fortest/ns-cfg-2.conf")
				time.Sleep(2 * time.Second)

				// 订阅新主题
				newTestTopic := "NowtestPubSubTopic006"
				startSubTestTool("subTestID", newTestTopic)
				time.Sleep(3 * time.Second)
				// 发布新主题
				startPubTestTool("pubTestID", newTestTopic, udp_address)
				time.Sleep(3 * time.Second)
				req_pubmsgs(udp_address, 3)
				time.Sleep(1 * time.Second)

				// 检查结果
				pair, _, _ := kv.Get("pubTestID", nil)
				value, _ := strconv.Atoi(string(pair.Value))
				So(value, ShouldEqual, 3)
				//
				pair, _, _ = kv.Get("subTestID", nil)
				value, _ = strconv.Atoi(string(pair.Value))
				So(value, ShouldEqual, 3)

				//clean kv
				kv.Delete("pubTestID", nil)
				kv.Delete("subTestID", nil)

				// 恢复双MQ结点
				startNS("./tools_fortest/ns-cfg-2.conf")
				time.Sleep(3 * time.Second)
			})
		})

		// tear down
		stopPubTestTool()
		stopSubTestTool()
	})
}
