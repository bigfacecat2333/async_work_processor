package main

import (
	"encoding/json"
	"fmt"
	"github.com/niuniumart/asyncflow/flowsvr/src/config"
	"github.com/niuniumart/asyncflow/flowsvr/src/db"
	"github.com/niuniumart/asyncflow/taskutils/rpc"
	"github.com/niuniumart/asyncflow/taskutils/rpc/model"
	"github.com/niuniumart/gosdk/tools"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateTask(t *testing.T) {
	config.TestFilePath = "../config/config-test.toml"
	config.Init()
	db.InitDB()
	fmt.Println("aaa   ************* ")
	// convey的功能是将测试用例的描述信息与测试用例的执行结果关联起来，从而更加直观的展示测试结果
	convey.Convey("TestCreateTask", t, func() {
		// case 1: input err
		var rpc rpc.TaskRpc
		rpc.Host = "http://127.0.0.1:41555"
		var reqBody = new(model.CreateTaskReq)
		reqBody.TaskData.TaskType = "lark"
		reqBody.TaskData.TaskStage = "sendmsg"
		reqBody.TaskData.UserId = "niuniu"
		reqBody.TaskData.Status = 1
		var ltctx = LarkTaskContext{
			ReqBody: &LarkReq{Msg: "nice to meet u", FromAddr: "fish", ToAddr: "cat"},
		}
		ltctxStr, _ := json.Marshal(ltctx)
		reqBody.TaskData.TaskContext = string(ltctxStr)
		resp, err := rpc.CreateTask(reqBody)
		fmt.Println(tools.GetFmtStr(resp))
		fmt.Println(err)
		convey.So(err, convey.ShouldBeNil)
	})
}
