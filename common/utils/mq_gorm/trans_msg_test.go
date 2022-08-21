package mq_gorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"testing"
)

var (
	ip       = "127.0.0.1"
	port     = 3306
	userName = "root"
	password = ""
	dbName   = "trans"
)

type Callable struct{}

func (c *Callable) ExecuteLocalTrans(tx *gorm.DB) error {
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

func TestPublishTransMsg(t *testing.T) {
	InitWithMySQLCfg(ip, port, userName, password, dbName)
	err := PublishTransMsg("TransTopic", []byte("This is a test message"), &Callable{})
	if err != nil {
		t.Failed()
	}
}
