package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/niuniumart/asyncflow/flowsvr/src/ctrl/task"
	"github.com/niuniumart/asyncflow/flowsvr/src/db"
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
		router.POST("/create_task", task.CreateTask)
		router.POST("/hold_tasks", task.HoldTasks)
		router.GET("/get_task_list", task.GetTaskList)
		router.GET("/get_task_schedule_cfg_list", task.GetTaskScheduleCfgList)
		router.GET("/get_task", task.GetTask)
		router.POST("/set_task", task.SetTask)
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
