package mq_gorm

//发送事务性消息共3步：
//1.调用 InitWithExistDB, InitWithMySQLCfg 或者 InitWithDBCfg 三个函数中的任何一个初始化数据源。
//  该数据源包含表trans_msg，建表语句参考"trans_msg.sql"文件
//2.在 ExecuteLocalTrans 函数内完成自己的本地事务逻辑，注意：并不需要提交本地事务；
//3.调用 PublishTransMsg 函数发送事务性消息。

import (
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/juju/errors"
	"github.com/nats-io/nuid"
	log "github.com/xiaomi-tc/log15"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	RetryTimes = 16 //消息最大补发次数

	TransStateUnknown  = 0 // 事务状态：未知（默认）
	TransStateCommit   = 1 // 事务状态：提交
	TransStateRollback = 2 // 事务状态：回滚

	SendStateUnsent = 0 // 消息状态：未送出（默认）
	SendStateSent   = 1 // 消息状态：已送出
	SendStateFail   = 2 // 消息状态：尝试指定次数后失败

	FlagNormal  = 0 // 状态标记：正常（默认）
	FlagDeleted = 1 // 状态标记：已删除
)

var (
	// 消息补发时间间隔
	RetryTntervals = [RetryTimes]time.Duration{
		10 * time.Second, 20 * time.Second, 30 * time.Second, time.Minute,
		2 * time.Minute, 3 * time.Minute, 4 * time.Minute, 5 * time.Minute,
		6 * time.Minute, 7 * time.Minute, 8 * time.Minute, 9 * time.Minute,
		10 * time.Minute, 20 * time.Minute, 30 * time.Minute, time.Hour}

	once    sync.Once // 初始化时调用，处理由于进程意外终止而补偿发送未完成的情况
	transDb *gorm.DB  // 业务数据源
)

// 执行本地事务回调接口
type TransCallback interface {
	ExecuteLocalTrans(tx *gorm.DB) error
}

