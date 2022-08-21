package main

import (
	"bufio"
	"flag"
	"fmt"
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	"gworld/git/GoFrameWork/common/utils/mq"
	"github.com/jmoiron/sqlx"
	log "github.com/xiaomi-tc/log15"
	"os"
	"os/signal"
	"strconv"
)

var (
	ip       = "127.0.0.1"
	port     = 3306
	userName = "root2"
	password = ""
	dbName   = "trans"

	pubC chan os.Signal
)

type Callable struct{}

func (c *Callable) ExecuteLocalTrans(tx *sqlx.Tx) error {
	fmt.Println("Execute ExecuteLocalTrans() ... ")
	//事务环境已经开启，直接使用tx更新DB即可
	// 2.1 扣减货款
	//  tx.Exec("update user_account set balance = balance - 100 where id = 101010")
	// 2.2 记录支付流水
	//  tx.Exec("insert into account_log(src,dst,target_id,amount,op) values(101010,1,2100,100,'订单扣款')")
	// 2.3 修改订单状态为已支付
	//  tx.Exec("update order set pay_status = paid where id = 2100")
	//无需调用tx.Commit()

	return nil
}

//发送事务性消息共3步：
//1.调用 InitWithExistDB, InitWithMySQLCfg 或者 InitWithDBCfg 三个函数中的任何一个初始化数据源。该数据源包含表trans_msg。
//2.在 ExecuteLocalTrans 函数内完成自己的本地事务逻辑，注意：并不需要提交本地事务；
//3.调用 PublishTransMsg 函数发送事务性消息；
//----------下面通过一个跨微服务的订单扣款后打印发票的业务来描述如何使用事务性消息----------
//用户账户表 user_account(id bigint,balance number(16,2))
//账户变化流水表 account_log(id bigint,src bigint,dst bigint,target_id bigint,amount number(16,2),op varchar(256),add_time date_time)
//订单表 order(id bigint,uid bigint,amount number(16,2),status tinyint)
//user_id= 101010, order_id=2100, 收款方账户id=1
//1.调用err := mq.InitWithMySQLCfg(ip, port, userName, password, dbName)或者使用已有的数据源调用 InitWithExistDB 函数
//  完成数据源的初始化；
//
//2.在回调函数内完成本地扣款业务。
// 2.1 扣减货款
//  tx.Exec("update user_account set balance = balance - 100 where id = 101010")
// 2.2 记录支付流水
//  tx.Exec("insert into account_log(src,dst,target_id,amount,op) values(101010,1,2100,100,'订单扣款')")
// 2.3 修改订单状态为已支付
//  tx.Exec("update order set pay_status = paid where id = 2100")
//
//3.调用 err := mq.PublishTransMsg(topic, data, &Callable{}) 函数发送事务性消息通知订阅者打印发票。

func main() {
	var clientID, topic string
	flag.StringVar(&clientID, "id", "TransTopicID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic, "tp", "TransTopic", "The topic name")
	flag.Parse()

	log.SetOutLevel(log.LvlInfo)

	qbh.SetDelayMsgDir("data/pub")
	serviceUtil := su.GetInstanceWithOptions()
	qbh.Init(serviceUtil, clientID)
	//初始化DB配置
	err := mq.InitWithMySQLCfg(ip, port, userName, password, dbName)
	if err != nil {
		log.Error("Init db failed!", "error", err)
		return
	}

	waitPubToExit()

	j := 0
	for {
		log.Info("Press any key to send:")
		readstring := bufio.NewReader(os.Stdin)
		readstring.ReadString('\n')

		for i := 0; i < 2; i++ {
			msg := "msg-" + topic + "-" + clientID + "-" + strconv.Itoa(j)
			j++
			data := []byte(msg)
			//发送事务性消息
			if err := mq.PublishTransMsg(topic, data, &Callable{}); err != nil {
				log.Error(" \""+string(data[:])+"\"", "error=", err)
			} else {
				log.Info(msg)
			}
		}
	}

}

func waitPubToExit() {
	pubC = make(chan os.Signal, 1)
	signal.Notify(pubC, os.Interrupt, os.Kill)

	go func() {
		for _ = range pubC {
			// just close connection, no need unsubscribe !!
			qbh.CloseConn()
			os.Exit(1)
		}
	}()
}
