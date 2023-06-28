package tasksdk

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/niuniumart/asyncflow/taskutils/constant"
	"github.com/niuniumart/asyncflow/taskutils/rpc"
	"github.com/niuniumart/asyncflow/taskutils/rpc/model"
	"github.com/niuniumart/gosdk/martlog"
	"github.com/niuniumart/gosdk/tools"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

const (
	DEFAULT_TIME_INTERVAL = 20 // for second
)

const (
	MAX_ERR_MSG_LEN = 256
)

var taskSvrHost, lockSvrHost string //new is host: for example http://127.0.0.1:41555

//InitSvr task svr host
func InitSvr(taskServerHost, lockServerHost string) {
	taskSvrHost, lockSvrHost = taskServerHost, lockServerHost
}

//TaskMgr struct short task mgr
type TaskMgr struct {
	InternelTime  time.Duration
	TaskType      string
	ScheduleLimit int
}

var mu sync.RWMutex
var MaxConcurrentRunTimes = 20
var concurrentRunTimes = MaxConcurrentRunTimes
var once sync.Once

var scheduleCfgDic map[string]*model.TaskScheduleCfg

func init() {
	scheduleCfgDic = make(map[string]*model.TaskScheduleCfg, 0)
}

//CycleReloadCfg func cycle reload cfg
func CycleReloadCfg() {
	for {
		now := time.Now()
		internelTime := time.Second * DEFAULT_TIME_INTERVAL
		next := now.Add(internelTime)
		martlog.Infof("schedule load cfg")
		sub := next.Sub(now)
		t := time.NewTimer(sub)
		<-t.C
		LoadCfg()
	}
}

//LoadCfg func load cfg
func LoadCfg() error {
	cfgList, err := taskRpc.GetTaskScheduleCfgList()
	if err != nil {
		martlog.Errorf("reload task schedule cfg err %s", err.Error())
		return err
	}
	for _, cfg := range cfgList.ScheduleCfgList {
		scheduleCfgDic[cfg.TaskType] = cfg
	}
	return nil
}

//Schedule func schedule
func (p *TaskMgr) Schedule() {
	taskRpc.Host = taskSvrHost
	once.Do(func() {
		// 初始化
		if p.ScheduleLimit != 0 {
			martlog.Infof("init ScheduleLimit : %d", p.ScheduleLimit)
			concurrentRunTimes = p.ScheduleLimit
			MaxConcurrentRunTimes = p.ScheduleLimit
		}
		if err := LoadCfg(); err != nil {
			msg := "load task cfg schedule err" + err.Error()
			martlog.Errorf(msg)
			fmt.Println(msg)
			os.Exit(1)
		}
		go func() {
			CycleReloadCfg()
		}()
	})
	rand.Seed(time.Now().Unix())
	for {
		cfg, ok := scheduleCfgDic[p.TaskType]
		if !ok {
			martlog.Errorf("scheduleCfgDic %s, not have taskType %s", tools.GetFmtStr(scheduleCfgDic), p.TaskType)
			return
		}
		internelTime := time.Second * time.Duration(cfg.ScheduleInterval)
		if cfg.ScheduleInterval == 0 {
			internelTime = time.Second * DEFAULT_TIME_INTERVAL
		}
		// 前后波动500ms
		step := RandNum(500)
		internelTime += time.Duration(step) * time.Millisecond
		martlog.Infof("taskType %s internelTime %v", p.TaskType, internelTime)
		fmt.Printf("taskType %s internelTime %v \n", p.TaskType, internelTime)
		t := time.NewTimer(internelTime)
		<-t.C
		martlog.Infof("schedule run %s task", p.TaskType)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					martlog.Errorf("In PanicRecover,Error:%s", err)
					//打印调用栈信息
					debug.PrintStack()
					buf := make([]byte, 2048)
					n := runtime.Stack(buf, false)
					stackInfo := fmt.Sprintf("%s", buf[:n])
					martlog.Errorf("panic stack info %s\n", stackInfo)
				}
			}()
			p.schedule()
		}()
	}
}

