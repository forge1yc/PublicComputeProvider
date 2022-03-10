package model

type ExceptionJob struct {
	JobId     int    `json:"job_id"`
	Exception string `json:"exception"`
}
