package model

// Task 任务对象
// @Author: hyc
// @Description:
// @Date: 2021/11/16 16:06
type Task struct {
	ImagePath     string   `json:"task"`
	TaskDeveloper string   `json:"task_developer"`
	JobId         int      `json:"job"`
	Ports         []string `json:"ports"`
}