// 调度逻辑所在
func (p *TaskMgr) schedule() {
	defer func() {
		if err := recover(); err != nil {
			martlog.Errorf("In PanicRecover,Error:%s", err)
			//打印调用栈信息
			debug.PrintStack()
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			stackInfo := fmt.Sprintf("%s", buf[:n])
			martlog.Errorf("panic stack info %s\n", stackInfo)
		}
	}()
	// 这里是开始抢任务，分布式锁也应该加这里
	martlog.Infof("Start hold")
	// 占据一批任务
	taskIntfList, err := p.hold()
	if err != nil {
		martlog.Errorf("p.hold err %s", err.Error())
		return
	}
	martlog.Infof("End hold.")
	if len(taskIntfList) == 0 {
		martlog.Infof("no task to deal")
		return
	}
	// 获取这个任务类型的配置
	cfg, ok := scheduleCfgDic[p.TaskType]
	if !ok {
		martlog.Errorf("scheduleCfgDic %s, not have taskType %s", tools.GetFmtStr(scheduleCfgDic), p.TaskType)
		return
	}
	martlog.Infof("will do %d num task", len(taskIntfList))
	// 并发执行每个任务
	for _, taskIntf := range taskIntfList {
		taskInterface := taskIntf
		go func() {
			defer func() {
				if reErr := recover(); reErr != nil {
					martlog.Errorf("In PanicRecover,Error:%s", reErr)
					//打印调用栈信息
					debug.PrintStack()
					buf := make([]byte, 2048)
					n := runtime.Stack(buf, false)
					stackInfo := fmt.Sprintf("%s", buf[:n])
					martlog.Errorf("panic stack info %s\n", stackInfo)
				}
			}()
			run(taskInterface, cfg)
		}()
	}
}

var taskRpc rpc.TaskRpc
var ownerId string

func init() {
	ownerId = fmt.Sprintf("%v", uuid.New())
}

// 占据任务
func (p *TaskMgr) hold() ([]TaskIntf, error) {
	taskIntfList := make([]TaskIntf, 0)
	/**** Step1:拿到scheduleCfgDic中缓存的任务配置 ****/
	cfg, ok := scheduleCfgDic[p.TaskType]
	if !ok {
		martlog.Errorf("scheduleCfgDic %s, not have taskType %s", tools.GetFmtStr(scheduleCfgDic), p.TaskType)
		return nil, errors.New("tasktype not exist")
	}
	// 构造拉取任务列表的请求，其中拉取多少个，由cfg中的ScheduleLimit决定
	var reqBody = &model.HoldTasksReq{
		TaskType: p.TaskType,
		Limit:    cfg.ScheduleLimit,
	}
	/**** Step2:调用http请求，从flowsvr拉任务 ****/
	rpcTaskResp, err := taskRpc.HoldTasks(reqBody)
	if err != nil {
		martlog.Errorf("taskRpc.GetTaskList %s", err.Error())
		return taskIntfList, err
	}
	martlog.Infof("rpcTaskResp %+v", rpcTaskResp)
	if rpcTaskResp.Code != 0 {
		errMsg := fmt.Sprintf("taskRpc.GetTaskList resp code %d", rpcTaskResp.Code)
		martlog.Errorf(errMsg)
		return taskIntfList, errors.New(errMsg)
	}
	storageTaskList := rpcTaskResp.TaskList
	if len(storageTaskList) == 0 {
		return taskIntfList, nil
	}
	// 日志记录拉到了多少任务
	martlog.Infof("schedule will deal %d task", len(storageTaskList))
	taskIdList := make([]string, 0)
	/**** Step 3: 将数据库返回任务结构，转换为TaskIntf这个接口，方面后续操作 ****/
	for _, st := range storageTaskList {
		task, err := GetTaskInfoFromStorage(st)
		if err != nil {
			martlog.Errorf("GetTaskInfoFromStorage err %s", err.Error())
			return taskIntfList, err
		}
		task.Base().Status = int(constant.TASK_STATUS_PROCESSING)
		taskIntfList = append(taskIntfList, task)
		taskIdList = append(taskIdList, task.Base().TaskId)
	}
	if len(taskIdList) == 0 {
		return taskIntfList, nil
	}
	martlog.Infof("TaskType len(taskIntfList) %s %d", p.TaskType, len(taskIntfList))
	return taskIntfList, nil
}

/**
 * @Description: 处理单任务
 * @param taskInterface
 */
