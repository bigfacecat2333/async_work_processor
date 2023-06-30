package task

import (
	"fmt"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/ctrl/ctrlmodel"
	"github.com/bigfacecat2333/async_work_processor/taskutils/rpc/model"
	"net/http"
	"time"

	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/constant"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/db"

	"github.com/niuniumart/gosdk/martlog"

	"github.com/niuniumart/gosdk/tools"

	"github.com/gin-gonic/gin"
	"github.com/niuniumart/gosdk/handler"
)

// GetTaskListHandler 接口处理handler
type GetTaskListHandler struct {
	Req    model.GetTaskListReq
	Resp   model.GetTaskListResp
	UserId string
}

// GetTaskList 接口
func GetTaskList(c *gin.Context) {
	var hd GetTaskListHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		martlog.Infof("GetTaskList "+
			"resp code %d, msg %s, taskCount %d", hd.Resp.Code, hd.Resp.Msg, len(hd.Resp.TaskList))
		c.JSON(http.StatusOK, hd.Resp)
	}()
	// 获取用户Id
	hd.UserId = c.Request.Header.Get(constant.HEADER_USERID)
	if err := c.ShouldBind(&hd.Req); err != nil {
		martlog.Errorf("GetTaskList shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}
	martlog.Infof("GetTaskList hd.Req %s", tools.GetFmtStr(hd.Req))
	handler.Run(&hd)
}

// HandleInput 参数检查
func (p *GetTaskListHandler) HandleInput() error {
	if p.Req.TaskType == "" {
		martlog.Errorf("input invalid")
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return constant.ERR_HANDLE_INPUT
	}
	if !db.IsValidStatus(db.TaskEnum(p.Req.Status)) {
		martlog.Errorf("input invalid status")
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return constant.ERR_HANDLE_INPUT
	}
	return nil
}

// HandleProcess 处理函数
func (p *GetTaskListHandler) HandleProcess() error {
	limit := p.Req.Limit
	if limit > constant.MAX_TASK_LIST_LIMIT {
		limit = constant.MAX_TASK_LIST_LIMIT
	}
	if limit == 0 {
		limit = constant.DEFAULT_TASK_LIST_LIMIT
	}
	taskTableName := db.GetTaskTableName(p.Req.TaskType)
	taskPos, err := db.TaskPosNsp.GetTaskPos(db.DB, taskTableName)
	if err != nil {
		martlog.Errorf("db.TaskPosNsp.GetRandomSchedulePos %s", err.Error())
		p.Resp.Code = constant.ERR_GET_TASK_SET_POS_FROM_DB
		return err
	}
	taskList, err := db.TaskNsp.GetTaskList(db.DB, fmt.Sprintf(
		"%d", taskPos.ScheduleBeginPos), p.Req.TaskType, db.TaskEnum(p.Req.Status), limit)
	if err != nil {
		martlog.Errorf("GetTaskList %s", err.Error())
		p.Resp.Code = constant.ERR_GET_TASK_LIST_FROM_DB
		return err
	}
	for _, dbTask := range taskList {
		// 是重试任务，并且存在重试间隔，且还没到就排序时间，这种就过滤掉
		if dbTask.CrtRetryNum != 0 && dbTask.MaxRetryInterval != 0 &&
			dbTask.OrderTime > time.Now().Unix() {
			continue
		}
		var task = &model.TaskData{}
		ctrlmodel.FillTaskResp(dbTask, task)
		p.Resp.TaskList = append(p.Resp.TaskList, task)
	}

	return nil
}
