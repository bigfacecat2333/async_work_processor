package main

import (
	"fmt"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/config"
	"github.com/bigfacecat2333/async_work_processor/flowsvr/src/initialize"
	"github.com/niuniumart/gosdk/gin"

	"github.com/niuniumart/gosdk/martlog"
)

// flowsvr:
// 1. 读取配置文件
// 2. 初始化资源，主要是MySQL连接
// 3. 创建一个web服务 (gin)
// 4. 注册路由，包括CreateTask,HoldTasks,GetTaskList,GetTaskScheduleCfgList,GetTask,SetTask... 等待worker调用
// 5. 启动web server，这一步之后这个主协程启动会阻塞在这里，请求可以通过gin的子协程进来

func main() {
	// 初始化配置 主要是读取配置文件
	config.Init()
	// 初始资源，主要是MySQL连接
	err := initialize.InitResource()
	if err != nil {
		fmt.Printf("initialize.InitResource err %s", err.Error())
		martlog.Errorf("initialize.InitResource err %s", err.Error())
		return
	}
	// 创建一个web服务
	router := gin.CreateGin()
	// 这里跳进去就能看到有哪些接口
	initialize.RegisterRouter(router)
	fmt.Println("before router run")
	// 启动web server，这一步之后这个主协程启动会阻塞在这里，请求可以通过gin的子协程进来
	err = gin.RunByPort(router, config.Conf.Common.Port)
	fmt.Println(err)
}
