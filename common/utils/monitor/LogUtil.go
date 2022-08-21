package monitor

import (
	"encoding/json"
	"gworld/git/GoFrameWork/pb/ComMessage"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/xiaomi-tc/log15"
	"strconv"
	"strings"
	"sync"
	"time"
)

/////////****** BEGIN定义监控标量类型，四种BEGIN ******////////
type MetricType string // 标量类型,参考 https://www.jianshu.com/p/3a76d329feaf

// 对形如[*_begin|*_end|*_err|*_suc]这些后缀结束的标量默认增加性能（耗时）监控，标量名称为*_duration。
// 例如：监控了user_login_begin，如果再监控user_login_end会自动监控user_login_duration；
// 监控了sending_sms_begin，如果再监控sending_sms_end或sending_sms_err会自动监控sending_sms_duration。
const (
	Counter   MetricType = "counter"   // 随时间累计，只增不减。后缀一般为：_total,_count
	Gauge     MetricType = "gauge"     // 随时间变化，可增可减。后缀一般为：_duration,_interval
	Histogram MetricType = "histogram" // 直方图
	Summary   MetricType = "summary"   // 分位图
)

func (m MetricType) String() string {
	switch m {
	case Counter:
		return "counter"
	case Gauge:
		return "gauge"
	case Histogram:
		return "histogram"
	case Summary:
		return "summary"
	default:
		panic("Unknown metric type!")
	}
}

/////////****** END定义监控标量类型END ******////////
const (
	ThresholdExpiredDuration = 300000 // 阈值：失效周期（毫秒）
	ThresholdDeletedCount    = 100000 // 阈值：Map中删除记录数（达到或超过此值重新分配Map）
)

var MetricProject = ComMessage.CBSType_Base //项目名
var MetricService = ""                      //服务名
var hostname = "unknown"                    //主机名

var performanceMap map[string]int64 // k:traceId,v:时间（毫秒）
var deletedCount = 0                // 标记性能监控Map中删除的记录数
var once sync.Once
var mutex sync.Mutex

type Metric struct {
	Metric     string  `json:"metric"`
	MetricType string  `json:"metricType"`
	Time       int64   `json:"time"`
	Value      float64 `json:"value"`
	Host       string  `json:"host"`
	Project    string  `json:"project"`
	Service    string  `json:"service"`
	Method     string  `json:"method"`
	Filter     string  `json:"filter"`
	BizCode    string  `json:"bizCode"`
	ErrCode    int     `json:"errCode"`
	Msg        string  `json:"msg"`
	TraceId    string  `json:"traceId"`
	Remark     string  `json:"remark"`
}

func init() {
	info, _ := host.Info()
	hostname = info.Hostname
}

// 初始化项目和服务名称。
// project:项目名称，[CBSType_Own|CBSType_Woda|CBSType_Jifanfei|CBSType_ZXX|...]
// service:服务名称，["base"|"auth"|"sms"|"pay"|...]
func InitCfg(project ComMessage.CBSType, service string) {
	MetricProject = project
	MetricService = service
	performanceMap = make(map[string]int64, 16)

	once.Do(StartTicker)
}

// 启动定时任务
func StartTicker() {
	mTicker := time.NewTicker(30 * time.Second) // 每30秒扫描一次
	cTicker := time.NewTicker(5 * time.Minute)  // 每5分钟清理一次

	go func(mt, ct *time.Ticker) {
		for {
			select {
			case <-mt.C:
				MonitorHost() // 系统监控

			case <-ct.C:
				clearPerformanceMap() // 清理性能监控Map中过期的数据
			}
		}
	}(mTicker, cTicker)
}

// 只输出监控标量的必要信息
// metric:必填，标量名称，如：req_login_count,req_login_duration,req_auth_count
// method:必填，方法名称，如：DALGrpcHandler,EC_USCH_GetUserSignList
// mType:必填，标量类型[Counter|Gauge|Histogram|Summary]
// value:必填，标量值
func MonitorBriefly(metric, method string, mType MetricType, value float64) error {
	return Monitor(metric, method, mType, value, "", "", "", 0, "", "")
}

// 输出监控标量
// metric:必填，标量名称，如：req_login_count,req_login_duration,req_auth_count
// method:必填，方法名称，如：DALGrpcHandler,EC_USCH_GetUserSignList
// mType:必填，标量类型[Counter|Gauge|Histogram|Summary]
// value:必填，标量值
// msg:可选，日志内容
// filter:可选，过滤规则，[!db|db|third|...]
// bizCode:可选，6位业务编吗，[000000|100000|200000|...]
// errCode:错误编码，0-正确，非0-错误
// traceId:函数调用跟踪ID
// remark:标注，扩展字段
func Monitor(metric, method string, mType MetricType, value float64, msg, filter, bizCode string, errCode int, traceId, remark string) error {
	// *Host在MQ接受消息时设置
	m := Metric{Metric: metric, Method: method, MetricType: mType.String(), Value: value, Msg: msg, Filter: filter,
		BizCode: bizCode, ErrCode: errCode, TraceId: traceId, Remark: remark}

	return MonitorS(&m)
}