func run(taskInterface TaskIntf, cfg *model.TaskScheduleCfg) {
	martlog.Infof("Start run taskId %s... ", taskInterface.Base().TaskId)
	// defer函数会在当前函数结束时调用，用来更新Task状态，以及做一些异常处理
	defer func() {
		// 如果任务失败了
		if taskInterface.Base().Status == int(constant.TASK_STATUS_FAILED) {
			// HandleFailedMust是说这个收尾函数必须成功，不然不让关掉任务
			// 但此时其实任务重试次数已经结束了，所以如果这个操作失败，
			// 就把任务保持在执行中，等待执行时间过长重试，希望下次成功
			// 相当于是给了时间人工介入处理，不是关联逻辑
			err := taskInterface.HandleFailedMust()
			if err != nil {
				taskInterface.Base().Status = int(constant.TASK_STATUS_PROCESSING)
				martlog.Errorf("handle failed must err %s", err.Error())
				return
			}

			// HandleFinishError是失败处理函数，但这个处理无论是否生效，都可以结束任务
			err = taskInterface.HandleFinishError()
			if err != nil {
				martlog.Errorf("handle finish err %s", err.Error())
				return
			}
		}
		// 结束时无论成功和失败，都调用HandleFinish, 用来收尾
		if taskInterface.Base().Status == int(constant.TASK_STATUS_FAILED) ||
			taskInterface.Base().Status == int(constant.TASK_STATUS_SUCC) {
			taskInterface.HandleFinish()
		}
		// 更新任务状态
		err := taskInterface.SetTask()
		if err != nil {
			martlog.Errorf("schedule set task err %s", err.Error())
			// 再尝试一次，非必要流程
			err = taskInterface.SetTask()
			if err != nil {
				martlog.Errorf("schedule set task err twice.Err %s", err.Error())
			}
		}
		martlog.Infof("End run. releaseProcessRight")
	}()
	// 加载任务上下文
	err := taskInterface.ContextLoad()
	if err != nil {
		martlog.Errorf("taskid %s reload err %s", taskInterface.Base().TaskId, err.Error())
		taskInterface.Base().Status = int(constant.TASK_STATUS_PENDING)
		return
	}
	beginTime := time.Now()
	// 执行HandleProcess业务逻辑
	err = taskInterface.HandleProcess()
	// 若用户调用过SetContextLocal, 则会自动更新状态
	// taskInterface.ScheduleSetContext()
	// 记录调度信息
	taskInterface.Base().ScheduleLog.HistoryDatas = append(taskInterface.Base().ScheduleLog.HistoryDatas,
		taskInterface.Base().ScheduleLog.LastData)
	// 只记录最近三次操作信息
	if len(taskInterface.Base().ScheduleLog.HistoryDatas) > 3 {
		taskInterface.Base().ScheduleLog.HistoryDatas = taskInterface.Base().ScheduleLog.HistoryDatas[1:]
	}
	cost := time.Since(beginTime)
	martlog.Infof("taskId %s HandleProcess cost %v", taskInterface.Base().TaskId, cost)
	// 任务没设置为结果，就重置状态以待调度
	if taskInterface.Base().Status == int(constant.TASK_STATUS_PROCESSING) {
		taskInterface.Base().Status = int(constant.TASK_STATUS_PENDING)
	}
	taskInterface.Base().ScheduleLog.LastData.TraceId = fmt.Sprintf("%v", uuid.New())
	taskInterface.Base().ScheduleLog.LastData.Cost = fmt.Sprintf("%dms", cost.Milliseconds())
	taskInterface.Base().ScheduleLog.LastData.ErrMsg = ""
	// 计算延迟多少秒，从1,2,4....MaxRetryInterval，最大翻倍30次，再大怕溢出，同时减去优先时间
	taskInterface.Base().OrderTime = time.Now().Unix() - taskInterface.Base().Priority
	if err != nil {
		delayTime := cfg.MaxRetryInterval
		// 延时加到orderTime上去
		if delayTime != 0 {
			taskInterface.Base().OrderTime = time.Now().Unix() + int64(delayTime)
		}
		msgLen := tools.Min(len(err.Error()), MAX_ERR_MSG_LEN)
		errMsg := err.Error()[:msgLen]
		taskInterface.Base().ScheduleLog.LastData.ErrMsg = errMsg
		martlog.Errorf("task.HandleProcess err %s", err.Error())
		if taskInterface.Base().MaxRetryNum == 0 || taskInterface.Base().CrtRetryNum >= taskInterface.Base().MaxRetryNum {
			taskInterface.Base().Status = int(constant.TASK_STATUS_FAILED)
			return
		}
		if taskInterface.Base().Status != int(constant.TASK_STATUS_FAILED) {
			taskInterface.Base().CrtRetryNum++
		}
		return
	}
}

//RandNum func for rand num
func RandNum(num int64) int64 {
	step := rand.Int63n(num) + int64(1)
	flag := rand.Int63n(2)
	if flag == 0 {
		return -step
	}
	return step
}
