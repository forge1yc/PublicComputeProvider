package model

// Job Job信息
// @Author: hyc
// @Description: 部分字段不需要
// @Date: 2021/11/16 12:58
type Job struct {
	Id        int    `json:"job_id"`
	// developer用来指明接收到的用户名
	Developer string `json:"developer"`
	StartTime int64  `json:"start_time"`
	AckTime   string `json:"ack_time"`
	PullTime  int64  `json:"pull_time"`
	RunTime   int64  `json:"run_time"`
	TotalTime int64  `json:"total_time"`
	Cost      string `json:"cost"`
	Finished  bool   `json:"finished"`
	Response  string `json:"response"`
}