// 输出监控标量
func MonitorS(m *Metric) error {
	m.Project = ComMessage.CBSType_name[int32(MetricProject)]
	m.Service = MetricService

	if strings.TrimSpace(m.Service) == "" {
		return errors.New("Set MetricService's value first!")
	}

	if strings.TrimSpace(m.Metric) == "" {
		return errors.New("Invalid param[metric]!")
	}

	if strings.TrimSpace(m.Method) == "" {
		return errors.New("Invalid param[method]!")
	}

	m.Time = time.Now().UnixNano() / 1e6 // 日志时间（毫秒）
	if m.BizCode == "" {
		m.BizCode = "000000"
	}
	if m.TraceId == "" {
		m.TraceId = "0"
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	log15.MetaDebug(string(jsonBytes), log15.BaseMonitor, log15.BaseMonitor.String())
	durationFilter(m.Metric, m.Method, m.TraceId)

	return nil
}

// 根据监控标量名称自动加入性能监控。对形如[*_begin|*_end|*_err|*_suc]这些后缀结束的标量默认增加性能（耗时）监控，
// 标量名称为*_duration。例如：监控了user_login_begin，如果再监控user_login_end会自动监控user_login_duration；
// 监控了sending_sms_begin，如果再监控sending_sms_end或sending_sms_err会自动监控sending_sms_duration。
func durationFilter(metric, method, traceId string) {
	if strings.TrimSpace(metric) == "" {
		return
	}

	// 后缀匹配性能监控设置开始时间
	if strings.HasSuffix(metric, "_begin") || strings.HasSuffix(metric, "_bgn") || strings.HasSuffix(metric, "_start") || strings.HasSuffix(metric, "_in") {
		mutex.Lock()
		_, ok := performanceMap[traceId]
		if !ok {
			performanceMap[traceId] = time.Now().UnixNano() / 1e6 // 当前时间（毫秒）
		}
		mutex.Unlock()
	}

	// 后缀匹配性能监控计算耗时
	if strings.HasSuffix(metric, "_end") || strings.HasSuffix(metric, "_success") || strings.HasSuffix(metric, "_suc") || strings.HasSuffix(metric, "_error") || strings.HasSuffix(metric, "_err") || strings.HasSuffix(metric, "_out") {
		v, ok := performanceMap[traceId]
		if ok {
			duration := time.Now().UnixNano()/1e6 - v
			index := strings.LastIndex(metric, "_")
			dMetric := string([]byte(metric[0:index])) + "_duration" //new metric *_duration
			//自动加入性能监控
			vMethod := "auto_duration_monitor"
			if method != "" {
				vMethod = method
			}
			MonitorBriefly(dMetric, vMethod, Gauge, float64(duration))
			mutex.Lock()
			delete(performanceMap, traceId) // 性能监测结束
			deletedCount++
			mutex.Unlock()
		}
	}
}

// 清理性能监控Map中过期的数据
func clearPerformanceMap() {
	now := time.Now().UnixNano() / 1e6 // 当前时间（毫秒）
	mutex.Lock()
	for k, v := range performanceMap {
		if now - v >= ThresholdExpiredDuration { // 超时阈值:5分钟（300000毫秒）
			delete(performanceMap, k) // 清理超时的无效数据
			deletedCount++
		}
	}
	if deletedCount >= ThresholdDeletedCount && len(performanceMap) == 0 {
		performanceMap = nil // for GC
		performanceMap = make(map[string]int64, 16) // 重新初始化Map
		deletedCount = 0 // 重置计数
	}
	mutex.Unlock()
}

// 监控CPU使用情况
func MonitorCpu() {
	infoSata, _ := cpu.Info() //总体信息
	msg := strconv.FormatInt(int64(infoSata[0].Cores), 10)
	filter := hostname

	usedInfo, _ := cpu.Percent(time.Duration(time.Second), false)
	value := usedInfo[0]
	Monitor("m_base_process", "cpu", Gauge, value, msg, filter, "000010", 0, "0", "")
}

// 监控内存使用情况
func MonitorMem() {
	info, _ := mem.VirtualMemory()
	msg := strconv.FormatInt(int64(info.Used), 10) + "/" + strconv.FormatInt(int64(info.Total), 10)
	filter := hostname
	value := info.UsedPercent

	Monitor("m_base_process", "mem", Gauge, value, msg, filter, "000010", 0, "0", "")
}

// 监控磁盘使用情况
func MonitorDisk() {
	info, _ := disk.Usage("/")
	msg := strconv.FormatInt(int64(info.Used), 10) + "/" + strconv.FormatInt(int64(info.Total), 10)
	filter := hostname
	value := info.UsedPercent

	Monitor("m_base_process", "disk", Gauge, value, msg, filter, "000010", 0, "0", "")
}

// 系统监控
func MonitorHost() {
	MonitorCpu()
	MonitorMem()
	MonitorDisk()
}