// mapping table trans_msg
type TransMsg struct {
	Id         int64     `gorm:"type:bigint(20);column:id;primary_key;auto_increment"`
	TransId    string    `gorm:"type:varchar(64);column:trans_id"`
	Topic      string    `gorm:"type:varchar(64);column:topic"`
	Tag        string    `gorm:"type:varchar(64);column:tag"`
	Data       []byte    `gorm:"type:mediumblob;column:data"`
	TransState int8      `gorm:"type:tinyint(4);column:trans_state"`
	SendState  int8      `gorm:"type:tinyint(4);column:send_state"`
	RetryTimes int       `gorm:"type:tinyint(4);column:retry_times"`
	DelFlag    int8      `gorm:"type:tinyint(4);column:del_flag"`
	AddTime    time.Time `gorm:"type:timestamp;column:add_time;default:CURRENT_TIMESTAMP"`
	UpdateTime time.Time `gorm:"type:timestamp;column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (TransMsg) TableName() string {
	return "trans_msg"
}

// 使用已经存在的DB初始化
func InitWithExistDB(db *gorm.DB) error {
	if transDb != nil {
		transDb.Close()
	}

	transDb = db
	transDb.LogMode(true)
	if !transDb.HasTable(&TransMsg{}) {
		return errors.New("Create table `trans_msg` first!")
	}
	once.Do(resendMsgFromDb)

	return nil
}

//使用配置参数初始化mysql
func InitWithMySQLCfg(ip string, port int, userName, password, dbName string) error {
	return InitWithDBCfg("mysql", ip, port, userName, password, dbName+"?charset=utf8&parseTime=true")
}

//使用配置参数初始化DB
func InitWithDBCfg(driverName, ip string, port int, userName, password, dbName string) error {
	if transDb != nil {
		transDb.Close()
	}

	url := strings.Join([]string{userName, ":", password, "@tcp(", ip, ":", strconv.Itoa(port), ")/", dbName}, "")
	db, err := gorm.Open(driverName, url)
	if err != nil {
		return err
	}

	transDb = db
	transDb.LogMode(true)
	if !transDb.HasTable(&TransMsg{}) {
		return errors.New("Create table `trans_msg` first!")
	}
	once.Do(resendMsgFromDb)

	return nil
}

// 发送事务性消息
func PublishTransMsg(topic string, data []byte, callback TransCallback) error {
	if transDb == nil {
		return errors.New("Init DB first!")
	}

	if callback == nil {
		return errors.New("TransCallback is nil!")
	}

	// 开启事务环境
	tx := transDb.Begin()
	// 先执行本地事务并保存事务性消息
	err := callback.ExecuteLocalTrans(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 保存事务性消息
	rowId, err := saveTransMsg(tx, nuid.Next(), topic, "*", data, TransStateCommit)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	// 事务执行成功，立即发送消息
	err = qbh.Publish(topic, data)
	if err != nil {
		log.Error("Send msg error, try to resend it! ", "topic", topic, "data", string(data), "error", err.Error())
		go resendMsg(rowId, topic, data, 0) // 启动异步补偿发送
	} else {
		markMsgSent(rowId)
	}

	return err
}

// 在事务环境下保存事务性消息
func saveTransMsg(tx *gorm.DB, transId, topic, tag string, data []byte, transState int8) (int64, error) {
	var msg = TransMsg{TransId: transId, Topic: topic, Tag: tag, Data: data, TransState: transState}
	db := tx.Create(&msg)
	if db.Error != nil {
		log.Error("In saveTransMsg() error! ", "error", db.Error.Error())
	}

	return msg.Id, db.Error
}

// 补偿发送未送出的事务性消息，一共补偿发送[16]次。
// id:表trans_msg行id
// topic:消息的主题
// data:消息体
// retriedTimes:已经补偿发送过的次数
func resendMsg(id int64, topic string, data []byte, retriedTimes int) {
	for i := 0; i < RetryTimes-retriedTimes; i++ {
		time.Sleep(RetryTntervals[i])   // 等待指定的时间间隔
		err := qbh.Publish(topic, data) // 补偿发送
		incMsgRetryTimes(id)            // 消息重试次数+1
		if err == nil {
			markMsgSent(id) // 标记消息已补偿送出

			return
		} else {
			log.Error("In resendMsg() error!", "times", i+1, "topic", topic, "data", string(data), "err", err)
		}
	}

	markMsgDeleted(id) // 删除补偿发送[16]次后依然失败的消息
}

// 消息重试次数+1
// id:表trans_msg行id
func incMsgRetryTimes(id int64) (int64, error) {
	db := transDb.Model(&TransMsg{Id: id}).UpdateColumn("retry_times", gorm.Expr("retry_times + ?", 1))
	if db.Error != nil {
		log.Error("In incMsgRetryTimes() error!", "msg id", id, "error", db.Error.Error())
	}

	return db.RowsAffected, db.Error
}

// 标记消息已送出
// id:表trans_msg行id
func markMsgSent(id int64) (int64, error) {
	db := transDb.Model(&TransMsg{Id: id}).UpdateColumn("send_state", SendStateSent)
	if db.Error != nil {
		log.Error("In markMsgSent() error!", "msg id", id, "error", db.Error.Error())
	}

	return db.RowsAffected, db.Error
}

// 标记消息已删除
// id:表trans_msg行id
func markMsgDeleted(id int64) (int64, error) {
	db := transDb.Model(&TransMsg{Id: id}).UpdateColumn("del_flag", FlagDeleted)
	if db.Error != nil {
		log.Error("In markMsgDeleted() error!", "error", db.Error.Error())
	}

	return db.RowsAffected, db.Error
}

// 回查那些本地事务已经执行成功但是消息还未送出的事务性消息并进行补偿发送
func resendMsgFromDb() {
	unsentMsgs := queryActiveUnsentMsg()
	for _, v := range unsentMsgs {
		log.Info("In resendMsgFromDb().", "msg", v)
		go resendMsg(v.Id, v.Topic, v.Data, v.RetryTimes) // 启动异步补偿发送
	}
}

// 查询未送出的事务性消息
func queryActiveUnsentMsg() []TransMsg {
	var msgs []TransMsg
	db := transDb.Where("trans_state = ? and send_state = ? and retry_times < ?", TransStateCommit, SendStateUnsent, RetryTimes).Find(&msgs)
	if db.Error != nil {
		log.Error("In queryActiveUnsentMsg() error!", "error", db.Error.Error())
	}

	return msgs
}
