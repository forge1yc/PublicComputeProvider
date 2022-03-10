package model

type Result struct {
	Result    string `json:"result"`
	PullTime  int64  `json:"pull_time"`
	RunTime   int64  `json:"run_time"`
	TotalTime int64  `json:"total_time"`
	JobId     int    `json:"job_id"`
}
