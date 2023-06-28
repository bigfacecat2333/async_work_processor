package model

import (
	"time"
)

// CreateTaskReq 请求消息
type CreateTaskReq struct {
	TaskData TaskData `json:"taskData"`
}

// RespComm 通用的响应消息
type RespComm struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// CreateTaskResp 响应消息
type CreateTaskResp struct {
	RespComm
	TaskId string `json:"taskId"`
}

// GetTaskListReq 请求消息
type GetTaskListReq struct {
	TaskType string `json:"taskType" form:"taskType"`
	Status   int    `json:"status" form:"status"`
	Limit    int    `json:"limit" form:"limit"`
}

// GetTaskListResp 响应消息
type GetTaskListResp struct {
	RespComm
	TaskList []*TaskData `json:"taskList"`
}

// GetTaskListReq 请求消息
type HoldTasksReq struct {
	TaskType string `json:"taskType" form:"taskType"`
	Limit    int    `json:"limit" form:"limit"`
}

// GetTaskListResp 响应消息
type HoldTasksResp struct {
	RespComm
	TaskList []*TaskData `json:"taskList"`
}

// GetTaskReq 请求消息
type GetTaskReq struct {
	TaskId string `json:"taskId" form:"taskId"`
}

// GetTaskResp 响应消息
type GetTaskResp struct {
	RespComm
	TaskData *TaskData `json:"taskData"`
}

// GetTaskCountByStatusReq 请求消息
type GetTaskCountByStatusReq struct {
	TaskType string `json:"taskType" form:"taskType"`
	Status   int    `json:"status" form:"status"`
}

// GetTaskCountByStatusResp 响应消息
type GetTaskCountByStatusResp struct {
	RespComm
	Count int `json:"count"`
}

// GetTaskScheduleCfgListReq 请求消息
type GetTaskScheduleCfgListReq struct {
}

// GetTaskScheduleCfgListResp 响应消息
type GetTaskScheduleCfgListResp struct {
	RespComm
	ScheduleCfgList []*TaskScheduleCfg `json:"scheduleCfgList"`
}

// TaskScheduleCfg 任务调度信息
type TaskScheduleCfg struct {
	TaskType          string
	ScheduleLimit     int
	ScheduleInterval  int
	MaxProcessingTime int64
	MaxRetryNum       int
	MaxRetryInterval  int
	CreateTime        *time.Time
	ModifyTime        *time.Time
}

// SetTaskStatusReq 请求消息
type SetTaskStatusReq struct {
	TaskId       string `json:"taskId"`
	Status       int    `json:"status"`
	NoModifyTime bool   `json:"noModifyTime"`
}

// SetTaskStatusResp 响应消息
type SetTaskStatusResp struct {
	RespComm
}

// SetTaskReq 请求消息
type SetTaskReq struct {
	TaskId   string `json:"taskId"`
	TaskData `json:"TaskData"`
	Context  string `json:"context"`
}

// SetTaskResp 响应消息
type SetTaskResp struct {
	RespComm
}

// TaskData 任务调度数据
type TaskData struct {
	UserId           string     `json:"userId"`
	TaskId           string     `json:"taskId"`
	TaskType         string     `json:"taskType"`
	TaskStage        string     `json:"taskStage"`
	Status           int        `json:"status"`
	Priority         *int       `json:"priority"`
	CrtRetryNum      int        `json:"crtRetryNum"`
	MaxRetryNum      int        `json:"maxRetryNum"`
	MaxRetryInterval int        `json:"maxRetryInterval"`
	ScheduleLog      string     `json:"scheduleLog"`
	TaskContext      string     `json:"context"`
	OrderTime        int64      `json:"orderTime"`
	CreateTime       *time.Time `json:"createTime"`
	ModifyTime       *time.Time `json:"modifyTime"`
}
