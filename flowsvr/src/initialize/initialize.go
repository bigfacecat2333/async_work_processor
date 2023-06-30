package initialize

import (
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/ctrl/task"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/db"
	"github.com/bigfacecat2333/async_work_processor/taskutils/rpc/model"
	"github.com/gin-gonic/gin"
	"github.com/niuniumart/gosdk/martlog"
)

// InitResource 初始化服务资源
// 包括：数据库连接， 注册路由
func InitResource() error {
	err := InitInfra()
	if err != nil {
		martlog.Errorf("InitInfra err %s", err.Error())
		return err
	}
	return nil
}

// RegisterRouter 注册路由
// 路由的功能是把请求转发到对应的处理函数
func RegisterRouter(router *gin.Engine) {
	{
		// 创建任务接口，前面是路径，后面是执行的函数，跳进去
		// 解析成对应的handler.Run函数，即HandleInput和HandleProcess。
		// tips:这里改用 RPC 效率更高
		router.POST(model.CREATE_TASK_SUFFIX, task.CreateTask)
		router.POST(model.HOLD_TASKS, task.HoldTasks)
		router.GET(model.GET_TASK_LIST_SUFFIX, task.GetTaskList)
		router.GET(model.GET_TASK_SCHEDULE_CFG_SUFFIX, task.GetTaskScheduleCfgList)
		router.GET(model.GET_TASK_SUFFIX, task.GetTask)
		router.POST(model.SET_TASK_SUFFIX, task.SetTask)
		//logprint.RegisterIgnoreRespLogUrl("/get_task_list")
	}
}

// InitInfra 初始化基础设施
func InitInfra() error {
	err := db.InitDB()
	if err != nil {
		return err
	}
	return nil
}
