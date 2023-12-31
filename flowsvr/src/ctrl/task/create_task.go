package task

import (
	"errors"
	"fmt"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/constant"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/ctrl/ctrlmodel"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/db"
	"github.com/bigfacecat2333/async_work_processor/taskutils/rpc/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niuniumart/gosdk/handler"
	"github.com/niuniumart/gosdk/martlog"
)

// CreateTaskHandler 接口处理handler request和response都在这里
type CreateTaskHandler struct {
	Req    model.CreateTaskReq
	Resp   model.CreateTaskResp
	UserId string
}

// CreateTask 接口
func CreateTask(ctx *gin.Context) {
	var hd CreateTaskHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		ctx.JSON(http.StatusOK, hd.Resp)
	}()
	// 获取用户Id
	hd.UserId = ctx.Request.Header.Get(constant.HEADER_USERID)
	// 解析请求包
	// shouldBind会根据请求头中的Content-Type(JSON)自动推断请求数据类型，然后做相应的解析
	if err := ctx.ShouldBind(&hd.Req); err != nil {
		martlog.Errorf("CreateTask shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}
	// 执行处理函数, 这里会调用对应的HandleInput和HandleProcess，往下看
	handler.Run(&hd)
}

// HandleInput 参数检查, 比如检查任务类型，优先级参数，实际每个参数都要检查，代码为了省事，只留了2个做例子
func (p *CreateTaskHandler) HandleInput() error {
	if p.Req.TaskData.TaskType == "" {
		martlog.Errorf("input invalid")
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return constant.ERR_HANDLE_INPUT
	}
	if p.Req.TaskData.Priority != nil {
		if *p.Req.TaskData.Priority > db.MAX_PRIORITY || *p.Req.TaskData.Priority < 0 {
			p.Resp.Code = constant.ERR_INPUT_INVALID
			return constant.ERR_HANDLE_INPUT
		}
	}
	return nil
}

// HandleProcess 处理函数
// 根据SQL，构成对应的需要的结构体
func (p *CreateTaskHandler) HandleProcess() error {
	var err error
	var taskTableName string
	// 拿到任务位置信息，这里其实是预先考虑了分表，将数据插入pos表中ScheduleEndPos对应的位置。
	// 目前我们并没有实现分表，所以 ScheduleEndPos 和 ScheduleBeginPos始终都等于1
	var taskPos *db.TaskPos
	taskTableName = db.GetTaskTableName(p.Req.TaskData.TaskType)
	taskPos, err = db.TaskPosNsp.GetTaskPos(db.DB, taskTableName)
	if err != nil {
		p.Resp.Code = constant.ERR_GET_TASK_POS
		martlog.Errorf("db.TaskPosNsp.GetTaskPos err: %s", err.Error())
		return err
	}
	if taskPos == nil {
		martlog.Errorf("db.TaskPosNsp.GetTaskPos failed. TaskTableName : %s", taskTableName)
		return errors.New("Get task pos failed.  TaskTableName : " + taskTableName)
	}
	// 拿到任务类型配置信息
	taskCfg, err := db.TaskTypeCfgNsp.GetTaskTypeCfg(db.DB, p.Req.TaskData.TaskType)
	if err != nil {
		p.Resp.Code = constant.ERR_GET_TASK_SET_POS_FROM_DB
		martlog.Errorf("visit t_task_type_cfg err %s", err.Error())
		return err
	}
	// 拿到任务位置信息，这里其实是预先考虑了分表，未来将数据插入pos表中ScheduleEndPos对应的位置。
	scheduleEndPosStr := fmt.Sprintf("%d", taskPos.ScheduleEndPos)
	if err != nil {
		martlog.Errorf("db.TaskPosNsp.GetTaskPos %s", err.Error())
		return err
	}
	// 构建要插入db的任务信息
	var task = new(db.Task)
	p.Req.TaskData.MaxRetryNum = taskCfg.MaxRetryNum
	p.Req.TaskData.MaxRetryInterval = taskCfg.MaxRetryInterval
	// 创建时的时间，就是一开始的调度顺序，调度查询时会根据orderTime由小到大排序
	p.Req.TaskData.OrderTime = time.Now().Unix()
	// 如果有优先级，就减去优先级，这样orderTime就会变小，查询时就会优先查询
	if p.Req.TaskData.Priority != nil {
		p.Req.TaskData.OrderTime -= int64(*p.Req.TaskData.Priority)
	}
	// 填充了任务信息
	err = ctrlmodel.FillTaskModel(&p.Req.TaskData, task, scheduleEndPosStr)
	if err != nil {
		p.Resp.Code = constant.ERR_CREATE_TASK
		martlog.Errorf("db.TaskPosNsp.GetTaskPos %s", err.Error())
		return err
	}
	// 调用封装好的数据库操作，在数据库插入了一行任务记录，也就是将task插入db insert into t_task_1 values(...)
	// 创建任务记录
	err = db.TaskNsp.Create(db.DB, p.Req.TaskData.TaskType, scheduleEndPosStr, task)
	if err != nil {
		martlog.Errorf("db.TaskNsp.Create %s", err.Error())
		p.Resp.Code = constant.ERR_CREATE_TASK
		return err
	}
	// 返回用户任务id
	p.Resp.TaskId = task.TaskId
	return nil
}
